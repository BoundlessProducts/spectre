package temporal

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// FairnessChecker checks fairness conditions (WF and SF) over execution traces
type FairnessChecker struct {
	stateMachine *exec.StateMachine
	enablednessChecker *ActionEnablednessChecker
}

// NewFairnessChecker creates a new fairness checker
func NewFairnessChecker(stateMachine *exec.StateMachine) *FairnessChecker {
	return &FairnessChecker{
		stateMachine: stateMachine,
		enablednessChecker: NewActionEnablednessChecker(stateMachine),
	}
}

// EvaluateFairness evaluates a fairness expression (WF or SF) over a trace
func (fc *FairnessChecker) EvaluateFairness(expr ast.Expr, trace *Trace) (bool, error) {
	switch e := expr.(type) {
	case *ast.WFExpr:
		return fc.evaluateWeakFairness(e, trace)
	case *ast.SFExpr:
		return fc.evaluateStrongFairness(e, trace)
	default:
		return false, fmt.Errorf("expected WF or SF expression, got %T", expr)
	}
}

// evaluateWeakFairness evaluates WF(action) - action executes if continuously enabled
// Weak Fairness: If an action is continuously enabled from some point onwards, it must eventually execute
func (fc *FairnessChecker) evaluateWeakFairness(expr *ast.WFExpr, trace *Trace) (bool, error) {
	// Extract action name from target expression
	actionName, err := fc.extractActionName(expr.Target)
	if err != nil {
		return false, err
	}

	// Find all positions where action is enabled
	enabledPositions, err := fc.enablednessChecker.GetEnabledPositions(actionName, trace)
	if err != nil {
		return false, err
	}

	if len(enabledPositions) == 0 {
		// Action never enabled, weak fairness trivially satisfied
		return true, nil
	}

	// Find all positions where action executes
	executionPositions := fc.enablednessChecker.GetExecutionPositions(actionName, trace)

	// Check for each continuous enabled interval
	for i := 0; i < len(enabledPositions); i++ {
		start := enabledPositions[i]
		
		// Find the end of this continuous enabled interval
		end := start
		for j := i + 1; j < len(enabledPositions); j++ {
			if enabledPositions[j] == end+1 {
				end = enabledPositions[j]
				i = j // Update outer loop index
			} else {
				break
			}
		}

		// Check if action executes during or after this continuous interval
		executed := false
		for _, execPos := range executionPositions {
			// Action executes at execPos, which corresponds to transition from state execPos to execPos+1
			// So if execPos >= start, action executed during or after the interval
			if execPos >= start {
				executed = true
				break
			}
		}

		if !executed {
			// Action was continuously enabled but never executed - weak fairness violated
			return false, nil
		}
	}

	// All continuous enabled intervals had executions - weak fairness satisfied
	return true, nil
}

// evaluateStrongFairness evaluates SF(action) - action executes if enabled infinitely often
// Strong Fairness: If an action is enabled infinitely often, it must execute infinitely often
func (fc *FairnessChecker) evaluateStrongFairness(expr *ast.SFExpr, trace *Trace) (bool, error) {
	// Extract action name from target expression
	actionName, err := fc.extractActionName(expr.Target)
	if err != nil {
		return false, err
	}

	// Find all positions where action is enabled
	enabledPositions, err := fc.enablednessChecker.GetEnabledPositions(actionName, trace)
	if err != nil {
		return false, err
	}

	if len(enabledPositions) == 0 {
		// Action never enabled, strong fairness trivially satisfied
		return true, nil
	}

	// Find all positions where action executes
	executionPositions := fc.enablednessChecker.GetExecutionPositions(actionName, trace)

	// For strong fairness, we need to check that for every enabled position,
	// there is an execution at that position or later
	// In a finite trace, "infinitely often" means "many times" or "repeatedly"
	
	// Check if action executes after each enabled position
	for _, enabledPos := range enabledPositions {
		executed := false
		
		// Check if action executes at enabledPos or later
		for _, execPos := range executionPositions {
			if execPos >= enabledPos {
				executed = true
				break
			}
		}
		
		if !executed {
			// Action was enabled at enabledPos but never executed afterwards
			// Strong fairness violated
			return false, nil
		}
	}

	// All enabled positions had executions afterwards - strong fairness satisfied
	return true, nil
}

// extractActionName extracts the action name from an expression (usually an Ident)
func (fc *FairnessChecker) extractActionName(expr ast.Expr) (string, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name, nil
	default:
		return "", fmt.Errorf("expected identifier for action name, got %T", expr)
	}
}

