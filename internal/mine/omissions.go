package mine

import (
	"regexp"
	"strings"
)

// omissions.go — Item 19: extraction omission tracking and supported-subset report.

// OmissionKind classifies what kind of Rust construct was dropped or abstracted.
type OmissionKind int

const (
	OmissionUnsupportedType     OmissionKind = iota // Rust type has no Spectre equivalent
	OmissionSkippedMethod                           // Method skipped (constructor, async, etc.)
	OmissionTupleVariant                            // Enum variant with data — dropped to unit
	OmissionOverflowOp                              // Saturating/wrapping arithmetic — abstracted
	OmissionPanicPattern                            // panic!/unwrap/expect — dropped
	OmissionResultReturn                            // Result<T,E> return / ? operator — abstracted
	OmissionComplexAssign                           // Assignment RHS too complex to translate
	OmissionUnsupportedPattern                      // Misc unrecognised pattern
)

func (k OmissionKind) String() string {
	switch k {
	case OmissionUnsupportedType:
		return "UNSUPPORTED_TYPE"
	case OmissionSkippedMethod:
		return "SKIPPED_METHOD"
	case OmissionTupleVariant:
		return "TUPLE_VARIANT"
	case OmissionOverflowOp:
		return "OVERFLOW_OP"
	case OmissionPanicPattern:
		return "PANIC_PATTERN"
	case OmissionResultReturn:
		return "RESULT_RETURN"
	case OmissionComplexAssign:
		return "COMPLEX_ASSIGN"
	case OmissionUnsupportedPattern:
		return "UNSUPPORTED_PATTERN"
	default:
		return "UNKNOWN"
	}
}

// ExtractionOmission records one construct that was skipped or abstracted.
type ExtractionOmission struct {
	Kind    OmissionKind
	Context string // Where this omission occurred (method name, field name, etc.)
	Detail  string // Human-readable description of what was skipped and why
}

func (o ExtractionOmission) String() string {
	if o.Context != "" {
		return o.Kind.String() + " in " + o.Context + ": " + o.Detail
	}
	return o.Kind.String() + ": " + o.Detail
}

// SupportedSubset returns the canonical list of Rust constructs that the
// miner can translate to Spectre, grouped by category.
func SupportedSubset() []SupportCategory {
	return []SupportCategory{
		{
			Category: "Struct fields",
			Supported: []string{
				"i8, i16, i32, i64, i128 — mapped to int",
				"u8, u16, u32, u64, u128, usize, isize — mapped to int",
				"bool — mapped to bool",
				"f32, f64 — mapped to float",
				"String, &str — mapped to str",
				"Vec<T> — mapped to List[T]",
				"HashSet<T>, BTreeSet<T> — mapped to Set[T]",
				"HashMap<K,V>, BTreeMap<K,V> — mapped to Map[K,V]",
				"Option<T> — inner type extracted, None abstracted away",
				"Arc<T>, Rc<T>, Box<T>, Mutex<T>, RefCell<T> — wrapper stripped, inner type extracted",
				"Named enums — mapped to Spectre enum",
			},
			Omitted: []string{
				"PhantomData<T> — skipped",
				"fn() / dyn Trait / impl Trait — mapped to int with omission",
				"Tuple types (T1, T2) — mapped to int with omission",
				"Struct types (nested) — mapped to int with omission",
			},
		},
		{
			Category: "Enums",
			Supported: []string{
				"Unit variants (Name,) — mapped to Spectre enum values",
			},
			Omitted: []string{
				"Tuple variants (Name(T)) — variant name kept, data dropped",
				"Struct variants (Name { field: T }) — variant name kept, data dropped",
			},
		},
		{
			Category: "Methods / actions",
			Supported: []string{
				"&mut self methods — extracted as actions",
				"assert!(cond) / debug_assert!(cond) — extracted as require",
				"if cond { return Err(..) } — negated condition extracted as require",
				"self.field = expr — extracted as field assignment",
				"self.field += expr — extracted as field increment (x' = x + rhs)",
				"self.field -= expr — extracted as field decrement",
				"self.field *= expr, /= expr — extracted as multiply/divide assign",
			},
			Omitted: []string{
				"new(), default(), clone(), fmt(), drop() — skipped (constructors/utilities)",
				"async fn — extracted but async qualifier ignored",
				"panic!() / unwrap() / expect() — reported as omission, not extracted",
				"checked_add / wrapping_add / saturating_add — reported as omission, abstracted to +",
				"? operator / Result return — reported as omission, error path dropped",
				"assert_eq! / assert_ne! — skipped (too complex to translate)",
				"for loops / while loops — body scanned for assignments, structure dropped",
				"match expressions — not translated, inner assignments extracted heuristically",
			},
		},
	}
}

// SupportCategory groups supported and omitted constructs under a label.
type SupportCategory struct {
	Category  string
	Supported []string
	Omitted   []string
}

// ── ScanOmissions ─────────────────────────────────────────────────────────────

var (
	reOverflowOp     = regexp.MustCompile(`\.(checked|wrapping|saturating)_(add|sub|mul|div|shl|shr)\s*\(`)
	rePanicPattern   = regexp.MustCompile(`\bpanic!\s*\(|\.unwrap\(\)|\.expect\s*\(`)
	reResultPattern  = regexp.MustCompile(`\?\s*[;,\n\r]|\bResult\s*<|\.ok\(\)|\bErr\s*\(|\bOk\s*\(`)
	reTupleVariant   = regexp.MustCompile(`(?m)^\s*(\w+)\s*\(`)
	reStructVariant  = regexp.MustCompile(`(?m)^\s*(\w+)\s*\{`)
	reSkippedFnName  = regexp.MustCompile(`(?m)(?:pub\s+)?(?:async\s+)?fn\s+(new|default|clone|fmt|drop)\s*\(`)
)

// ScanOmissions analyzes cleaned Rust source and the already-extracted spec
// to produce a list of constructs that were skipped or abstracted away.
func ScanOmissions(src string, m *MinedSpec) []ExtractionOmission {
	var omissions []ExtractionOmission

	// Unsupported field types mapped to int as fallback
	for _, f := range m.Fields {
		if f.SpectreType == "int" && !isLegitimateIntMapping(f.RustType) {
			omissions = append(omissions, ExtractionOmission{
				Kind:    OmissionUnsupportedType,
				Context: m.StructName + "." + f.Name,
				Detail:  "type " + f.RustType + " has no Spectre equivalent; mapped to int",
			})
		}
	}

	// Enum variants with data (tuple/struct) that were silently dropped to unit
	omissions = append(omissions, scanDroppedVariants(src, m.Enums)...)

	// Per-method: overflow, panic, Result patterns
	for _, method := range m.Methods {
		omissions = append(omissions, scanMethodBody(method)...)
	}

	// Skipped methods (constructors/utilities) present in source but not extracted
	omissions = append(omissions, scanSkippedMethods(src, m.Methods)...)

	return omissions
}

// isLegitimateIntMapping returns true when a Rust type naturally maps to int
// (i.e., it is a primitive integer type, possibly behind a reference).
func isLegitimateIntMapping(rustType string) bool {
	rt := strings.TrimSpace(rustType)
	// Strip lifetimes and reference markers
	for strings.HasPrefix(rt, "&") {
		rt = strings.TrimPrefix(rt, "&mut ")
		rt = strings.TrimPrefix(rt, "&")
		rt = strings.TrimSpace(rt)
	}
	return isIntType(rt)
}

// scanDroppedVariants finds enum variants with data (tuple or struct form) that
// were omitted from the extracted variant list.
func scanDroppedVariants(src string, enums []RustEnum) []ExtractionOmission {
	var omissions []ExtractionOmission

	// Build set of already-extracted variant names per enum
	extracted := make(map[string]map[string]bool)
	for _, e := range enums {
		extracted[e.Name] = make(map[string]bool)
		for _, v := range e.Variants {
			extracted[e.Name][v] = true
		}
	}

	locs := reEnum.FindAllStringIndex(src, -1)
	for _, loc := range locs {
		match := reEnum.FindStringSubmatch(src[loc[0]:loc[1]])
		if len(match) < 2 {
			continue
		}
		enumName := match[1]
		body, _ := extractBraceBlock(src, loc[1]-1)
		ev := extracted[enumName] // may be nil if enum had no unit variants

		// Tuple variants: Name(...)
		for _, vm := range reTupleVariant.FindAllStringSubmatch(body, -1) {
			varName := vm[1]
			if isRustKeyword(varName) {
				continue
			}
			if ev == nil || !ev[varName] {
				omissions = append(omissions, ExtractionOmission{
					Kind:    OmissionTupleVariant,
					Context: enumName + "::" + varName,
					Detail:  "tuple variant data dropped; variant name preserved as unit variant",
				})
			}
		}

		// Struct variants: Name { ... } — only if name looks like a variant (UpperCase)
		for _, vm := range reStructVariant.FindAllStringSubmatch(body, -1) {
			varName := vm[1]
			if isRustKeyword(varName) || !isUpperCase(varName) {
				continue
			}
			if ev == nil || !ev[varName] {
				omissions = append(omissions, ExtractionOmission{
					Kind:    OmissionTupleVariant,
					Context: enumName + "::" + varName,
					Detail:  "struct variant data dropped; variant name preserved as unit variant",
				})
			}
		}
	}

	return omissions
}

// scanMethodBody checks a method's raw body for overflow ops, panics, and
// Result/? patterns that were abstracted or dropped.
func scanMethodBody(method RustMethod) []ExtractionOmission {
	var omissions []ExtractionOmission
	body := method.BodyRaw

	if reOverflowOp.MatchString(body) {
		omissions = append(omissions, ExtractionOmission{
			Kind:    OmissionOverflowOp,
			Context: method.Name,
			Detail:  "overflow-checked arithmetic abstracted to plain arithmetic",
		})
	}

	if rePanicPattern.MatchString(body) {
		omissions = append(omissions, ExtractionOmission{
			Kind:    OmissionPanicPattern,
			Context: method.Name,
			Detail:  "panic/unwrap/expect dropped; error path not modelled",
		})
	}

	if reResultPattern.MatchString(body) {
		omissions = append(omissions, ExtractionOmission{
			Kind:    OmissionResultReturn,
			Context: method.Name,
			Detail:  "Result return / ? operator: error path abstracted away",
		})
	}

	return omissions
}

// scanSkippedMethods reports constructor/utility methods that appear in the
// source but were intentionally excluded from the extracted method list.
func scanSkippedMethods(src string, methods []RustMethod) []ExtractionOmission {
	var omissions []ExtractionOmission
	extracted := make(map[string]bool)
	for _, m := range methods {
		extracted[m.Name] = true
	}

	seen := make(map[string]bool)
	for _, match := range reSkippedFnName.FindAllStringSubmatch(src, -1) {
		name := match[1]
		if !extracted[name] && !seen[name] {
			seen[name] = true
			omissions = append(omissions, ExtractionOmission{
				Kind:    OmissionSkippedMethod,
				Context: name,
				Detail:  "constructor/utility method skipped (not translated to action)",
			})
		}
	}
	return omissions
}
