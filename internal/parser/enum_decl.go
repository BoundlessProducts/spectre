package parser

import (
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// parseEnumDecl parses an enum type declaration
// Format: description "text" enum Name { Value1, Value2, ... }
func (p *Parser) parseEnumDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "enum" keyword - curToken should already be ENUM (since that's what triggered this case)
	if !p.curTokenIs(lexer.ENUM) {
		p.addErrorf("expected enum, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "enum"

	// Parse enum name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after enum, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse "{" - curToken should now be LBRACE
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { after enum name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"
	var values []string

	// Parse enum values
	if !p.curTokenIs(lexer.RBRACE) {
		if !p.curTokenIs(lexer.IDENT) {
			p.addErrorf("expected identifier in enum, got %s", p.curToken.Type)
			return nil
		}
		values = append(values, p.curToken.Literal)
		p.nextToken()

		for p.curTokenIs(lexer.COMMA) {
			p.nextToken()
			if !p.curTokenIs(lexer.IDENT) {
				p.addErrorf("expected identifier in enum, got %s", p.curToken.Type)
				return nil
			}
			values = append(values, p.curToken.Literal)
			p.nextToken()
		}
	}

	// curToken should be on "}" now
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected }, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.EnumDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Values:      values,
	}
}

