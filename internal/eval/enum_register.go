package eval

import (
	"github.com/akkeshavan/spectre/pkg/ast"
)

// RegisterEnumTypes registers all enum types from a parsed file into an environment
// This includes enums from top-level declarations and from within modules
func RegisterEnumTypes(env *Environment, file *ast.File) {
	var registerFromDecl func(ast.Decl)
	registerFromDecl = func(d ast.Decl) {
		switch decl := d.(type) {
		case *ast.EnumDecl:
			enumDef := &EnumTypeDef{
				Name:   decl.Name,
				Values: decl.Values,
			}
			env.SetEnumType(decl.Name, enumDef)
		case *ast.ModuleDecl:
			// Also register enums from inside modules
			for _, moduleDecl := range decl.Decls {
				registerFromDecl(moduleDecl)
			}
		}
	}
	
	for _, decl := range file.Decls {
		registerFromDecl(decl)
	}
}

