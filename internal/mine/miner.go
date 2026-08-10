// Package mine extracts Spectre spec skeletons from Rust source code.
// It uses regex-based static analysis to mine struct fields, enum types,
// method signatures, assert conditions, and assignment patterns.
package mine

import (
	"regexp"
	"strings"
	"unicode"
)

// ── Extracted types ───────────────────────────────────────────────────────────

// RustField is one field in the target struct.
type RustField struct {
	Name        string
	RustType    string
	SpectreType string
	Default     string // Spectre default value
}

// RustEnum is a Rust enum that maps to a Spectre enum.
type RustEnum struct {
	Name     string
	Variants []string // only unit variants (no tuple/struct variants)
}

// RustParam is one parameter of a method.
type RustParam struct {
	Name        string
	SpectreType string
}

// RustAssign is a field assignment detected in a method body.
type RustAssign struct {
	Field   string
	Op      string // "=", "+=", "-=", "*=", "/="
	RHSRust string // original Rust RHS (before translation)
}

// RustMethod is a method extracted from an impl block.
type RustMethod struct {
	Name     string
	MutSelf  bool
	Params   []RustParam
	Requires []string    // conditions from assert!/early-return patterns
	Assigns  []RustAssign
	BodyRaw  string
}

// MinedSpec is the result of analyzing a Rust source file.
type MinedSpec struct {
	SpecName   string
	StructName string
	Fields     []RustField
	Enums      []RustEnum
	Methods    []RustMethod
	SourceFile string
}

// ── Regex patterns ────────────────────────────────────────────────────────────

var (
	reStruct    = regexp.MustCompile(`(?m)(?:pub\s+)?struct\s+(\w+)\s*\{`)
	reEnum      = regexp.MustCompile(`(?m)(?:pub\s+)?enum\s+(\w+)\s*\{`)
	reField     = regexp.MustCompile(`(?m)^\s*(?:pub\s+)?(\w+)\s*:\s*([^,/\n}]+)`)
	reVariant   = regexp.MustCompile(`(?m)^\s*(\w+)\s*(?:[,\n]|$)`)
	reImpl      = regexp.MustCompile(`(?m)impl(?:<[^>]+>)?\s+(\w+)\s*(?:for\s+\w+\s*)?\{`)
	reFn        = regexp.MustCompile(`(?m)(?:pub\s+)?(?:async\s+)?fn\s+(\w+)\s*\((&(?:mut\s+)?self)([^)]*)\)`)
	reAssert        = regexp.MustCompile(`debug_assert!|assert!`)
	reEarlyRet      = regexp.MustCompile(`(?m)if\s+(.+?)\s*\{[^}]*return\s+Err`)
	reAssign        = regexp.MustCompile(`self\.(\w+)\s*([\+\-\*\/]?=)\s*([^;{]+);`)
	reParamPair     = regexp.MustCompile(`(\w+)\s*:\s*([^,)]+)`)
	reWhitespace    = regexp.MustCompile(`\s+`)
	reLifetime      = regexp.MustCompile(`'[a-z_]+\s+`)
	reSelfDot       = regexp.MustCompile(`self\.`)
	reColonColon    = regexp.MustCompile(`(\w+)::(\w+)`)
	reNumUnderscore = regexp.MustCompile(`(\d)_(\d)`)
)

// MineFromRust analyzes a Rust source file and extracts a MinedSpec.
// specName is used as the Spectre spec name; if empty, it's inferred from the
// first public struct found.
func MineFromRust(source, specName, filename string) *MinedSpec {
	// Strip line comments so they don't confuse regex patterns.
	cleaned := stripLineComments(source)

	enums := extractEnums(cleaned)
	enumNames := enumNameSet(enums)

	fields, structName := extractMainStruct(cleaned, specName, enumNames)
	if specName == "" {
		specName = structName
	}

	methods := extractMethods(cleaned, structName, fields, enumNames)

	return &MinedSpec{
		SpecName:   specName,
		StructName: structName,
		Fields:     fields,
		Enums:      enums,
		Methods:    methods,
		SourceFile: filename,
	}
}

// ── Struct extraction ─────────────────────────────────────────────────────────

// extractMainStruct finds the struct matching specName (case-insensitive),
// or the first public struct, and returns its fields.
func extractMainStruct(src, specName string, enumNames map[string]bool) ([]RustField, string) {
	matches := reStruct.FindAllStringIndex(src, -1)
	if len(matches) == 0 {
		return nil, specName
	}

	// Prefer a struct whose name matches specName (case-insensitive).
	var chosen int = -1
	var chosenName string
	for _, m := range matches {
		nameLoc := reStruct.FindStringSubmatch(src[m[0]:m[1]])
		if len(nameLoc) < 2 {
			continue
		}
		n := nameLoc[1]
		if specName != "" && strings.EqualFold(n, specName) {
			chosen = m[1]
			chosenName = n
			break
		}
		if chosen < 0 {
			chosen = m[1]
			chosenName = n
		}
	}
	if chosen < 0 {
		return nil, specName
	}

	body, _ := extractBraceBlock(src, chosen-1)
	fields := parseStructFields(body, enumNames)
	return fields, chosenName
}

func parseStructFields(body string, enumNames map[string]bool) []RustField {
	var fields []RustField
	for _, m := range reField.FindAllStringSubmatch(body, -1) {
		name := m[1]
		if isRustKeyword(name) {
			continue
		}
		rustType := strings.TrimSpace(m[2])
		// Skip PhantomData and similar internal fields.
		if strings.Contains(rustType, "PhantomData") {
			continue
		}
		st, def := rustTypeToSpectre(rustType, enumNames)
		fields = append(fields, RustField{
			Name:        name,
			RustType:    rustType,
			SpectreType: st,
			Default:     def,
		})
	}
	return fields
}

// ── Enum extraction ───────────────────────────────────────────────────────────

func extractEnums(src string) []RustEnum {
	var enums []RustEnum
	locs := reEnum.FindAllStringIndex(src, -1)
	for _, loc := range locs {
		m := reEnum.FindStringSubmatch(src[loc[0]:loc[1]])
		if len(m) < 2 {
			continue
		}
		name := m[1]
		body, _ := extractBraceBlock(src, loc[1]-1)
		variants := parseEnumVariants(body)
		if len(variants) > 0 {
			enums = append(enums, RustEnum{Name: name, Variants: variants})
		}
	}
	return enums
}

func parseEnumVariants(body string) []string {
	var variants []string
	seen := map[string]bool{}
	for _, m := range reVariant.FindAllStringSubmatch(body, -1) {
		v := m[1]
		if isRustKeyword(v) || seen[v] {
			continue
		}
		// Skip tuple/struct variants — they contain parens or braces.
		variants = append(variants, v)
		seen[v] = true
	}
	return variants
}

// ── Method extraction ─────────────────────────────────────────────────────────

func extractMethods(src, structName string, fields []RustField, enumNames map[string]bool) []RustMethod {
	fieldNames := fieldNameSet(fields)

	// Find impl blocks for structName.
	var methods []RustMethod
	locs := reImpl.FindAllStringIndex(src, -1)
	for _, loc := range locs {
		m := reImpl.FindStringSubmatch(src[loc[0]:loc[1]])
		if len(m) < 2 {
			continue
		}
		implTarget := m[1]
		if structName != "" && !strings.EqualFold(implTarget, structName) {
			continue // skip impl blocks for other types
		}

		implBody, _ := extractBraceBlock(src, loc[1]-1)
		methods = append(methods, parseMethodsFromImpl(implBody, fieldNames, enumNames)...)
	}

	// If no impl block matched by name, take methods from the first impl block.
	if len(methods) == 0 && structName == "" && len(locs) > 0 {
		implBody, _ := extractBraceBlock(src, locs[0][1]-1)
		methods = parseMethodsFromImpl(implBody, fieldNames, enumNames)
	}

	return methods
}

func parseMethodsFromImpl(implBody string, fieldNames, enumNames map[string]bool) []RustMethod {
	var methods []RustMethod
	fns := reFn.FindAllStringIndex(implBody, -1)
	for _, fnLoc := range fns {
		m := reFn.FindStringSubmatch(implBody[fnLoc[0]:fnLoc[1]])
		if len(m) < 3 {
			continue
		}
		name := m[1]
		selfRef := m[2]
		restParams := strings.TrimSpace(m[3])

		// Skip constructors and common non-action methods.
		if name == "new" || name == "default" || name == "clone" || name == "fmt" || name == "drop" {
			continue
		}

		mutSelf := strings.Contains(selfRef, "mut")

		var params []RustParam
		if restParams != "" {
			restParams = strings.TrimPrefix(restParams, ",")
			for _, pm := range reParamPair.FindAllStringSubmatch(restParams, -1) {
				pname := strings.TrimSpace(pm[1])
				ptype := strings.TrimSpace(pm[2])
				if pname == "self" {
					continue
				}
				st, _ := rustTypeToSpectre(ptype, enumNames)
				params = append(params, RustParam{Name: pname, SpectreType: st})
			}
		}

		// Find and extract the method body.
		bodyStart := fnLoc[1]
		// Advance to the opening brace.
		for bodyStart < len(implBody) && implBody[bodyStart] != '{' {
			bodyStart++
		}
		if bodyStart >= len(implBody) {
			continue
		}
		body, _ := extractBraceBlock(implBody, bodyStart)

		requires := extractRequires(body, fieldNames, enumNames)
		assigns := extractAssigns(body, fieldNames)

		methods = append(methods, RustMethod{
			Name:     name,
			MutSelf:  mutSelf,
			Params:   params,
			Requires: requires,
			Assigns:  assigns,
			BodyRaw:  body,
		})
	}
	return methods
}

// extractRequires collects preconditions from assert! calls and early-return
// guards in the method body.
func extractRequires(body string, fieldNames, enumNames map[string]bool) []string {
	var reqs []string
	seen := map[string]bool{}

	// assert!(...) and debug_assert!(...)
	for _, loc := range reAssert.FindAllStringIndex(body, -1) {
		// Find opening paren.
		i := loc[1]
		for i < len(body) && body[i] != '(' {
			i++
		}
		if i >= len(body) {
			continue
		}
		inner, _ := extractParenBlock(body, i)
		// For assert_eq!/assert_ne!, skip (too complex to translate simply).
		if strings.HasPrefix(body[loc[0]:loc[1]], "assert_eq") ||
			strings.HasPrefix(body[loc[0]:loc[1]], "assert_ne") {
			continue
		}
		cond := rustCondToSpectre(inner)
		if cond != "" && !seen[cond] {
			seen[cond] = true
			reqs = append(reqs, cond)
		}
	}

	// if condition { return Err(..) } — early-return guards
	for _, m := range reEarlyRet.FindAllStringSubmatch(body, -1) {
		raw := strings.TrimSpace(m[1])
		// These are negated (the condition causes a return), so negate them.
		cond := negateRustCond(rustCondToSpectre(raw))
		if cond != "" && !seen[cond] {
			seen[cond] = true
			reqs = append(reqs, cond)
		}
	}

	return reqs
}

// extractAssigns collects field assignments from the method body.
func extractAssigns(body string, fieldNames map[string]bool) []RustAssign {
	var assigns []RustAssign
	seen := map[string]bool{}
	for _, m := range reAssign.FindAllStringSubmatch(body, -1) {
		field := m[1]
		if !fieldNames[field] {
			continue
		}
		op := strings.TrimSpace(m[2])
		rhs := strings.TrimSpace(m[3])
		// If op is "=" and rhs starts with "=", this is a "==" comparison, not assignment.
		if op == "=" && strings.HasPrefix(rhs, "=") {
			continue
		}
		key := field + op + rhs
		if seen[key] {
			continue
		}
		seen[key] = true
		assigns = append(assigns, RustAssign{Field: field, Op: op, RHSRust: rhs})
	}
	return assigns
}

// ── Type mapping ──────────────────────────────────────────────────────────────

// rustTypeToSpectre converts a Rust type string to (spectreType, defaultValue).
func rustTypeToSpectre(rt string, enumNames map[string]bool) (string, string) {
	rt = strings.TrimSpace(rt)
	// Strip lifetimes: &'a str -> str
	rt = reLifetime.ReplaceAllString(rt, "")
	// Strip reference markers: &mut T -> T, &T -> T
	for strings.HasPrefix(rt, "&") {
		rt = strings.TrimPrefix(rt, "&mut ")
		rt = strings.TrimPrefix(rt, "&")
		rt = strings.TrimSpace(rt)
	}

	switch {
	case isIntType(rt):
		return "int", "0"
	case isFloatType(rt):
		return "float", "0.0"
	case rt == "bool":
		return "bool", "false"
	case rt == "String" || rt == "&str" || rt == "str":
		return "str", `""`
	case strings.HasPrefix(rt, "Vec<"):
		inner := innerGeneric(rt)
		st, _ := rustTypeToSpectre(inner, enumNames)
		return "List[" + st + "]", "{}"
	case strings.HasPrefix(rt, "HashSet<") || strings.HasPrefix(rt, "BTreeSet<"):
		inner := innerGeneric(rt)
		st, _ := rustTypeToSpectre(inner, enumNames)
		return "Set[" + st + "]", "{}"
	case strings.HasPrefix(rt, "HashMap<") || strings.HasPrefix(rt, "BTreeMap<"):
		k, v := splitMapGeneric(rt)
		sk, _ := rustTypeToSpectre(k, enumNames)
		sv, _ := rustTypeToSpectre(v, enumNames)
		return "Map[" + sk + ", " + sv + "]", "{}"
	case strings.HasPrefix(rt, "Option<"):
		inner := innerGeneric(rt)
		st, def := rustTypeToSpectre(inner, enumNames)
		return st, def // treat Option<T> as T for spec purposes
	case strings.HasPrefix(rt, "Arc<") || strings.HasPrefix(rt, "Mutex<") ||
		strings.HasPrefix(rt, "Rc<") || strings.HasPrefix(rt, "RefCell<") ||
		strings.HasPrefix(rt, "Box<"):
		inner := innerGeneric(rt)
		return rustTypeToSpectre(inner, enumNames)
	default:
		// Known enum?
		if enumNames[rt] {
			return rt, rt + "." + "Default" // caller will fix with first variant
		}
		// Unknown type — fall back to int. Original Rust type is in RustField.RustType.
		return "int", "0"
	}
}

func isIntType(t string) bool {
	switch t {
	case "i8", "i16", "i32", "i64", "i128",
		"u8", "u16", "u32", "u64", "u128",
		"isize", "usize":
		return true
	}
	return false
}

func isFloatType(t string) bool {
	return t == "f32" || t == "f64"
}

func innerGeneric(rt string) string {
	i := strings.Index(rt, "<")
	if i < 0 {
		return rt
	}
	// Find matching >.
	depth := 0
	for j := i; j < len(rt); j++ {
		if rt[j] == '<' {
			depth++
		}
		if rt[j] == '>' {
			depth--
			if depth == 0 {
				return strings.TrimSpace(rt[i+1 : j])
			}
		}
	}
	return rt[i+1:]
}

func splitMapGeneric(rt string) (string, string) {
	inner := innerGeneric(rt)
	// Split on the first comma not inside angle brackets.
	depth := 0
	for i, c := range inner {
		switch c {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				return strings.TrimSpace(inner[:i]), strings.TrimSpace(inner[i+1:])
			}
		}
	}
	return inner, "str"
}

// ── Condition translation ─────────────────────────────────────────────────────

// rustCondToSpectre translates a Rust condition string to Spectre syntax.
func rustCondToSpectre(cond string) string {
	s := strings.TrimSpace(cond)
	if s == "" {
		return ""
	}

	// Remove message arg from assert: assert!(cond, "msg") → cond
	if i := indexOfCommaAtDepth0(s); i >= 0 {
		if looksLikeMessage(s[i+1:]) {
			s = strings.TrimSpace(s[:i])
		}
	}

	// Remove self. prefix.
	s = reSelfDot.ReplaceAllString(s, "")
	// Rust :: → Spectre . for enum variants.
	s = reColonColon.ReplaceAllString(s, "$1.$2")
	// == → = (Spectre equality)
	s = strings.ReplaceAll(s, " == ", " = ")
	// != stays as-is
	// Remove _ thousand-separators in numbers.
	s = reNumUnderscore.ReplaceAllStringFunc(s, func(m string) string {
		return strings.ReplaceAll(m, "_", "")
	})
	// Clean up extra whitespace.
	s = reWhitespace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// negateRustCond produces the logical negation of a simple Rust condition.
func negateRustCond(cond string) string {
	if cond == "" {
		return ""
	}
	// Already negated: strip the leading !
	if strings.HasPrefix(cond, "!") {
		return strings.TrimPrefix(cond, "!")
	}
	// Reverse comparison operators (handles the most common early-return guards).
	pairs := [][2]string{
		{" != ", " = "}, {" = ", " != "},
		{" < ", " >= "}, {" >= ", " < "},
		{" > ", " <= "}, {" <= ", " > "},
	}
	for _, p := range pairs {
		if strings.Contains(cond, p[0]) {
			return strings.Replace(cond, p[0], p[1], 1)
		}
	}
	// For simple identifiers (boolean flags) just prepend !.
	if isSimpleIdent(cond) {
		return "!" + cond
	}
	// For complex expressions, prefix with ! (no parens — Spectre doesn't support !(...)).
	return "!" + cond
}

// isSimpleIdent reports whether s is a plain identifier (letters, digits, underscores).
func isSimpleIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}

// rustAssignToSpectre converts a Rust assignment to a Spectre assignment RHS.
func rustAssignToSpectre(field, op, rhs string) string {
	rhs = rustCondToSpectre(rhs) // reuse condition translator
	switch op {
	case "+=":
		return field + " + " + rhs
	case "-=":
		return field + " - " + rhs
	case "*=":
		return field + " * " + rhs
	case "/=":
		return field + " / " + rhs
	default: // "="
		return rhs
	}
}

// ── Bracket helpers ───────────────────────────────────────────────────────────

// extractBraceBlock returns the content inside the brace block starting at
// position start (which must point to the opening '{').
func extractBraceBlock(src string, start int) (string, int) {
	if start >= len(src) || src[start] != '{' {
		return "", start
	}
	depth := 0
	for i := start; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[start+1 : i], i + 1
			}
		}
	}
	return src[start+1:], len(src)
}

// extractParenBlock returns the content inside the paren block starting at
// position start (which must point to the opening '(').
func extractParenBlock(src string, start int) (string, int) {
	if start >= len(src) || src[start] != '(' {
		return "", start
	}
	depth := 0
	for i := start; i < len(src); i++ {
		switch src[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return src[start+1 : i], i + 1
			}
		}
	}
	return src[start+1:], len(src)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func enumNameSet(enums []RustEnum) map[string]bool {
	s := make(map[string]bool)
	for _, e := range enums {
		s[e.Name] = true
	}
	return s
}

func fieldNameSet(fields []RustField) map[string]bool {
	s := make(map[string]bool)
	for _, f := range fields {
		s[f.Name] = true
	}
	return s
}

func stripLineComments(src string) string {
	var b strings.Builder
	lines := strings.Split(src, "\n")
	for _, line := range lines {
		// Remove // comments but keep strings (rough heuristic: find // outside quotes).
		out := removeLineComment(line)
		b.WriteString(out)
		b.WriteByte('\n')
	}
	return b.String()
}

func removeLineComment(line string) string {
	inStr := false
	for i := 0; i < len(line)-1; i++ {
		if inStr && line[i] == '\\' {
			i++ // skip escaped character (e.g. \")
			continue
		}
		if line[i] == '"' {
			inStr = !inStr
		}
		if !inStr && line[i] == '/' && line[i+1] == '/' {
			return line[:i]
		}
	}
	return line
}

func indexOfCommaAtDepth0(s string) int {
	depth := 0
	for i, c := range s {
		// Only track bracket pairs, not < > (which are comparison operators in conditions).
		switch c {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func looksLikeMessage(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, `"`) || strings.HasPrefix(s, "format!")
}

func isRustKeyword(s string) bool {
	switch s {
	case "pub", "fn", "let", "mut", "ref", "use", "mod", "struct", "enum",
		"impl", "trait", "type", "where", "for", "in", "if", "else", "match",
		"loop", "while", "return", "break", "continue", "self", "super", "crate":
		return true
	}
	return false
}

// firstRune returns the first rune of s, or 0 if empty.
func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

// isUpperCase reports whether s starts with an uppercase letter (used to
// distinguish enum variants from field names in some heuristics).
func isUpperCase(s string) bool {
	r := firstRune(s)
	return unicode.IsUpper(r)
}

// enumDefaultValue returns the first variant of a known enum, or "".
func enumDefaultValue(typeName string, enums []RustEnum) string {
	for _, e := range enums {
		if e.Name == typeName && len(e.Variants) > 0 {
			return typeName + "." + e.Variants[0]
		}
	}
	return ""
}
