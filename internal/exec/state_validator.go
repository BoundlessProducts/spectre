package exec

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/internal/eval"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// StateValidator validates states against invariants and postconditions
type StateValidator struct {
	constraintModel *state.ConstraintModel
	evaluator       *eval.Evaluator
	file            *ast.File // Keep reference to file for enum type registration
	additionalFiles []*ast.File // Additional files (imported modules) for constant/enum registration
}

// getAllFiles returns all files including the main file and additional files
func (sv *StateValidator) getAllFiles() []*ast.File {
	return append([]*ast.File{sv.file}, sv.additionalFiles...)
}

// NewStateValidator creates a new state validator
// additionalFiles can be provided to register constants/enums from imported modules
func NewStateValidator(constraintModel *state.ConstraintModel, file *ast.File, additionalFiles ...*ast.File) *StateValidator {
	env := eval.NewEnvironment()
	// Register enum types and constants from all files
	allFiles := append([]*ast.File{file}, additionalFiles...)
	for _, f := range allFiles {
		eval.RegisterEnumTypes(env, f)
		eval.RegisterConstants(env, f, allFiles...)
	}
	evaluator := eval.NewEvaluator(env)

	return &StateValidator{
		constraintModel: constraintModel,
		evaluator:       evaluator,
		file:            file,
		additionalFiles: additionalFiles,
	}
}

// ValidateState validates a state against all invariants
func (sv *StateValidator) ValidateState(s *state.State) ([]*ValidationError, error) {
	var errors []*ValidationError

	// Create environment with state variables
	env := eval.NewEnvironment()
	// Register enum types and constants from all files
	allFiles := sv.getAllFiles()
	for _, f := range allFiles {
		eval.RegisterEnumTypes(env, f)
		eval.RegisterConstants(env, f, allFiles...)
	}
	for varName, varValue := range s.Variables {
		env.SetVariable(varName, varValue)
	}

	evaluator := eval.NewEvaluator(env)

	// Check all invariants
	invariants := sv.constraintModel.GetInvariants()
	for _, inv := range invariants {
		value, err := evaluator.Eval(inv.Condition)
		if err != nil {
			return nil, fmt.Errorf("error evaluating invariant %s: %w", inv.Name, err)
		}

		// Invariant must evaluate to true
		if pv, ok := value.(*state.PrimitiveValue); ok && pv.TypeName == "bool" {
			if pv.BoolValue == nil || !*pv.BoolValue {
				msg := fmt.Sprintf("invariant %s violated", inv.Name)
				if inv.Description != "" {
					msg = fmt.Sprintf("invariant %s violated: (%s)", inv.Name, inv.Description)
				}
				errors = append(errors, &ValidationError{
					Type:        ErrorTypeInvariant,
					Name:        inv.Name,
					Message:     msg,
					Description: inv.Description,
					Condition:   inv.Condition,
					Position:    inv.Position,
				})
			}
		} else {
			return nil, fmt.Errorf("invariant %s must evaluate to boolean", inv.Name)
		}
	}

	return errors, nil
}

// ValidatePostconditions validates postconditions after an action execution
func (sv *StateValidator) ValidatePostconditions(actionName string, currentState, nextState *state.State) ([]*ValidationError, error) {
	var errors []*ValidationError

	// Get postconditions for this action (if any)
	// Note: Postconditions are typically associated with actions via `ensure` statements
	postconditions := sv.constraintModel.GetPostconditions(actionName)

	if len(postconditions) == 0 {
		return errors, nil
	}

	// Create environment with both current and next state variables
	env := eval.NewEnvironment()
	// Register enum types and constants from all files
	allFiles := sv.getAllFiles()
	for _, f := range allFiles {
		eval.RegisterEnumTypes(env, f)
		eval.RegisterConstants(env, f, allFiles...)
	}

	// Add current state variables (unprimed)
	for varName, varValue := range currentState.Variables {
		env.SetVariable(varName, varValue)
	}

	// Add next state variables (primed)
	for varName, varValue := range nextState.Variables {
		primedName := varName + "'"
		env.SetVariable(primedName, varValue)
		// Also add unprimed version for convenience
		env.SetVariable(varName, varValue)
	}

	evaluator := eval.NewEvaluator(env)

	// Check all postconditions
	for i, post := range postconditions {
		value, err := evaluator.Eval(post.Condition)
		if err != nil {
			return nil, fmt.Errorf("error evaluating postcondition: %w", err)
		}

		// Postcondition must evaluate to true
		if pv, ok := value.(*state.PrimitiveValue); ok && pv.TypeName == "bool" {
			if pv.BoolValue == nil || !*pv.BoolValue {
				errors = append(errors, &ValidationError{
					Type:      ErrorTypePostcondition,
					Name:      fmt.Sprintf("postcondition_%d", i),
					Message:   fmt.Sprintf("postcondition violated after action %s", actionName),
					Condition: post.Condition,
					Position:  post.Position,
				})
			}
		} else {
			return nil, fmt.Errorf("postcondition must evaluate to boolean")
		}
	}

	return errors, nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Type        ErrorType
	Name        string
	Message     string
	Description string // Human-readable description from the declaration
	Condition   ast.Expr
	Position    ast.Position
}

// ErrorType represents the type of validation error
type ErrorType int

const (
	ErrorTypeInvariant ErrorType = iota
	ErrorTypePostcondition
	ErrorTypePrecondition
)

func (e *ValidationError) Error() string {
	return e.Message
}

