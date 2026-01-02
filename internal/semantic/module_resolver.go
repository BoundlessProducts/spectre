package semantic

import (
	"fmt"

	"github.com/spectre-lang/spectre/pkg/ast"
)

// ModuleResolver resolves module imports and extensions
type ModuleResolver struct {
	symbolTable *SymbolTable
	modules     map[string]*ast.ModuleDecl // Module name -> Module declaration
	imports     map[string]*ast.ImportDecl // Import name -> Import declaration
	errors      []string
}

// NewModuleResolver creates a new module resolver
func NewModuleResolver(symbolTable *SymbolTable) *ModuleResolver {
	return &ModuleResolver{
		symbolTable: symbolTable,
		modules:     make(map[string]*ast.ModuleDecl),
		imports:     make(map[string]*ast.ImportDecl),
		errors:      []string{},
	}
}

// ResolveModules resolves all modules, imports, and extensions in a file
func (mr *ModuleResolver) ResolveModules(file *ast.File) []string {
	mr.errors = []string{}
	mr.modules = make(map[string]*ast.ModuleDecl)
	mr.imports = make(map[string]*ast.ImportDecl)

	// First pass: collect all modules and imports
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.ModuleDecl:
			mr.collectModule(d)
		case *ast.ImportDecl:
			mr.collectImport(d)
		}
	}

	// Second pass: resolve module extensions and imports
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.ModuleDecl:
			mr.resolveModuleExtension(d)
		case *ast.ImportDecl:
			mr.resolveImport(d)
		}
	}

	return mr.errors
}

// collectModule collects a module declaration
func (mr *ModuleResolver) collectModule(decl *ast.ModuleDecl) {
	if _, exists := mr.modules[decl.Name]; exists {
		mr.addError(decl.Pos(), "duplicate module declaration: %s", decl.Name)
		return
	}
	mr.modules[decl.Name] = decl
}

// collectImport collects an import declaration
func (mr *ModuleResolver) collectImport(decl *ast.ImportDecl) {
	// Note: ImportDecl.Module contains the module name to import
	if _, exists := mr.imports[decl.Module]; exists {
		mr.addError(decl.Pos(), "duplicate import: %s", decl.Module)
		return
	}
	mr.imports[decl.Module] = decl
}

// resolveModuleExtension resolves a module's extension clause
func (mr *ModuleResolver) resolveModuleExtension(decl *ast.ModuleDecl) {
	if decl.Extends == "" {
		return // No extension
	}

	// Check if the extended module exists
	if _, exists := mr.modules[decl.Extends]; !exists {
		mr.addError(decl.Pos(), "module %s extends undefined module: %s", decl.Name, decl.Extends)
		return
	}

	// Check for circular dependencies
	// A circular dependency exists if the extended module (or any of its ancestors) extends back to this module
	if mr.hasCircularDependency(decl.Extends, decl.Name, make(map[string]bool)) {
		mr.addError(decl.Pos(), "circular module dependency detected: %s extends %s", decl.Name, decl.Extends)
		return
	}
}

// resolveImport resolves an import declaration
func (mr *ModuleResolver) resolveImport(decl *ast.ImportDecl) {
	// Check if the imported module exists
	if _, exists := mr.modules[decl.Module]; !exists {
		mr.addError(decl.Pos(), "imported module not found: %s", decl.Module)
		return
	}
}

// hasCircularDependency checks for circular module dependencies
// Returns true if following the extension chain from 'current' leads back to 'target'
func (mr *ModuleResolver) hasCircularDependency(current, target string, visited map[string]bool) bool {
	if current == target {
		return true // Found a cycle
	}
	if visited[current] {
		return false // Already checked this path, no cycle found
	}

	visited[current] = true

	// Get the current module
	module, exists := mr.modules[current]
	if !exists {
		return false // Module doesn't exist, no cycle
	}

	// If this module extends something, check if that leads to target
	if module.Extends != "" {
		return mr.hasCircularDependency(module.Extends, target, visited)
	}

	return false // No extension, no cycle
}

// GetModule returns a module by name
func (mr *ModuleResolver) GetModule(name string) (*ast.ModuleDecl, bool) {
	module, exists := mr.modules[name]
	return module, exists
}

// GetModules returns all modules
func (mr *ModuleResolver) GetModules() map[string]*ast.ModuleDecl {
	return mr.modules
}

// GetImports returns all imports
func (mr *ModuleResolver) GetImports() map[string]*ast.ImportDecl {
	return mr.imports
}

// addError adds an error message
func (mr *ModuleResolver) addError(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		mr.errors = append(mr.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		mr.errors = append(mr.errors, msg)
	}
}

