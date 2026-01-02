package parser

import (
	"strings"
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
)

func TestErrorReportingWithPositions(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors []string // Substrings that should appear in error messages
		description    string
	}{
		{
			name: "Error with line number for missing colon",
			input: `var x: int
var y int
var z: int`,
			expectedErrors: []string{"2:", "expected", ":"},
			description:    "Error should include line number",
		},
		{
			name: "Error with position for missing identifier",
			input: `var x: int
var : int`,
			expectedErrors: []string{"2:", "expected identifier"},
			description:    "Error should include line number",
		},
		{
			name: "Multiple errors with positions",
			input: `var a: int
var b int
var c int`,
			expectedErrors: []string{"2:", "3:"},
			description:    "Multiple errors should each have line numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseFile()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Fatalf("expected errors but got none")
			}

			// Check that errors contain expected substrings
			errorText := strings.Join(errors, " ")
			for _, expected := range tt.expectedErrors {
				if !strings.Contains(errorText, expected) {
					t.Errorf("expected error to contain %q, but errors were: %v", expected, errors)
				}
			}

			// Verify errors have line numbers (format: "line:column: message")
			for _, err := range errors {
				if !strings.Contains(err, ":") {
					t.Errorf("expected error to contain line number (format 'line:column: message'), got: %q", err)
				}
			}

			t.Logf("%s: Errors: %v", tt.description, errors)
		})
	}
}

func TestErrorPositionAccuracy(t *testing.T) {
	input := `var x: int
var y int
var z: int`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseFile()

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected at least one error")
	}

	// The error should be on line 2 (where "var y int" is missing the colon)
	foundLine2 := false
	for _, err := range errors {
		if strings.Contains(err, "2:") {
			foundLine2 = true
			break
		}
	}

	if !foundLine2 {
		t.Errorf("expected error on line 2, but errors were: %v", errors)
	}
}

