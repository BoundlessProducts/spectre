package semantic

import (
	"fmt"

	"github.com/spectre-lang/spectre/pkg/ast"
)

// VisibilityChecker checks public/private access rules
type VisibilityChecker struct {
	symbolTable    *SymbolTable
	moduleResolver *ModuleResolver
	errors         []string
}

// NewVisibilityChecker creates a new visibility checker
func NewVisibilityChecker(symbolTable *SymbolTable, moduleResolver *ModuleResolver) *VisibilityChecker {
	return &VisibilityChecker{
		symbolTable:    symbolTable,
		moduleResolver: moduleResolver,
		errors:         []string{},
	}
}

// CheckVisibility checks visibility rules for all declarations
func (vc *VisibilityChecker) CheckVisibility(file *ast.File) []string {
	vc.errors = []string{}

	for _, decl := range file.Decls {
		vc.checkDeclVisibility(decl, "")
	}

	return vc.errors
}

// checkDeclVisibility checks visibility for a declaration
func (vc *VisibilityChecker) checkDeclVisibility(decl ast.Decl, currentModule string) {
	switch d := decl.(type) {
	case *ast.ModuleDecl:
		vc.checkModuleVisibility(d)
	case *ast.VariableDecl:
		vc.checkVariableVisibility(d, currentModule)
	case *ast.ConstantDecl:
		vc.checkConstantVisibility(d, currentModule)
	case *ast.FunctionDecl:
		vc.checkFunctionVisibility(d, currentModule)
	case *ast.ActionDecl:
		vc.checkActionVisibility(d, currentModule)
	case *ast.InvariantDecl:
		vc.checkInvariantVisibility(d, currentModule)
	case *ast.TemporalDecl:
		vc.checkTemporalVisibility(d, currentModule)
	default:
		return
	}
}

// checkModuleVisibility checks visibility within a module
func (vc *VisibilityChecker) checkModuleVisibility(decl *ast.ModuleDecl) {
	// Check all declarations within the module
	for _, innerDecl := range decl.Decls {
		vc.checkDeclVisibility(innerDecl, decl.Name)
	}
}

// checkVariableVisibility checks variable visibility
func (vc *VisibilityChecker) checkVariableVisibility(decl *ast.VariableDecl, currentModule string) {
	// Variables are private by default (can only be accessed within their module)
	// Public variables can be accessed from other modules
	// For now, we just track visibility - actual access checking will be done during name resolution
	if decl.Visibility == ast.Private && currentModule == "" {
		// Private variables at top-level are fine (they're effectively module-private)
		return
	}
}

// checkConstantVisibility checks constant visibility
func (vc *VisibilityChecker) checkConstantVisibility(decl *ast.ConstantDecl, currentModule string) {
	// Constants follow the same visibility rules as variables
	if decl.Visibility == ast.Private && currentModule == "" {
		return
	}
}

// checkFunctionVisibility checks function visibility
func (vc *VisibilityChecker) checkFunctionVisibility(decl *ast.FunctionDecl, currentModule string) {
	// Functions are public by default
	// Public functions can be called from other modules
	// Private functions can only be called within their module
	if decl.Visibility == ast.Private && currentModule == "" {
		return
	}
}

// checkActionVisibility checks action visibility
func (vc *VisibilityChecker) checkActionVisibility(decl *ast.ActionDecl, currentModule string) {
	// Actions are public by default
	// Public actions can be called from other modules
	// Private actions can only be called within their module
	if decl.Visibility == ast.Private && currentModule == "" {
		return
	}
}

// checkInvariantVisibility checks invariant visibility
func (vc *VisibilityChecker) checkInvariantVisibility(decl *ast.InvariantDecl, currentModule string) {
	// Invariants are public by default
	// Public invariants are checked globally
	// Private invariants are only checked within their module
	if decl.Visibility == ast.Private && currentModule == "" {
		return
	}
}

// checkTemporalVisibility checks temporal property visibility
func (vc *VisibilityChecker) checkTemporalVisibility(decl *ast.TemporalDecl, currentModule string) {
	// Temporal properties are public by default
	// Public temporal properties are checked globally
	// Private temporal properties are only checked within their module
	if decl.Visibility == ast.Private && currentModule == "" {
		return
	}
}

// CheckAccess checks if a symbol can be accessed from a given context
func (vc *VisibilityChecker) CheckAccess(symbol *Symbol, accessorModule string, symbolModule string) bool {
	// If accessing from the same module, always allowed
	if accessorModule == symbolModule {
		return true
	}

	// If accessing from top-level (no module), check if symbol is public
	if accessorModule == "" {
		// Top-level can access public symbols from any module
		// For now, we allow access - actual checking will be done during qualified name resolution
		return true
	}

	// If accessing from a different module, symbol must be public
	switch symbol.Kind {
	case SymbolVariable:
		if varDecl, ok := symbol.Decl.(*ast.VariableDecl); ok {
			return varDecl.Visibility == ast.Public
		}
	case SymbolConstant:
		if constDecl, ok := symbol.Decl.(*ast.ConstantDecl); ok {
			return constDecl.Visibility == ast.Public
		}
	case SymbolFunction:
		if funcDecl, ok := symbol.Decl.(*ast.FunctionDecl); ok {
			return funcDecl.Visibility == ast.Public
		}
	case SymbolAction:
		if actionDecl, ok := symbol.Decl.(*ast.ActionDecl); ok {
			return actionDecl.Visibility == ast.Public
		}
	// Note: Invariants and temporal properties don't have separate symbol kinds
	// They're checked through their declarations directly
	}

	return false
}

// addError adds an error message
func (vc *VisibilityChecker) addError(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		vc.errors = append(vc.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		vc.errors = append(vc.errors, msg)
	}
}

