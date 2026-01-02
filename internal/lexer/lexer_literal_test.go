package lexer

import "testing"

func TestNextTokenInteger(t *testing.T) {
	input := `0 123 456789`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{INT, "0"},
		{INT, "123"},
		{INT, "456789"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenFloat(t *testing.T) {
	input := `0.0 123.456 0.5 10.0`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{FLOAT, "0.0"},
		{FLOAT, "123.456"},
		{FLOAT, "0.5"},
		{FLOAT, "10.0"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenString(t *testing.T) {
	input := `"hello" "world" "test string" ""`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{STRING, "hello"},
		{STRING, "world"},
		{STRING, "test string"},
		{STRING, ""},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenBoolean(t *testing.T) {
	input := `true false`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{BOOL, "true"},
		{BOOL, "false"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenMixed(t *testing.T) {
	input := `var counter = 42 const name = "test"`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{IDENT, "counter"},
		{ASSIGN, "="},
		{INT, "42"},
		{CONST, "const"},
		{IDENT, "name"},
		{ASSIGN, "="},
		{STRING, "test"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

