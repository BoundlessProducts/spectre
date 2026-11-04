// Package tla translates Spectre specs to TLA+ and runs TLC / Apalache.
package tla

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// Translate converts a parsed Spectre file into a TLA+ module string and a
// TLC config string.  params overrides param defaults (name → integer string).
func Translate(file *ast.File, params map[string]string) (tlaModule string, cfgFile string, err error) {
	t := &translator{file: file, params: params}
	return t.run()
}

type translator struct {
	file   *ast.File
	params map[string]string

	vars       []string
	varTypes   map[string]ast.Type
	paramDecls []*ast.ParamDecl
	inits      []*ast.InitDecl
	actions    []*ast.ActionDecl
	invs       []*ast.InvariantDecl
	temps      []*ast.TemporalDecl
	enums      []*ast.EnumDecl
}

func (t *translator) run() (string, string, error) {
	t.varTypes = make(map[string]ast.Type)

	for _, d := range t.file.Decls {
		switch v := d.(type) {
		case *ast.VariableDecl:
			t.vars = append(t.vars, v.Name)
			t.varTypes[v.Name] = v.Type
		case *ast.ParamDecl:
			t.paramDecls = append(t.paramDecls, v)
		case *ast.InitDecl:
			t.inits = append(t.inits, v)
		case *ast.ActionDecl:
			t.actions = append(t.actions, v)
		case *ast.InvariantDecl:
			t.invs = append(t.invs, v)
		case *ast.TemporalDecl:
			t.temps = append(t.temps, v)
		case *ast.EnumDecl:
			t.enums = append(t.enums, v)
		}
	}

	// Use a fixed module name; TLA+ module names cannot contain hyphens.
	modName := "SpectreSpec"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---- MODULE %s ----\n", modName))
	sb.WriteString("EXTENDS Integers, Sequences, FiniteSets, TLC\n\n")

	// CONSTANTS (params).
	if len(t.paramDecls) > 0 {
		names := make([]string, len(t.paramDecls))
		for i, p := range t.paramDecls {
			names[i] = p.Name
		}
		sb.WriteString("CONSTANTS " + strings.Join(names, ", ") + "\n\n")
	}

	// VARIABLES with Apalache @type annotations (ignored by TLC/SANY).
	if len(t.vars) > 0 {
		sb.WriteString("VARIABLES\n")
		for i, name := range t.vars {
			sb.WriteString("  \\* @type: " + apalacheType(t.varTypes[name]) + ";\n")
			if i < len(t.vars)-1 {
				sb.WriteString("  " + name + ",\n")
			} else {
				sb.WriteString("  " + name + "\n")
			}
		}
		sb.WriteString("\n")
		// @type annotation disambiguates tuple vs sequence for Apalache's Snowcat.
		tupleTypes := make([]string, len(t.vars))
		for i, name := range t.vars {
			tupleTypes[i] = apalacheType(t.varTypes[name])
		}
		sb.WriteString("\\* @type: <<" + strings.Join(tupleTypes, ", ") + ">>;\n")
		sb.WriteString("vars == <<" + strings.Join(t.vars, ", ") + ">>\n\n")
	}

	// TypeOK.
	if len(t.vars) > 0 {
		sb.WriteString("TypeOK ==\n")
		for _, name := range t.vars {
			sb.WriteString("  /\\ " + name + " \\in " + typeSet(t.varTypes[name]) + "\n")
		}
		sb.WriteString("\n")
	}

	// Init.
	sb.WriteString("Init ==\n")
	if len(t.inits) > 0 {
		for _, init := range t.inits {
			if init.Body != nil {
				for _, stmt := range init.Body.Statements {
					if as, ok := stmt.(*ast.AssignStmt); ok {
						sb.WriteString("  /\\ " + exprToTLA(as.Left) + " = " + exprToTLA(as.Right) + "\n")
					}
				}
			}
		}
	} else {
		sb.WriteString("  /\\ TRUE\n")
	}
	sb.WriteString("\n")

	// Actions.
	for _, a := range t.actions {
		sb.WriteString(t.translateAction(a))
		sb.WriteString("\n")
	}

	// Next.
	if len(t.actions) > 0 {
		names := make([]string, len(t.actions))
		for i, a := range t.actions {
			names[i] = capitalize(a.Name)
		}
		sb.WriteString("Next == " + strings.Join(names, " \\/ ") + "\n\n")
	} else {
		sb.WriteString("Next == FALSE\n\n")
	}

	// Invariants.
	for _, inv := range t.invs {
		sb.WriteString(fmt.Sprintf("%s == %s\n\n", inv.Name, exprToTLA(inv.Condition)))
	}

	// Temporal properties.
	for _, tp := range t.temps {
		sb.WriteString(fmt.Sprintf("%s == %s\n\n", tp.Name, exprToTLA(tp.Expression)))
	}

	// Enums as TLA+ model values.
	for _, e := range t.enums {
		sb.WriteString(fmt.Sprintf("%s == {%s}\n\n", e.Name, strings.Join(e.Values, ", ")))
	}

	// Spec definition with fairness extracted from temporal declarations.
	fairness := t.collectFairness()
	sb.WriteString("Spec == Init /\\ [][Next]_vars")
	if len(fairness) > 0 {
		sb.WriteString("\n")
		for _, f := range fairness {
			sb.WriteString("  /\\ " + f + "\n")
		}
	} else {
		sb.WriteString("\n")
	}
	sb.WriteString("\n====\n")

	cfg := t.buildConfig()
	return sb.String(), cfg, nil
}

func (t *translator) translateAction(a *ast.ActionDecl) string {
	var sb strings.Builder
	aName := capitalize(a.Name)
	sb.WriteString(aName + " ==\n")

	// Collect primed var names from assignment targets.
	primed := map[string]bool{}
	if a.Body != nil {
		for _, stmt := range a.Body.Statements {
			if as, ok := stmt.(*ast.AssignStmt); ok {
				collectPrimed(as.Left, primed)
			}
		}
	}

	// Guard condition (when clause).
	if a.Guard != nil {
		sb.WriteString("  /\\ " + exprToTLA(a.Guard) + "\n")
	}

	// Require and assign statements.
	if a.Body != nil {
		for _, stmt := range a.Body.Statements {
			switch s := stmt.(type) {
			case *ast.RequireStmt:
				sb.WriteString("  /\\ " + exprToTLA(s.Condition) + "\n")
			case *ast.AssignStmt:
				sb.WriteString("  /\\ " + exprToTLA(s.Left) + " = " + exprToTLA(s.Right) + "\n")
			}
		}
	}

	// UNCHANGED for vars not assigned.
	var unchanged []string
	for _, v := range t.vars {
		if !primed[v] {
			unchanged = append(unchanged, v)
		}
	}
	switch len(unchanged) {
	case 0:
		// all vars assigned — no UNCHANGED needed
	case 1:
		sb.WriteString("  /\\ UNCHANGED " + unchanged[0] + "\n")
	default:
		sb.WriteString("  /\\ UNCHANGED <<" + strings.Join(unchanged, ", ") + ">>\n")
	}

	return sb.String()
}

func (t *translator) collectFairness() []string {
	var result []string
	for _, tp := range t.temps {
		result = append(result, collectFairnessFromExpr(tp.Expression)...)
	}
	return result
}

func collectFairnessFromExpr(e ast.Expr) []string {
	if e == nil {
		return nil
	}
	var result []string
	switch v := e.(type) {
	case *ast.WFExpr:
		result = append(result, fmt.Sprintf("WF_vars(%s)", exprToTLA(v.Target)))
	case *ast.SFExpr:
		result = append(result, fmt.Sprintf("SF_vars(%s)", exprToTLA(v.Target)))
	case *ast.BinaryExpr:
		result = append(result, collectFairnessFromExpr(v.Left)...)
		result = append(result, collectFairnessFromExpr(v.Right)...)
	case *ast.UnaryExpr:
		result = append(result, collectFairnessFromExpr(v.Expr)...)
	case *ast.AlwaysExpr:
		result = append(result, collectFairnessFromExpr(v.Expr)...)
	case *ast.EventuallyExpr:
		result = append(result, collectFairnessFromExpr(v.Expr)...)
	}
	return result
}

func (t *translator) buildConfig() string {
	var sb strings.Builder
	sb.WriteString("INIT Init\n")
	sb.WriteString("NEXT Next\n")
	for _, inv := range t.invs {
		sb.WriteString("INVARIANT " + inv.Name + "\n")
	}
	for _, tp := range t.temps {
		sb.WriteString("PROPERTY " + tp.Name + "\n")
	}
	if len(t.paramDecls) > 0 {
		sb.WriteString("CONSTANTS\n")
		for _, p := range t.paramDecls {
			if t.params != nil {
				if v, ok := t.params[p.Name]; ok {
					sb.WriteString(fmt.Sprintf("  %s = %s\n", p.Name, v))
					continue
				}
			}
			// No default on ParamDecl — use 0 as a safe fallback.
			sb.WriteString(fmt.Sprintf("  %s = 0\n", p.Name))
		}
	}
	return sb.String()
}

// exprToTLA converts a Spectre expression to TLA+ syntax.
func exprToTLA(e ast.Expr) string {
	if e == nil {
		return "TRUE"
	}
	switch v := e.(type) {
	case *ast.Ident:
		if v.Prime {
			return v.Name + "'"
		}
		return v.Name
	case *ast.BasicLit:
		switch v.Kind {
		case ast.BoolLit:
			if strings.ToLower(v.Value) == "true" {
				return "TRUE"
			}
			return "FALSE"
		default:
			return v.Value
		}
	case *ast.BinaryExpr:
		l := exprToTLA(v.Left)
		r := exprToTLA(v.Right)
		switch v.Op {
		case ast.Add:
			return fmt.Sprintf("(%s + %s)", l, r)
		case ast.Sub:
			return fmt.Sprintf("(%s - %s)", l, r)
		case ast.Mul:
			return fmt.Sprintf("(%s * %s)", l, r)
		case ast.Div:
			return fmt.Sprintf("(%s \\div %s)", l, r)
		case ast.Eq:
			return fmt.Sprintf("%s = %s", l, r)
		case ast.Neq:
			return fmt.Sprintf("%s # %s", l, r)
		case ast.Lt:
			return fmt.Sprintf("%s < %s", l, r)
		case ast.Gt:
			return fmt.Sprintf("%s > %s", l, r)
		case ast.Leq:
			return fmt.Sprintf("%s =< %s", l, r)
		case ast.Geq:
			return fmt.Sprintf("%s >= %s", l, r)
		case ast.And:
			return fmt.Sprintf("(%s /\\ %s)", l, r)
		case ast.Or:
			return fmt.Sprintf("(%s \\/ %s)", l, r)
		case ast.In:
			return fmt.Sprintf("%s \\in %s", l, r)
		}
	case *ast.UnaryExpr:
		if v.Op == ast.Not {
			return fmt.Sprintf("~(%s)", exprToTLA(v.Expr))
		}
	case *ast.AlwaysExpr:
		return fmt.Sprintf("[](%s)", exprToTLA(v.Expr))
	case *ast.EventuallyExpr:
		return fmt.Sprintf("<>(%s)", exprToTLA(v.Expr))
	case *ast.UntilExpr:
		return fmt.Sprintf("(%s U %s)", exprToTLA(v.Left), exprToTLA(v.Right))
	case *ast.WFExpr:
		return fmt.Sprintf("WF_vars(%s)", exprToTLA(v.Target))
	case *ast.SFExpr:
		return fmt.Sprintf("SF_vars(%s)", exprToTLA(v.Target))
	case *ast.LeadsToExpr:
		return fmt.Sprintf("(%s ~> %s)", exprToTLA(v.Left), exprToTLA(v.Right))
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", exprToTLA(v.X))
	case *ast.SetLiteral:
		if len(v.Elements) == 0 {
			return "{}"
		}
		elems := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			elems[i] = exprToTLA(el)
		}
		return "{" + strings.Join(elems, ", ") + "}"
	case *ast.ListLiteral:
		if len(v.Elements) == 0 {
			return "<<>>"
		}
		elems := make([]string, len(v.Elements))
		for i, el := range v.Elements {
			elems[i] = exprToTLA(el)
		}
		return "<<" + strings.Join(elems, ", ") + ">>"
	case *ast.CallExpr:
		// Detect method calls on collections: obj.method(args...)
		if sel, ok := v.Fun.(*ast.SelectorExpr); ok {
			obj := exprToTLA(sel.X)
			args := make([]string, len(v.Args))
			for i, a := range v.Args {
				args[i] = exprToTLA(a)
			}
			switch sel.Sel {
			case "union":
				if len(args) == 1 {
					return fmt.Sprintf("(%s \\union %s)", obj, args[0])
				}
			case "intersection":
				if len(args) == 1 {
					return fmt.Sprintf("(%s \\intersect %s)", obj, args[0])
				}
			case "difference":
				if len(args) == 1 {
					return fmt.Sprintf("(%s \\ %s)", obj, args[0])
				}
			case "contains":
				if len(args) == 1 {
					return fmt.Sprintf("%s \\in %s", args[0], obj)
				}
			case "size":
				return fmt.Sprintf("Cardinality(%s)", obj)
			case "put":
				if len(args) == 2 {
					return fmt.Sprintf("[%s EXCEPT ![%s] = %s]", obj, args[0], args[1])
				}
			case "get":
				if len(args) == 1 {
					return fmt.Sprintf("%s[%s]", obj, args[0])
				}
			case "append":
				if len(args) == 1 {
					return fmt.Sprintf("Append(%s, %s)", obj, args[0])
				}
			case "head":
				return fmt.Sprintf("Head(%s)", obj)
			case "tail":
				return fmt.Sprintf("Tail(%s)", obj)
			case "toList":
				return obj // best-effort
			case "toSet":
				return fmt.Sprintf("Range(%s)", obj)
			}
		}
		// Generic function call fallback
		fun := exprToTLA(v.Fun)
		args := make([]string, len(v.Args))
		for i, a := range v.Args {
			args[i] = exprToTLA(a)
		}
		return fmt.Sprintf("%s(%s)", fun, strings.Join(args, ", "))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", exprToTLA(v.X), v.Sel)
	}
	return fmt.Sprintf("TRUE (* unsupported: %T *)", e)
}

// collectPrimed records variable names that appear as primed idents in an lhs expr.
func collectPrimed(e ast.Expr, primed map[string]bool) {
	if e == nil {
		return
	}
	if id, ok := e.(*ast.Ident); ok && id.Prime {
		primed[id.Name] = true
		return
	}
	switch v := e.(type) {
	case *ast.SelectorExpr:
		collectPrimed(v.X, primed)
	case *ast.IndexExpr:
		collectPrimed(v.X, primed)
	}
}

// typeSet returns the TLA+ set membership expression for a Spectre type.
func typeSet(typ ast.Type) string {
	if typ == nil {
		return "Int"
	}
	switch t := typ.(type) {
	case *ast.PrimitiveType:
		switch t.Name {
		case "int":
			return "Int"
		case "bool":
			return "BOOLEAN"
		case "str", "string":
			return "STRING"
		case "float":
			return "Real"
		}
	case *ast.EnumType:
		return t.Name
	case *ast.SetType:
		return "SUBSET " + typeSet(t.Element)
	case *ast.MapType:
		return "[" + typeSet(t.Key) + " -> " + typeSet(t.Value) + "]"
	case *ast.ListType:
		return "Seq(" + typeSet(t.Element) + ")"
	}
	return "Int"
}

// apalacheType returns the Apalache/Snowcat type annotation string for a Spectre type.
func apalacheType(typ ast.Type) string {
	if typ == nil {
		return "Int"
	}
	switch t := typ.(type) {
	case *ast.PrimitiveType:
		switch t.Name {
		case "int":
			return "Int"
		case "bool":
			return "Bool"
		case "str", "string":
			return "Str"
		}
	case *ast.EnumType:
		return "Str"
	case *ast.SetType:
		return "Set(" + apalacheType(t.Element) + ")"
	case *ast.MapType:
		return apalacheType(t.Key) + " -> " + apalacheType(t.Value)
	case *ast.ListType:
		return "Seq(" + apalacheType(t.Element) + ")"
	}
	return "Int"
}

// capitalize uppercases the first rune of s.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
