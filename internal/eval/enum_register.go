package eval

import (
	"github.com/akkeshavan/spectre/pkg/ast"
)

// RegisterEnumTypes registers all enum types from a parsed file into an environment
func RegisterEnumTypes(env *Environment, file *ast.File) {
	for _, decl := range file.Decls {
		if enumDecl, ok := decl.(*ast.EnumDecl); ok {
			enumDef := &EnumTypeDef{
				Name:   enumDecl.Name,
				Values: enumDecl.Values,
			}
			env.SetEnumType(enumDecl.Name, enumDef)
		}
	}
}

