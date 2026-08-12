package state

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

// DemonstrateStateMachineModel shows what we can extract from a spec file
func DemonstrateStateMachineModel(t *testing.T) {
	examplePath := filepath.Join("..", "..", "examples", "counter.spec")

	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("failed to read example file: %v", err)
	}

	p := parser.New(lexer.New(string(content)))
	file := p.ParseFile()

	// Build all models
	variableModel := NewVariableModel(file)
	initialStateModel, _ := NewInitialStateModel(file)
	actionModel := NewActionModel(file)
	constraintModel := NewConstraintModel(file, actionModel)

	// Print what we extracted
	t.Log("=== State Machine Model ===")
	t.Logf("\nVariables (%d):", len(variableModel.Variables))
	for name := range variableModel.Variables {
		t.Logf("  - %s", name)
	}

	t.Logf("\nInitial States:")
	if initialStateModel.IsDeterministic() {
		t.Logf("  - Deterministic (1 initial state)")
	} else if initialStateModel.IsOneOf() {
		t.Logf("  - Non-deterministic (oneOf with %d options)", initialStateModel.Count())
	}

	t.Logf("\nActions (%d):", len(actionModel.Actions))
	for name, info := range actionModel.Actions {
		actionStr := fmt.Sprintf("  - %s", name)
		if info.HasGuard() {
			actionStr += " (with guard)"
		}
		if len(info.Parameters) > 0 {
			actionStr += fmt.Sprintf(" (parameters: %d)", len(info.Parameters))
		}
		t.Log(actionStr)
	}

	t.Logf("\nInvariants (%d):", constraintModel.GetInvariantCount())
	for _, inv := range constraintModel.Invariants {
		t.Logf("  - %s", inv.Name)
	}

	t.Logf("\nPreconditions:")
	for actionName, preconds := range constraintModel.Preconditions {
		t.Logf("  - %s: %d precondition(s)", actionName, len(preconds))
	}

	t.Logf("\nPostconditions:")
	for actionName, postconds := range constraintModel.Postconditions {
		t.Logf("  - %s: %d postcondition(s)", actionName, len(postconds))
	}
}

// TestStateMachineModelDemo demonstrates the current capabilities
func TestStateMachineModelDemo(t *testing.T) {
	// This test demonstrates what we CAN do now
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

invariant nonNegative {
    counter >= 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// We CAN extract the model
	variableModel := NewVariableModel(file)
	initialStateModel, _ := NewInitialStateModel(file)
	actionModel := NewActionModel(file)
	constraintModel := NewConstraintModel(file, actionModel)

	// Verify we extracted everything correctly
	if len(variableModel.Variables) != 2 {
		t.Errorf("expected 2 variables, got %d", len(variableModel.Variables))
	}
	if !initialStateModel.IsDeterministic() {
		t.Error("should have deterministic initial state")
	}
	if len(actionModel.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(actionModel.Actions))
	}
	if constraintModel.GetInvariantCount() != 1 {
		t.Errorf("expected 1 invariant, got %d", constraintModel.GetInvariantCount())
	}

	// What we CANNOT do yet (commented out - would require Phase 13/14):
	// - Evaluate the init block to create a concrete State with counter=0, flag=false
	// - Apply the increment action to generate a new state with counter=1
	// - Check if the invariant holds in a given state
	// - Explore all reachable states

	t.Log("✓ Model extraction works")
	t.Log("✗ State execution not yet implemented (requires Phase 13/14)")
}

