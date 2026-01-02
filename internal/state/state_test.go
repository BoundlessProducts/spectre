package state

import (
	"testing"

	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestNewState(t *testing.T) {
	s := NewState()
	if s == nil {
		t.Fatal("NewState returned nil")
	}
	if s.Variables == nil {
		t.Error("Variables map should not be nil")
	}
	if len(s.Variables) != 0 {
		t.Errorf("expected empty state, got %d variables", len(s.Variables))
	}
}

func TestSetAndGetVariable(t *testing.T) {
	s := NewState()

	// Set an integer variable
	intVal := NewIntValue(42)
	s.SetVariable("counter", intVal)

	// Get the variable
	value, exists := s.GetVariable("counter")
	if !exists {
		t.Error("variable 'counter' should exist")
	}
	if value == nil {
		t.Error("value should not be nil")
	}

	// Verify it's the correct value
	if pv, ok := value.(*PrimitiveValue); ok {
		if pv.Type() != "int" {
			t.Errorf("expected type 'int', got '%s'", pv.Type())
		}
		if pv.IntValue == nil || *pv.IntValue != 42 {
			t.Errorf("expected value 42, got %v", pv.IntValue)
		}
	} else {
		t.Error("value should be *PrimitiveValue")
	}
}

func TestPrimitiveValues(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{
			name:     "int value",
			value:    NewIntValue(42),
			expected: "42",
		},
		{
			name:     "float value",
			value:    NewFloatValue(3.14),
			expected: "3.140000",
		},
		{
			name:     "string value",
			value:    NewStringValue("hello"),
			expected: `"hello"`,
		},
		{
			name:     "bool value true",
			value:    NewBoolValue(true),
			expected: "true",
		},
		{
			name:     "bool value false",
			value:    NewBoolValue(false),
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.value.String())
			}
		})
	}
}

func TestStateCopy(t *testing.T) {
	s1 := NewState()
	s1.SetVariable("counter", NewIntValue(10))
	s1.SetVariable("flag", NewBoolValue(true))

	s2 := s1.Copy()

	// Verify both states have the same variables
	if len(s1.Variables) != len(s2.Variables) {
		t.Errorf("copy should have same number of variables: %d vs %d", len(s1.Variables), len(s2.Variables))
	}

	// Verify values are copied
	val1, _ := s1.GetVariable("counter")
	val2, _ := s2.GetVariable("counter")
	if val1.String() != val2.String() {
		t.Errorf("copied value should match: %s vs %s", val1.String(), val2.String())
	}

	// Modify original and verify copy is independent
	s1.SetVariable("counter", NewIntValue(20))
	val1, _ = s1.GetVariable("counter")
	val2, _ = s2.GetVariable("counter")
	if val1.String() == val2.String() {
		t.Error("modifying original should not affect copy")
	}
}

func TestStateEquals(t *testing.T) {
	s1 := NewState()
	s1.SetVariable("counter", NewIntValue(10))
	s1.SetVariable("flag", NewBoolValue(true))

	s2 := NewState()
	s2.SetVariable("counter", NewIntValue(10))
	s2.SetVariable("flag", NewBoolValue(true))

	if !s1.Equals(s2) {
		t.Error("states with same variables and values should be equal")
	}

	// Modify one state
	s2.SetVariable("counter", NewIntValue(20))
	if s1.Equals(s2) {
		t.Error("states with different values should not be equal")
	}

	// Different number of variables
	s3 := NewState()
	s3.SetVariable("counter", NewIntValue(10))
	if s1.Equals(s3) {
		t.Error("states with different number of variables should not be equal")
	}
}

func TestPrimitiveValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		v1       *PrimitiveValue
		v2       *PrimitiveValue
		expected bool
	}{
		{
			name:     "equal ints",
			v1:       NewIntValue(42),
			v2:       NewIntValue(42),
			expected: true,
		},
		{
			name:     "unequal ints",
			v1:       NewIntValue(42),
			v2:       NewIntValue(43),
			expected: false,
		},
		{
			name:     "equal floats",
			v1:       NewFloatValue(3.14),
			v2:       NewFloatValue(3.14),
			expected: true,
		},
		{
			name:     "equal strings",
			v1:       NewStringValue("hello"),
			v2:       NewStringValue("hello"),
			expected: true,
		},
		{
			name:     "unequal strings",
			v1:       NewStringValue("hello"),
			v2:       NewStringValue("world"),
			expected: false,
		},
		{
			name:     "equal bools",
			v1:       NewBoolValue(true),
			v2:       NewBoolValue(true),
			expected: true,
		},
		{
			name:     "unequal bools",
			v1:       NewBoolValue(true),
			v2:       NewBoolValue(false),
			expected: false,
		},
		{
			name:     "different types",
			v1:       NewIntValue(42),
			v2:       NewStringValue("42"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := primitiveValuesEqual(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestStatePosition(t *testing.T) {
	s := NewState()
	s.Position = ast.Position{Line: 10, Column: 5, Offset: 100}

	if s.Position.Line != 10 {
		t.Errorf("expected line 10, got %d", s.Position.Line)
	}
	if s.Position.Column != 5 {
		t.Errorf("expected column 5, got %d", s.Position.Column)
	}
}

