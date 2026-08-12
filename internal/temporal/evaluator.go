package temporal

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/internal/eval"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// TemporalEvaluator evaluates temporal properties over execution traces
type TemporalEvaluator struct {
	evaluator *eval.Evaluator
}

// NewTemporalEvaluator creates a new temporal evaluator
func NewTemporalEvaluator() *TemporalEvaluator {
	env := eval.NewEnvironment()
	evaluator := eval.NewEvaluator(env)
	return &TemporalEvaluator{
		evaluator: evaluator,
	}
}

// EvaluateTemporalProperty evaluates a temporal property over a trace
func (te *TemporalEvaluator) EvaluateTemporalProperty(expr ast.Expr, trace *Trace) (bool, error) {
	switch e := expr.(type) {
	case *ast.AlwaysExpr:
		return te.evaluateAlways(e, trace)
	case *ast.EventuallyExpr:
		return te.evaluateEventually(e, trace)
	case *ast.UntilExpr:
		return te.evaluateUntil(e, trace)
	case *ast.LeadsToExpr:
		return te.evaluateLeadsTo(e, trace)
	case *ast.BinaryExpr:
		// Handle logical operators (AND, OR) in temporal expressions
		return te.evaluateBinaryTemporal(e, trace)
	default:
		// For non-temporal expressions, evaluate at current position
		return te.evaluateExpression(e, trace.CurrentState())
	}
}

// evaluateAlways evaluates "always P" - P must hold in all states
func (te *TemporalEvaluator) evaluateAlways(expr *ast.AlwaysExpr, trace *Trace) (bool, error) {
	originalPos := trace.Position
	trace.Reset()
	
	for !trace.IsComplete() {
		state := trace.CurrentState()
		holds, err := te.evaluateExpression(expr.Expr, state)
		if err != nil {
			trace.Position = originalPos
			return false, err
		}
		if !holds {
			trace.Position = originalPos
			return false, nil
		}
		trace.NextState()
	}
	
	// Check last state
	state := trace.CurrentState()
	holds, err := te.evaluateExpression(expr.Expr, state)
	trace.Position = originalPos
	return holds, err
}

// evaluateEventually evaluates "eventually P" - P must hold in at least one future state
func (te *TemporalEvaluator) evaluateEventually(expr *ast.EventuallyExpr, trace *Trace) (bool, error) {
	originalPos := trace.Position
	
	// Check from current position onwards
	for !trace.IsComplete() {
		state := trace.CurrentState()
		holds, err := te.evaluateExpression(expr.Expr, state)
		if err != nil {
			trace.Position = originalPos
			return false, err
		}
		if holds {
			trace.Position = originalPos
			return true, nil
		}
		trace.NextState()
	}
	
	// Check last state
	state := trace.CurrentState()
	holds, err := te.evaluateExpression(expr.Expr, state)
	trace.Position = originalPos
	return holds, err
}

// evaluateUntil evaluates "P until Q" - P must hold until Q becomes true
func (te *TemporalEvaluator) evaluateUntil(expr *ast.UntilExpr, trace *Trace) (bool, error) {
	originalPos := trace.Position
	
	for !trace.IsComplete() {
		state := trace.CurrentState()
		
		// Check if Q holds
		qHolds, err := te.evaluateExpression(expr.Right, state)
		if err != nil {
			trace.Position = originalPos
			return false, err
		}
		if qHolds {
			// Q holds, so P until Q is satisfied
			trace.Position = originalPos
			return true, nil
		}
		
		// Q doesn't hold, so P must hold
		pHolds, err := te.evaluateExpression(expr.Left, state)
		if err != nil {
			trace.Position = originalPos
			return false, err
		}
		if !pHolds {
			// P doesn't hold, so P until Q is violated
			trace.Position = originalPos
			return false, nil
		}
		
		trace.NextState()
	}
	
	// Check last state
	state := trace.CurrentState()
	
	// Check if Q holds in last state
	qHolds, err := te.evaluateExpression(expr.Right, state)
	if err != nil {
		trace.Position = originalPos
		return false, err
	}
	if qHolds {
		trace.Position = originalPos
		return true, nil
	}
	
	// Q doesn't hold in last state, so P must hold
	pHolds, err := te.evaluateExpression(expr.Left, state)
	trace.Position = originalPos
	return pHolds, err
}

// evaluateLeadsTo evaluates "P → Q" - if P becomes true, Q will eventually become true
func (te *TemporalEvaluator) evaluateLeadsTo(expr *ast.LeadsToExpr, trace *Trace) (bool, error) {
	originalPos := trace.Position
	trace.Reset()
	
	// For each position where P holds, check if Q eventually holds
	for !trace.IsComplete() {
		state := trace.CurrentState()
		
		// Check if P holds
		pHolds, err := te.evaluateExpression(expr.Left, state)
		if err != nil {
			trace.Position = originalPos
			return false, err
		}
		
		if pHolds {
			// P holds, check if Q eventually holds from this point
			checkPos := trace.Position
			foundQ := false
			
			for !trace.IsComplete() {
				checkState := trace.CurrentState()
				qHolds, err := te.evaluateExpression(expr.Right, checkState)
				if err != nil {
					trace.Position = originalPos
					return false, err
				}
				if qHolds {
					foundQ = true
					break
				}
				trace.NextState()
			}
			
			// Check last state if we didn't find Q
			if !foundQ {
				lastState := trace.CurrentState()
				qHolds, err := te.evaluateExpression(expr.Right, lastState)
				if err != nil {
					trace.Position = originalPos
					return false, err
				}
				if !qHolds {
					// P holds but Q never becomes true
					trace.Position = originalPos
					return false, nil
				}
			}
			
			// Reset to checkPos + 1 to continue checking
			trace.Position = checkPos
		}
		
		trace.NextState()
	}
	
	// Check last state
	state := trace.CurrentState()
	pHolds, err := te.evaluateExpression(expr.Left, state)
	if err != nil {
		trace.Position = originalPos
		return false, err
	}
	
	if pHolds {
		// P holds in last state, check if Q holds
		qHolds, err := te.evaluateExpression(expr.Right, state)
		trace.Position = originalPos
		return qHolds, err
	}
	
	trace.Position = originalPos
	return true, nil
}

// evaluateBinaryTemporal evaluates binary temporal expressions (AND, OR)
func (te *TemporalEvaluator) evaluateBinaryTemporal(expr *ast.BinaryExpr, trace *Trace) (bool, error) {
	left, err := te.EvaluateTemporalProperty(expr.Left, trace)
	if err != nil {
		return false, err
	}
	
	right, err := te.EvaluateTemporalProperty(expr.Right, trace)
	if err != nil {
		return false, err
	}
	
	switch expr.Op {
	case ast.And:
		return left && right, nil
	case ast.Or:
		return left || right, nil
	default:
		return false, fmt.Errorf("unsupported binary operator in temporal expression: %v", expr.Op)
	}
}

// evaluateExpression evaluates a non-temporal expression in a given state
func (te *TemporalEvaluator) evaluateExpression(expr ast.Expr, s *state.State) (bool, error) {
	// Set up environment with state variables
	env := eval.NewEnvironment()
	
	// Add state variables to environment
	for varName, varValue := range s.Variables {
		env.SetVariable(varName, varValue)
	}
	
	// Create evaluator with this environment
	evaluator := eval.NewEvaluator(env)
	
	// Evaluate expression
	result, err := evaluator.Eval(expr)
	if err != nil {
		return false, err
	}
	
	// Convert result to boolean
	return te.valueToBool(result)
}

// valueToBool converts a value to boolean
func (te *TemporalEvaluator) valueToBool(val state.Value) (bool, error) {
	switch v := val.(type) {
	case *state.PrimitiveValue:
		if v.TypeName == "bool" && v.BoolValue != nil {
			return *v.BoolValue, nil
		}
		return false, fmt.Errorf("expected boolean value, got %v", v.TypeName)
	default:
		return false, fmt.Errorf("cannot convert value to boolean: %T", val)
	}
}

