package parser

import (
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	LOGICAL     // && || (lowest precedence for logical operators)
	EQUALS      // == != = (equality/comparison)
	LESSGREATER // < > <= >=
	SUM         // + -
	PRODUCT     // * /
	PREFIX      // ! -x
	CALL        // function calls
	INDEX       // array[index]
)

// Precedence map for binary operators
var precedences = map[lexer.TokenType]int{
	lexer.EQ:     EQUALS,
	lexer.ASSIGN: EQUALS, // = is used for equality in expressions (e.g., ensure counter' = counter + 1)
	lexer.NEQ:    EQUALS,
	lexer.LT:     LESSGREATER,
	lexer.GT:     LESSGREATER,
	lexer.LEQ:    LESSGREATER,
	lexer.GEQ:    LESSGREATER,
	lexer.PLUS:   SUM,
	lexer.MINUS:  SUM,
	lexer.ASTERISK: PRODUCT,
	lexer.SLASH:  PRODUCT,
	lexer.AND:    LOGICAL, // Logical AND has lower precedence than equality
	lexer.OR:     LOGICAL, // Logical OR has lower precedence than equality
	lexer.LPAREN: CALL,
	lexer.LBRACKET: INDEX,
	lexer.DOT:    CALL, // Method calls have same precedence as function calls
}

// parseExpression parses an expression with the given precedence
func (p *Parser) parseExpression(precedence int) ast.Expr {
	left := p.parsePrefixExpression()

	// parsePrefixExpression already advanced, so curToken is now on the potential operator
	// Stop parsing if we hit EOF, SEMICOLON, or LBRACE (start of block)
	// Note: ASSIGN is now treated as an infix operator (equality) in expressions
	// Note: RBRACE is checked after parsing infix expressions to allow expressions like "users.size() > 0"
	if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.LBRACE) {
		return left
	}

	// Continue parsing infix expressions while precedence allows
	for {
		// Check if current token is an infix operator
		curPrec := p.curPrecedence()
		if curPrec <= precedence {
			// Current operator has lower or equal precedence, stop
			break
		}
		
		// Stop if we hit EOF, SEMICOLON, LBRACE, or RBRACE
		if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.LBRACE) || p.curTokenIs(lexer.RBRACE) {
			break
		}
		
		infix := p.parseInfixExpression(left)
		if infix == nil {
			// No infix operator recognized, stop
			break
		}
		left = infix
		
		// Stop if we hit EOF, SEMICOLON, LBRACE, or RBRACE after parsing infix
		if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.LBRACE) || p.curTokenIs(lexer.RBRACE) {
			break
		}
	}

	return left
}

// parseExpressionUntil parses an expression until it encounters a specific token
// This is useful for parsing expressions inside blocks (e.g., invariant conditions)
func (p *Parser) parseExpressionUntil(stopToken lexer.TokenType) ast.Expr {
	left := p.parsePrefixExpression()

	// parsePrefixExpression already advanced, so curToken is now on the potential operator
	// Stop parsing if we hit EOF, SEMICOLON, or LBRACE
	if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.LBRACE) {
		return left
	}
	
	// Check for stop token after parsing prefix
	if p.curTokenIs(stopToken) {
		return left
	}

	// Continue parsing infix expressions while precedence allows
	for LOWEST < p.curPrecedence() {
		// Don't parse infix if we hit the stop token
		if p.curTokenIs(stopToken) {
			break
		}
		
		infix := p.parseInfixExpression(left)
		if infix == nil {
			return left
		}
		left = infix
		
		// Stop if we hit EOF, SEMICOLON, or LBRACE
		if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.LBRACE) {
			break
		}
		
		// Check for stop token after parsing infix
		if p.curTokenIs(stopToken) {
			break
		}
	}

	return left
}

// parsePrefixExpression parses a prefix expression
func (p *Parser) parsePrefixExpression() ast.Expr {
	var expr ast.Expr
	switch p.curToken.Type {
	case lexer.IDENT, lexer.SET, lexer.MAP, lexer.LIST, lexer.OPTION:
		// These keywords can be used as identifiers in expressions
		// Check if this is a lambda: param => expr (single param without parens)
		if p.peekTokenIs(lexer.FATARROW) {
			return p.parseLambdaExpression()
		}
		// parseIdentifier already advances past the identifier (and prime if present)
		return p.parseIdentifier()
	case lexer.INT, lexer.FLOAT, lexer.STRING, lexer.BOOL:
		expr = p.parseLiteral()
		p.nextToken() // Advance past literal
		return expr
	case lexer.NOT, lexer.MINUS:
		return p.parseUnaryExpression()
	case lexer.LPAREN:
		// Could be a grouped expression or lambda with parentheses
		// Try to parse as grouped expression first. If we see => after ), it's a lambda
		return p.parseGroupedOrLambda()
	case lexer.IF:
		return p.parseIfExpression()
	case lexer.LET:
		return p.parseLetExpression()
	case lexer.SUPER:
		return p.parseSuperExpression()
	case lexer.RETURN:
		// Handle return statements in expression context (e.g., in if expressions)
		// Parse return and extract the value
		p.nextToken() // consume "return"
		return p.parseExpression(LOWEST)
	case lexer.LBRACE:
		// Distinguish between record literals and set literals
		// Record literals: { field: value, ... } or { ...spread }
		// Set literals: { value1, value2, ... }
		// Use a try-parse approach: attempt to parse as record literal first
		// If that fails (no colon found), parse as set literal
		return p.parseRecordOrSetLiteral()
	case lexer.LBRACKET:
		// List literal: [ value1, value2, ... ]
		return p.parseListLiteral()
	default:
		p.addErrorf("no prefix parse function for %s found", p.curToken.Type)
		return nil
	}
}

// parseInfixExpression parses an infix expression
func (p *Parser) parseInfixExpression(left ast.Expr) ast.Expr {
	switch p.curToken.Type {
	case lexer.PLUS, lexer.MINUS, lexer.ASTERISK, lexer.SLASH,
		lexer.EQ, lexer.ASSIGN, lexer.NEQ, lexer.LT, lexer.GT, lexer.LEQ, lexer.GEQ,
		lexer.AND, lexer.OR:
		return p.parseBinaryExpression(left)
	case lexer.LPAREN:
		return p.parseCallExpression(left)
	case lexer.LBRACKET:
		return p.parseIndexExpression(left)
	case lexer.DOT:
		return p.parseSelectorExpression(left)
	default:
		return nil
	}
}

// parseIdentifier parses an identifier (including prime notation)
func (p *Parser) parseIdentifier() ast.Expr {
	pos := tokenPos(p.curToken)
	name := p.curToken.Literal
	
	// Check for prime notation (')
	isPrime := false
	if p.peekTokenIs(lexer.PRIME) {
		p.nextToken() // consume identifier
		p.nextToken() // consume "'"
		isPrime = true
	} else {
		p.nextToken() // consume identifier
	}
	
	return &ast.Ident{
		Position: pos,
		Name:     name,
		Prime:    isPrime,
	}
}

// parseLiteral parses a literal
func (p *Parser) parseLiteral() ast.Expr {
	var kind ast.LitKind
	switch p.curToken.Type {
	case lexer.INT:
		kind = ast.IntLit
	case lexer.FLOAT:
		kind = ast.FloatLit
	case lexer.STRING:
		kind = ast.StringLit
	case lexer.BOOL:
		kind = ast.BoolLit
	default:
		p.addErrorf("unexpected literal type: %s", p.curToken.Type)
		return nil
	}

	return &ast.BasicLit{
		Position: tokenPos(p.curToken),
		Kind:     kind,
		Value:    p.curToken.Literal,
	}
}

// parseUnaryExpression parses a unary expression
func (p *Parser) parseUnaryExpression() ast.Expr {
	op := p.curToken.Type
	p.nextToken()

	var unaryOp ast.UnaryOp
	switch op {
	case lexer.NOT:
		unaryOp = ast.Not
	case lexer.MINUS:
		unaryOp = ast.Neg
	default:
		p.addErrorf("unknown unary operator: %s", op)
		return nil
	}

	return &ast.UnaryExpr{
		Position: tokenPos(p.curToken),
		Op:       unaryOp,
		Expr:     p.parseExpression(PREFIX),
	}
}

// parseBinaryExpression parses a binary expression
func (p *Parser) parseBinaryExpression(left ast.Expr) ast.Expr {
	op := p.curToken.Type
	precedence := p.curPrecedence()
	
	var binaryOp ast.BinaryOp
	switch op {
	case lexer.PLUS:
		binaryOp = ast.Add
	case lexer.MINUS:
		binaryOp = ast.Sub
	case lexer.ASTERISK:
		binaryOp = ast.Mul
	case lexer.SLASH:
		binaryOp = ast.Div
	case lexer.EQ:
		binaryOp = ast.Eq
	case lexer.ASSIGN:
		// In expression contexts, = is equality (not assignment)
		binaryOp = ast.Eq
	case lexer.NEQ:
		binaryOp = ast.Neq
	case lexer.LT:
		binaryOp = ast.Lt
	case lexer.GT:
		binaryOp = ast.Gt
	case lexer.LEQ:
		binaryOp = ast.Leq
	case lexer.GEQ:
		binaryOp = ast.Geq
	case lexer.AND:
		binaryOp = ast.And
	case lexer.OR:
		binaryOp = ast.Or
	default:
		p.addErrorf("unknown binary operator: %s", op)
		return nil
	}

	p.nextToken() // Advance past operator
	right := p.parseExpression(precedence)

	return &ast.BinaryExpr{
		Position: tokenPos(p.curToken),
		Op:       binaryOp,
		Left:     left,
		Right:    right,
	}
}

// parseGroupedOrLambda parses either a grouped expression or a lambda
// We parse the parenthesized content and check if => follows to determine which it is
func (p *Parser) parseGroupedOrLambda() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "("

	// Check for empty parens followed by =>
	if p.curTokenIs(lexer.RPAREN) && p.peekTokenIs(lexer.FATARROW) {
		// It's a lambda with no parameters: () => expr
		p.nextToken() // consume ")"
		p.nextToken() // consume "=>"
		body := p.parseExpression(LOWEST)
		return &ast.LambdaExpr{
			Position: pos,
			Params:   []ast.Parameter{},
			Body:     body,
		}
	}

	// Check if it looks like lambda parameters (identifier optionally with type)
	// A lambda must have: (param) => or (param1, param2) => or (param: Type) =>
	// Only try lambda detection when peek is ), :, or , — the only valid lambda parameter separators.
	// Any other peek token (arithmetic/comparison operators) means it's a grouped expression.
	if p.curTokenIs(lexer.IDENT) && (p.peekTokenIs(lexer.RPAREN) || p.peekTokenIs(lexer.COLON) || p.peekTokenIs(lexer.COMMA)) {
		// Could be a lambda - try to parse as lambda parameters
		// We check if we can parse parameters and then see ) followed by =>
		params, isLambda := p.tryParseLambdaParams()
		if isLambda {
			// It's a lambda
			if !p.curTokenIs(lexer.RPAREN) {
				p.addErrorf("expected ) after lambda parameters, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken() // consume ")"
			
			if !p.curTokenIs(lexer.FATARROW) {
				p.addErrorf("expected => after lambda parameters, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken() // consume "=>"
			
			body := p.parseExpression(LOWEST)
			return &ast.LambdaExpr{
				Position: pos,
				Params:   params,
				Body:     body,
			}
		}
		// If tryParseLambdaParams failed, we've consumed some tokens
		// But that's okay - we'll parse as grouped expression from current position
		// The tokens consumed were just identifiers and commas, which are valid in expressions
	}

	// Not a lambda, parse as grouped expression
	expr := p.parseExpression(LOWEST)

	if !p.curTokenIs(lexer.RPAREN) {
		p.addErrorf("expected ), got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ")"

	return &ast.ParenExpr{
		Position: pos,
		X:        expr,
	}
}

// tryParseLambdaParams tries to parse lambda parameters and returns them if successful
// Returns (params, true) if it's lambda parameters, (nil, false) otherwise
// This function consumes tokens, so caller must restore state if it returns false
func (p *Parser) tryParseLambdaParams() ([]ast.Parameter, bool) {
	var params []ast.Parameter
	
	for p.curTokenIs(lexer.IDENT) {
		// Check if next token is = (assignment) or == (equality) - this means it's not a lambda parameter
		if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.EQ) {
			// This is an assignment or equality expression, not a lambda
			return nil, false
		}
		
		paramName := p.curToken.Literal
		p.nextToken() // consume parameter name
		
		// Check for type annotation
		var paramType ast.Type
		if p.curTokenIs(lexer.COLON) {
			p.nextToken() // consume ":"
			paramType = p.parseType()
			if paramType == nil {
				return nil, false
			}
		}
		
		params = append(params, ast.Parameter{
			Name: paramName,
			Type: paramType,
		})
		
		// Check for comma (more params) or closing paren
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
			continue
		}
		
		// Must be at closing paren followed by =>
		if p.curTokenIs(lexer.RPAREN) && p.peekTokenIs(lexer.FATARROW) {
			// It's a lambda!
			return params, true
		}
		
		// Not a lambda (could be grouped expression like (a = b) or (a + b))
		return nil, false
	}
	
	return nil, false
}

// parseGroupedExpression parses a parenthesized expression
func (p *Parser) parseGroupedExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "("

	expr := p.parseExpression(LOWEST)

	// parseExpression already advanced, so curToken should be on ")"
	if !p.curTokenIs(lexer.RPAREN) {
		p.addErrorf("expected ), got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ")"

	return &ast.ParenExpr{
		Position: pos,
		X:        expr,
	}
}

// parseCallExpression parses a function call
func (p *Parser) parseCallExpression(fn ast.Expr) ast.Expr {
	// curToken should be on LPAREN
	if !p.curTokenIs(lexer.LPAREN) {
		p.addErrorf("expected ( for function call, got %s", p.curToken.Type)
		return nil
	}
	
	args := p.parseExpressionList(lexer.RPAREN)
	// parseExpressionList returns nil only on error, empty slice [] for empty lists
	if args == nil {
		return nil
	}

	return &ast.CallExpr{
		Position: tokenPos(p.curToken),
		Fun:      fn,
		Args:     args,
	}
}

// parseIndexExpression parses an index expression
func (p *Parser) parseIndexExpression(left ast.Expr) ast.Expr {
	p.nextToken() // consume '['

	index := p.parseExpressionUntil(lexer.RBRACKET)
	if index == nil {
		return nil
	}
	// parseExpressionUntil stops at RBRACKET, so curToken should be on RBRACKET now

	if !p.curTokenIs(lexer.RBRACKET) {
		return nil
	}
	p.nextToken() // consume ']'

	return &ast.IndexExpr{
		Position: tokenPos(p.curToken),
		X:        left,
		Index:    index,
	}
}

// parseSelectorExpression parses a selector expression (e.g., record.field)
func (p *Parser) parseSelectorExpression(left ast.Expr) ast.Expr {
	// curToken should be on DOT
	if !p.curTokenIs(lexer.DOT) {
		p.addErrorf("expected . for selector, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "."
	
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after ., got %s", p.curToken.Type)
		return nil
	}

	sel := p.curToken.Literal
	p.nextToken() // consume identifier

	return &ast.SelectorExpr{
		Position: tokenPos(p.curToken),
		X:        left,
		Sel:      sel,
	}
}

// parseExpressionList parses a comma-separated list of expressions
func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expr {
	list := []ast.Expr{} // Initialize as empty slice, not nil

	// curToken should be on the opening delimiter (LPAREN or LBRACKET)
	p.nextToken() // consume opening delimiter

	// Empty list
	if p.curTokenIs(end) {
		p.nextToken() // consume closing delimiter
		return list
	}

	// Parse first expression
	list = append(list, p.parseExpression(LOWEST))

	// Parse remaining expressions
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken() // consume ","
		list = append(list, p.parseExpression(LOWEST))
	}

	// curToken should be on closing delimiter after parseExpression
	if !p.curTokenIs(end) {
		p.addErrorf("expected %s, got %s", end, p.curToken.Type)
		return nil
	}
	p.nextToken() // consume closing delimiter

	return list
}

// Helper functions

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func tokenPos(tok lexer.Token) ast.Position {
	return ast.Position{
		Line:   tok.Position.Line,
		Column: tok.Position.Column,
		Offset: tok.Position.Offset,
	}
}

// parseIfExpression parses an if-else expression
// Format: if (condition) { thenExpr } else { elseExpr }
func (p *Parser) parseIfExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "if"

	// Parse condition in parentheses
	if !p.curTokenIs(lexer.LPAREN) {
		p.addErrorf("expected ( after if, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "("

	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return nil
	}

	if !p.curTokenIs(lexer.RPAREN) {
		p.addErrorf("expected ) after condition, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ")"

	// Check for optional "then" keyword
	if p.curTokenIs(lexer.THEN) {
		p.nextToken() // consume "then"
	}

	// Parse then block
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { after if condition, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	// Check if block contains statements (like return) or expressions
	// Peek ahead to see if first token is RETURN
	var thenExpr ast.Expr
	if p.curTokenIs(lexer.RETURN) {
		// Parse as return statement and extract its value
		// We need to manually parse the return statement since we're in expression context
		p.nextToken() // consume "return"
		
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		thenExpr = value
	} else {
		// Parse as expression
		thenExpr = p.parseExpression(LOWEST)
		if thenExpr == nil {
			return nil
		}
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } after then block, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	// Parse else block
	if !p.curTokenIs(lexer.ELSE) {
		p.addErrorf("expected else after then block, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "else"

	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { after else, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	// Check if block contains statements (like return) or expressions
	var elseExpr ast.Expr
	if p.curTokenIs(lexer.RETURN) {
		// Parse as return statement and extract its value
		p.nextToken() // consume "return"
		
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		elseExpr = value
	} else {
		// Parse as expression
		elseExpr = p.parseExpression(LOWEST)
		if elseExpr == nil {
			return nil
		}
	}

	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } after else block, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.IfExpr{
		Position:  pos,
		Condition: condition,
		Then:      thenExpr,
		Else:      elseExpr,
	}
}

// parseLetExpression parses a let binding expression
// Format: let name = value
// parseSuperExpression parses a super call expression
// Format: super.methodName() or super.methodName
// Note: super calls are typically method calls, so we parse as a selector followed by a call
func (p *Parser) parseSuperExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "super"

	// Parse dot
	if !p.curTokenIs(lexer.DOT) {
		p.addErrorf("expected . after super, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "."

	// Parse method name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after super., got %s", p.curToken.Type)
		return nil
	}
	methodName := p.curToken.Literal
	p.nextToken() // consume method name

	// Create SuperExpr
	superExpr := &ast.SuperExpr{
		Position: pos,
		Method:   methodName,
	}

	// Check if this is a call (has parentheses)
	if p.curTokenIs(lexer.LPAREN) {
		// Parse call arguments
		p.nextToken() // consume "("
		var args []ast.Expr
		
		// Check for empty argument list
		if !p.curTokenIs(lexer.RPAREN) {
			// Parse first argument
			arg := p.parseExpression(LOWEST)
			if arg != nil {
				args = append(args, arg)
			}
			
			// Parse remaining arguments
			for p.curTokenIs(lexer.COMMA) {
				p.nextToken() // consume ","
				arg := p.parseExpression(LOWEST)
				if arg != nil {
					args = append(args, arg)
				}
			}
		}
		
		// Parse closing parenthesis
		if !p.curTokenIs(lexer.RPAREN) {
			p.addErrorf("expected ) after super call arguments, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume ")"
		
		return &ast.CallExpr{
			Position: pos,
			Fun:      superExpr,
			Args:     args,
		}
	}

	// If no parentheses, return just the SuperExpr
	// (though in practice super calls usually have parentheses)
	return superExpr
}

func (p *Parser) parseLetExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "let"

	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after let, got %s", p.curToken.Type)
		return nil
	}

	name := p.curToken.Literal
	p.nextToken() // consume identifier

	if !p.curTokenIs(lexer.ASSIGN) {
		p.addErrorf("expected = after let binding name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "="

	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil
	}

	// For now, let expressions don't have an explicit "in" keyword
	// The body is the rest of the expression
	// This is a simplified version - we'll enhance it later if needed
	return &ast.LetExpr{
		Position: pos,
		Name:     name,
		Value:    value,
		Body:     nil, // Will be set by the caller if needed
	}
}

// parseRecordLiteral parses a record literal value
// Format: { field1: value1, field2: value2, ...identifier }
func (p *Parser) parseRecordLiteral() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "{"
	
	var fields []ast.RecordField
	
	// Parse fields until closing brace
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Check for spread operator
		if p.curTokenIs(lexer.ELLIPSIS) {
			p.nextToken() // consume "..."
			
			if !p.curTokenIs(lexer.IDENT) {
				p.addErrorf("expected identifier after ..., got %s", p.curToken.Type)
				return nil
			}
			
			fields = append(fields, ast.RecordField{
				Name:   p.curToken.Literal,
				Value:  nil, // Spread fields don't have values
				Spread: true,
			})
			p.nextToken() // consume identifier
		} else {
			// Regular field: name: value
			if !p.curTokenIs(lexer.IDENT) {
				p.addErrorf("expected identifier or ... in record literal, got %s", p.curToken.Type)
				return nil
			}
			
			fieldName := p.curToken.Literal
			p.nextToken() // consume identifier
			
			if !p.curTokenIs(lexer.COLON) {
				p.addErrorf("expected : after field name, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken() // consume ":"
			
			value := p.parseExpression(LOWEST)
			if value == nil {
				return nil
			}
			
			fields = append(fields, ast.RecordField{
				Name:   fieldName,
				Value:  value,
				Spread: false,
			})
		}
		
		// Check for comma or closing brace
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
		} else if !p.curTokenIs(lexer.RBRACE) {
			p.addErrorf("expected , or } after field, got %s", p.curToken.Type)
			return nil
		}
	}
	
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close record literal, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"
	
	return &ast.RecordLiteral{
		Position: pos,
		Fields:   fields,
	}
}

// parseRecordOrSetLiteral attempts to parse as record literal first,
// then falls back to set literal if no colons are found
// This function is called when we encounter a { token and need to decide
// if it's a record literal { field: value } or a set literal { value1, value2 }
func (p *Parser) parseRecordOrSetLiteral() ast.Expr {
	pos := tokenPos(p.curToken)
	
	// Check if we have a colon or ellipsis - indicates record literal
	if p.peekTokenIs(lexer.ELLIPSIS) {
		// Spread operator - must be record
		return p.parseRecordLiteral()
	}
	
	if p.peekTokenIs(lexer.RBRACE) {
		// Empty: {} - treat as empty set
		p.nextToken() // consume "{"
		p.nextToken() // consume "}"
		return &ast.SetLiteral{
			Position: pos,
			Elements: []ast.Expr{},
		}
	}
	
	// Check if next token after { is IDENT followed by COLON (record) or not (set)
	if p.peekTokenIs(lexer.LBRACE) {
		// Nested brace: { { ... } }
		// The outer must be a set literal (sets can contain records)
		// When we parse elements, parseExpression will correctly identify
		// the inner { ... } as a record if it has the pattern { id: ... }
		return p.parseSetLiteral()
	} else if p.peekTokenIs(lexer.IDENT) {
		// Check if IDENT is followed by COLON (record) or not (set)
		// We need to look ahead by consuming { to get to IDENT, then checking
		// what comes after IDENT
		p.nextToken() // consume {, now curToken is IDENT, peekToken is what comes after IDENT
		isRecord := p.peekTokenIs(lexer.COLON)
		
		if isRecord {
			// It's a record literal. We've consumed {, so we're at IDENT.
			// Parse record fields starting from IDENT.
			return p.parseRecordLiteralFromIdent(pos)
		}
		
		// It's a set literal. We've consumed {, so we're at IDENT.
		// IDENT is the first element of the set. Parse it as an expression
		// (which may include method calls, etc.), then continue with set parsing.
		var elements []ast.Expr
		
		// Parse the first element as an expression (not just identifier)
		// This handles cases like { queue.head().id } where the element
		// is a complex expression
		firstElement := p.parseExpression(LOWEST)
		if firstElement == nil {
			return nil
		}
		elements = append(elements, firstElement)
		
		// Continue parsing remaining elements
		for p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
			value := p.parseExpression(LOWEST)
			if value == nil {
				return nil
			}
			elements = append(elements, value)
		}
		
		if !p.curTokenIs(lexer.RBRACE) {
			p.addErrorf("expected } to close set literal, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume "}"
		
		return &ast.SetLiteral{
			Position: pos,
			Elements: elements,
		}
	}
	
	// For nested braces or when we can't determine, default to set literal
	return p.parseSetLiteral()
}

// parseRecordLiteralFromIdent parses a record literal starting from IDENT
// (assuming { was already consumed). This is used when we've determined
// that { IDENT ... is a record literal.
func (p *Parser) parseRecordLiteralFromIdent(pos ast.Position) ast.Expr {
	// We're at IDENT, which is the first field name
	var fields []ast.RecordField
	
	// Parse fields until closing brace
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Check for spread operator
		if p.curTokenIs(lexer.ELLIPSIS) {
			p.nextToken() // consume "..."
			
			if !p.curTokenIs(lexer.IDENT) {
				p.addErrorf("expected identifier after ..., got %s", p.curToken.Type)
				return nil
			}
			
			fields = append(fields, ast.RecordField{
				Name:   p.curToken.Literal,
				Value:  nil,
				Spread: true,
			})
			p.nextToken() // consume identifier
		} else {
			// Regular field: name: value
			if !p.curTokenIs(lexer.IDENT) {
				p.addErrorf("expected identifier or ... in record literal, got %s", p.curToken.Type)
				return nil
			}
			
			fieldName := p.curToken.Literal
			p.nextToken() // consume identifier
			
			if !p.curTokenIs(lexer.COLON) {
				p.addErrorf("expected : after field name, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken() // consume ":"
			
			value := p.parseExpression(LOWEST)
			if value == nil {
				return nil
			}
			
			fields = append(fields, ast.RecordField{
				Name:   fieldName,
				Value:  value,
				Spread: false,
			})
		}
		
		// Check for comma or closing brace
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
		} else if !p.curTokenIs(lexer.RBRACE) {
			p.addErrorf("expected , or } after field, got %s", p.curToken.Type)
			return nil
		}
	}
	
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close record literal, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"
	
	return &ast.RecordLiteral{
		Position: pos,
		Fields:   fields,
	}
}

// parseSetLiteral parses a set literal value
// Format: { value1, value2, ... }
func (p *Parser) parseSetLiteral() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "{"
	
	var elements []ast.Expr
	
	// Parse elements until closing brace
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		
		elements = append(elements, value)
		
		// Check for comma or closing brace
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
		} else if !p.curTokenIs(lexer.RBRACE) {
			p.addErrorf("expected , or } after set element, got %s", p.curToken.Type)
			return nil
		}
	}
	
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close set literal, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"
	
	return &ast.SetLiteral{
		Position: pos,
		Elements: elements,
	}
}

// parseListLiteral parses a list literal value
// Format: [ value1, value2, ... ]
func (p *Parser) parseListLiteral() ast.Expr {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume "["
	
	var elements []ast.Expr
	
	// Parse elements until closing bracket
	for !p.curTokenIs(lexer.RBRACKET) && !p.curTokenIs(lexer.EOF) {
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		
		elements = append(elements, value)
		
		// Check for comma or closing bracket
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
		} else if !p.curTokenIs(lexer.RBRACKET) {
			p.addErrorf("expected , or ] after list element, got %s", p.curToken.Type)
			return nil
		}
	}
	
	if !p.curTokenIs(lexer.RBRACKET) {
		p.addErrorf("expected ] to close list literal, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "]"
	
	return &ast.ListLiteral{
		Position: pos,
		Elements: elements,
	}
}

// isLambdaStart checks if the current position starts a lambda expression
// Lambda can be: param => expr or (param) => expr or (param1, param2) => expr
func (p *Parser) isLambdaStart() bool {
	// We're at LPAREN, check if next tokens suggest a lambda
	// Save current position
	savedCur := p.curToken
	savedPeek := p.peekToken
	
	// Advance past LPAREN
	p.nextToken()
	
	// Check if we have identifier(s) followed by => or )
	if p.curTokenIs(lexer.IDENT) {
		// Could be lambda parameter
		p.nextToken() // consume IDENT
		
		// Check for comma (multiple params) or closing paren
		if p.curTokenIs(lexer.COMMA) {
			// Multiple params, continue checking
			p.nextToken() // consume COMMA
			if p.curTokenIs(lexer.IDENT) {
				p.nextToken() // consume second IDENT
				// Check for closing paren and fat arrow
				if p.curTokenIs(lexer.RPAREN) && p.peekTokenIs(lexer.FATARROW) {
					// Restore and return true
					p.curToken = savedCur
					p.peekToken = savedPeek
					return true
				}
			}
		} else if p.curTokenIs(lexer.RPAREN) && p.peekTokenIs(lexer.FATARROW) {
			// Single param with parens: (param) =>
			// Restore and return true
			p.curToken = savedCur
			p.peekToken = savedPeek
			return true
		} else if p.curTokenIs(lexer.COLON) {
			// Typed parameter: (param: Type) =>
			p.nextToken() // consume COLON
			// Skip type parsing - just check if we eventually hit RPAREN FATARROW
			depth := 1
			for depth > 0 && !p.curTokenIs(lexer.EOF) {
				if p.curTokenIs(lexer.LPAREN) {
					depth++
				} else if p.curTokenIs(lexer.RPAREN) {
					depth--
					if depth == 0 && p.peekTokenIs(lexer.FATARROW) {
						// Restore and return true
						p.curToken = savedCur
						p.peekToken = savedPeek
						return true
					}
				}
				p.nextToken()
			}
		}
	}
	
	// Restore position
	p.curToken = savedCur
	p.peekToken = savedPeek
	return false
}

// parseLambdaExpression parses a lambda expression
// Format: param => expr or (param) => expr or (param1, param2) => expr
// Also handles: (param: Type) => expr
func (p *Parser) parseLambdaExpression() ast.Expr {
	pos := tokenPos(p.curToken)
	var params []ast.Parameter
	
	// Check if lambda starts with parentheses
	if p.curTokenIs(lexer.LPAREN) {
		p.nextToken() // consume "("
		
		// Parse parameters until closing paren
		for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
			if !p.curTokenIs(lexer.IDENT) {
				p.addErrorf("expected parameter name in lambda, got %s", p.curToken.Type)
				return nil
			}
			
			paramName := p.curToken.Literal
			p.nextToken() // consume parameter name
			
			// Check for optional type annotation
			var paramType ast.Type
			if p.curTokenIs(lexer.COLON) {
				p.nextToken() // consume ":"
				paramType = p.parseType()
				if paramType == nil {
					return nil
				}
			}
			
			params = append(params, ast.Parameter{
				Name: paramName,
				Type: paramType,
			})
			
			// Check for comma or closing paren
			if p.curTokenIs(lexer.COMMA) {
				p.nextToken() // consume ","
			} else if !p.curTokenIs(lexer.RPAREN) {
				p.addErrorf("expected , or ) after lambda parameter, got %s", p.curToken.Type)
				return nil
			}
		}
		
		if !p.curTokenIs(lexer.RPAREN) {
			p.addErrorf("expected ) to close lambda parameters, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume ")"
	} else if p.curTokenIs(lexer.IDENT) {
		// Single parameter without parentheses: param => expr
		paramName := p.curToken.Literal
		p.nextToken() // consume parameter name
		
		params = append(params, ast.Parameter{
			Name: paramName,
			Type: nil, // Type inferred from usage
		})
	} else {
		p.addErrorf("expected parameter name or ( for lambda expression, got %s", p.curToken.Type)
		return nil
	}
	
	// Parse fat arrow
	if !p.curTokenIs(lexer.FATARROW) {
		p.addErrorf("expected => after lambda parameters, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "=>"
	
	// Parse lambda body expression
	body := p.parseExpression(LOWEST)
	if body == nil {
		return nil
	}
	
	return &ast.LambdaExpr{
		Position: pos,
		Params:   params,
		Body:     body,
	}
}

