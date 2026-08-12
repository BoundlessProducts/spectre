package errors

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestErrorFormatterBasic(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 10, Column: 5}
	context := NewErrorContext(pos, "A counter variable", "counter", "variable")

	message := "variable not found"
	formatted := formatter.FormatError(message, context)

	expected := "10:5: variable 'counter' (A counter variable) variable not found"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterWithoutDescription(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 5, Column: 3}
	context := NewErrorContext(pos, "", "myVar", "variable")

	message := "type mismatch"
	formatted := formatter.FormatError(message, context)

	expected := "5:3: variable 'myVar' type mismatch"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterInvariantViolation(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 20, Column: 10}
	context := NewErrorContext(pos, "Counter must be non-negative", "counterNonNegative", "invariant")

	message := "condition evaluated to false"
	formatted := formatter.FormatInvariantViolation("counterNonNegative", message, context)

	expected := "20:10: Invariant 'counterNonNegative' violated: (Counter must be non-negative) condition evaluated to false"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterPostconditionViolation(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 15, Column: 8}
	context := NewErrorContext(pos, "Counter should increase", "counterIncreases", "postcondition")

	message := "counter' <= counter"
	formatted := formatter.FormatPostconditionViolation("increment", "counterIncreases", message, context)

	expected := "15:8: Postcondition 'counterIncreases' of action 'increment' violated: (Counter should increase) counter' <= counter"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterPreconditionViolation(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 12, Column: 4}
	context := NewErrorContext(pos, "Counter must be positive", "counterPositive", "precondition")

	message := "counter <= 0"
	formatted := formatter.FormatPreconditionViolation("decrement", "counterPositive", message, context)

	expected := "12:4: Precondition 'counterPositive' of action 'decrement' violated: (Counter must be positive) counter <= 0"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterTemporalViolation(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 30, Column: 15}
	context := NewErrorContext(pos, "Counter eventually reaches 10", "eventuallyReachesTen", "temporal property")

	message := "property does not hold"
	formatted := formatter.FormatTemporalViolation("eventuallyReachesTen", message, context)

	expected := "30:15: Temporal property 'eventuallyReachesTen' violated: (Counter eventually reaches 10) property does not hold"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterTypeError(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 8, Column: 12}
	formatted := formatter.FormatTypeError("cannot assign int to string", pos)

	expected := "8:12: Type error: cannot assign int to string"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterParseError(t *testing.T) {
	formatter := NewErrorFormatter()

	pos := ast.Position{Line: 3, Column: 7}
	formatted := formatter.FormatParseError("unexpected token", pos)

	expected := "3:7: Parse error: unexpected token"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestErrorFormatterWithoutDescriptions(t *testing.T) {
	formatter := NewErrorFormatter()
	formatter.SetIncludeDescriptions(false)

	pos := ast.Position{Line: 10, Column: 5}
	context := NewErrorContext(pos, "A counter variable", "counter", "variable")

	message := "variable not found"
	formatted := formatter.FormatError(message, context)

	expected := "10:5: variable 'counter' variable not found"
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}

	// Description should not be included
	if strings.Contains(formatted, "A counter variable") {
		t.Error("description should not be included when disabled")
	}
}

