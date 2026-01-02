package parser

import (
	"github.com/spectre-lang/spectre/internal/lexer"
)

// syncToDeclaration advances the parser to the next declaration start
// This is used for error recovery - when we encounter an error, we skip
// to the next declaration to continue parsing
func (p *Parser) syncToDeclaration() {
	for !p.curTokenIs(lexer.EOF) {
		// Check if current token is a declaration start
		if isDeclarationStart(p.curToken.Type) {
			return
		}

		// Skip tokens until we find a declaration start
		p.nextToken()
	}
}

// isDeclarationStart checks if the token type indicates the start of a declaration
func isDeclarationStart(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.VAR, lexer.CONST, lexer.FUN, lexer.ACTION,
		lexer.INIT, lexer.INVARIANT, lexer.TEMPORAL,
		lexer.MODULE, lexer.IMPORT, lexer.DESCRIPTION:
		return true
	default:
		return false
	}
}

// syncToStatement advances the parser to the next statement start
// Used for error recovery within blocks (actions, functions, init)
func (p *Parser) syncToStatement() {
	for !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.RBRACE) {
		// Check if current token is a statement start
		if isStatementStart(p.curToken.Type) {
			return
		}

		// If we encounter a declaration start, we've gone too far (outside the block)
		if isDeclarationStart(p.curToken.Type) {
			return
		}

		// Skip tokens until we find a statement start or block end
		p.nextToken()
	}
}

// isStatementStart checks if the token type indicates the start of a statement
func isStatementStart(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.IDENT, lexer.RETURN, lexer.REQUIRE, lexer.ENSURE,
		lexer.IF, lexer.LET, lexer.DESCRIPTION:
		return true
	default:
		return false
	}
}

// syncToExpressionEnd advances the parser to the end of an expression
// Used for error recovery when parsing expressions
func (p *Parser) syncToExpressionEnd() {
	for !p.curTokenIs(lexer.EOF) {
		// Stop at common expression terminators
		if p.curTokenIs(lexer.SEMICOLON) || p.curTokenIs(lexer.COMMA) ||
			p.curTokenIs(lexer.RPAREN) || p.curTokenIs(lexer.RBRACKET) ||
			p.curTokenIs(lexer.RBRACE) || isStatementStart(p.curToken.Type) ||
			isDeclarationStart(p.curToken.Type) {
			return
		}

		p.nextToken()
	}
}

