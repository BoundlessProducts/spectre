package lexer

import "testing"

func TestNextTokenUnterminatedString(t *testing.T) {
	input := `var name = "unterminated`

	l := New(input)
	
	tokens := []Token{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF || tok.Type == ILLEGAL {
			break
		}
	}

	// Should have tokens before the error
	if len(tokens) < 3 {
		t.Fatal("Expected at least some tokens before error")
	}

	// Last token should be ILLEGAL or we should handle the unterminated string
	lastToken := tokens[len(tokens)-1]
	if lastToken.Type != EOF && lastToken.Type != ILLEGAL {
		t.Errorf("Expected ILLEGAL or EOF for unterminated string, got %v", lastToken.Type)
	}
}

func TestNextTokenUnterminatedMultiLineComment(t *testing.T) {
	input := `var counter /* unterminated comment`

	l := New(input)
	
	tokens := []Token{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF || tok.Type == ILLEGAL {
			break
		}
		if len(tokens) > 10 {
			t.Fatal("Too many tokens, possible infinite loop")
			break
		}
	}

	// Should detect the error
	foundError := false
	for _, tok := range tokens {
		if tok.Type == ILLEGAL {
			foundError = true
			break
		}
	}
	
	if !foundError {
		t.Error("Expected ILLEGAL token for unterminated multi-line comment")
	}
}

func TestNextTokenIllegalCharacter(t *testing.T) {
	input := `var counter @ invalid`

	l := New(input)
	
	tokens := []Token{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}

	// Should have ILLEGAL token for '@'
	foundIllegal := false
	for _, tok := range tokens {
		if tok.Type == ILLEGAL && tok.Literal == "@" {
			foundIllegal = true
			break
		}
	}
	
	if !foundIllegal {
		t.Error("Expected ILLEGAL token for '@' character")
	}
}

