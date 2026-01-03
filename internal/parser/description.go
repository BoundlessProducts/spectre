package parser

import "github.com/akkeshavan/spectre/internal/lexer"

// parseDescription parses an optional description field
// Format: description "text"
// Returns the description string and whether a description was found
// If a pending description was set (from a description on a previous line),
// it will be used instead of parsing from the current token.
func (p *Parser) parseDescription() (string, bool) {
	// Check if there's a pending description (from a description on previous line)
	if p.pendingDescription != "" {
		desc := p.pendingDescription
		p.pendingDescription = "" // Clear after use
		return desc, true
	}
	
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

