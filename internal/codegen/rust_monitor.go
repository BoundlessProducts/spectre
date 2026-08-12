package codegen

import (
	"fmt"
	"strings"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// monitorInvariant is an invariant or always-temporal that must hold after every step.
type monitorInvariant struct {
	Name        string
	Description string
	RustExpr    string // translated Rust Boolean expression
}

// monitorLiveness is an eventually-temporal that must become true before step_limit.
type monitorLiveness struct {
	Name        string
	Description string
	RustExpr    string
}

// RustMonitor holds everything needed to generate the embedded monitor crate.
type RustMonitor struct {
	SpecName   string
	Vars       []varInfo
	Enums      []enumInfo
	Invariants []monitorInvariant // from invariant decls + always temporals
	Liveness   []monitorLiveness  // from eventually temporals
}

// ExtractMonitorFromAST builds a RustMonitor from a parsed spec.
func ExtractMonitorFromAST(file *ast.File, specName string) *RustMonitor {
	driver := ExtractFromAST(file, specName)
	mon := &RustMonitor{
		SpecName: specName,
		Vars:     driver.Vars,
		Enums:    driver.Enums,
	}

	var walk func([]ast.Decl)
	walk = func(decls []ast.Decl) {
		for _, d := range decls {
			switch decl := d.(type) {
			case *ast.InvariantDecl:
				mon.Invariants = append(mon.Invariants, monitorInvariant{
					Name:        decl.Name,
					Description: decl.Description,
					RustExpr:    monitorExpr(decl.Condition),
				})

			case *ast.TemporalDecl:
				switch te := decl.Expression.(type) {
				case *ast.AlwaysExpr:
					mon.Invariants = append(mon.Invariants, monitorInvariant{
						Name:        decl.Name,
						Description: decl.Description,
						RustExpr:    monitorExpr(te.Expr),
					})
				case *ast.EventuallyExpr:
					mon.Liveness = append(mon.Liveness, monitorLiveness{
						Name:        decl.Name,
						Description: decl.Description,
						RustExpr:    monitorExpr(te.Expr),
					})
				case *ast.LeadsToExpr:
					// Extract the right-side eventually, if present.
					if ev, ok := te.Right.(*ast.EventuallyExpr); ok {
						mon.Liveness = append(mon.Liveness, monitorLiveness{
							Name:        decl.Name,
							Description: decl.Description,
							RustExpr:    monitorExpr(ev.Expr),
						})
					}
				}

			case *ast.ModuleDecl:
				walk(decl.Decls)
			}
		}
	}
	walk(file.Decls)
	return mon
}

// Generate produces the complete Rust source for the embedded monitor.
func (m *RustMonitor) Generate() string {
	var b strings.Builder

	// ── file header ─────────────────────────────────────────────────────────
	b.WriteString(fmt.Sprintf("//! Embedded runtime monitor generated from `%s`.\n", m.SpecName))
	b.WriteString("//!\n")
	b.WriteString("//! Drop this file into your project to check spec invariants on every state\n")
	b.WriteString("//! transition in production or staging — no trace files, no offline replay.\n")
	b.WriteString("//!\n")
	b.WriteString("//! # Quick start\n")
	b.WriteString("//! ```rust,ignore\n")
	b.WriteString("//! let mut monitor = Monitor::new(SpecState { /* initial values */ });\n")
	b.WriteString("//! monitor.on_violation(|v| eprintln!(\"[SPEC VIOLATION] {}\", v));\n")
	b.WriteString("//!\n")
	b.WriteString("//! // After every state transition in your application:\n")
	b.WriteString("//! if let Err(violations) = monitor.step(\"action_name\", SpecState { /* new values */ }) {\n")
	b.WriteString("//!     // violations is Vec<Violation> — handle or log them\n")
	b.WriteString("//! }\n")
	b.WriteString("//!\n")
	b.WriteString("//! // At shutdown, check liveness obligations:\n")
	b.WriteString("//! for prop in monitor.unmet_liveness_properties() {\n")
	b.WriteString("//!     eprintln!(\"[SPEC] liveness property `{}` was never satisfied\", prop);\n")
	b.WriteString("//! }\n")
	b.WriteString("//! ```\n")
	b.WriteString("\n")
	b.WriteString("#![allow(dead_code)]\n\n")

	// ── enums ────────────────────────────────────────────────────────────────
	if len(m.Enums) > 0 {
		b.WriteString("// ================================================================\n")
		b.WriteString("// Enums\n")
		b.WriteString("// ================================================================\n\n")
		for _, e := range m.Enums {
			b.WriteString("#[derive(Debug, Clone, PartialEq, Eq, Hash)]\n")
			b.WriteString(fmt.Sprintf("pub enum %s {\n", e.Name))
			for _, v := range e.Values {
				b.WriteString(fmt.Sprintf("    %s,\n", v))
			}
			b.WriteString("}\n\n")
		}
	}

	// ── SpecState ────────────────────────────────────────────────────────────
	b.WriteString("// ================================================================\n")
	b.WriteString("// State\n")
	b.WriteString("// ================================================================\n\n")
	b.WriteString("/// Mirrors the spec variables. Fill in values from your application state\n")
	b.WriteString("/// before calling `Monitor::step`.\n")
	b.WriteString("#[derive(Debug, Clone)]\n")
	b.WriteString("pub struct SpecState {\n")
	for _, v := range m.Vars {
		b.WriteString(fmt.Sprintf("    pub %s: %s,\n", v.Name, rustTypeOf(v.ASTType)))
	}
	b.WriteString("}\n\n")

	// ── Violation ────────────────────────────────────────────────────────────
	b.WriteString("// ================================================================\n")
	b.WriteString("// Monitor\n")
	b.WriteString("// ================================================================\n\n")
	b.WriteString("/// A spec violation detected at runtime.\n")
	b.WriteString("#[derive(Debug, Clone)]\n")
	b.WriteString("pub struct Violation {\n")
	b.WriteString("    /// Sequential step number when this violation was detected.\n")
	b.WriteString("    pub step: u64,\n")
	b.WriteString("    /// Name of the action that triggered the check.\n")
	b.WriteString("    pub action: String,\n")
	b.WriteString("    /// Name of the invariant or temporal property that was violated.\n")
	b.WriteString("    pub property: String,\n")
	b.WriteString("    /// Human-readable description from the spec.\n")
	b.WriteString("    pub message: String,\n")
	b.WriteString("}\n\n")
	b.WriteString("impl std::fmt::Display for Violation {\n")
	b.WriteString("    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {\n")
	b.WriteString("        write!(\n")
	b.WriteString("            f,\n")
	b.WriteString("            \"[step {}] `{}` violated after `{}`: {}\",\n")
	b.WriteString("            self.step, self.property, self.action, self.message\n")
	b.WriteString("        )\n")
	b.WriteString("    }\n")
	b.WriteString("}\n\n")
	b.WriteString("struct LivenessTracker {\n")
	b.WriteString("    property: &'static str,\n")
	b.WriteString("    satisfied: bool,\n")
	b.WriteString("}\n\n")

	// ── Monitor struct ───────────────────────────────────────────────────────
	b.WriteString(fmt.Sprintf("/// Embedded runtime monitor for the `%s` specification.\n", m.SpecName))
	b.WriteString("///\n")
	b.WriteString("/// Checks all invariants after every `Monitor::step` call.\n")
	b.WriteString("pub struct Monitor {\n")
	b.WriteString("    state: SpecState,\n")
	b.WriteString("    step_count: u64,\n")
	b.WriteString("    violations: Vec<Violation>,\n")
	b.WriteString("    liveness: Vec<LivenessTracker>,\n")
	b.WriteString("    on_violation: Option<Box<dyn Fn(&Violation) + Send>>,\n")
	b.WriteString("}\n\n")

	// ── impl Monitor ─────────────────────────────────────────────────────────
	b.WriteString("impl Monitor {\n")

	// new()
	b.WriteString("    /// Create a monitor starting from `initial`.\n")
	b.WriteString("    pub fn new(initial: SpecState) -> Self {\n")
	b.WriteString("        Self {\n")
	b.WriteString("            state: initial,\n")
	b.WriteString("            step_count: 0,\n")
	b.WriteString("            violations: Vec::new(),\n")
	if len(m.Liveness) > 0 {
		b.WriteString("            liveness: vec![\n")
		for _, lp := range m.Liveness {
			b.WriteString(fmt.Sprintf("                LivenessTracker { property: %q, satisfied: false },\n", lp.Name))
		}
		b.WriteString("            ],\n")
	} else {
		b.WriteString("            liveness: Vec::new(),\n")
	}
	b.WriteString("            on_violation: None,\n")
	b.WriteString("        }\n")
	b.WriteString("    }\n\n")

	// on_violation()
	b.WriteString("    /// Register a callback invoked immediately on every violation.\n")
	b.WriteString("    /// Only one callback is supported; calling again replaces the previous one.\n")
	b.WriteString("    pub fn on_violation<F: Fn(&Violation) + Send + 'static>(&mut self, f: F) {\n")
	b.WriteString("        self.on_violation = Some(Box::new(f));\n")
	b.WriteString("    }\n\n")

	// step()
	b.WriteString("    /// Apply `action` -> `new_state` and check all invariants.\n")
	b.WriteString("    ///\n")
	b.WriteString("    /// Returns `Err` with new violations; they are also accumulated in\n")
	b.WriteString("    /// `Monitor::all_violations` regardless of the return value.\n")
	b.WriteString("    pub fn step(&mut self, action: &str, new_state: SpecState) -> Result<(), Vec<Violation>> {\n")
	b.WriteString("        self.step_count += 1;\n")
	b.WriteString("        self.state = new_state;\n")
	b.WriteString("        let new_violations = self.check_invariants(action);\n")
	b.WriteString("        for v in &new_violations {\n")
	b.WriteString("            self.violations.push(v.clone());\n")
	b.WriteString("            if let Some(cb) = &self.on_violation {\n")
	b.WriteString("                cb(v);\n")
	b.WriteString("            }\n")
	b.WriteString("        }\n")
	if len(m.Liveness) > 0 {
		b.WriteString("        self.update_liveness();\n")
	}
	b.WriteString("        if new_violations.is_empty() { Ok(()) } else { Err(new_violations) }\n")
	b.WriteString("    }\n\n")

	// accessors
	b.WriteString("    /// Current spec state.\n")
	b.WriteString("    pub fn state(&self) -> &SpecState { &self.state }\n\n")
	b.WriteString("    /// All violations recorded since the monitor was created.\n")
	b.WriteString("    pub fn all_violations(&self) -> &[Violation] { &self.violations }\n\n")
	b.WriteString("    /// Returns `true` if any invariant violation has been recorded.\n")
	b.WriteString("    pub fn has_violations(&self) -> bool { !self.violations.is_empty() }\n\n")
	b.WriteString("    /// Current step count (number of `Monitor::step` calls).\n")
	b.WriteString("    pub fn step_count(&self) -> u64 { self.step_count }\n\n")

	if len(m.Liveness) > 0 {
		b.WriteString("    /// Names of `eventually` properties not yet satisfied.\n")
		b.WriteString("    /// Call before shutdown to detect unmet liveness obligations.\n")
		b.WriteString("    pub fn unmet_liveness_properties(&self) -> Vec<&str> {\n")
		b.WriteString("        self.liveness.iter().filter(|t| !t.satisfied).map(|t| t.property).collect()\n")
		b.WriteString("    }\n\n")
	}

	// check_invariants()
	b.WriteString("    fn check_invariants(&self, action: &str) -> Vec<Violation> {\n")
	b.WriteString("        let mut out = Vec::new();\n")
	if len(m.Invariants) > 0 {
		b.WriteString("\n")
		for _, inv := range m.Invariants {
			desc := inv.Description
			if desc == "" {
				desc = inv.Name
			}
			b.WriteString(fmt.Sprintf("        // invariant `%s`\n", inv.Name))
			b.WriteString(fmt.Sprintf("        if !(%s) {\n", inv.RustExpr))
			b.WriteString("            out.push(Violation {\n")
			b.WriteString("                step: self.step_count,\n")
			b.WriteString("                action: action.to_string(),\n")
			b.WriteString(fmt.Sprintf("                property: %q.to_string(),\n", inv.Name))
			b.WriteString(fmt.Sprintf("                message: %q.to_string(),\n", desc))
			b.WriteString("            });\n")
			b.WriteString("        }\n")
		}
	} else {
		b.WriteString("        // No invariants declared in this spec.\n")
	}
	b.WriteString("        out\n")
	b.WriteString("    }\n")

	// update_liveness()
	if len(m.Liveness) > 0 {
		b.WriteString("\n    fn update_liveness(&mut self) {\n")
		b.WriteString("        // Clone state to avoid borrow conflicts during mutation.\n")
		b.WriteString("        let state = self.state.clone();\n")
		b.WriteString("        for tracker in &mut self.liveness {\n")
		b.WriteString("            if tracker.satisfied { continue; }\n")
		b.WriteString("            let satisfied = match tracker.property {\n")
		for _, lp := range m.Liveness {
			b.WriteString(fmt.Sprintf("                // eventually `%s`\n", lp.Name))
			// Replace "self.state." with "state." since state is now a local clone.
			rustExpr := strings.ReplaceAll(lp.RustExpr, "self.state.", "state.")
			b.WriteString(fmt.Sprintf("                %q => %s,\n", lp.Name, rustExpr))
		}
		b.WriteString("                _ => false,\n")
		b.WriteString("            };\n")
		b.WriteString("            if satisfied { tracker.satisfied = true; }\n")
		b.WriteString("        }\n")
		b.WriteString("    }\n")
	}

	b.WriteString("}\n")
	return b.String()
}

// monitorExpr translates a Spectre AST expression into a Rust boolean expression
// where state variables are referenced as `self.state.name`.
func monitorExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Prime {
			return fmt.Sprintf("/* primed `%s` not valid in monitor */ false", e.Name)
		}
		return "self.state." + e.Name

	case *ast.BasicLit:
		return e.Value // "42", "true", "false", `"hello"`

	case *ast.BinaryExpr:
		left := monitorExpr(e.Left)
		right := monitorExpr(e.Right)
		op := monitorBinOp(e.Op)
		return fmt.Sprintf("(%s %s %s)", left, op, right)

	case *ast.UnaryExpr:
		inner := monitorExpr(e.Expr)
		if e.Op == ast.Not {
			return fmt.Sprintf("!(%s)", inner)
		}
		return fmt.Sprintf("-(%s)", inner)

	case *ast.ParenExpr:
		return "(" + monitorExpr(e.X) + ")"

	case *ast.SelectorExpr:
		// Enum variant: EnumType.Value -> EnumType::Value
		if ident, ok := e.X.(*ast.Ident); ok {
			return ident.Name + "::" + e.Sel
		}
		// Record field: obj.field
		return monitorExpr(e.X) + "." + e.Sel

	default:
		return "/* unsupported expression */ false"
	}
}

func monitorBinOp(op ast.BinaryOp) string {
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
		return "==" // Spectre `=` is equality; Rust uses `==`
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
	default:
		return "?"
	}
}
