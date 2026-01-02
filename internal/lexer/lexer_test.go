package lexer

import "testing"

func TestNextTokenBasic(t *testing.T) {
	input := `var counter action increment`

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

func TestNextTokenOperators(t *testing.T) {
	input := `= == != < > <= >= + - * / && || ! ' ->`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{ASSIGN, "="},
		{EQ, "=="},
		{NEQ, "!="},
		{LT, "<"},
		{GT, ">"},
		{LEQ, "<="},
		{GEQ, ">="},
		{PLUS, "+"},
		{MINUS, "-"},
		{ASTERISK, "*"},
		{SLASH, "/"},
		{AND, "&&"},
		{OR, "||"},
		{NOT, "!"},
		{PRIME, "'"},
		{ARROW, "->"},
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

func TestNextTokenDelimiters(t *testing.T) {
	input := `, . ; : ( ) { } [ ] ...`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{COMMA, ","},
		{DOT, "."},
		{SEMICOLON, ";"},
		{COLON, ":"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{ELLIPSIS, "..."},
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

func TestNextTokenKeywords(t *testing.T) {
	input := `var const action fun invariant temporal module import extends`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{CONST, "const"},
		{ACTION, "action"},
		{FUN, "fun"},
		{INVARIANT, "invariant"},
		{TEMPORAL, "temporal"},
		{MODULE, "module"},
		{IMPORT, "import"},
		{EXTENDS, "extends"},
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

func TestNextTokenIdentifiers(t *testing.T) {
	input := `counter myVariable user123 _private`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENT, "counter"},
		{IDENT, "myVariable"},
		{IDENT, "user123"},
		{IDENT, "_private"},
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

func TestNextTokenWhitespace(t *testing.T) {
	input := `var   counter
	
	action`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{IDENT, "counter"},
		{ACTION, "action"},
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

