package eval

import (
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// RegisterConstants registers all constants from a parsed file into an environment
// This includes constants from top-level declarations and from within modules
// If additionalFiles is provided, constants from those files are also registered
func RegisterConstants(env *Environment, file *ast.File, additionalFiles ...*ast.File) {
	allFiles := append([]*ast.File{file}, additionalFiles...)
	
	var registerFromDecl func(ast.Decl, *Environment, []*ast.File)
	registerFromDecl = func(d ast.Decl, e *Environment, files []*ast.File) {
		switch decl := d.(type) {
		case *ast.ConstantDecl:
			// Evaluate the constant value
			// Create a temporary evaluator to evaluate the constant's initializer
			tempEnv := NewEnvironment()
			// Register enum types first (constants might reference enums)
			for _, f := range files {
				RegisterEnumTypes(tempEnv, f)
			}
			// Also copy constants we've already registered in this pass
			for name, val := range e.constants {
				tempEnv.SetConstant(name, val)
			}
			tempEvaluator := NewEvaluator(tempEnv)
			
			if decl.Value != nil {
				value, err := tempEvaluator.Eval(decl.Value)
				if err == nil {
					e.SetConstant(decl.Name, value)
				}
				// If evaluation fails, skip this constant (might depend on state variables)
			}
		case *ast.ModuleDecl:
			// Also register constants from inside modules
			for _, moduleDecl := range decl.Decls {
				registerFromDecl(moduleDecl, e, files)
			}
		}
	}
	
	// Do multiple passes to handle dependencies between constants
	for pass := 0; pass < 5; pass++ {
		prevCount := len(env.constants)
		for _, f := range allFiles {
			for _, decl := range f.Decls {
				registerFromDecl(decl, env, allFiles)
			}
		}
		// If no new constants were registered, we're done
		if len(env.constants) == prevCount {
			break
		}
	}
}
