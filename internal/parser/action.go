package parser

import (
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// parseActionDecl parses an action declaration
// Format: description "text" action name(param1: Type1, param2: Type2) { body }
func (p *Parser) parseActionDecl() ast.Decl {
	pos := tokenPos(p.curToken)
	
	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "action" keyword
	if !p.curTokenIs(lexer.ACTION) {
		p.addErrorf("expected action, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "action"

	// Parse action name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after action, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse parameters (optional)
	var parameters []ast.Parameter
	if p.curTokenIs(lexer.LPAREN) {
		p.nextToken() // consume "("
		parameters = p.parseParameters()
		if parameters == nil && len(p.Errors()) > 0 {
			return nil
		}
		if !p.curTokenIs(lexer.RPAREN) {
			p.addErrorf("expected ) after parameters, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume ")"
	}

	// Parse optional guard (when clause)
	var guard ast.Expr
	if p.curTokenIs(lexer.WHEN) {
		p.nextToken() // consume "when"
		guard = p.parseExpression(LOWEST)
		if guard == nil {
			return nil
		}
		// parseExpression already advanced past the guard expression
	}

	// Parse action body
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { after action signature, got %s", p.curToken.Type)
		return nil
	}
	// parseBlock expects curToken to be on LBRACE and will consume it
	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.ActionDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Parameters:  parameters,
		Guard:       guard,
		Body:        body,
	}
}

