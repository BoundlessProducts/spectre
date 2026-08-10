package parser

import (
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// parseVariableDecl parses a variable declaration
// Format: description "text" var name: Type
func (p *Parser) parseVariableDecl() ast.Decl {
	pos := tokenPos(p.curToken)
	
	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "var" keyword
	if !p.curTokenIs(lexer.VAR) {
		p.addErrorf("expected var, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "var"

	// Parse variable name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after var, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse ":"
	if !p.curTokenIs(lexer.COLON) {
		p.addErrorf("expected : after variable name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ":"

	// Parse type
	varType := p.parseType()
	if varType == nil {
		return nil
	}
	// parseType() already advanced past the type

	return &ast.VariableDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Type:        varType,
	}
}

// parseConstantDecl parses a constant declaration
// Format: description "text" const name: Type = value
func (p *Parser) parseConstantDecl() ast.Decl {
	pos := tokenPos(p.curToken)
	
	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "const" keyword
	if !p.curTokenIs(lexer.CONST) {
		p.addErrorf("expected const, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "const"

	// Parse constant name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after const, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse ":"
	if !p.curTokenIs(lexer.COLON) {
		p.addErrorf("expected : after constant name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ":"

	// Parse type
	constType := p.parseType()
	if constType == nil {
		return nil
	}
	// parseType() already advanced past the type

	// Parse "="
	if !p.curTokenIs(lexer.ASSIGN) {
		p.addErrorf("expected = after constant type, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "="

	// Parse value expression
	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}
	// parseExpression() already advanced past the expression
	// But we need to ensure we're past any whitespace/newlines
	
	return &ast.ConstantDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Type:        constType,
		Value:       value,
	}
}

// parseParamDecl parses a specification parameter declaration.
// Format: param <name>: <type>
// Parameters are read-only integer variables bound via --param Name=Value at run time.
func (p *Parser) parseParamDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	if !p.curTokenIs(lexer.PARAM) {
		p.addErrorf("expected param, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "param"

	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after param, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	if !p.curTokenIs(lexer.COLON) {
		p.addErrorf("expected : after param name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ":"

	paramType := p.parseType()
	if paramType == nil {
		return nil
	}

	return &ast.ParamDecl{
		Position: pos,
		Name:     name,
		Type:     paramType,
	}
}

