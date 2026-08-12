package state

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// VariableModel represents the model of all state variables in a specification
// It tracks variable names, types, and their declarations
type VariableModel struct {
	Variables map[string]*VariableInfo // Variable name -> Variable info
}

// VariableInfo contains information about a state variable
type VariableInfo struct {
	Name        string
	Type        ast.Type
	Declaration *ast.VariableDecl
	Position    ast.Position
	Description string
}

// NewVariableModel creates a new variable model from an AST file
func NewVariableModel(file *ast.File) *VariableModel {
	model := &VariableModel{
		Variables: make(map[string]*VariableInfo),
	}

	// Extract all state variables from the file
	model.extractVariables(file)

	return model
}

// extractVariables extracts all state variables from the AST
func (vm *VariableModel) extractVariables(file *ast.File) {
	for _, decl := range file.Decls {
		vm.extractVariablesFromDecl(decl)
	}
}

// extractVariablesFromDecl extracts variables from a declaration
func (vm *VariableModel) extractVariablesFromDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.VariableDecl:
		// Top-level variable declaration
		vm.addVariable(d)
	case *ast.ModuleDecl:
		// Extract variables from module declarations
		for _, innerDecl := range d.Decls {
			if varDecl, ok := innerDecl.(*ast.VariableDecl); ok {
				vm.addVariable(varDecl)
			}
		}
	default:
		// Other declarations don't contain state variables
		return
	}
}

// addVariable adds a variable to the model
func (vm *VariableModel) addVariable(decl *ast.VariableDecl) {
	if _, exists := vm.Variables[decl.Name]; exists {
		// Variable already exists - this should have been caught by semantic analysis
		// But we'll still add it to avoid crashes
		return
	}

	vm.Variables[decl.Name] = &VariableInfo{
		Name:        decl.Name,
		Type:        decl.Type,
		Declaration: decl,
		Position:    decl.Position,
		Description: decl.Description,
	}
}

// GetVariable returns information about a variable
func (vm *VariableModel) GetVariable(name string) (*VariableInfo, bool) {
	info, exists := vm.Variables[name]
	return info, exists
}

// GetAllVariables returns all variables in the model
func (vm *VariableModel) GetAllVariables() map[string]*VariableInfo {
	return vm.Variables
}

// ValidateState validates that a state contains all required variables
// and that their types match the model
func (vm *VariableModel) ValidateState(s *State) []string {
	var errors []string

	// Check that all variables in the model are present in the state
	for name, info := range vm.Variables {
		value, exists := s.GetVariable(name)
		if !exists {
			errors = append(errors, fmt.Sprintf("state missing required variable: %s", name))
			continue
		}

		// Validate type compatibility
		if !vm.validateType(value, info.Type) {
			errors = append(errors, fmt.Sprintf("variable %s has incorrect type: expected %s, got %s",
				name, vm.typeString(info.Type), value.Type()))
		}
	}

	// Check that state doesn't have extra variables
	for name := range s.Variables {
		if _, exists := vm.Variables[name]; !exists {
			errors = append(errors, fmt.Sprintf("state contains unknown variable: %s", name))
		}
	}

	return errors
}

// validateType validates that a value matches the expected type
func (vm *VariableModel) validateType(value Value, expectedType ast.Type) bool {
	valueType := value.Type()

	switch t := expectedType.(type) {
	case *ast.PrimitiveType:
		return valueType == t.Name
	case *ast.SetType:
		// Set types not yet fully implemented in value system
		return valueType == "set"
	case *ast.MapType:
		// Map types not yet fully implemented in value system
		return valueType == "map"
	case *ast.ListType:
		// List types not yet fully implemented in value system
		return valueType == "list"
	case *ast.OptionType:
		// Option types not yet fully implemented in value system
		return valueType == "option"
	case *ast.RecordType:
		// Record types not yet fully implemented in value system
		return valueType == "record"
	case *ast.EnumType:
		// Enum types not yet fully implemented in value system
		return valueType == "enum"
	case *ast.NamedType:
		// Named types - check if it's a known type name
		return valueType == t.Name
	default:
		return false
	}
}

// typeString returns a string representation of a type
func (vm *VariableModel) typeString(t ast.Type) string {
	switch t := t.(type) {
	case *ast.PrimitiveType:
		return t.Name
	case *ast.SetType:
		return fmt.Sprintf("Set<%s>", vm.typeString(t.Element))
	case *ast.MapType:
		return fmt.Sprintf("Map<%s, %s>", vm.typeString(t.Key), vm.typeString(t.Value))
	case *ast.ListType:
		return fmt.Sprintf("List<%s>", vm.typeString(t.Element))
	case *ast.OptionType:
		return fmt.Sprintf("Option<%s>", vm.typeString(t.Element))
	case *ast.RecordType:
		return "record"
	case *ast.EnumType:
		return fmt.Sprintf("enum(%s)", t.Name)
	case *ast.NamedType:
		return t.Name
	default:
		return "unknown"
	}
}

// GetVariableNames returns a list of all variable names
func (vm *VariableModel) GetVariableNames() []string {
	names := make([]string, 0, len(vm.Variables))
	for name := range vm.Variables {
		names = append(names, name)
	}
	return names
}

// HasVariable checks if a variable exists in the model
func (vm *VariableModel) HasVariable(name string) bool {
	_, exists := vm.Variables[name]
	return exists
}

