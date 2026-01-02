package parser

import (
	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

// parseInitDecl parses an init declaration
// Formats:
//   init { assignments... }
//   init expression
//   init oneOf { option1, option2, ... }
func (p *Parser) parseInitDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "init" keyword
	if !p.curTokenIs(lexer.INIT) {
		p.addErrorf("expected init, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "init"

	// Check for oneOf
	if p.curTokenIs(lexer.ONEOF) {
		return p.parseOneOfInitDecl(pos, description)
	}

	// Check if it's a block form or single expression form
	if p.curTokenIs(lexer.LBRACE) {
		// Block form: init { assignments... }
		body := p.parseBlock()
		if body == nil {
			return nil
		}

		return &ast.InitDecl{
			Position:    pos,
			Description: description,
			Body:        body,
			Expression:  nil,
		}
	}

	// Single expression form: init expression
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	return &ast.InitDecl{
		Position:    pos,
		Description: description,
		Body:        nil,
		Expression:  expr,
	}
}

// parseOneOfInitDecl parses a oneOf init declaration
// Formats:
//   init oneOf { option1, option2, ... }
//   where each option can be:
//     - Single assignment: counter = 0
//     - Tuple: { counter = 0, status = ... }
//     - Block: { counter = 0; status = ... }
func (p *Parser) parseOneOfInitDecl(pos ast.Position, description string) ast.Decl {
	// curToken should be on ONEOF
	if !p.curTokenIs(lexer.ONEOF) {
		p.addErrorf("expected oneOf, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "oneOf"

	// Parse opening brace
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { after oneOf, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	var options []*ast.BlockStmt

	// Parse options until closing brace
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Parse a single option (can be single assignment, tuple, or block)
		option := p.parseOneOfOption()
		if option == nil {
			// Skip to next comma or closing brace
			for !p.curTokenIs(lexer.COMMA) && !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			}
			if p.curTokenIs(lexer.COMMA) {
				p.nextToken() // consume comma
			}
			continue
		}
		options = append(options, option)

		// Check for comma or closing brace
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
		} else if !p.curTokenIs(lexer.RBRACE) {
			// Error: expected comma or closing brace
			p.addErrorf("expected , or }, got %s", p.curToken.Type)
			break
		}
	}

	// Parse closing brace
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } after oneOf options, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.OneOfInitDecl{
		Position:    pos,
		Description: description,
		Options:    options,
	}
}

// parseOneOfOption parses a single option in a oneOf init declaration
// Can be:
//   - Single assignment: counter = 0
//   - Tuple: { counter = 0, status = ... }
//   - Block: { counter = 0; status = ... }
func (p *Parser) parseOneOfOption() *ast.BlockStmt {
	optionPos := tokenPos(p.curToken)
	var statements []ast.Stmt

	// Check if it's a tuple/block form (starts with {)
	if p.curTokenIs(lexer.LBRACE) {
		p.nextToken() // consume "{"

		// Parse statements until closing brace
		for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
			stmt := p.parseStatement()
			if stmt == nil {
				// Skip to closing brace
				for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
					p.nextToken()
				}
				break
			}
			statements = append(statements, stmt)

			// In block form, statements are separated by newlines (no semicolons)
			// Check if next token is identifier (next statement) or closing brace
			if p.curTokenIs(lexer.IDENT) || p.curTokenIs(lexer.SET) || p.curTokenIs(lexer.MAP) || p.curTokenIs(lexer.LIST) {
				// Next statement - continue loop
				continue
			}
			// Otherwise, should be closing brace
			if !p.curTokenIs(lexer.RBRACE) {
				// Try to recover - skip unexpected token
				p.nextToken()
			}
		}

		// Parse closing brace
		if !p.curTokenIs(lexer.RBRACE) {
			p.addErrorf("expected } after oneOf option, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume "}"

		return &ast.BlockStmt{
			Position:   optionPos,
			Statements: statements,
		}
	}

	// Single assignment form: counter = 0
	stmt := p.parseStatement()
	if stmt == nil {
		return nil
	}

	return &ast.BlockStmt{
		Position:   optionPos,
		Statements: []ast.Stmt{stmt},
	}
}

// parseAssignStatement parses an assignment statement
// Format: identifier [''] = expression
// Note: curToken should be on IDENT (or SET/MAP/LIST/OPTION) when called
func (p *Parser) parseAssignStatement() ast.Stmt {
	pos := tokenPos(p.curToken)

	// Parse left-hand side (identifier with optional prime, or selector)
	// parseIdentifier will consume the identifier and prime if present
	// Use EQUALS+1 precedence to stop before ASSIGN (which has EQUALS precedence)
	// This prevents ASSIGN from being parsed as an infix operator (equality) in assignment contexts
	left := p.parseExpression(EQUALS + 1)
	if left == nil {
		return nil
	}

	// After parseExpression, curToken should be on ASSIGN (if identifier was primed)
	// or we need to check if it's ASSIGN
	if !p.curTokenIs(lexer.ASSIGN) {
		p.addErrorf("expected = after identifier, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "="

	// Parse right-hand side (expression)
	// In the right-hand side, ASSIGN can be treated as equality if needed
	right := p.parseExpression(LOWEST)
	if right == nil {
		return nil
	}

	return &ast.AssignStmt{
		Position: pos,
		Left:     left,
		Right:    right,
	}
}

