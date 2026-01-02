package lexer

import "testing"

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"var", VAR},
		{"const", CONST},
		{"action", ACTION},
		{"fun", FUN},
		{"invariant", INVARIANT},
		{"temporal", TEMPORAL},
		{"module", MODULE},
		{"import", IMPORT},
		{"extends", EXTENDS},
		{"public", PUBLIC},
		{"private", PRIVATE},
		{"description", DESCRIPTION},
		{"init", INIT},
		{"oneOf", ONEOF},
		{"require", REQUIRE},
		{"ensure", ENSURE},
		{"if", IF},
		{"else", ELSE},
		{"then", THEN},
		{"let", LET},
		{"return", RETURN},
		{"type", TYPE},
		{"enum", ENUM},
		{"when", WHEN},
		{"super", SUPER},
		{"with", WITH},
		{"always", ALWAYS},
		{"eventually", EVENTUALLY},
		{"until", UNTIL},
		{"WF", WF},
		{"SF", SF},
		{"next", NEXT},
		{"Set", SET},
		{"Map", MAP},
		{"List", LIST},
		{"Option", OPTION},
		{"true", BOOL},
		{"false", BOOL},
		{"counter", IDENT},
		{"myVariable", IDENT},
		{"unknown", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := LookupIdent(tt.input)
			if got != tt.expected {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTokenTypes(t *testing.T) {
	// Verify all token types are defined
	expectedTypes := []TokenType{
		ILLEGAL, EOF, IDENT, INT, FLOAT, STRING, BOOL,
		ASSIGN, EQ, NEQ, LT, GT, LEQ, GEQ, PLUS, MINUS, ASTERISK, SLASH,
		AND, OR, NOT, PRIME, ARROW,
		COMMA, DOT, SEMICOLON, COLON, LPAREN, RPAREN, LBRACE, RBRACE,
		LBRACKET, RBRACKET, ELLIPSIS,
		MODULE, IMPORT, EXTENDS, PUBLIC, PRIVATE, CONST, VAR, DESCRIPTION,
		INIT, ONEOF, ACTION, FUN, INVARIANT, TEMPORAL, REQUIRE, ENSURE,
		IF, ELSE, THEN, LET, RETURN, TYPE, ENUM, WHEN, SUPER, WITH,
		ALWAYS, EVENTUALLY, UNTIL, WF, SF, NEXT,
		SET, MAP, LIST, OPTION,
	}

	for _, tokType := range expectedTypes {
		if tokType == "" {
			t.Errorf("Token type %v is empty", tokType)
		}
	}
}

