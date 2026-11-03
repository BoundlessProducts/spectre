package main

// commands_fmt.go — Item 18: spec formatter and machine-readable output.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/diagnose"
	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func init() {
	Commands["fmt"] = &Command{
		Name:        "fmt",
		Description: "Format (pretty-print) a Spectre spec file in canonical style",
		Run:         runFmt,
	}
	Commands["check-json"] = &Command{
		Name:        "check-json",
		Description: "Run verification and emit machine-readable JSON output (exit 1 on violations)",
		Run:         runCheckJSON,
	}
}

// ── spectre fmt ───────────────────────────────────────────────────────────────

func runFmt(args []string) error {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	write := fs.Bool("w", false, "Write result back to the input file (default: print to stdout)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: spectre fmt [-w] <spec.spec> [<spec2.spec>...]")
	}

	for _, filename := range fs.Args() {
		src, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("reading %s: %w", filename, err)
		}
		l := lexer.New(string(src))
		p := parser.New(l)
		file := p.ParseFile()
		if len(p.Errors()) > 0 {
			return fmt.Errorf("parse errors in %s:\n  %s", filename, strings.Join(p.Errors(), "\n  "))
		}
		formatted := FormatSpec(file)
		if *write {
			if err := os.WriteFile(filename, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("writing %s: %w", filename, err)
			}
		} else {
			fmt.Print(formatted)
		}
	}
	return nil
}

// FormatSpec pretty-prints a Spectre AST in canonical style.
// The output is valid Spectre that re-parses to the same AST.
func FormatSpec(file *ast.File) string {
	var sb strings.Builder
	f := &formatter{w: &sb}
	f.formatFile(file)
	return sb.String()
}

type formatter struct {
	w *strings.Builder
}

func (f *formatter) line(s string) { f.w.WriteString(s); f.w.WriteByte('\n') }
func (f *formatter) nl()           { f.w.WriteByte('\n') }

func (f *formatter) formatFile(file *ast.File) {
	// Group declarations by kind for canonical ordering.
	var params []*ast.ParamDecl
	var consts []*ast.ConstantDecl
	var enums []*ast.EnumDecl
	var types []*ast.TypeAliasDecl
	var vars []*ast.VariableDecl
	var funs []*ast.FunctionDecl
	var modules []*ast.ModuleDecl
	var actions []*ast.ActionDecl
	var inits []ast.Decl
	var invariants []*ast.InvariantDecl
	var temporals []*ast.TemporalDecl
	var imports []*ast.ImportDecl

	for _, d := range file.Decls {
		switch v := d.(type) {
		case *ast.ParamDecl:
			params = append(params, v)
		case *ast.ConstantDecl:
			consts = append(consts, v)
		case *ast.EnumDecl:
			enums = append(enums, v)
		case *ast.TypeAliasDecl:
			types = append(types, v)
		case *ast.VariableDecl:
			vars = append(vars, v)
		case *ast.FunctionDecl:
			funs = append(funs, v)
		case *ast.ModuleDecl:
			modules = append(modules, v)
		case *ast.ActionDecl:
			actions = append(actions, v)
		case *ast.InitDecl:
			inits = append(inits, v)
		case *ast.OneOfInitDecl:
			inits = append(inits, v)
		case *ast.InvariantDecl:
			invariants = append(invariants, v)
		case *ast.TemporalDecl:
			temporals = append(temporals, v)
		case *ast.ImportDecl:
			imports = append(imports, v)
		}
	}

	for _, d := range imports {
		f.line(fmt.Sprintf("import %q", d.Path))
	}
	if len(imports) > 0 {
		f.nl()
	}

	for _, d := range params {
		f.line(fmt.Sprintf("param %s: %s", d.Name, fmtType(d.Type)))
	}
	if len(params) > 0 {
		f.nl()
	}

	for _, d := range consts {
		f.line(fmt.Sprintf("const %s", d.Name))
	}
	if len(consts) > 0 {
		f.nl()
	}

	for _, d := range enums {
		f.line(fmt.Sprintf("enum %s { %s }", d.Name, strings.Join(d.Values, ", ")))
	}
	if len(enums) > 0 {
		f.nl()
	}

	for _, d := range types {
		f.line(fmt.Sprintf("type %s = %s", d.Name, fmtType(d.Type)))
	}
	if len(types) > 0 {
		f.nl()
	}

	for _, d := range vars {
		f.line(fmt.Sprintf("var %s: %s", d.Name, fmtType(d.Type)))
	}
	if len(vars) > 0 {
		f.nl()
	}

	for _, d := range modules {
		f.formatModule(d)
		f.nl()
	}

	for _, d := range funs {
		f.formatFun(d)
		f.nl()
	}

	for _, d := range inits {
		f.formatInit(d)
		f.nl()
	}

	for _, d := range actions {
		f.formatAction(d)
		f.nl()
	}

	for _, d := range invariants {
		f.formatInvariant(d)
		f.nl()
	}

	for _, d := range temporals {
		f.formatTemporal(d)
		f.nl()
	}
}

func (f *formatter) formatModule(m *ast.ModuleDecl) {
	f.line(fmt.Sprintf("module %s {", m.Name))
	for _, d := range m.Decls {
		switch v := d.(type) {
		case *ast.FunctionDecl:
			vis := ""
			if v.Visibility == ast.Public {
				vis = "public "
			}
			params := fmtParams(v.Parameters)
			ret := ""
			if v.ReturnType != nil {
				ret = ": " + fmtType(v.ReturnType)
			}
			f.line(fmt.Sprintf("  %sfun %s(%s)%s {", vis, v.Name, params, ret))
			f.formatBlock(v.Body, "    ")
			f.line("  }")
		default:
			// Fallback: skip unsupported module members silently.
		}
	}
	f.line("}")
}

func (f *formatter) formatFun(d *ast.FunctionDecl) {
	vis := ""
	if d.Visibility == ast.Public {
		vis = "public "
	}
	params := fmtParams(d.Parameters)
	ret := ""
	if d.ReturnType != nil {
		ret = ": " + fmtType(d.ReturnType)
	}
	f.line(fmt.Sprintf("%sfun %s(%s)%s {", vis, d.Name, params, ret))
	f.formatBlock(d.Body, "  ")
	f.line("}")
}

func (f *formatter) formatAction(d *ast.ActionDecl) {
	params := fmtParams(d.Parameters)
	if params != "" {
		f.line(fmt.Sprintf("action %s(%s) {", d.Name, params))
	} else {
		f.line(fmt.Sprintf("action %s {", d.Name))
	}
	f.formatBlock(d.Body, "  ")
	f.line("}")
}

func (f *formatter) formatInit(d ast.Decl) {
	switch v := d.(type) {
	case *ast.InitDecl:
		f.line("init {")
		if v.Body != nil {
			f.formatBlock(v.Body, "  ")
		} else if v.Expression != nil {
			f.line("  " + fmtExpr(v.Expression))
		}
		f.line("}")
	case *ast.OneOfInitDecl:
		f.line("init oneOf {")
		for i, opt := range v.Options {
			sep := ","
			if i == len(v.Options)-1 {
				sep = ""
			}
			f.line("  {")
			f.formatBlock(opt, "    ")
			f.line("  }" + sep)
		}
		f.line("}")
	}
}

func (f *formatter) formatInvariant(d *ast.InvariantDecl) {
	f.line(fmt.Sprintf("invariant %s { %s }", d.Name, fmtExpr(d.Condition)))
}

func (f *formatter) formatTemporal(d *ast.TemporalDecl) {
	f.line(fmt.Sprintf("temporal %s { %s }", d.Name, fmtExpr(d.Expression)))
}

func (f *formatter) formatBlock(b *ast.BlockStmt, indent string) {
	if b == nil {
		return
	}
	for _, s := range b.Statements {
		f.w.WriteString(indent)
		f.w.WriteString(fmtStmt(s))
		f.w.WriteByte('\n')
	}
}

// ── expression / statement printers ──────────────────────────────────────────

func fmtStmt(s ast.Stmt) string {
	switch v := s.(type) {
	case *ast.AssignStmt:
		return fmt.Sprintf("%s = %s", fmtExpr(v.Left), fmtExpr(v.Right))
	case *ast.RequireStmt:
		return fmt.Sprintf("require %s", fmtExpr(v.Condition))
	case *ast.EnsureStmt:
		return fmt.Sprintf("ensure %s", fmtExpr(v.Condition))
	case *ast.ReturnStmt:
		return fmt.Sprintf("return %s", fmtExpr(v.Value))
	case *ast.ExprStmt:
		return fmtExpr(v.Expr)
	default:
		return fmt.Sprintf("/* unsupported stmt %T */", s)
	}
}

func fmtExpr(e ast.Expr) string {
	if e == nil {
		return ""
	}
	switch v := e.(type) {
	case *ast.Ident:
		if v.Prime {
			return v.Name + "'"
		}
		return v.Name
	case *ast.BasicLit:
		return v.Value
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", fmtExpr(v.Left), fmtBinOp(v.Op), fmtExpr(v.Right))
	case *ast.UnaryExpr:
		switch v.Op {
		case ast.Not:
			return fmt.Sprintf("!%s", fmtExprParens(v.Expr))
		case ast.Neg:
			return fmt.Sprintf("-%s", fmtExprParens(v.Expr))
		}
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", fmtExpr(v.X))
	case *ast.CallExpr:
		args := make([]string, len(v.Args))
		for i, a := range v.Args {
			args[i] = fmtExpr(a)
		}
		return fmt.Sprintf("%s(%s)", fmtExpr(v.Fun), strings.Join(args, ", "))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", fmtExpr(v.X), v.Sel)
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", fmtExpr(v.X), fmtExpr(v.Index))
	case *ast.IfExpr:
		return fmt.Sprintf("if %s { %s } else { %s }", fmtExpr(v.Condition), fmtExpr(v.Then), fmtExpr(v.Else))
	case *ast.LetExpr:
		return fmt.Sprintf("let %s = %s in %s", v.Name, fmtExpr(v.Value), fmtExpr(v.Body))
	case *ast.SetLiteral:
		elems := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			elems[i] = fmtExpr(el)
		}
		return fmt.Sprintf("{%s}", strings.Join(elems, ", "))
	case *ast.ListLiteral:
		elems := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			elems[i] = fmtExpr(el)
		}
		return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
	case *ast.RecordLiteral:
		fields := make([]string, len(v.Fields))
		for i, field := range v.Fields {
			if field.Spread {
				fields[i] = "..." + fmtExpr(&ast.Ident{Name: field.Name})
			} else {
				fields[i] = field.Name + ": " + fmtExpr(field.Value)
			}
		}
		return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
	case *ast.AlwaysExpr:
		return fmt.Sprintf("always(%s)", fmtExpr(v.Expr))
	case *ast.EventuallyExpr:
		return fmt.Sprintf("eventually(%s)", fmtExpr(v.Expr))
	case *ast.UntilExpr:
		return fmt.Sprintf("(%s until %s)", fmtExpr(v.Left), fmtExpr(v.Right))
	case *ast.WFExpr:
		return fmt.Sprintf("WF(%s)", fmtExpr(v.Target))
	case *ast.SFExpr:
		return fmt.Sprintf("SF(%s)", fmtExpr(v.Target))
	case *ast.LeadsToExpr:
		return fmt.Sprintf("(%s ~> %s)", fmtExpr(v.Left), fmtExpr(v.Right))
	case *ast.LambdaExpr:
		params := make([]string, len(v.Params))
		for i, p := range v.Params {
			params[i] = p.Name
		}
		return fmt.Sprintf("(%s) => %s", strings.Join(params, ", "), fmtExpr(v.Body))
	}
	return fmt.Sprintf("/* unsupported expr %T */", e)
}

func fmtExprParens(e ast.Expr) string {
	switch e.(type) {
	case *ast.BinaryExpr:
		return fmt.Sprintf("(%s)", fmtExpr(e))
	}
	return fmtExpr(e)
}

func fmtBinOp(op ast.BinaryOp) string {
	switch op {
	case ast.Add:
		return "+"
	case ast.Sub:
		return "-"
	case ast.Mul:
		return "*"
	case ast.Div:
		return "/"
	case ast.Eq:
		return "=="
	case ast.Neq:
		return "!="
	case ast.Lt:
		return "<"
	case ast.Gt:
		return ">"
	case ast.Leq:
		return "<="
	case ast.Geq:
		return ">="
	case ast.And:
		return "&&"
	case ast.Or:
		return "||"
	}
	return "?"
}

func fmtType(t ast.Type) string {
	if t == nil {
		return "unknown"
	}
	switch v := t.(type) {
	case *ast.PrimitiveType:
		return v.Name
	case *ast.SetType:
		return fmt.Sprintf("Set[%s]", fmtType(v.Element))
	case *ast.ListType:
		return fmt.Sprintf("List[%s]", fmtType(v.Element))
	case *ast.MapType:
		return fmt.Sprintf("Map[%s, %s]", fmtType(v.Key), fmtType(v.Value))
	case *ast.NamedType:
		return v.Name
	case *ast.EnumType:
		return v.Name
	case *ast.OptionType:
		return fmt.Sprintf("Option[%s]", fmtType(v.Element))
	case *ast.RecordType:
		fields := make([]string, len(v.Fields))
		for i, f := range v.Fields {
			fields[i] = f.Name + ": " + fmtType(f.Type)
		}
		return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
	}
	return "unknown"
}

func fmtParams(params []ast.Parameter) string {
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.Name + ": " + fmtType(p.Type)
	}
	return strings.Join(parts, ", ")
}

// ── spectre check-json ────────────────────────────────────────────────────────

// JSONResult is the machine-readable output of check-json.
type JSONResult struct {
	Version        int                `json:"version"`
	SpecFile       string             `json:"spec_file"`
	Kind           string             `json:"kind"`
	StatesExplored int                `json:"states_explored"`
	Violations     []JSONViolation    `json:"violations"`
	Errors         []string           `json:"errors,omitempty"`
	DomainWarnings []string           `json:"domain_warnings,omitempty"`
}

// JSONViolation is the machine-readable form of an invariant violation.
type JSONViolation struct {
	Invariant   string           `json:"invariant"`
	Description string           `json:"description"`
	TraceLength int              `json:"trace_length"`
	Trace       []JSONTraceStep  `json:"trace"`
	FinalState  map[string]string `json:"final_state"`
	StateDiffs  []JSONStateDiff  `json:"state_diffs"`
}

// JSONTraceStep is one step in a counterexample trace.
type JSONTraceStep struct {
	Step   int               `json:"step"`
	Action string            `json:"action"`
	Diffs  []JSONStateDiff   `json:"diffs"`
}

// JSONStateDiff records a single variable change.
type JSONStateDiff struct {
	Variable string `json:"variable"`
	Before   string `json:"before"`
	After    string `json:"after"`
}

func runCheckJSON(args []string) error {
	fs := flag.NewFlagSet("check-json", flag.ContinueOnError)
	maxStates := fs.Int("max-states", 10000, "Maximum states to explore")
	maxDepth := fs.Int("max-depth", 100, "Maximum exploration depth")
	output := fs.String("output", "", "Write JSON to file (default: stdout)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: spectre check-json [--max-states N] [--max-depth N] [--output file.json] <spec.spec>")
	}
	filename := fs.Arg(0)

	content, err := os.ReadFile(filename)
	if err != nil {
		return emitJSONError(filename, err.Error(), *output)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return emitJSONError(filename, strings.Join(p.Errors(), "; "), *output)
	}

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		return emitJSONError(filename, err.Error(), *output)
	}

	exp := explore.NewExplorer(sm)
	exp.SetMaxStates(*maxStates)
	exp.SetMaxDepth(*maxDepth)
	exp.SetSilent(true)

	result, err := exp.ExploreBFS()
	if err != nil {
		return emitJSONError(filename, err.Error(), *output)
	}

	jr := JSONResult{
		Version:        1,
		SpecFile:       filename,
		Kind:           result.Kind.String(),
		StatesExplored: result.StatesExplored,
		DomainWarnings: result.DomainWarnings,
	}

	for _, v := range result.Violations {
		jv := JSONViolation{
			Invariant:   v.Invariant,
			Description: v.Description,
			TraceLength: len(v.Path),
		}
		if v.State != nil {
			jv.FinalState = make(map[string]string)
			for k, val := range v.State.Variables {
				jv.FinalState[k] = val.String()
			}
		}
		for i, tr := range v.Path {
			step := JSONTraceStep{
				Step:   i + 1,
				Action: tr.Action,
			}
			if tr.FromState != nil && tr.ToState != nil {
				rawDiffs := diagnose.StateDiff(tr.FromState, tr.ToState)
				for _, d := range rawDiffs {
					before, after := "<absent>", "<absent>"
					if d.Before != nil {
						before = d.Before.String()
					}
					if d.After != nil {
						after = d.After.String()
					}
					step.Diffs = append(step.Diffs, JSONStateDiff{
						Variable: d.Name,
						Before:   before,
						After:    after,
					})
				}
			}
			jv.Trace = append(jv.Trace, step)
		}

		// Top-level per-step diffs (flattened across whole trace).
		for _, tr := range v.Path {
			if tr.FromState != nil && tr.ToState != nil {
				rawDiffs := diagnose.StateDiff(tr.FromState, tr.ToState)
				for _, d := range rawDiffs {
					before, after := "<absent>", "<absent>"
					if d.Before != nil {
						before = d.Before.String()
					}
					if d.After != nil {
						after = d.After.String()
					}
					jv.StateDiffs = append(jv.StateDiffs, JSONStateDiff{
						Variable: d.Name,
						Before:   before,
						After:    after,
					})
				}
			}
		}

		jr.Violations = append(jr.Violations, jv)
	}

	// Stable sort violations by invariant name.
	sort.Slice(jr.Violations, func(i, j int) bool {
		return jr.Violations[i].Invariant < jr.Violations[j].Invariant
	})

	data, _ := json.MarshalIndent(jr, "", "  ")

	var w io.Writer = os.Stdout
	if *output != "" {
		f, createErr := os.Create(*output)
		if createErr != nil {
			return fmt.Errorf("creating output file: %w", createErr)
		}
		defer f.Close()
		w = f
	}
	fmt.Fprintf(w, "%s\n", data)

	if len(jr.Violations) > 0 {
		return fmt.Errorf("verification failed: %d violation(s)", len(jr.Violations))
	}
	return nil
}

func emitJSONError(filename, errMsg, output string) error {
	jr := JSONResult{
		Version:  1,
		SpecFile: filename,
		Kind:     "error",
		Errors:   []string{errMsg},
	}
	data, _ := json.MarshalIndent(jr, "", "  ")
	w := os.Stdout
	if output != "" {
		if f, err := os.Create(output); err == nil {
			defer f.Close()
			fmt.Fprintf(f, "%s\n", data)
			return fmt.Errorf("error: %s", errMsg)
		}
	}
	fmt.Fprintf(w, "%s\n", data)
	return fmt.Errorf("error: %s", errMsg)
}
