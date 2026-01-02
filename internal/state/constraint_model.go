package state

import (
	"github.com/spectre-lang/spectre/pkg/ast"
)

// ConstraintModel represents the model of all constraints in a specification
// Constraints include invariants, preconditions (require), and postconditions (ensure)
type ConstraintModel struct {
	Invariants    []*InvariantInfo
	Preconditions map[string][]*PreconditionInfo  // Action name -> preconditions
	Postconditions map[string][]*PostconditionInfo // Action name -> postconditions
}

// InvariantInfo contains information about an invariant
type InvariantInfo struct {
	Name        string
	Condition   ast.Expr
	Declaration *ast.InvariantDecl
	Position    ast.Position
	Description string
	Visibility  ast.Visibility
}

// PreconditionInfo contains information about a precondition (require statement)
type PreconditionInfo struct {
	Condition ast.Expr
	Position  ast.Position
	Action    string // Name of the action this precondition belongs to
}

// PostconditionInfo contains information about a postcondition (ensure statement)
type PostconditionInfo struct {
	Condition ast.Expr
	Position  ast.Position
	Action    string // Name of the action this postcondition belongs to
}

// NewConstraintModel creates a new constraint model from an AST file
func NewConstraintModel(file *ast.File, actionModel *ActionModel) *ConstraintModel {
	model := &ConstraintModel{
		Invariants:     []*InvariantInfo{},
		Preconditions:  make(map[string][]*PreconditionInfo),
		Postconditions: make(map[string][]*PostconditionInfo),
	}

	// Extract all constraints from the file
	model.extractConstraints(file, actionModel)

	return model
}

// extractConstraints extracts all constraints from the AST
func (cm *ConstraintModel) extractConstraints(file *ast.File, actionModel *ActionModel) {
	for _, decl := range file.Decls {
		cm.extractConstraintsFromDecl(decl, actionModel)
	}
}

// extractConstraintsFromDecl extracts constraints from a declaration
func (cm *ConstraintModel) extractConstraintsFromDecl(decl ast.Decl, actionModel *ActionModel) {
	switch d := decl.(type) {
	case *ast.InvariantDecl:
		// Top-level invariant declaration
		cm.addInvariant(d)
	case *ast.ActionDecl:
		// Extract preconditions and postconditions from action body
		cm.extractActionConstraints(d, actionModel)
	case *ast.ModuleDecl:
		// Extract constraints from module declarations
		for _, innerDecl := range d.Decls {
			switch innerD := innerDecl.(type) {
			case *ast.InvariantDecl:
				cm.addInvariant(innerD)
			case *ast.ActionDecl:
				cm.extractActionConstraints(innerD, actionModel)
			}
		}
	default:
		// Other declarations don't contain constraints
		return
	}
}

// addInvariant adds an invariant to the model
func (cm *ConstraintModel) addInvariant(decl *ast.InvariantDecl) {
	cm.Invariants = append(cm.Invariants, &InvariantInfo{
		Name:        decl.Name,
		Condition:   decl.Condition,
		Declaration: decl,
		Position:    decl.Position,
		Description: decl.Description,
		Visibility:  decl.Visibility,
	})
}

// extractActionConstraints extracts preconditions and postconditions from an action body
func (cm *ConstraintModel) extractActionConstraints(actionDecl *ast.ActionDecl, actionModel *ActionModel) {
	if actionDecl.Body == nil {
		return
	}

	actionName := actionDecl.Name

	// Extract require and ensure statements from the action body
	for _, stmt := range actionDecl.Body.Statements {
		switch s := stmt.(type) {
		case *ast.RequireStmt:
			// Precondition (require statement)
			cm.addPrecondition(actionName, s)
		case *ast.EnsureStmt:
			// Postcondition (ensure statement)
			cm.addPostcondition(actionName, s)
		}
	}
}

// addPrecondition adds a precondition to the model
func (cm *ConstraintModel) addPrecondition(actionName string, stmt *ast.RequireStmt) {
	if cm.Preconditions[actionName] == nil {
		cm.Preconditions[actionName] = []*PreconditionInfo{}
	}

	cm.Preconditions[actionName] = append(cm.Preconditions[actionName], &PreconditionInfo{
		Condition: stmt.Condition,
		Position:  stmt.Position,
		Action:    actionName,
	})
}

// addPostcondition adds a postcondition to the model
func (cm *ConstraintModel) addPostcondition(actionName string, stmt *ast.EnsureStmt) {
	if cm.Postconditions[actionName] == nil {
		cm.Postconditions[actionName] = []*PostconditionInfo{}
	}

	cm.Postconditions[actionName] = append(cm.Postconditions[actionName], &PostconditionInfo{
		Condition: stmt.Condition,
		Position:  stmt.Position,
		Action:    actionName,
	})
}

// GetInvariants returns all invariants
func (cm *ConstraintModel) GetInvariants() []*InvariantInfo {
	return cm.Invariants
}

// GetInvariant returns an invariant by name
func (cm *ConstraintModel) GetInvariant(name string) *InvariantInfo {
	for _, inv := range cm.Invariants {
		if inv.Name == name {
			return inv
		}
	}
	return nil
}

// GetPreconditions returns all preconditions for an action
func (cm *ConstraintModel) GetPreconditions(actionName string) []*PreconditionInfo {
	return cm.Preconditions[actionName]
}

// GetPostconditions returns all postconditions for an action
func (cm *ConstraintModel) GetPostconditions(actionName string) []*PostconditionInfo {
	return cm.Postconditions[actionName]
}

// HasInvariants returns true if there are any invariants
func (cm *ConstraintModel) HasInvariants() bool {
	return len(cm.Invariants) > 0
}

// HasPreconditions returns true if an action has preconditions
func (cm *ConstraintModel) HasPreconditions(actionName string) bool {
	preconds := cm.Preconditions[actionName]
	return preconds != nil && len(preconds) > 0
}

// HasPostconditions returns true if an action has postconditions
func (cm *ConstraintModel) HasPostconditions(actionName string) bool {
	postconds := cm.Postconditions[actionName]
	return postconds != nil && len(postconds) > 0
}

// GetInvariantCount returns the number of invariants
func (cm *ConstraintModel) GetInvariantCount() int {
	return len(cm.Invariants)
}

