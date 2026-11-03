package eval

import (
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// RegisterFunctions registers all function definitions from a file into the environment.
//
// Top-level functions are registered under their plain name (e.g. "add").
// Functions inside a module declaration are registered under both the
// module-qualified name "ModuleName.FuncName" (for explicit calls like
// Math.double(x)) and their plain name (for unqualified calls within the
// same module, e.g. double(x)).
//
// If additionalFiles is provided, functions from those files are also
// registered using the same rules.
func RegisterFunctions(env *Environment, file *ast.File, additionalFiles ...*ast.File) {
	allFiles := append([]*ast.File{file}, additionalFiles...)
	for _, f := range allFiles {
		for _, decl := range f.Decls {
			registerFunctionFromDecl(decl, env, "")
		}
	}
}

func registerFunctionFromDecl(decl ast.Decl, env *Environment, modulePrefix string) {
	switch d := decl.(type) {
	case *ast.FunctionDecl:
		def := &FunctionDef{
			Decl:       d,
			Params:     d.Parameters,
			Body:       d.Body,
			ReturnType: d.ReturnType,
		}
		// Register under plain name
		_ = env.DefineFunction(d.Name, def) // ignore duplicate errors
		// If inside a module, also register under qualified name
		if modulePrefix != "" {
			qualifiedName := modulePrefix + "." + d.Name
			_ = env.DefineFunction(qualifiedName, def)
		}
	case *ast.ModuleDecl:
		for _, inner := range d.Decls {
			registerFunctionFromDecl(inner, env, d.Name)
		}
	}
}
