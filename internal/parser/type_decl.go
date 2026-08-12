package parser

import (
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// parseTypeAliasDecl parses a type alias declaration
// Format: description "text" type Name = Type
func (p *Parser) parseTypeAliasDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "type" keyword
	if !p.curTokenIs(lexer.TYPE) {
		p.addErrorf("expected type, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "type"

	// Parse type name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after type, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse "="
	if !p.curTokenIs(lexer.ASSIGN) {
		p.addErrorf("expected = after type name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "="

	// Parse the type being aliased
	aliasType := p.parseType()
	if aliasType == nil {
		return nil
	}
	// parseType() already advanced past the type

	return &ast.TypeAliasDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Type:        aliasType,
	}
}

