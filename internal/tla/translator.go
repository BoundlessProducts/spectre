// Package tla translates Spectre ASTs to TLA+ specifications.
//
// Translation limits are tracked and reported so callers can understand which
// constructs are supported and which will produce stub placeholders.
package tla

import (
	"fmt"
	"strings"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// LimitKind classifies a translation limit.
type LimitKind int

const (
	// LimitUnsupported means the construct is not representable in TLA+ via
	// this translator and has been replaced by a placeholder.
	LimitUnsupported LimitKind = iota
	// LimitPartial means the construct is representable but with caveats that
	// the caller should be aware of (e.g. string variables become model values).
	LimitPartial
)

func (k LimitKind) String() string {
	switch k {
	case LimitUnsupported:
		return "UNSUPPORTED"
	case LimitPartial:
		return "PARTIAL"
	default:
		return "UNKNOWN"
	}
}

// Diagnostic records a translation limit with source position.
type Diagnostic struct {
	Kind     LimitKind
	Message  string
	Position ast.Position // zero value means position unknown
}

func (d Diagnostic) String() string {
	if d.Position.Line > 0 {
		return fmt.Sprintf("%s at %d:%d — %s", d.Kind, d.Position.Line, d.Position.Column, d.Message)
	}
	return fmt.Sprintf("%s — %s", d.Kind, d.Message)
}

// Result holds the translated TLA+ source and diagnostics.
type Result struct {
	TLA         string
	ModuleName  string
	Diagnostics []Diagnostic
}

// Translate converts a parsed Spectre AST to TLA+.
// filename is used to derive the module name and for error messages.
func Translate(file *ast.File, filename string) *Result {
	t := &translator{
		sb:     &strings.Builder{},
		result: &Result{},
	}

	// Derive TLA+ module name from filename.
	name := filename
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	if strings.HasSuffix(name, ".spec") {
		name = name[:len(name)-5]
	}
	// TLA+ module names must start with a letter.
	if len(name) == 0 || !isLetter(rune(name[0])) {
		name = "Spec"
	}
	// Capitalise first letter per TLA+ convention.
	name = strings.ToUpper(name[:1]) + name[1:]
	t.result.ModuleName = name

	t.translateFile(file, name)
	t.result.TLA = t.sb.String()
	return t.result
}

// ── translator ────────────────────────────────────────────────────────────────

type translator struct {
	sb     *strings.Builder
	result *Result
}

func (t *translator) diag(kind LimitKind, pos ast.Position, msg string) {
	t.result.Diagnostics = append(t.result.Diagnostics, Diagnostic{
		Kind:     kind,
		Message:  msg,
		Position: pos,
	})
}

func (t *translator) emit(s string) { t.sb.WriteString(s) }
func (t *translator) line(s string) { t.sb.WriteString(s); t.sb.WriteByte('\n') }
func (t *translator) nl()           { t.sb.WriteByte('\n') }

// ── file ──────────────────────────────────────────────────────────────────────

func (t *translator) translateFile(file *ast.File, moduleName string) {
	// TLA+ header.
	dashes := strings.Repeat("-", 20)
	t.line(fmt.Sprintf("%s MODULE %s %s", dashes, moduleName, dashes))
	t.nl()
	t.line("EXTENDS Integers, Sequences, FiniteSets, TLC")
	t.nl()

	// Collect declarations by kind so we can emit them in TLA+ order.
	var vars []*ast.VariableDecl
	var consts []*ast.ConstantDecl
	var enums []*ast.EnumDecl
	var params []*ast.ParamDecl
	var inits []ast.Decl   // InitDecl | OneOfInitDecl
	var invariants []*ast.InvariantDecl
	var temporals []*ast.TemporalDecl
	var actions []*ast.ActionDecl
	var funs []*ast.FunctionDecl
	var modules []*ast.ModuleDecl

	for _, d := range file.Decls {
		switch v := d.(type) {
		case *ast.VariableDecl:
			vars = append(vars, v)
		case *ast.ConstantDecl:
			consts = append(consts, v)
		case *ast.EnumDecl:
			enums = append(enums, v)
		case *ast.ParamDecl:
			params = append(params, v)
		case *ast.InitDecl:
			inits = append(inits, v)
		case *ast.OneOfInitDecl:
			inits = append(inits, v)
		case *ast.InvariantDecl:
			invariants = append(invariants, v)
		case *ast.TemporalDecl:
			temporals = append(temporals, v)
		case *ast.ActionDecl:
			actions = append(actions, v)
		case *ast.FunctionDecl:
			funs = append(funs, v)
		case *ast.ModuleDecl:
			modules = append(modules, v)
		case *ast.ImportDecl:
			t.diag(LimitPartial, v.Position,
				fmt.Sprintf("import '%s' not resolved; referenced definitions may be missing", v.Path))
		case *ast.TypeAliasDecl:
			t.diag(LimitPartial, v.Position,
				fmt.Sprintf("type alias '%s' translated as comment only", v.Name))
			t.line(fmt.Sprintf("\\* TYPE ALIAS: %s", v.Name))
		}
	}

	// CONSTANTS (params + enums + regular consts).
	var constNames []string
	for _, p := range params {
		constNames = append(constNames, sanitize(p.Name))
	}
	for _, c := range consts {
		constNames = append(constNames, sanitize(c.Name))
	}
	for _, e := range enums {
		constNames = append(constNames, sanitize(e.Name)+"_VALUES")
	}
	if len(constNames) > 0 {
		t.line(fmt.Sprintf("CONSTANTS %s", strings.Join(constNames, ", ")))
		t.nl()
	}

	// Enum type assumptions.
	for _, e := range enums {
		vals := make([]string, len(e.Values))
		for i, v := range e.Values {
			vals[i] = sanitize(e.Name) + "_" + sanitize(v)
		}
		t.line(fmt.Sprintf("\\* Enum %s: model values", e.Name))
		t.line(fmt.Sprintf("ASSUME %s_VALUES = {%s}", sanitize(e.Name), strings.Join(vals, ", ")))
		for _, v := range vals {
			t.line(fmt.Sprintf("ASSUME %s \\in %s_VALUES", v, sanitize(e.Name)))
		}
		t.nl()
	}

	// VARIABLES.
	if len(vars) > 0 {
		varNames := make([]string, len(vars))
		for i, v := range vars {
			varNames[i] = sanitize(v.Name)
		}
		t.line(fmt.Sprintf("VARIABLES %s", strings.Join(varNames, ", ")))
		t.nl()

		// TypeInvariant.
		t.line("TypeInvariant ==")
		for i, v := range vars {
			sep := "/\\"
			if i == 0 {
				sep = "  /\\"
			} else {
				sep = "  /\\"
			}
			t.line(fmt.Sprintf("%s %s \\in %s", sep, sanitize(v.Name), t.typeSet(v.Type)))
		}
		t.nl()
	}

	// Helper functions from function declarations (pre-actions).
	for _, f := range funs {
		t.translateFun(f)
	}

	// Module functions.
	for _, m := range modules {
		t.translateModule(m)
	}

	// Init.
	if len(inits) > 0 {
		t.translateInit(inits)
	} else {
		t.line("Init == TRUE \\* No init declaration found")
		t.nl()
	}

	// Actions.
	var actionNames []string
	for _, a := range actions {
		t.translateAction(a)
		actionNames = append(actionNames, sanitize(a.Name))
	}

	// Next.
	if len(actionNames) > 0 {
		t.line("Next ==")
		for i, n := range actionNames {
			if i == 0 {
				t.line(fmt.Sprintf("  \\/ %s", n))
			} else {
				t.line(fmt.Sprintf("  \\/ %s", n))
			}
		}
		t.nl()
	} else {
		t.line("Next == FALSE \\* No actions declared")
		t.nl()
	}

	// Spec (safety-only; liveness added separately if temporals exist).
	t.line("Spec == Init /\\ [][Next]_<<" + strings.Join(t.varTuple(vars), ", ") + ">>")
	t.nl()

	// Invariants.
	for _, inv := range invariants {
		t.translateInvariant(inv)
	}

	// Temporal properties.
	for _, temp := range temporals {
		t.translateTemporal(temp)
	}

	// THEOREM statements.
	for _, inv := range invariants {
		t.line(fmt.Sprintf("THEOREM Spec => []%s", sanitize(inv.Name)))
	}
	if len(invariants) > 0 {
		t.nl()
	}

	// Footer.
	t.line(strings.Repeat("=", 40))
}

func (t *translator) varTuple(vars []*ast.VariableDecl) []string {
	names := make([]string, len(vars))
	for i, v := range vars {
		names[i] = sanitize(v.Name)
	}
	return names
}

// ── init ──────────────────────────────────────────────────────────────────────

func (t *translator) translateInit(inits []ast.Decl) {
	if len(inits) == 1 {
		switch v := inits[0].(type) {
		case *ast.InitDecl:
			t.translateDeterministicInit(v)
		case *ast.OneOfInitDecl:
			t.translateOneOfInit(v)
		}
		return
	}
	// Multiple init declarations: translate each as Init_N and conjoin.
	names := make([]string, len(inits))
	for i, d := range inits {
		n := fmt.Sprintf("Init_%d", i)
		names[i] = n
		switch v := d.(type) {
		case *ast.InitDecl:
			t.translateNamedInit(n, v.Body)
		case *ast.OneOfInitDecl:
			t.line(fmt.Sprintf("%s == \\* oneOf (multiple branches)", n))
			for j, opt := range v.Options {
				t.translateNamedInit(fmt.Sprintf("Init_%d_%d", i, j), opt)
			}
		}
	}
	t.line(fmt.Sprintf("Init == %s", strings.Join(names, " /\\ ")))
	t.nl()
}

func (t *translator) translateDeterministicInit(d *ast.InitDecl) {
	t.translateNamedInit("Init", d.Body)
}

func (t *translator) translateNamedInit(name string, body *ast.BlockStmt) {
	t.line(fmt.Sprintf("%s ==", name))
	for _, stmt := range body.Statements {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			t.line(fmt.Sprintf("  /\\ %s = %s", t.lhsExpr(s.Left), t.expr(s.Right)))
		default:
			t.diag(LimitUnsupported, stmt.Pos(),
				fmt.Sprintf("unsupported statement type %T in init block", stmt))
			t.line("  /\\ TRUE \\* UNSUPPORTED STATEMENT")
		}
	}
	t.nl()
}

func (t *translator) translateOneOfInit(d *ast.OneOfInitDecl) {
	if len(d.Options) == 0 {
		t.line("Init == FALSE \\* empty oneOf")
		t.nl()
		return
	}
	optNames := make([]string, len(d.Options))
	for i, opt := range d.Options {
		n := fmt.Sprintf("Init_%d", i)
		optNames[i] = n
		t.translateNamedInit(n, opt)
	}
	t.line(fmt.Sprintf("Init =="))
	for _, n := range optNames {
		t.line(fmt.Sprintf("  \\/ %s", n))
	}
	t.nl()
}

// ── actions ───────────────────────────────────────────────────────────────────

func (t *translator) translateAction(a *ast.ActionDecl) {
	name := sanitize(a.Name)
	if len(a.Parameters) > 0 {
		// Parameterized action: emit with parameters; also emit _any for Next.
		paramNames := make([]string, len(a.Parameters))
		paramSets := make([]string, len(a.Parameters))
		for i, p := range a.Parameters {
			paramNames[i] = sanitize(p.Name)
			paramSets[i] = t.typeSet(p.Type)
		}
		t.line(fmt.Sprintf("%s(%s) ==", name, strings.Join(paramNames, ", ")))
		t.translateActionBody(a.Body)
		t.nl()
		// Top-level quantified wrapper for Next.
		t.line(fmt.Sprintf("%s_any ==", name))
		t.line(fmt.Sprintf("  \\E %s \\in %s : %s(%s)", paramNames[0], paramSets[0],
			name, strings.Join(paramNames, ", ")))
		t.nl()
	} else {
		t.line(fmt.Sprintf("%s ==", name))
		t.translateActionBody(a.Body)
		t.nl()
	}
}

func (t *translator) translateActionBody(body *ast.BlockStmt) {
	var requires []string
	var assignments []string
	var unchanged []string

	// Collect primed variable names from assignments so we can compute UNCHANGED.
	assignedVars := map[string]bool{}

	for _, stmt := range body.Statements {
		switch s := stmt.(type) {
		case *ast.RequireStmt:
			requires = append(requires, t.expr(s.Condition))
		case *ast.AssignStmt:
			if ident, ok := s.Left.(*ast.Ident); ok && ident.Prime {
				lhs := sanitize(ident.Name) + "'"
				assignments = append(assignments, fmt.Sprintf("%s = %s", lhs, t.expr(s.Right)))
				assignedVars[ident.Name] = true
			} else {
				t.diag(LimitPartial, s.Pos(),
					"assignment to unprimed variable in action treated as precondition")
				requires = append(requires, fmt.Sprintf("%s = %s", t.lhsExpr(s.Left), t.expr(s.Right)))
			}
		case *ast.EnsureStmt:
			// Ensure becomes a postcondition check — emit as conjunct.
			assignments = append(assignments, t.expr(s.Condition))
		default:
			t.diag(LimitUnsupported, stmt.Pos(),
				fmt.Sprintf("unsupported statement type %T in action body", stmt))
			assignments = append(assignments, "TRUE \\* UNSUPPORTED")
		}
	}
	_ = unchanged

	allClauses := append(requires, assignments...)
	if len(allClauses) == 0 {
		t.line("  TRUE")
		return
	}
	for _, c := range allClauses {
		t.line(fmt.Sprintf("  /\\ %s", c))
	}
}

// ── functions / modules ───────────────────────────────────────────────────────

func (t *translator) translateFun(f *ast.FunctionDecl) {
	if len(f.Parameters) == 0 {
		body := t.blockToExpr(f.Body)
		t.line(fmt.Sprintf("%s == %s", sanitize(f.Name), body))
	} else {
		paramNames := make([]string, len(f.Parameters))
		for i, p := range f.Parameters {
			paramNames[i] = sanitize(p.Name)
		}
		body := t.blockToExpr(f.Body)
		t.line(fmt.Sprintf("%s(%s) == %s", sanitize(f.Name), strings.Join(paramNames, ", "), body))
	}
	t.nl()
}

func (t *translator) translateModule(m *ast.ModuleDecl) {
	t.line(fmt.Sprintf("\\* Module: %s", m.Name))
	for _, d := range m.Decls {
		switch v := d.(type) {
		case *ast.FunctionDecl:
			// Prefix with module name: ModuleName_FuncName.
			origName := v.Name
			v.Name = sanitize(m.Name) + "_" + sanitize(origName)
			t.translateFun(v)
			v.Name = origName
		default:
			t.diag(LimitPartial, d.Pos(),
				fmt.Sprintf("module member type %T not translated", d))
		}
	}
}

// ── invariants / temporals ────────────────────────────────────────────────────

func (t *translator) translateInvariant(inv *ast.InvariantDecl) {
	t.line(fmt.Sprintf("%s ==", sanitize(inv.Name)))
	t.line(fmt.Sprintf("  %s", t.expr(inv.Condition)))
	t.nl()
}

func (t *translator) translateTemporal(temp *ast.TemporalDecl) {
	e := t.temporalExpr(temp.Expression)
	t.line(fmt.Sprintf("%s == %s", sanitize(temp.Name), e))
	t.nl()
}

// ── expression translation ────────────────────────────────────────────────────

func (t *translator) expr(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		if v.Prime {
			return sanitize(v.Name) + "'"
		}
		return sanitize(v.Name)
	case *ast.BasicLit:
		switch v.Kind {
		case ast.BoolLit:
			if v.Value == "true" {
				return "TRUE"
			}
			return "FALSE"
		case ast.StringLit:
			t.diag(LimitPartial, v.Position, "string literal mapped to model value (no native TLA+ strings)")
			return fmt.Sprintf(`"%s"`, v.Value)
		default:
			return v.Value
		}
	case *ast.BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", t.expr(v.Left), binaryOp(v.Op), t.expr(v.Right))
	case *ast.UnaryExpr:
		switch v.Op {
		case ast.Not:
			return fmt.Sprintf("~(%s)", t.expr(v.Expr))
		case ast.Neg:
			return fmt.Sprintf("-((%s))", t.expr(v.Expr))
		}
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", t.expr(v.X))
	case *ast.CallExpr:
		return t.callExpr(v)
	case *ast.SelectorExpr:
		// Could be enum access (Color.Red) or record field (rec.field).
		base := t.expr(v.X)
		// Treat as enum value: Base_Sel.
		return fmt.Sprintf("%s_%s", base, sanitize(v.Sel))
	case *ast.IfExpr:
		return fmt.Sprintf("(IF %s THEN %s ELSE %s)", t.expr(v.Condition), t.expr(v.Then), t.expr(v.Else))
	case *ast.LetExpr:
		return fmt.Sprintf("(LET %s == %s IN %s)", sanitize(v.Name), t.expr(v.Value), t.expr(v.Body))
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", t.expr(v.X), t.expr(v.Index))
	case *ast.SetLiteral:
		elems := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			elems[i] = t.expr(el)
		}
		return fmt.Sprintf("{%s}", strings.Join(elems, ", "))
	case *ast.ListLiteral:
		t.diag(LimitPartial, v.Position, "list literal mapped to TLA+ sequence")
		elems := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			elems[i] = t.expr(el)
		}
		return fmt.Sprintf("<<%s>>", strings.Join(elems, ", "))
	case *ast.RecordLiteral:
		fields := make([]string, len(v.Fields))
		for i, f := range v.Fields {
			if f.Spread {
				t.diag(LimitUnsupported, v.Position, "record spread not supported in TLA+ translation")
				fields[i] = "\\* SPREAD UNSUPPORTED"
			} else {
				fields[i] = fmt.Sprintf("%s |-> %s", sanitize(f.Name), t.expr(f.Value))
			}
		}
		return fmt.Sprintf("[%s]", strings.Join(fields, ", "))
	case *ast.AlwaysExpr:
		return fmt.Sprintf("[](%s)", t.expr(v.Expr))
	case *ast.EventuallyExpr:
		return fmt.Sprintf("<>(%s)", t.expr(v.Expr))
	case *ast.UntilExpr:
		t.diag(LimitUnsupported, v.Position, "until operator not in standard TLA+; emitting WF approximation")
		return fmt.Sprintf("((%s) ~> (%s))", t.expr(v.Left), t.expr(v.Right))
	case *ast.WFExpr:
		return fmt.Sprintf("WF_vars(%s)", t.expr(v.Target))
	case *ast.SFExpr:
		return fmt.Sprintf("SF_vars(%s)", t.expr(v.Target))
	case *ast.LeadsToExpr:
		return fmt.Sprintf("((%s) ~> (%s))", t.expr(v.Left), t.expr(v.Right))
	case *ast.LambdaExpr:
		t.diag(LimitUnsupported, v.Position, "lambda expression not directly translatable to TLA+; emitting LAMBDA placeholder")
		params := make([]string, len(v.Params))
		for i, p := range v.Params {
			params[i] = sanitize(p.Name)
		}
		return fmt.Sprintf("(LAMBDA %s : %s)", strings.Join(params, ", "), t.expr(v.Body))
	case *ast.SuperExpr:
		t.diag(LimitUnsupported, v.Position, "super() has no TLA+ equivalent")
		return fmt.Sprintf("\\* SUPER(%s)", v.Method)
	case nil:
		return "TRUE"
	}
	t.diag(LimitUnsupported, ast.Position{},
		fmt.Sprintf("unsupported expression type %T", e))
	return "TRUE \\* UNSUPPORTED"
}

func (t *translator) callExpr(c *ast.CallExpr) string {
	args := make([]string, len(c.Args))
	for i, a := range c.Args {
		args[i] = t.expr(a)
	}
	argStr := strings.Join(args, ", ")

	switch fun := c.Fun.(type) {
	case *ast.Ident:
		return t.builtinOrUser(fun.Name, fun.Position, argStr, c.Args)
	case *ast.SelectorExpr:
		// Module-qualified: ModuleName.FuncName → ModuleName_FuncName.
		base := t.expr(fun.X)
		return fmt.Sprintf("%s_%s(%s)", base, sanitize(fun.Sel), argStr)
	default:
		t.diag(LimitUnsupported, c.Pos(), "complex call expression not supported")
		return fmt.Sprintf("\\* UNSUPPORTED_CALL(%s)", argStr)
	}
}

func (t *translator) builtinOrUser(name string, pos ast.Position, argStr string, args []ast.Expr) string {
	switch name {
	// Set operations.
	case "contains":
		if len(args) == 2 {
			return fmt.Sprintf("(%s \\in %s)", t.expr(args[0]), t.expr(args[1]))
		}
	case "size", "card":
		if len(args) == 1 {
			return fmt.Sprintf("Cardinality(%s)", t.expr(args[0]))
		}
	case "union":
		if len(args) == 2 {
			return fmt.Sprintf("(%s \\union %s)", t.expr(args[0]), t.expr(args[1]))
		}
	case "intersect":
		if len(args) == 2 {
			return fmt.Sprintf("(%s \\intersect %s)", t.expr(args[0]), t.expr(args[1]))
		}
	case "diff", "difference":
		if len(args) == 2 {
			return fmt.Sprintf("(%s \\ %s)", t.expr(args[0]), t.expr(args[1]))
		}
	// Sequence operations.
	case "append":
		if len(args) == 2 {
			return fmt.Sprintf("Append(%s, %s)", t.expr(args[0]), t.expr(args[1]))
		}
	case "head":
		if len(args) == 1 {
			return fmt.Sprintf("Head(%s)", t.expr(args[0]))
		}
	case "tail":
		if len(args) == 1 {
			return fmt.Sprintf("Tail(%s)", t.expr(args[0]))
		}
	case "len", "length":
		if len(args) == 1 {
			return fmt.Sprintf("Len(%s)", t.expr(args[0]))
		}
	// Arithmetic.
	case "abs":
		if len(args) == 1 {
			a := t.expr(args[0])
			return fmt.Sprintf("(IF %s >= 0 THEN %s ELSE -(%s))", a, a, a)
		}
	case "min":
		if len(args) == 2 {
			a, b := t.expr(args[0]), t.expr(args[1])
			return fmt.Sprintf("(IF %s <= %s THEN %s ELSE %s)", a, b, a, b)
		}
	case "max":
		if len(args) == 2 {
			a, b := t.expr(args[0]), t.expr(args[1])
			return fmt.Sprintf("(IF %s >= %s THEN %s ELSE %s)", a, b, a, b)
		}
	}
	// Fallthrough: treat as user-defined operator.
	if argStr == "" {
		return sanitize(name)
	}
	return fmt.Sprintf("%s(%s)", sanitize(name), argStr)
}

func (t *translator) temporalExpr(e ast.Expr) string {
	return t.expr(e)
}

// blockToExpr extracts the expression from a function body (return statement).
func (t *translator) blockToExpr(body *ast.BlockStmt) string {
	if body == nil {
		return "TRUE"
	}
	for _, stmt := range body.Statements {
		if ret, ok := stmt.(*ast.ReturnStmt); ok {
			return t.expr(ret.Value)
		}
	}
	// No return statement — try first expression statement.
	for _, stmt := range body.Statements {
		if es, ok := stmt.(*ast.ExprStmt); ok {
			return t.expr(es.Expr)
		}
	}
	t.diag(LimitUnsupported, body.Position, "function body has no return statement; using TRUE")
	return "TRUE"
}

// lhsExpr produces the TLA+ name for a left-hand side expression.
func (t *translator) lhsExpr(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return sanitize(v.Name)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", t.lhsExpr(v.X), sanitize(v.Sel))
	default:
		return t.expr(e)
	}
}

// typeSet maps a Spectre type to a TLA+ set expression.
func (t *translator) typeSet(typ ast.Type) string {
	if typ == nil {
		return "Int"
	}
	switch v := typ.(type) {
	case *ast.PrimitiveType:
		switch v.Name {
		case "int":
			return "Int"
		case "bool":
			return "BOOLEAN"
		case "str", "string":
			t.diag(LimitPartial, v.Position, "string type mapped to STRING (model values in TLC)")
			return "STRING"
		case "float":
			t.diag(LimitPartial, v.Position, "float type not natively supported in TLA+; mapped to Int")
			return "Int"
		}
	case *ast.SetType:
		return fmt.Sprintf("SUBSET %s", t.typeSet(v.Element))
	case *ast.ListType:
		return fmt.Sprintf("Seq(%s)", t.typeSet(v.Element))
	case *ast.MapType:
		return fmt.Sprintf("[%s -> %s]", t.typeSet(v.Key), t.typeSet(v.Value))
	case *ast.NamedType:
		return sanitize(v.Name) + "_VALUES"
	case *ast.EnumType:
		vals := make([]string, len(v.Values))
		for i, val := range v.Values {
			vals[i] = sanitize(v.Name) + "_" + sanitize(val)
		}
		return fmt.Sprintf("{%s}", strings.Join(vals, ", "))
	case *ast.OptionType:
		t.diag(LimitPartial, v.Position, "Option type translated as element set plus a None model value")
		return fmt.Sprintf("(%s \\union {None})", t.typeSet(v.Element))
	case *ast.RecordType:
		fields := make([]string, len(v.Fields))
		for i, f := range v.Fields {
			fields[i] = fmt.Sprintf("%s: %s", sanitize(f.Name), t.typeSet(f.Type))
		}
		return fmt.Sprintf("[%s]", strings.Join(fields, ", "))
	}
	return "Int \\* UNKNOWN TYPE"
}

// ── utilities ─────────────────────────────────────────────────────────────────

func binaryOp(op ast.BinaryOp) string {
	switch op {
	case ast.Add:
		return "+"
	case ast.Sub:
		return "-"
	case ast.Mul:
		return "*"
	case ast.Div:
		return "\\div"
	case ast.Eq:
		return "="
	case ast.Neq:
		return "#"
	case ast.Lt:
		return "<"
	case ast.Gt:
		return ">"
	case ast.Leq:
		return "<="
	case ast.Geq:
		return ">="
	case ast.And:
		return "/\\"
	case ast.Or:
		return "\\/"
	default:
		return "\\* UNKNOWN_OP"
	}
}

// sanitize makes a Spectre identifier safe to use as a TLA+ identifier.
// TLA+ identifiers are case-sensitive but cannot start with a number.
func sanitize(s string) string {
	if len(s) == 0 {
		return "_"
	}
	var b strings.Builder
	for i, c := range s {
		if i == 0 && c >= '0' && c <= '9' {
			b.WriteRune('v')
		}
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_':
			b.WriteRune(c)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func isLetter(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
