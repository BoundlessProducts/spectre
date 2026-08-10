package exec

import (
	"fmt"
	"sort"

	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// StateMachine represents a complete state machine executor
type StateMachine struct {
	variableModel     *state.VariableModel
	initialStateModel *state.InitialStateModel
	actionModel       *state.ActionModel
	constraintModel   *state.ConstraintModel
	initializer       *StateInitializer
	executor          *ActionExecutor
	validator         *StateValidator
	params            map[string]int64 // spec params bound via --param flags
}

// NewStateMachine creates a new state machine from a parsed file
// additionalFiles can be provided to register constants/enums from imported modules
func NewStateMachine(file *ast.File, additionalFiles ...*ast.File) (*StateMachine, error) {
	vm := state.NewVariableModel(file)
	ism, err := state.NewInitialStateModel(file)
	if err != nil {
		return nil, fmt.Errorf("error creating initial state model: %w", err)
	}

	am := state.NewActionModel(file)
	cm := state.NewConstraintModel(file, am)

	// Pass all files (including imported modules) to components for constant/enum registration
	allFiles := append([]*ast.File{file}, additionalFiles...)
	
	return &StateMachine{
		variableModel:     vm,
		initialStateModel: ism,
		actionModel:       am,
		constraintModel:   cm,
		initializer:       NewStateInitializer(vm, ism, file, allFiles...),
		executor:          NewActionExecutor(vm, am, cm, file, allFiles...),
		validator:         NewStateValidator(cm, file, allFiles...),
	}, nil
}

// ExecuteAction executes an action on a state and returns the next state
func (sm *StateMachine) ExecuteAction(actionName string, currentState *state.State, args []state.Value) (*state.State, error) {
	// Check if action can be executed (guards/preconditions)
	canExecute, err := sm.executor.CanExecute(actionName, currentState, args)
	if err != nil {
		return nil, fmt.Errorf("error checking if action can execute: %w", err)
	}

	if !canExecute {
		return nil, fmt.Errorf("action %s cannot be executed in current state", actionName)
	}

	// Validate current state invariants
	errors, err := sm.validator.ValidateState(currentState)
	if err != nil {
		return nil, fmt.Errorf("error validating current state: %w", err)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("current state violates invariants: %v", errors)
	}

	// Execute action
	nextState, err := sm.executor.ExecuteAction(actionName, currentState, args)
	if err != nil {
		return nil, fmt.Errorf("error executing action: %w", err)
	}
	sm.injectParams(nextState)

	// Validate next state invariants
	errors, err = sm.validator.ValidateState(nextState)
	if err != nil {
		return nil, fmt.Errorf("error validating next state: %w", err)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("next state violates invariants: %v", errors)
	}

	// Validate postconditions
	errors, err = sm.validator.ValidatePostconditions(actionName, currentState, nextState)
	if err != nil {
		return nil, fmt.Errorf("error validating postconditions: %w", err)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("postconditions violated: %v", errors)
	}

	return nextState, nil
}

// GetAvailableActions returns all actions that can be executed in the current state
// For parameterized actions, it generates argument combinations and checks which ones are valid
func (sm *StateMachine) GetAvailableActions(currentState *state.State) ([]string, error) {
	var available []string

	// Create environment to access constants for argument generation
	env := eval.NewEnvironment()
	allFiles := sm.executor.getAllFiles()
	for _, f := range allFiles {
		eval.RegisterEnumTypes(env, f)
		eval.RegisterConstants(env, f, allFiles...)
	}

	// Extract and sort action names for deterministic iteration order
	actionNames := make([]string, 0, len(sm.actionModel.Actions))
	for actionName := range sm.actionModel.Actions {
		actionNames = append(actionNames, actionName)
	}
	sort.Strings(actionNames)

	for _, actionName := range actionNames {
		actionInfo, exists := sm.actionModel.GetAction(actionName)
		if !exists {
			continue
		}
		
		if len(actionInfo.Parameters) == 0 {
			// Non-parameterized action - check directly
			canExecute, err := sm.executor.CanExecute(actionName, currentState, nil)
			if err != nil {
				return nil, fmt.Errorf("error checking action %s: %w", actionName, err)
			}

			if canExecute {
				available = append(available, actionName)
			}
		} else {
			// Parameterized action - generate argument combinations
			argCombos := generateArgumentCombinations(actionInfo, env, sm.executor.file)
			
			// Check if at least one combination is valid
			hasValidCombo := false
			for _, args := range argCombos {
				// Skip if any parameter has no possible values (empty combo)
				if len(args) != len(actionInfo.Parameters) {
					continue
				}
				
				canExecute, err := sm.executor.CanExecute(actionName, currentState, args)
				if err != nil {
					// Skip this combination if there's an error
					continue
				}

				if canExecute {
					hasValidCombo = true
					break
				}
			}
			
			if hasValidCombo {
				available = append(available, actionName)
			}
		}
	}

	return available, nil
}

// GetAvailableActionsWithArgs returns all actions with their valid argument combinations
func (sm *StateMachine) GetAvailableActionsWithArgs(currentState *state.State) ([]*ActionWithArgs, error) {
	var available []*ActionWithArgs

	// Create environment to access constants for argument generation
	env := eval.NewEnvironment()
	allFiles := sm.executor.getAllFiles()
	for _, f := range allFiles {
		eval.RegisterEnumTypes(env, f)
		eval.RegisterConstants(env, f, allFiles...)
	}

	// Extract and sort action names for deterministic iteration order
	actionNames := make([]string, 0, len(sm.actionModel.Actions))
	for actionName := range sm.actionModel.Actions {
		actionNames = append(actionNames, actionName)
	}
	sort.Strings(actionNames)

	for _, actionName := range actionNames {
		actionInfo, exists := sm.actionModel.GetAction(actionName)
		if !exists {
			continue
		}
		
		if len(actionInfo.Parameters) == 0 {
			// Non-parameterized action - check directly
			canExecute, err := sm.executor.CanExecute(actionName, currentState, nil)
			if err != nil {
				return nil, fmt.Errorf("error checking action %s: %w", actionName, err)
			}

			if canExecute {
				available = append(available, &ActionWithArgs{
					ActionName: actionName,
					Args:       nil,
				})
			}
		} else {
			// Parameterized action - generate argument combinations
			argCombos := generateArgumentCombinations(actionInfo, env, sm.executor.file)
			
			for _, args := range argCombos {
				// Skip if any parameter has no possible values (empty combo)
				if len(args) != len(actionInfo.Parameters) {
					continue
				}
				
				canExecute, err := sm.executor.CanExecute(actionName, currentState, args)
				if err != nil {
					// Skip this combination if there's an error
					continue
				}

				if canExecute {
					available = append(available, &ActionWithArgs{
						ActionName: actionName,
						Args:       args,
					})
				}
			}
		}
	}

	return available, nil
}

// ValidateState validates a state against all invariants
func (sm *StateMachine) ValidateState(s *state.State) ([]*ValidationError, error) {
	return sm.validator.ValidateState(s)
}

// GetConstraintModel returns the constraint model
func (sm *StateMachine) GetConstraintModel() *state.ConstraintModel {
	return sm.constraintModel
}

// SetParam binds a spec parameter to a concrete integer value.
// The parameter becomes visible as a read-only variable in all initial states,
// which makes it accessible during invariant and guard evaluation.
func (sm *StateMachine) SetParam(name string, value int64) {
	if sm.params == nil {
		sm.params = make(map[string]int64)
	}
	sm.params[name] = value
}

// injectParams adds bound spec parameters into a state's variable map.
// It is called by GetInitialStates so that params are available throughout exploration.
func (sm *StateMachine) injectParams(s *state.State) {
	for k, v := range sm.params {
		n := v
		s.Variables[k] = &state.PrimitiveValue{TypeName: "int", IntValue: &n}
	}
}

// GetInitialStates returns all possible initial states with params injected.
func (sm *StateMachine) GetInitialStates() ([]*state.State, error) {
	states, err := sm.initializer.GenerateInitialStates()
	if err != nil {
		return nil, err
	}
	for _, s := range states {
		sm.injectParams(s)
	}
	return states, nil
}

