package lexer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLexerWithExampleFiles(t *testing.T) {
	exampleDir := "../../examples"
	
	// Get all .spec files
	specFiles, err := filepath.Glob(filepath.Join(exampleDir, "*.spec"))
	if err != nil {
		t.Fatalf("Failed to find example files: %v", err)
	}

	if len(specFiles) == 0 {
		t.Fatal("No example .spec files found")
	}

	for _, specFile := range specFiles {
		t.Run(filepath.Base(specFile), func(t *testing.T) {
			content, err := os.ReadFile(specFile)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", specFile, err)
			}

			l := New(string(content))
			
			tokenCount := 0
			illegalCount := 0
			maxTokens := 10000 // Safety limit
			
			for {
				tok := l.NextToken()
				tokenCount++
				
				if tok.Type == EOF {
					break
				}
				
				if tok.Type == ILLEGAL {
					illegalCount++
					// Don't fail immediately, collect all errors
					t.Logf("Illegal token at line %d, col %d: %q", 
						tok.Position.Line, tok.Position.Column, tok.Literal)
				}
				
				if tokenCount > maxTokens {
					t.Fatalf("Too many tokens, possible infinite loop")
				}
			}
			
			// Report results
			t.Logf("Total tokens: %d, Illegal tokens: %d", tokenCount, illegalCount)
			
			// Fail if we found illegal tokens
			if illegalCount > 0 {
				t.Errorf("Found %d illegal tokens in %s", illegalCount, specFile)
			}
		})
	}
}

func TestLexerPositionTracking(t *testing.T) {
	input := `var counter
action increment`

	l := New(input)
	
	tokens := []Token{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}

	// Verify positions are tracked
	if len(tokens) < 3 {
		t.Fatal("Expected at least 3 tokens")
	}

	// First token should be at line 1
	if tokens[0].Position.Line != 1 {
		t.Errorf("Expected first token at line 1, got %d", tokens[0].Position.Line)
	}

	// Should have tokens on different lines
	foundLine2 := false
	for _, tok := range tokens {
		if tok.Position.Line == 2 {
			foundLine2 = true
			break
		}
	}
	
	if !foundLine2 {
		t.Error("Expected tokens on line 2")
	}
}

func TestLexerCompleteExample(t *testing.T) {
	// Test with a complete simple example
	input := `// Simple Counter Example
description "Tracks a numeric counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}`

	l := New(input)
	
	expectedTokens := []TokenType{
		DESCRIPTION, STRING, VAR, IDENT, COLON, IDENT,
		DESCRIPTION, STRING, INIT, LBRACE,
		IDENT, ASSIGN, INT, RBRACE,
		DESCRIPTION, STRING, ACTION, IDENT, LBRACE,
		IDENT, PRIME, ASSIGN, IDENT, PLUS, INT, RBRACE,
		EOF,
	}

	tokenIndex := 0
	for {
		tok := l.NextToken()
		
		if tokenIndex >= len(expectedTokens) {
			t.Fatalf("More tokens than expected. Got %v at position %d", tok.Type, tokenIndex)
		}
		
		if tok.Type != expectedTokens[tokenIndex] {
			t.Errorf("Token %d: expected %v, got %v (literal: %q, line: %d)",
				tokenIndex, expectedTokens[tokenIndex], tok.Type, tok.Literal, tok.Position.Line)
		}
		
		if tok.Type == EOF {
			break
		}
		
		tokenIndex++
	}
	
	if tokenIndex != len(expectedTokens)-1 {
		t.Errorf("Expected %d tokens, got %d", len(expectedTokens)-1, tokenIndex)
	}
}

