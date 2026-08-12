package errors

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// ErrorContext captures context information for errors
type ErrorContext struct {
	Position    ast.Position
	Description string // Description from the declaration (if available)
	ElementName string // Name of the element (variable, action, invariant, etc.)
	ElementType string // Type of element (variable, action, invariant, temporal, etc.)
}

// NewErrorContext creates a new error context
func NewErrorContext(position ast.Position, description, elementName, elementType string) *ErrorContext {
	return &ErrorContext{
		Position:    position,
		Description: description,
		ElementName: elementName,
		ElementType: elementType,
	}
}

// ExtractContextFromDecl extracts error context from a declaration
func ExtractContextFromDecl(decl ast.Decl) *ErrorContext {
	switch d := decl.(type) {
	case *ast.VariableDecl:
		return NewErrorContext(d.Position, d.Description, d.Name, "variable")
	case *ast.ConstantDecl:
		return NewErrorContext(d.Position, d.Description, d.Name, "constant")
	case *ast.FunctionDecl:
		return NewErrorContext(d.Position, d.Description, d.Name, "function")
	case *ast.ActionDecl:
		return NewErrorContext(d.Position, d.Description, d.Name, "action")
	case *ast.InvariantDecl:
		return NewErrorContext(d.Position, d.Description, d.Name, "invariant")
	case *ast.TemporalDecl:
		return NewErrorContext(d.Position, d.Description, d.Name, "temporal property")
	case *ast.InitDecl:
		return NewErrorContext(d.Position, d.Description, "", "initial state")
	case *ast.OneOfInitDecl:
		return NewErrorContext(d.Position, d.Description, "", "oneOf initial state")
	case *ast.ModuleDecl:
		return NewErrorContext(d.Position, d.Description, d.Name, "module")
	default:
		return NewErrorContext(ast.Position{}, "", "", "unknown")
	}
}

// ExtractContextFromExpr extracts error context from an expression
func ExtractContextFromExpr(expr ast.Expr) *ErrorContext {
	return NewErrorContext(expr.Pos(), "", "", "expression")
}

// HasDescription returns true if the context has a description
func (ec *ErrorContext) HasDescription() bool {
	return ec.Description != ""
}

// FormatPosition formats the position as a string
func (ec *ErrorContext) FormatPosition() string {
	if ec.Position.Line > 0 {
		return fmt.Sprintf("%d:%d", ec.Position.Line, ec.Position.Column)
	}
	return "unknown position"
}

// FormatElement formats the element information as a string
func (ec *ErrorContext) FormatElement() string {
	if ec.ElementName != "" {
		return fmt.Sprintf("%s '%s'", ec.ElementType, ec.ElementName)
	}
	return ec.ElementType
}

