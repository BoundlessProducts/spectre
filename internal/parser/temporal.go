package parser

import (
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// parseTemporalDecl parses a temporal property declaration
// Format: description "text" temporal name { expression }
func (p *Parser) parseTemporalDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "temporal" keyword
	if !p.curTokenIs(lexer.TEMPORAL) {
		p.addErrorf("expected temporal, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "temporal"

	// Parse temporal property name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after temporal, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Parse expression block
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { for temporal expression, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	// Parse the temporal expression
	expression := p.parseTemporalExpression()
	if expression == nil {
		return nil
	}
	// parseTemporalExpression already advanced past the expression

	// Parse closing brace
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close temporal expression, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.TemporalDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Expression:  expression,
	}
}

// parseTemporalExpression parses a temporal expression
// Supports: always, eventually, until, WF, SF, and combinations
func (p *Parser) parseTemporalExpression() ast.Expr {
	// Check for always or eventually operators
	if p.curTokenIs(lexer.ALWAYS) {
		return p.parseAlwaysExpression()
	}
	if p.curTokenIs(lexer.EVENTUALLY) {
		return p.parseEventuallyExpression()
	}

	// Check for fairness operators
	var left ast.Expr
	if p.curTokenIs(lexer.WF) {
		left = p.parseWFExpression()
		if left == nil {
			return nil
		}
	} else if p.curTokenIs(lexer.SF) {
		left = p.parseSFExpression()
		if left == nil {
			return nil
		}
	} else {
		// Parse left-hand side expression (for until or leads-to)
		// We need to parse until we hit RBRACE, UNTIL, or ARROW
		left = p.parseExpressionUntilTemporal()
		if left == nil {
			return nil
		}
	}

	// Check for until operator
	if p.curTokenIs(lexer.UNTIL) {
		return p.parseUntilExpression(left)
	}

	// Check for leads-to operator (→)
	if p.curTokenIs(lexer.ARROW) {
		return p.parseLeadsToExpression(left)
	}

	// Check for logical AND operator (for expressions like WF(...) && SF(...))
	if p.curTokenIs(lexer.AND) {
		return p.parseBinaryTemporalExpression(left)
	}

	// If no temporal operator found, return the expression as-is
	// (This handles cases where the expression is just a regular expression)
	return left
}

// parseBinaryTemporalExpression parses a binary temporal expression with AND/OR
func (p *Parser) parseBinaryTemporalExpression(left ast.Expr) ast.Expr {
	pos := tokenPos(p.curToken)
	operator := p.curToken.Type
	p.nextToken() // consume AND or OR

	// Parse right-hand side using temporal expression parser
	right := p.parseTemporalExpression()
	if right == nil {
		return nil
	}

	// Map token type to BinaryOp
	var op ast.BinaryOp
	switch operator {
	case lexer.AND:
		op = ast.And
	case lexer.OR:
		op = ast.Or
	default:
		p.addErrorf("unsupported binary operator in temporal expression: %s", operator)
		return nil
	}

	// Create a binary expression
	return &ast.BinaryExpr{
		Position: pos,
		Left:     left,
		Op:        op,
		Right:     right,
	}
}

// parseExpressionUntilTemporal parses an expression until it hits a temporal operator or RBRACE
func (p *Parser) parseExpressionUntilTemporal() ast.Expr {
	// Check for EOF before parsing (but not RBRACE - we want to parse expressions that end with RBRACE)
	if p.curTokenIs(lexer.EOF) {
		p.addErrorf("unexpected EOF in temporal expression")
		return nil
	}

	left := p.parsePrefixExpression()
	if left == nil {
		p.addErrorf("failed to parse prefix expression in temporal context, got token %s", p.curToken.Type)
		return nil
	}

	// parsePrefixExpression already advanced, so curToken is now on the potential operator
	// Stop parsing if we hit EOF, SEMICOLON, LBRACE, RBRACE, UNTIL, or ARROW
	if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.LBRACE) || 
		p.curTokenIs(lexer.RBRACE) || p.curTokenIs(lexer.UNTIL) || p.curTokenIs(lexer.ARROW) {
		return left
	}

	// Continue parsing infix expressions while precedence allows
	for LOWEST < p.curPrecedence() {
		// Stop if we hit a temporal operator or RBRACE
		if p.curTokenIs(lexer.RBRACE) || p.curTokenIs(lexer.UNTIL) || p.curTokenIs(lexer.ARROW) {
			break
		}
		
		infix := p.parseInfixExpression(left)
		if infix == nil {
			return left
		}
		left = infix
		
		// Stop if we hit EOF, SEMICOLON, LBRACE, RBRACE, UNTIL, or ARROW
		if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.LBRACE) || 
			p.curTokenIs(lexer.RBRACE) || p.curTokenIs(lexer.UNTIL) || p.curTokenIs(lexer.ARROW) {
			break
		}
	}

	return left
}

// parseAlwaysExpression parses an always expression
// Format: always (expression) or always eventually (expression) for nested expressions
func (p *Parser) parseAlwaysExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "always"

	var expr ast.Expr

	// Check if next token is a temporal operator (for nested expressions)
	if p.curTokenIs(lexer.EVENTUALLY) || p.curTokenIs(lexer.ALWAYS) {
		// Parse nested temporal expression without parentheses
		expr = p.parseTemporalExpression()
		if expr == nil {
			return nil
		}
	} else if p.curTokenIs(lexer.LPAREN) {
		// Parse parenthesized expression
		p.nextToken() // consume "("

		expr = p.parseTemporalExpression()
		if expr == nil {
			return nil
		}

		// Parse closing parenthesis
		if !p.curTokenIs(lexer.RPAREN) {
			p.addErrorf("expected ) after always expression, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume ")"
	} else {
		// Parse expression without parentheses (for cases like "always users.size() > 0")
		expr = p.parseTemporalExpression()
		if expr == nil {
			return nil
		}
	}

	return &ast.AlwaysExpr{
		Position: pos,
		Expr:     expr,
	}
}

// parseEventuallyExpression parses an eventually expression
// Format: eventually (expression) or eventually always (expression) for nested expressions
// Also supports: eventually expression (without parentheses when not nested)
func (p *Parser) parseEventuallyExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "eventually"

	var expr ast.Expr

	// Check if next token is a temporal operator (for nested expressions)
	if p.curTokenIs(lexer.EVENTUALLY) || p.curTokenIs(lexer.ALWAYS) {
		// Parse nested temporal expression without parentheses
		expr = p.parseTemporalExpression()
		if expr == nil {
			return nil
		}
	} else if p.curTokenIs(lexer.LPAREN) {
		// Parse parenthesized expression
		p.nextToken() // consume "("

		expr = p.parseTemporalExpression()
		if expr == nil {
			return nil
		}

		// Parse closing parenthesis
		if !p.curTokenIs(lexer.RPAREN) {
			p.addErrorf("expected ) after eventually expression, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume ")"
	} else {
		// Parse expression without parentheses (for cases like "eventually counter = 10")
		expr = p.parseTemporalExpression()
		if expr == nil {
			return nil
		}
	}

	return &ast.EventuallyExpr{
		Position: pos,
		Expr:     expr,
	}
}

// parseUntilExpression parses an until expression
// Format: leftExpr until rightExpr
func (p *Parser) parseUntilExpression(left ast.Expr) ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "until"

	// Parse right-hand side expression
	right := p.parseExpression(LOWEST)
	if right == nil {
		return nil
	}

	return &ast.UntilExpr{
		Position: pos,
		Left:     left,
		Right:    right,
	}
}

// parseWFExpression parses a weak fairness expression
// Format: WF(actionName) or WF(variableName)
func (p *Parser) parseWFExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "WF"

	// Parse opening parenthesis
	if !p.curTokenIs(lexer.LPAREN) {
		p.addErrorf("expected ( after WF, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "("

	// Parse target (action name or variable name)
	target := p.parseExpression(LOWEST)
	if target == nil {
		return nil
	}

	// Parse closing parenthesis
	if !p.curTokenIs(lexer.RPAREN) {
		p.addErrorf("expected ) after WF target, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ")"

	return &ast.WFExpr{
		Position: pos,
		Target:   target,
	}
}

// parseSFExpression parses a strong fairness expression
// Format: SF(actionName) or SF(variableName)
func (p *Parser) parseSFExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "SF"

	// Parse opening parenthesis
	if !p.curTokenIs(lexer.LPAREN) {
		p.addErrorf("expected ( after SF, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "("

	// Parse target (action name or variable name)
	target := p.parseExpression(LOWEST)
	if target == nil {
		return nil
	}

	// Parse closing parenthesis
	if !p.curTokenIs(lexer.RPAREN) {
		p.addErrorf("expected ) after SF target, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ")"

	return &ast.SFExpr{
		Position: pos,
		Target:   target,
	}
}

// parseLeadsToExpression parses a leads-to expression
// Format: leftExpr → rightExpr
func (p *Parser) parseLeadsToExpression(left ast.Expr) ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "→" or "->"

	// Parse right-hand side expression
	// Use parseTemporalExpression to allow temporal operators like eventually
	right := p.parseTemporalExpression()
	if right == nil {
		p.addErrorf("failed to parse right-hand side of leads-to expression")
		return nil
	}

	return &ast.LeadsToExpr{
		Position: pos,
		Left:     left,
		Right:    right,
	}
}

