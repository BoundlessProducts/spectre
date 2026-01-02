package parser

import (
	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

// parseFunctionDecl parses a pure function declaration
// Format: description "text" fun name(param1: Type1, param2: Type2): ReturnType { body }
func (p *Parser) parseFunctionDecl() ast.Decl {
	pos := tokenPos(p.curToken)
	
	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "fun" keyword
	if !p.curTokenIs(lexer.FUN) {
		p.addErrorf("expected fun, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "fun"

	// Parse function name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after fun, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse parameters
	if !p.curTokenIs(lexer.LPAREN) {
		p.addErrorf("expected ( after function name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "("

	parameters := p.parseParameters()
	if parameters == nil && len(p.Errors()) > 0 {
		return nil
	}

	if !p.curTokenIs(lexer.RPAREN) {
		p.addErrorf("expected ) after parameters, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ")"

	// Parse return type
	var returnType ast.Type
	if p.curTokenIs(lexer.COLON) {
		p.nextToken() // consume ":"
		returnType = p.parseType()
		if returnType == nil {
			return nil
		}
		// parseType() already advanced past the type
	}

	// Parse function body
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { after function signature, got %s", p.curToken.Type)
		return nil
	}
	// parseBlock expects curToken to be on LBRACE and will consume it
	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.FunctionDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Parameters:  parameters,
		ReturnType:  returnType,
		Body:        body,
	}
}

// parseParameters parses a parameter list
func (p *Parser) parseParameters() []ast.Parameter {
	var params []ast.Parameter

	// Empty parameter list
	if p.curTokenIs(lexer.RPAREN) {
		return params
	}

	// Parse first parameter
	param := p.parseParameter()
	if param == nil {
		return nil
	}
	params = append(params, *param)

	// Parse remaining parameters
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken() // consume ","
		param := p.parseParameter()
		if param == nil {
			return nil
		}
		params = append(params, *param)
	}

	return params
}

// parseParameter parses a single parameter (name: Type)
func (p *Parser) parseParameter() *ast.Parameter {
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected parameter name, got %s", p.curToken.Type)
		return nil
	}

	paramPos := tokenPos(p.curToken)
	name := p.curToken.Literal
	p.nextToken() // consume parameter name

	if !p.curTokenIs(lexer.COLON) {
		p.addErrorf("expected : after parameter name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ":"

	paramType := p.parseType()
	if paramType == nil {
		return nil
	}
	// parseType() already advanced past the type

	return &ast.Parameter{
		Position: paramPos,
		Name:     name,
		Type:     paramType,
	}
}

// parseBlock parses a block of statements
func (p *Parser) parseBlock() *ast.BlockStmt {
	pos := tokenPos(p.curToken)
	
	// curToken should be on LBRACE
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected {, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"
	
	var statements []ast.Stmt

	// Parse statements until closing brace
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt == nil {
			// Try to recover by advancing
			if p.curTokenIs(lexer.RBRACE) {
				break
			}
			p.nextToken()
			continue
		}
		statements = append(statements, stmt)
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close block, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.BlockStmt{
		Position:   pos,
		Statements: statements,
	}
}

// parseStatement parses a statement
func (p *Parser) parseStatement() ast.Stmt {
	switch p.curToken.Type {
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.REQUIRE:
		return p.parseRequireStatement()
	case lexer.ENSURE:
		return p.parseEnsureStatement()
	case lexer.IDENT, lexer.SET, lexer.MAP, lexer.LIST, lexer.OPTION:
		// Check if it's an assignment (identifier ['] = expression)
		// We need to peek ahead to see if next token is PRIME or ASSIGN
		if p.peekTokenIs(lexer.PRIME) {
			// identifier' = expression
			// In Spectre, identifier' is almost always followed by = (assignment)
			// So we can safely assume it's an assignment without peeking further
			// parseAssignStatement will call parseExpression which will consume IDENT and PRIME
			return p.parseAssignStatement()
		} else if p.peekTokenIs(lexer.ASSIGN) {
			// identifier = expression
			return p.parseAssignStatement()
		}
		// Otherwise, parse as expression statement
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return nil
		}
		return &ast.ExprStmt{
			Position: tokenPos(p.curToken),
			Expr:     expr,
		}
	default:
		// Try to parse as expression statement
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return nil
		}
		return &ast.ExprStmt{
			Position: tokenPos(p.curToken),
			Expr:     expr,
		}
	}
}

// parseReturnStatement parses a return statement
func (p *Parser) parseReturnStatement() ast.Stmt {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "return"

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}
	// parseExpression() already advanced past the expression

	return &ast.ReturnStmt{
		Position: pos,
		Value:    value,
	}
}

// parseRequireStatement parses a require (precondition) statement
func (p *Parser) parseRequireStatement() ast.Stmt {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "require"

	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return nil
	}
	// parseExpression() already advanced past the expression

	return &ast.RequireStmt{
		Position:  pos,
		Condition: condition,
	}
}

// parseEnsureStatement parses an ensure (postcondition) statement
func (p *Parser) parseEnsureStatement() ast.Stmt {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "ensure"

	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return nil
	}
	// parseExpression() already advanced past the expression

	return &ast.EnsureStmt{
		Position:  pos,
		Condition: condition,
	}
}

