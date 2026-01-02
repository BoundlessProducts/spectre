package exec

import (
	"fmt"

	"github.com/spectre-lang/spectre/internal/eval"
	"github.com/spectre-lang/spectre/internal/state"
	"github.com/spectre-lang/spectre/pkg/ast"
)

// StateValidator validates states against invariants and postconditions
type StateValidator struct {
	constraintModel *state.ConstraintModel
	evaluator       *eval.Evaluator
}

// NewStateValidator creates a new state validator
func NewStateValidator(constraintModel *state.ConstraintModel) *StateValidator {
	env := eval.NewEnvironment()
	evaluator := eval.NewEvaluator(env)

	return &StateValidator{
		constraintModel: constraintModel,
		evaluator:       evaluator,
	}
}

// ValidateState validates a state against all invariants
func (sv *StateValidator) ValidateState(s *state.State) ([]*ValidationError, error) {
	var errors []*ValidationError

	// Create environment with state variables
	env := eval.NewEnvironment()
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
				errors = append(errors, &ValidationError{
					Type:      ErrorTypeInvariant,
					Name:      inv.Name,
					Message:   fmt.Sprintf("invariant %s violated", inv.Name),
					Condition: inv.Condition,
					Position:  inv.Position,
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
	Type      ErrorType
	Name      string
	Message   string
	Condition ast.Expr
	Position  ast.Position
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

