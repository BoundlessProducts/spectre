package exec

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// StateInitializer generates initial states from init/oneOf declarations
type StateInitializer struct {
	variableModel     *state.VariableModel
	initialStateModel *state.InitialStateModel
	evaluator         *eval.Evaluator
	file              *ast.File // Keep reference to file for enum type registration
}

// NewStateInitializer creates a new state initializer
func NewStateInitializer(variableModel *state.VariableModel, initialStateModel *state.InitialStateModel, file *ast.File) *StateInitializer {
	// Create an environment for evaluating initial state expressions
	env := eval.NewEnvironment()
	// Register enum types
	eval.RegisterEnumTypes(env, file)
	evaluator := eval.NewEvaluator(env)

	return &StateInitializer{
		variableModel:     variableModel,
		initialStateModel: initialStateModel,
		evaluator:         evaluator,
		file:              file,
	}
}

// GenerateInitialStates generates all possible initial states
func (si *StateInitializer) GenerateInitialStates() ([]*state.State, error) {
	if si.initialStateModel.IsOneOf() {
		return si.generateOneOfInitialStates()
	}
	return si.generateDeterministicInitialState()
}

// generateDeterministicInitialState generates a single initial state from init declaration
func (si *StateInitializer) generateDeterministicInitialState() ([]*state.State, error) {
	initState := si.initialStateModel.GetDeterministicInit()
	if initState == nil {
		return nil, fmt.Errorf("no initial state declaration found")
	}

	// Create a new state
	newState := state.NewState()

	// Create an environment that can access state variables
	env := eval.NewEnvironment()
	// Register enum types
	eval.RegisterEnumTypes(env, si.file)

	// Evaluate each assignment in the init block
	for _, stmt := range initState.Body.Statements {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			// Handle regular assignment: variable = expression
			// Update the evaluator's environment with current state values
			for varName, varValue := range newState.Variables {
				env.SetVariable(varName, varValue)
			}

			// Create evaluator with updated environment
			stateEvaluator := eval.NewEvaluator(env)

			value, err := stateEvaluator.Eval(s.Right)
			if err != nil {
				return nil, fmt.Errorf("error evaluating initial value for %v: %w", s.Left, err)
			}

			// Get variable name
			varName, err := si.getVariableName(s.Left)
			if err != nil {
				return nil, err
			}

			// Set variable in state
			newState.SetVariable(varName, value)

		default:
			return nil, fmt.Errorf("unsupported statement type in init block: %T", stmt)
		}
	}

	// Ensure all variables are initialized
	if err := si.validateState(newState); err != nil {
		return nil, err
	}

	return []*state.State{newState}, nil
}

// generateOneOfInitialStates generates all possible initial states from oneOf declaration
func (si *StateInitializer) generateOneOfInitialStates() ([]*state.State, error) {
	options := si.initialStateModel.GetOneOfOptions()
	if len(options) == 0 {
		return nil, fmt.Errorf("oneOf declaration has no options")
	}

	var states []*state.State

		for i, option := range options {
		// Create a new state for this option
		newState := state.NewState()

		// Create an environment that can access state variables
		env := eval.NewEnvironment()
		// Register enum types
		eval.RegisterEnumTypes(env, si.file)

		// Evaluate each assignment in this option
		for _, stmt := range option.Statements {
			switch s := stmt.(type) {
			case *ast.AssignStmt:
				// Update evaluator environment with current state values
				for varName, varValue := range newState.Variables {
					env.SetVariable(varName, varValue)
				}

				// Create evaluator with updated environment
				stateEvaluator := eval.NewEvaluator(env)

				value, err := stateEvaluator.Eval(s.Right)
				if err != nil {
					return nil, fmt.Errorf("error evaluating initial value in oneOf option %d: %w", i, err)
				}

				varName, err := si.getVariableName(s.Left)
				if err != nil {
					return nil, err
				}

				newState.SetVariable(varName, value)

			default:
				return nil, fmt.Errorf("unsupported statement type in oneOf option: %T", stmt)
			}
		}

		// Validate state
		if err := si.validateState(newState); err != nil {
			return nil, fmt.Errorf("invalid state in oneOf option %d: %w", i, err)
		}

		states = append(states, newState)
	}

	return states, nil
}

// getVariableName extracts the variable name from an expression
func (si *StateInitializer) getVariableName(expr ast.Expr) (string, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name, nil
	case *ast.SelectorExpr:
		// For selector expressions like record.field, we need the base variable
		// For now, we'll handle simple cases
		if ident, ok := e.X.(*ast.Ident); ok {
			return ident.Name, nil
		}
		return "", fmt.Errorf("complex selector expressions not yet supported")
	default:
		return "", fmt.Errorf("unsupported left-hand side in assignment: %T", expr)
	}
}

// validateState ensures all state variables are initialized
func (si *StateInitializer) validateState(s *state.State) error {
	// Check that all variables declared in the model are present in the state
	for _, varName := range si.variableModel.GetVariableNames() {
		_, exists := s.GetVariable(varName)
		if !exists {
			return fmt.Errorf("state variable '%s' is not initialized", varName)
		}
	}
	return nil
}

