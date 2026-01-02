package parser

import (
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	EQUALS      // == !=
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
	lexer.AND:    EQUALS, // Logical AND has same precedence as ==
	lexer.OR:     EQUALS, // Logical OR has same precedence as ==
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
	for precedence < p.curPrecedence() {
		infix := p.parseInfixExpression(left)
		if infix == nil {
			return left
		}
		left = infix
		
		// Stop if we hit EOF, SEMICOLON, LBRACE, or RBRACE
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
		// parseIdentifier already advances past the identifier (and prime if present)
		return p.parseIdentifier()
	case lexer.INT, lexer.FLOAT, lexer.STRING, lexer.BOOL:
		expr = p.parseLiteral()
		p.nextToken() // Advance past literal
		return expr
	case lexer.NOT, lexer.MINUS:
		return p.parseUnaryExpression()
	case lexer.LPAREN:
		return p.parseGroupedExpression()
	case lexer.IF:
		return p.parseIfExpression()
	case lexer.LET:
		return p.parseLetExpression()
	case lexer.SUPER:
		return p.parseSuperExpression()
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

// parseGroupedExpression parses a parenthesized expression
func (p *Parser) parseGroupedExpression() ast.Expr {
	p.nextToken() // consume "("

	expr := p.parseExpression(LOWEST)

	// parseExpression already advanced, so curToken should be on ")"
	if !p.curTokenIs(lexer.RPAREN) {
		p.addErrorf("expected ), got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ")"

	return &ast.ParenExpr{
		Position: tokenPos(p.curToken),
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
	p.nextToken()

	index := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

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
	var list []ast.Expr

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
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
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

	// Parse then block
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { after if condition, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	thenExpr := p.parseExpression(LOWEST)
	if thenExpr == nil {
		return nil
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

	elseExpr := p.parseExpression(LOWEST)
	if elseExpr == nil {
		return nil
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

