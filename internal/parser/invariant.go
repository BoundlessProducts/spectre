package parser

import (
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// parseInvariantDecl parses an invariant declaration
// Format: description "text" invariant name { condition }
func (p *Parser) parseInvariantDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "invariant" keyword
	if !p.curTokenIs(lexer.INVARIANT) {
		p.addErrorf("expected invariant, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "invariant"

	// Parse invariant name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after invariant, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse condition block
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { for invariant condition, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	// Parse the condition expression
	// Use parseExpressionUntil to parse until RBRACE, allowing comparison operators
	condition := p.parseExpressionUntil(lexer.RBRACE)
	if condition == nil {
		return nil
	}
	// parseExpressionUntil already advanced past the expression and stopped at RBRACE

	// Parse closing brace
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close invariant condition, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.InvariantDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Condition:   condition,
	}
}

