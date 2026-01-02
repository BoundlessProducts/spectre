package lexer

import "testing"

func TestNextTokenSingleLineComment(t *testing.T) {
	input := `var counter // This is a comment
action increment`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{IDENT, "counter"},
		{ACTION, "action"},
		{IDENT, "increment"},
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

func TestNextTokenMultiLineComment(t *testing.T) {
	input := `var counter /* This is a
multi-line comment */ action increment`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{IDENT, "counter"},
		{ACTION, "action"},
		{IDENT, "increment"},
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

func TestNextTokenCommentAtEnd(t *testing.T) {
	input := `var counter // comment at end`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{IDENT, "counter"},
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

func TestNextTokenMultipleComments(t *testing.T) {
	input := `// First comment
var counter // Second comment
/* Third comment */
action increment`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{IDENT, "counter"},
		{ACTION, "action"},
		{IDENT, "increment"},
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

