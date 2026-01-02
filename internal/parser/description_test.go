package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
)

func TestParseDescription(t *testing.T) {
	input := `description "This is a test description"`

	l := lexer.New(input)
	p := New(l)
	desc, found := p.parseDescription()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	if !found {
		t.Error("expected description to be found")
	}

	if desc != "This is a test description" {
		t.Errorf("description not 'This is a test description'. got=%s", desc)
	}
}

func TestParseDescriptionMissing(t *testing.T) {
	input := `var counter: int`

	l := lexer.New(input)
	p := New(l)
	desc, found := p.parseDescription()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	if found {
		t.Error("expected description not to be found")
	}

	if desc != "" {
		t.Errorf("description should be empty. got=%s", desc)
	}
}

func TestParseDescriptionWithVariable(t *testing.T) {
	input := `description "Tracks counter value"
var counter: int`

	l := lexer.New(input)
	p := New(l)
	desc, found := p.parseDescription()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	if !found {
		t.Error("expected description to be found")
	}

	if desc != "Tracks counter value" {
		t.Errorf("description not 'Tracks counter value'. got=%s", desc)
	}

	// After parsing description, curToken should be on "var"
	if !p.curTokenIs(lexer.VAR) {
		t.Errorf("expected curToken to be VAR after description, got %s", p.curToken.Type)
	}
}

func TestParseDescriptionWithConstant(t *testing.T) {
	input := `description "Maximum value"
const MAX: int = 100`

	l := lexer.New(input)
	p := New(l)
	desc, found := p.parseDescription()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	if !found {
		t.Error("expected description to be found")
	}

	if desc != "Maximum value" {
		t.Errorf("description not 'Maximum value'. got=%s", desc)
	}

	// After parsing description, curToken should be on "const"
	if !p.curTokenIs(lexer.CONST) {
		t.Errorf("expected curToken to be CONST after description, got %s", p.curToken.Type)
	}
}

func TestParseDescriptionError(t *testing.T) {
	input := `description 123`

	l := lexer.New(input)
	p := New(l)
	desc, found := p.parseDescription()

	if len(p.Errors()) == 0 {
		t.Error("expected parser error for invalid description")
	}

	if found {
		t.Error("expected description not to be found due to error")
	}

	if desc != "" {
		t.Errorf("description should be empty on error. got=%s", desc)
	}
}

func TestMultipleDescriptions(t *testing.T) {
	input := `description "First description"
description "Second description"
var counter: int`

	l := lexer.New(input)
	p := New(l)
	
	desc1, found1 := p.parseDescription()
	if !found1 {
		t.Error("expected first description to be found")
	}
	if desc1 != "First description" {
		t.Errorf("first description not 'First description'. got=%s", desc1)
	}

	desc2, found2 := p.parseDescription()
	if !found2 {
		t.Error("expected second description to be found")
	}
	if desc2 != "Second description" {
		t.Errorf("second description not 'Second description'. got=%s", desc2)
	}

	// After parsing both descriptions, curToken should be on "var"
	if !p.curTokenIs(lexer.VAR) {
		t.Errorf("expected curToken to be VAR after descriptions, got %s", p.curToken.Type)
	}
}

