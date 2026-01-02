package parser

import "github.com/akkeshavan/spectre/internal/lexer"

// parseDescription parses an optional description field
// Format: description "text"
// Returns the description string and whether a description was found
func (p *Parser) parseDescription() (string, bool) {
	if !p.curTokenIs(lexer.DESCRIPTION) {
		return "", false
	}

	p.nextToken() // consume "description"

	if !p.curTokenIs(lexer.STRING) {
		p.addErrorf("expected string after description, got %s", p.curToken.Type)
		return "", false
	}

	description := p.curToken.Literal
	p.nextToken() // consume string

	return description, true
}

