package eval

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// PurityChecker verifies that pure functions do not access state variables
type PurityChecker struct {
	stateVariables map[string]bool // Set of state variable names
	errors          []string
}

// NewPurityChecker creates a new purity checker with the given state variables
func NewPurityChecker(stateVars *state.VariableModel) *PurityChecker {
	checker := &PurityChecker{
		stateVariables: make(map[string]bool),
		errors:          []string{},
	}

	// Populate state variables from the model
	if stateVars != nil {
		for _, name := range stateVars.GetVariableNames() {
			checker.stateVariables[name] = true
		}
	}

	return checker
}

// CheckFunction checks if a function is pure (doesn't access state variables)
func (pc *PurityChecker) CheckFunction(fnDecl *ast.FunctionDecl) error {
	pc.errors = []string{}
	pc.checkBlock(fnDecl.Body, "function body")
	
	if len(pc.errors) > 0 {
		return fmt.Errorf("purity violation in function %s: %v", fnDecl.Name, pc.errors)
	}
	return nil
}

// checkExpression checks an expression for state variable access
func (pc *PurityChecker) checkExpression(expr ast.Expr, context string) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.Ident:
		pc.checkIdentifier(e, context)
	case *ast.BinaryExpr:
		pc.checkExpression(e.Left, context)
		pc.checkExpression(e.Right, context)
	case *ast.UnaryExpr:
		pc.checkExpression(e.Expr, context)
	case *ast.CallExpr:
		pc.checkExpression(e.Fun, context)
		for _, arg := range e.Args {
			pc.checkExpression(arg, context)
		}
	case *ast.IfExpr:
		pc.checkExpression(e.Condition, context)
		pc.checkExpression(e.Then, context)
		if e.Else != nil {
			pc.checkExpression(e.Else, context)
		}
	case *ast.LetExpr:
		pc.checkExpression(e.Value, context)
		pc.checkExpression(e.Body, context)
	case *ast.ParenExpr:
		pc.checkExpression(e.X, context)
	case *ast.SelectorExpr:
		pc.checkExpression(e.X, context)
	case *ast.IndexExpr:
		pc.checkExpression(e.X, context)
		pc.checkExpression(e.Index, context)
	case *ast.BasicLit:
		// Literals are always pure
		return
	default:
		// Unknown expression type - assume it's okay for now
		return
	}
}

// checkIdentifier checks if an identifier is a state variable
func (pc *PurityChecker) checkIdentifier(ident *ast.Ident, context string) {
	// Check if this identifier is a state variable
	if pc.stateVariables[ident.Name] {
		pc.errors = append(pc.errors, fmt.Sprintf("access to state variable '%s' in %s", ident.Name, context))
	}
	// Note: We don't check for primed identifiers (ident.Prime) because
	// those are only valid in actions, not in pure functions
}

// checkBlock checks a block statement for state variable access
func (pc *PurityChecker) checkBlock(block *ast.BlockStmt, context string) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		pc.checkStatement(stmt, context)
	}
}

// checkStatement checks a statement for state variable access
func (pc *PurityChecker) checkStatement(stmt ast.Stmt, context string) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		pc.checkExpression(s.Value, context)
	case *ast.ExprStmt:
		pc.checkExpression(s.Expr, context)
	case *ast.AssignStmt:
		// Assignment statements should not appear in pure functions
		pc.errors = append(pc.errors, fmt.Sprintf("assignment statement in %s (pure functions cannot modify state)", context))
		pc.checkExpression(s.Left, context)
		pc.checkExpression(s.Right, context)
	default:
		// Unknown statement type
		return
	}
}

// Errors returns the list of purity violations found
func (pc *PurityChecker) Errors() []string {
	return pc.errors
}

