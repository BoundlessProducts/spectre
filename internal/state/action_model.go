package state

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// ActionModel represents the model of all actions in a specification
// Actions define state transitions with guards and updates
type ActionModel struct {
	Actions map[string]*ActionInfo // Action name -> Action info
}

// ActionInfo contains information about an action
type ActionInfo struct {
	Name        string
	Parameters  []ast.Parameter
	Guard       ast.Expr // Optional guard condition (when clause)
	Body        *ast.BlockStmt
	Declaration *ast.ActionDecl
	Position    ast.Position
	Description string
	Visibility  ast.Visibility
}

// NewActionModel creates a new action model from an AST file
func NewActionModel(file *ast.File) *ActionModel {
	model := &ActionModel{
		Actions: make(map[string]*ActionInfo),
	}

	// Extract all actions from the file
	model.extractActions(file)

	return model
}

// extractActions extracts all actions from the AST
func (am *ActionModel) extractActions(file *ast.File) {
	for _, decl := range file.Decls {
		am.extractActionsFromDecl(decl)
	}
}

// extractActionsFromDecl extracts actions from a declaration
func (am *ActionModel) extractActionsFromDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.ActionDecl:
		// Top-level action declaration
		am.addAction(d)
	case *ast.ModuleDecl:
		// Extract actions from module declarations
		for _, innerDecl := range d.Decls {
			if actionDecl, ok := innerDecl.(*ast.ActionDecl); ok {
				am.addAction(actionDecl)
			}
		}
	default:
		// Other declarations don't contain actions
		return
	}
}

// addAction adds an action to the model
func (am *ActionModel) addAction(decl *ast.ActionDecl) {
	if _, exists := am.Actions[decl.Name]; exists {
		// Action already exists - this should have been caught by semantic analysis
		// But we'll still add it to avoid crashes
		return
	}

	am.Actions[decl.Name] = &ActionInfo{
		Name:        decl.Name,
		Parameters:  decl.Parameters,
		Guard:       decl.Guard,
		Body:        decl.Body,
		Declaration: decl,
		Position:    decl.Position,
		Description: decl.Description,
		Visibility:  decl.Visibility,
	}
}

// GetAction returns information about an action
func (am *ActionModel) GetAction(name string) (*ActionInfo, bool) {
	info, exists := am.Actions[name]
	return info, exists
}

// GetAllActions returns all actions in the model
func (am *ActionModel) GetAllActions() map[string]*ActionInfo {
	return am.Actions
}

// GetActionNames returns a list of all action names
func (am *ActionModel) GetActionNames() []string {
	names := make([]string, 0, len(am.Actions))
	for name := range am.Actions {
		names = append(names, name)
	}
	return names
}

// HasAction checks if an action exists in the model
func (am *ActionModel) HasAction(name string) bool {
	_, exists := am.Actions[name]
	return exists
}

// HasGuard returns true if the action has a guard condition
func (ai *ActionInfo) HasGuard() bool {
	return ai.Guard != nil
}

// HasBody returns true if the action has a body
func (ai *ActionInfo) HasBody() bool {
	return ai.Body != nil
}

// GetParameterCount returns the number of parameters
func (ai *ActionInfo) GetParameterCount() int {
	return len(ai.Parameters)
}

// GetParameterNames returns a list of parameter names
func (ai *ActionInfo) GetParameterNames() []string {
	names := make([]string, len(ai.Parameters))
	for i, param := range ai.Parameters {
		names[i] = param.Name
	}
	return names
}

// String returns a string representation of the action
func (ai *ActionInfo) String() string {
	result := fmt.Sprintf("action %s", ai.Name)
	if len(ai.Parameters) > 0 {
		result += "("
		for i, param := range ai.Parameters {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%s: %s", param.Name, param.Type)
		}
		result += ")"
	}
	if ai.Guard != nil {
		result += " when <guard>"
	}
	if ai.Body != nil {
		result += " { ... }"
	}
	return result
}

