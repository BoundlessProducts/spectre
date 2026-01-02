package temporal

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/state"
)

// ActionEnablednessChecker checks when actions are enabled in states
type ActionEnablednessChecker struct {
	stateMachine *exec.StateMachine
}

// NewActionEnablednessChecker creates a new action enabledness checker
func NewActionEnablednessChecker(stateMachine *exec.StateMachine) *ActionEnablednessChecker {
	return &ActionEnablednessChecker{
		stateMachine: stateMachine,
	}
}

// IsActionEnabled checks if a specific action is enabled in a given state
func (aec *ActionEnablednessChecker) IsActionEnabled(actionName string, s *state.State) (bool, error) {
	availableActions, err := aec.stateMachine.GetAvailableActions(s)
	if err != nil {
		return false, fmt.Errorf("error getting available actions: %w", err)
	}

	for _, name := range availableActions {
		if name == actionName {
			return true, nil
		}
	}

	return false, nil
}

// GetEnabledActions returns all actions that are enabled in a given state
func (aec *ActionEnablednessChecker) GetEnabledActions(s *state.State) ([]string, error) {
	return aec.stateMachine.GetAvailableActions(s)
}

// IsActionEnabledInTrace checks if an action is enabled at a specific position in a trace
func (aec *ActionEnablednessChecker) IsActionEnabledInTrace(actionName string, trace *Trace, position int) (bool, error) {
	if position < 0 || position >= trace.Length() {
		return false, fmt.Errorf("position %d out of range [0, %d)", position, trace.Length())
	}

	state := trace.GetState(position)
	return aec.IsActionEnabled(actionName, state)
}

// GetEnabledActionsInTrace returns all actions enabled at a specific position in a trace
func (aec *ActionEnablednessChecker) GetEnabledActionsInTrace(trace *Trace, position int) ([]string, error) {
	if position < 0 || position >= trace.Length() {
		return nil, fmt.Errorf("position %d out of range [0, %d)", position, trace.Length())
	}

	state := trace.GetState(position)
	return aec.GetEnabledActions(state)
}

// IsContinuouslyEnabled checks if an action is continuously enabled from start to end position
func (aec *ActionEnablednessChecker) IsContinuouslyEnabled(actionName string, trace *Trace, start, end int) (bool, error) {
	if start < 0 || end > trace.Length() || start >= end {
		return false, fmt.Errorf("invalid range [%d, %d) for trace of length %d", start, end, trace.Length())
	}

	for i := start; i < end; i++ {
		enabled, err := aec.IsActionEnabledInTrace(actionName, trace, i)
		if err != nil {
			return false, err
		}
		if !enabled {
			return false, nil
		}
	}

	return true, nil
}

// GetEnabledPositions returns all positions in a trace where an action is enabled
func (aec *ActionEnablednessChecker) GetEnabledPositions(actionName string, trace *Trace) ([]int, error) {
	var positions []int

	for i := 0; i < trace.Length(); i++ {
		enabled, err := aec.IsActionEnabledInTrace(actionName, trace, i)
		if err != nil {
			return nil, err
		}
		if enabled {
			positions = append(positions, i)
		}
	}

	return positions, nil
}

// CountEnabledOccurrences counts how many times an action is enabled in a trace
func (aec *ActionEnablednessChecker) CountEnabledOccurrences(actionName string, trace *Trace) (int, error) {
	positions, err := aec.GetEnabledPositions(actionName, trace)
	if err != nil {
		return 0, err
	}
	return len(positions), nil
}

// ActionExecutesAt checks if an action executes at a specific position in a trace
func (aec *ActionEnablednessChecker) ActionExecutesAt(actionName string, trace *Trace, position int) bool {
	if position < 0 || position >= len(trace.Actions) {
		return false
	}
	return trace.GetAction(position) == actionName
}

// GetExecutionPositions returns all positions in a trace where an action executes
func (aec *ActionEnablednessChecker) GetExecutionPositions(actionName string, trace *Trace) []int {
	var positions []int

	for i := 0; i < len(trace.Actions); i++ {
		if trace.GetAction(i) == actionName {
			positions = append(positions, i)
		}
	}

	return positions
}

// CountExecutionOccurrences counts how many times an action executes in a trace
func (aec *ActionEnablednessChecker) CountExecutionOccurrences(actionName string, trace *Trace) int {
	return len(aec.GetExecutionPositions(actionName, trace))
}

