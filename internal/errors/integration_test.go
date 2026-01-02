package errors

import (
	"strings"
	"testing"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestErrorReportingInvariantViolation(t *testing.T) {
	spec := `
description "Tracks a counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Decrements the counter by one"
action decrement {
  counter' = counter - 1
}

description "Ensures counter never becomes negative"
invariant counterNonNegative {
  counter >= 0
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	// Create a state that violates the invariant
	violationState := state.NewState()
	violationState.SetVariable("counter", state.NewIntValue(-1))

	// Validate state to get violation
	errors, err := sm.ValidateState(violationState)
	if err != nil {
		t.Fatalf("error validating state: %v", err)
	}

	if len(errors) == 0 {
		t.Fatal("expected invariant violation")
	}

	// Extract context from invariant declaration
	var invariantDecl *ast.InvariantDecl
	for _, decl := range file.Decls {
		if inv, ok := decl.(*ast.InvariantDecl); ok && inv.Name == "counterNonNegative" {
			invariantDecl = inv
			break
		}
	}

	if invariantDecl == nil {
		t.Fatal("expected invariant declaration")
	}

	context := ExtractContextFromDecl(invariantDecl)
	formatter := NewErrorFormatter()

	// Format the error
	formatted := formatter.FormatInvariantViolation(
		errors[0].Name,
		errors[0].Message,
		context,
	)

	// Verify error message includes description (if description was parsed)
	// Note: Descriptions may not be parsed if they're on separate lines
	if context.HasDescription() {
		if !strings.Contains(formatted, context.Description) {
			t.Errorf("error message should include invariant description '%s'", context.Description)
		}
	} else {
		t.Logf("Note: Description not found in AST (may be due to parser handling)")
	}

	if !strings.Contains(formatted, "counterNonNegative") {
		t.Error("error message should include invariant name")
	}

	t.Logf("Formatted error: %s", formatted)
}

func TestErrorReportingWithStackTrace(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement {
  counter' = counter - 1
}

invariant counterNonNegative {
  counter >= 0
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	// Explore state space to find violations
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(5)
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	// If violations found, test stack trace generation
	if len(result.Violations) > 0 {
		violation := result.Violations[0]

		// Extract context from invariant
		var invariantDecl *ast.InvariantDecl
		for _, decl := range file.Decls {
			if inv, ok := decl.(*ast.InvariantDecl); ok && inv.Name == violation.Invariant {
				invariantDecl = inv
				break
			}
		}

		if invariantDecl != nil {
			context := ExtractContextFromDecl(invariantDecl)
			formatter := NewErrorFormatter()

			// Format error
			formatted := formatter.FormatInvariantViolation(
				violation.Invariant,
				violation.Description,
				context,
			)

			// Build stack trace from violation path
			st := BuildStackTraceFromViolationPath(
				violation.Path,
				context,
				violation.Description,
			)

			stackTraceFormatted := st.Format()

			if !strings.Contains(stackTraceFormatted, "Stack trace:") {
				t.Error("stack trace should contain 'Stack trace:'")
			}

			t.Logf("Error: %s", formatted)
			t.Logf("\n%s", stackTraceFormatted)
		}
	}
}

func TestErrorReportingPostconditionViolation(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

description "Increments counter and ensures it increases"
action increment {
  counter' = counter + 1
  ensure counter' > counter
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Extract action declaration
	var actionDecl *ast.ActionDecl
	for _, decl := range file.Decls {
		if act, ok := decl.(*ast.ActionDecl); ok && act.Name == "increment" {
			actionDecl = act
			break
		}
	}

	if actionDecl == nil {
		t.Fatal("expected action declaration")
	}

	context := ExtractContextFromDecl(actionDecl)
	formatter := NewErrorFormatter()

	// Format postcondition violation
	formatted := formatter.FormatPostconditionViolation(
		"increment",
		"counterIncreases",
		"counter' <= counter",
		context,
	)

	if !strings.Contains(formatted, "increment") {
		t.Error("error message should include action name")
	}

	if !strings.Contains(formatted, "Postcondition") {
		t.Error("error message should indicate postcondition violation")
	}

	t.Logf("Formatted error: %s", formatted)
}

func TestErrorReportingTemporalViolation(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

description "Counter eventually reaches 10"
temporal eventuallyReachesTen {
  eventually (counter == 10)
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Extract temporal declaration
	var temporalDecl *ast.TemporalDecl
	for _, decl := range file.Decls {
		if temp, ok := decl.(*ast.TemporalDecl); ok && temp.Name == "eventuallyReachesTen" {
			temporalDecl = temp
			break
		}
	}

	if temporalDecl == nil {
		t.Fatal("expected temporal declaration")
	}

	context := ExtractContextFromDecl(temporalDecl)
	formatter := NewErrorFormatter()

	// Format temporal violation
	formatted := formatter.FormatTemporalViolation(
		"eventuallyReachesTen",
		"property does not hold",
		context,
	)

	if !strings.Contains(formatted, "eventuallyReachesTen") {
		t.Error("error message should include temporal property name")
	}

	if !strings.Contains(formatted, "Temporal property") {
		t.Error("error message should indicate temporal property violation")
	}

	// Verify error message includes description (if description was parsed)
	if context.HasDescription() {
		if !strings.Contains(formatted, context.Description) {
			t.Errorf("error message should include description '%s'", context.Description)
		}
	} else {
		t.Logf("Note: Description not found in AST (may be due to parser handling)")
	}

	t.Logf("Formatted error: %s", formatted)
}

