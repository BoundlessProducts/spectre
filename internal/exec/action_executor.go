package exec

import (
	"fmt"

	"github.com/spectre-lang/spectre/internal/eval"
	"github.com/spectre-lang/spectre/internal/state"
	"github.com/spectre-lang/spectre/pkg/ast"
)

// ActionExecutor executes actions to generate next states
type ActionExecutor struct {
	variableModel *state.VariableModel
	actionModel   *state.ActionModel
	evaluator     *eval.Evaluator
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(variableModel *state.VariableModel, actionModel *state.ActionModel) *ActionExecutor {
	env := eval.NewEnvironment()
	evaluator := eval.NewEvaluator(env)

	return &ActionExecutor{
		variableModel: variableModel,
		actionModel:   actionModel,
		evaluator:     evaluator,
	}
}

// ExecuteAction executes an action on a state and returns the next state
func (ae *ActionExecutor) ExecuteAction(actionName string, currentState *state.State, args []state.Value) (*state.State, error) {
	// Get action information
	actionInfo, exists := ae.actionModel.GetAction(actionName)
	if !exists {
		return nil, fmt.Errorf("action %s not found", actionName)
	}

	// Create a copy of the current state for the next state
	nextState := currentState.Copy()

	// Create an environment with current state values and action parameters
	env := eval.NewEnvironment()

	// Add current state variables to environment
	for varName, varValue := range currentState.Variables {
		env.SetVariable(varName, varValue)
	}

	// Bind action parameters
	if len(args) != len(actionInfo.Parameters) {
		return nil, fmt.Errorf("action %s expects %d arguments, got %d", actionName, len(actionInfo.Parameters), len(args))
	}

	for i, param := range actionInfo.Parameters {
		env.SetVariable(param.Name, args[i])
	}

	// Create evaluator with the environment
	stateEvaluator := eval.NewEvaluator(env)

	// Execute action body
	if actionInfo.Body != nil {
		for _, stmt := range actionInfo.Body.Statements {
			switch s := stmt.(type) {
			case *ast.AssignStmt:
				// Handle assignment: variable' = expression or variable = expression
				value, err := ae.evaluateAssignment(s, currentState, nextState, stateEvaluator, env)
				if err != nil {
					return nil, fmt.Errorf("error executing assignment in action %s: %w", actionName, err)
				}

				// Get variable name (handle primed identifiers)
				varName, err := ae.getVariableName(s.Left)
				if err != nil {
					return nil, err
				}

				// Set variable in next state
				nextState.SetVariable(varName, value)

			case *ast.ExprStmt:
				// Expression statements are evaluated but don't modify state
				_, err := stateEvaluator.Eval(s.Expr)
				if err != nil {
					return nil, fmt.Errorf("error evaluating expression in action %s: %w", actionName, err)
				}

			default:
				return nil, fmt.Errorf("unsupported statement type in action body: %T", stmt)
			}
		}
	}

	return nextState, nil
}

// evaluateAssignment evaluates the right-hand side of an assignment
func (ae *ActionExecutor) evaluateAssignment(stmt *ast.AssignStmt, currentState, nextState *state.State, evaluator *eval.Evaluator, env *eval.Environment) (state.Value, error) {
	// Update environment with current state values (for unprimed variables)
	for varName, varValue := range currentState.Variables {
		env.SetVariable(varName, varValue)
	}

	// Update environment with next state values (for primed variables)
	// This allows expressions like counter' = counter + 1
	for varName, varValue := range nextState.Variables {
		// Create primed variable name
		primedName := varName + "'"
		env.SetVariable(primedName, varValue)
	}

	// Evaluate the right-hand side expression
	return evaluator.Eval(stmt.Right)
}

// getVariableName extracts the variable name from an expression, handling primed identifiers
func (ae *ActionExecutor) getVariableName(expr ast.Expr) (string, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name, nil
	case *ast.SelectorExpr:
		// For selector expressions like record.field, we need the base variable
		if ident, ok := e.X.(*ast.Ident); ok {
			return ident.Name, nil
		}
		return "", fmt.Errorf("complex selector expressions not yet supported")
	default:
		return "", fmt.Errorf("unsupported left-hand side in assignment: %T", expr)
	}
}

// CanExecute checks if an action can be executed in the current state
// This checks guards and preconditions
func (ae *ActionExecutor) CanExecute(actionName string, currentState *state.State, args []state.Value) (bool, error) {
	actionInfo, exists := ae.actionModel.GetAction(actionName)
	if !exists {
		return false, fmt.Errorf("action %s not found", actionName)
	}

	// Create environment with current state values
	env := eval.NewEnvironment()
	for varName, varValue := range currentState.Variables {
		env.SetVariable(varName, varValue)
	}

	// Bind action parameters
	for i, param := range actionInfo.Parameters {
		if i < len(args) {
			env.SetVariable(param.Name, args[i])
		}
	}

	evaluator := eval.NewEvaluator(env)

	// Check guard condition (if present)
	if actionInfo.Guard != nil {
		guardValue, err := evaluator.Eval(actionInfo.Guard)
		if err != nil {
			return false, fmt.Errorf("error evaluating guard: %w", err)
		}

		// Guard must evaluate to true
		if pv, ok := guardValue.(*state.PrimitiveValue); ok && pv.TypeName == "bool" {
			if pv.BoolValue == nil || !*pv.BoolValue {
				return false, nil
			}
		} else {
			return false, fmt.Errorf("guard must evaluate to boolean")
		}
	}

	return true, nil
}

