package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

// TestStateMachineModelOnCounterSpec tests the complete state machine model on counter.spec
func TestStateMachineModelOnCounterSpec(t *testing.T) {
	examplePath := filepath.Join("..", "..", "examples", "counter.spec")

	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("failed to read example file: %v", err)
	}

	p := parser.New(lexer.New(string(content)))
	file := p.ParseFile()

	// Some parse errors may be expected for unimplemented features
	if len(p.Errors()) > 0 {
		t.Logf("Parse errors in %s (may be expected): %v", examplePath, p.Errors())
	}

	// Build all models
	variableModel := NewVariableModel(file)
	initialStateModel, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("failed to create initial state model: %v", err)
	}
	actionModel := NewActionModel(file)
	constraintModel := NewConstraintModel(file, actionModel)

	// Verify variable model
	if len(variableModel.Variables) == 0 {
		t.Error("expected at least one variable in counter.spec")
	}
	if !variableModel.HasVariable("counter") {
		t.Error("expected variable 'counter' not found")
	}

	// Verify initial state model
	if !initialStateModel.IsDeterministic() {
		t.Error("counter.spec should have deterministic initial state")
	}
	if initialStateModel.Count() != 1 {
		t.Errorf("expected 1 initial state, got %d", initialStateModel.Count())
	}

	// Verify action model
	if len(actionModel.Actions) == 0 {
		t.Error("expected at least one action in counter.spec")
	}
	expectedActions := []string{"increment", "decrement", "reset"}
	for _, actionName := range expectedActions {
		if !actionModel.HasAction(actionName) {
			t.Errorf("expected action '%s' not found", actionName)
		}
	}

	// Verify constraint model
	if constraintModel.GetInvariantCount() == 0 {
		t.Error("expected at least one invariant in counter.spec")
	}
	if !constraintModel.HasInvariants() {
		t.Error("counter.spec should have invariants")
	}

	// Verify specific invariants
	expectedInvariants := []string{"nonNegative", "bounded"}
	for _, invName := range expectedInvariants {
		inv := constraintModel.GetInvariant(invName)
		if inv == nil {
			t.Errorf("expected invariant '%s' not found", invName)
		}
	}

	// Verify preconditions
	if !constraintModel.HasPreconditions("decrement") {
		t.Error("action 'decrement' should have preconditions")
	}
	decPreconds := constraintModel.GetPreconditions("decrement")
	if len(decPreconds) == 0 {
		t.Error("action 'decrement' should have at least one precondition")
	}
}

// TestStateMachineModelOnAllExamples tests the state machine model on all example files
func TestStateMachineModelOnAllExamples(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")
	files, err := filepath.Glob(filepath.Join(examplesDir, "*.spec"))
	if err != nil {
		t.Fatalf("failed to find example files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("no example files found")
	}

	for _, filePath := range files {
		t.Run(filepath.Base(filePath), func(t *testing.T) {
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", filePath, err)
			}

			// Parse the file
			p := parser.New(lexer.New(string(content)))
			file := p.ParseFile()

			// Some parse errors are expected for unimplemented features
			if len(p.Errors()) > 0 {
				t.Logf("Parse errors in %s (may be expected): %v", filePath, p.Errors())
			}

			// Build all models
			variableModel := NewVariableModel(file)
			initialStateModel, err := NewInitialStateModel(file)
			if err != nil {
				// Some files may not have init declarations (expected)
				t.Logf("Initial state model error in %s (may be expected): %v", filePath, err)
				return
			}
			actionModel := NewActionModel(file)
			constraintModel := NewConstraintModel(file, actionModel)

			// Basic sanity checks
			if variableModel == nil {
				t.Error("variable model should not be nil")
			}
			if initialStateModel == nil {
				t.Error("initial state model should not be nil")
			}
			if actionModel == nil {
				t.Error("action model should not be nil")
			}
			if constraintModel == nil {
				t.Error("constraint model should not be nil")
			}

			// Verify models are consistent
			// Variables should be extractable
			if len(variableModel.Variables) > 0 {
				// If there are variables, we should be able to get their names
				names := variableModel.GetVariableNames()
				if len(names) != len(variableModel.Variables) {
					t.Errorf("variable names count mismatch: %d vs %d", len(names), len(variableModel.Variables))
				}
			}

			// Actions should be extractable
			if len(actionModel.Actions) > 0 {
				// If there are actions, we should be able to get their names
				names := actionModel.GetActionNames()
				if len(names) != len(actionModel.Actions) {
					t.Errorf("action names count mismatch: %d vs %d", len(names), len(actionModel.Actions))
				}
			}

			// Initial state model should be valid
			if initialStateModel.IsDeterministic() && initialStateModel.IsOneOf() {
				t.Error("initial state cannot be both deterministic and oneOf")
			}

			// If there are invariants, they should be accessible
			if constraintModel.HasInvariants() {
				invariants := constraintModel.GetInvariants()
				if len(invariants) != constraintModel.GetInvariantCount() {
					t.Errorf("invariant count mismatch: %d vs %d", len(invariants), constraintModel.GetInvariantCount())
				}
			}
		})
	}
}

// TestStateMachineModelIntegration tests the integration of all models
func TestStateMachineModelIntegration(t *testing.T) {
	input := `
var counter: int
var flag: bool

init {
    counter = 0
    flag = false
}

action increment {
    require counter < 100
    counter' = counter + 1
    ensure counter > 0
}

action decrement {
    require counter > 0
    counter' = counter - 1
    ensure counter >= 0
}

invariant nonNegative {
    counter >= 0
}

invariant bounded {
    counter <= 100
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build all models
	variableModel := NewVariableModel(file)
	initialStateModel, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("failed to create initial state model: %v", err)
	}
	actionModel := NewActionModel(file)
	constraintModel := NewConstraintModel(file, actionModel)

	// Verify variable model
	if len(variableModel.Variables) != 2 {
		t.Errorf("expected 2 variables, got %d", len(variableModel.Variables))
	}
	if !variableModel.HasVariable("counter") {
		t.Error("variable 'counter' not found")
	}
	if !variableModel.HasVariable("flag") {
		t.Error("variable 'flag' not found")
	}

	// Verify initial state model
	if !initialStateModel.IsDeterministic() {
		t.Error("should have deterministic initial state")
	}
	if initialStateModel.Count() != 1 {
		t.Errorf("expected 1 initial state, got %d", initialStateModel.Count())
	}

	// Verify action model
	if len(actionModel.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actionModel.Actions))
	}
	if !actionModel.HasAction("increment") {
		t.Error("action 'increment' not found")
	}
	if !actionModel.HasAction("decrement") {
		t.Error("action 'decrement' not found")
	}

	// Verify constraint model
	if constraintModel.GetInvariantCount() != 2 {
		t.Errorf("expected 2 invariants, got %d", constraintModel.GetInvariantCount())
	}
	if !constraintModel.HasPreconditions("increment") {
		t.Error("action 'increment' should have preconditions")
	}
	if !constraintModel.HasPreconditions("decrement") {
		t.Error("action 'decrement' should have preconditions")
	}
	if !constraintModel.HasPostconditions("increment") {
		t.Error("action 'increment' should have postconditions")
	}
	if !constraintModel.HasPostconditions("decrement") {
		t.Error("action 'decrement' should have postconditions")
	}

	// Verify invariants
	if constraintModel.GetInvariant("nonNegative") == nil {
		t.Error("invariant 'nonNegative' not found")
	}
	if constraintModel.GetInvariant("bounded") == nil {
		t.Error("invariant 'bounded' not found")
	}
}

// TestStateMachineModelOneOf tests the model with oneOf initial states
func TestStateMachineModelOneOf(t *testing.T) {
	input := `
var counter: int

init oneOf {
    { counter = 0 },
    { counter = 5 },
    { counter = 10 }
}

action increment {
    counter' = counter + 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	initialStateModel, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("failed to create initial state model: %v", err)
	}

	if initialStateModel.IsDeterministic() {
		t.Error("should not be deterministic")
	}
	if !initialStateModel.IsOneOf() {
		t.Error("should be oneOf")
	}
	if initialStateModel.Count() != 3 {
		t.Errorf("expected 3 initial states, got %d", initialStateModel.Count())
	}

	options := initialStateModel.GetOneOfOptions()
	if len(options) != 3 {
		t.Errorf("expected 3 oneOf options, got %d", len(options))
	}
}

// TestStateMachineModelWithModules tests the model with modules
func TestStateMachineModelWithModules(t *testing.T) {
	input := `
module Counter {
    var counter: int
    
    init {
        counter = 0
    }
    
    action increment {
        counter' = counter + 1
    }
    
    invariant nonNegative {
        counter >= 0
    }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	variableModel := NewVariableModel(file)
	initialStateModel, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("failed to create initial state model: %v", err)
	}
	actionModel := NewActionModel(file)
	constraintModel := NewConstraintModel(file, actionModel)

	// Verify models extract from modules
	if !variableModel.HasVariable("counter") {
		t.Error("variable 'counter' not found in module")
	}
	if !initialStateModel.IsDeterministic() {
		t.Error("should have deterministic initial state in module")
	}
	if !actionModel.HasAction("increment") {
		t.Error("action 'increment' not found in module")
	}
	if constraintModel.GetInvariant("nonNegative") == nil {
		t.Error("invariant 'nonNegative' not found in module")
	}
}

