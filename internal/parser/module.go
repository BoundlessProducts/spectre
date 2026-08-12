package parser

import (
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// parseModuleDecl parses a module declaration or module instance
// Format: description "text" module Name { ... } or module Name extends Parent { ... }
//        or module InstanceName = ModuleName with { substitutions }
func (p *Parser) parseModuleDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse optional description
	description, _ := p.parseDescription()

	// Parse "module" keyword
	if !p.curTokenIs(lexer.MODULE) {
		p.addErrorf("expected module, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "module"

	// Parse module/instance name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after module, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken() // consume identifier

	// Check if this is a module instance (has =)
	if p.curTokenIs(lexer.ASSIGN) {
		return p.parseModuleInstanceDecl(pos, description, name)
	}

	// Parse optional extends clause
	var extends string
	if p.curTokenIs(lexer.EXTENDS) {
		p.nextToken() // consume "extends"
		if !p.curTokenIs(lexer.IDENT) {
			p.addErrorf("expected identifier after extends, got %s", p.curToken.Type)
			return nil
		}
		extends = p.curToken.Literal
		p.nextToken() // consume parent module name
	}

	// Parse module body
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { for module body, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	// Parse all declarations within the module
	decls := p.parseModuleBody()

	// Parse closing brace
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close module body, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.ModuleDecl{
		Position:    pos,
		Description: description,
		Name:        name,
		Extends:     extends,
		Decls:       decls,
	}
}

// parseModuleBody parses all declarations within a module body
func (p *Parser) parseModuleBody() []ast.Decl {
	var decls []ast.Decl

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Skip comments and standalone descriptions
		if p.curTokenIs(lexer.DESCRIPTION) {
			p.parseDescription()
			continue
		}

		var decl ast.Decl
		switch p.curToken.Type {
		case lexer.VAR:
			decl = p.parseVariableDecl()
		case lexer.CONST:
			decl = p.parseConstantDecl()
		case lexer.FUN:
			decl = p.parseFunctionDecl()
		case lexer.ACTION:
			decl = p.parseActionDecl()
		case lexer.INIT:
			decl = p.parseInitDecl()
		case lexer.INVARIANT:
			decl = p.parseInvariantDecl()
		case lexer.TEMPORAL:
			decl = p.parseTemporalDecl()
		case lexer.IMPORT:
			decl = p.parseImportDecl()
		case lexer.TYPE:
			decl = p.parseTypeAliasDecl()
		case lexer.ENUM:
			decl = p.parseEnumDecl()
		case lexer.PUBLIC, lexer.PRIVATE:
			// Parse visibility modifier followed by declaration
			decl = p.parseVisibilityDecl()
		case lexer.PARAM:
			decl = p.parseParamDecl()
		default:
			// Skip unknown tokens (comments, etc.)
			p.nextToken()
			continue
		}

		if decl != nil {
			decls = append(decls, decl)
		} else {
			// Error recovery: if parsing failed, sync to next declaration or block end
			// This allows us to continue parsing even after errors
			if len(p.errors) > 0 {
				// Check if we should sync to statement (within block) or declaration (top-level)
				if p.curTokenIs(lexer.RBRACE) || p.curTokenIs(lexer.EOF) {
					// We're at the end, break out
					break
				}
				p.syncToStatement()
			}
		}
	}

	return decls
}

// parseVisibilityDecl parses a declaration with a visibility modifier
// Format: public|private var|const|action|invariant|temporal|fun ...
func (p *Parser) parseVisibilityDecl() ast.Decl {
	visibility := ast.Public
	if p.curTokenIs(lexer.PRIVATE) {
		visibility = ast.Private
	}
	p.nextToken() // consume "public" or "private"

	// Parse the actual declaration
	var decl ast.Decl
	switch p.curToken.Type {
	case lexer.VAR:
		decl = p.parseVariableDecl()
	case lexer.CONST:
		decl = p.parseConstantDecl()
	case lexer.ACTION:
		decl = p.parseActionDecl()
	case lexer.INVARIANT:
		decl = p.parseInvariantDecl()
	case lexer.TEMPORAL:
		decl = p.parseTemporalDecl()
	case lexer.FUN:
		decl = p.parseFunctionDecl()
	default:
		p.addErrorf("expected declaration after visibility modifier, got %s", p.curToken.Type)
		return nil
	}

	// Set visibility on the declaration
	if decl != nil {
		switch d := decl.(type) {
		case *ast.VariableDecl:
			d.Visibility = visibility
		case *ast.ConstantDecl:
			d.Visibility = visibility
		case *ast.ActionDecl:
			d.Visibility = visibility
		case *ast.InvariantDecl:
			d.Visibility = visibility
		case *ast.TemporalDecl:
			d.Visibility = visibility
		case *ast.FunctionDecl:
			d.Visibility = visibility
		}
	}

	return decl
}

// parseModuleInstanceDecl parses a module instance declaration
// Format: module InstanceName = ModuleName with { originalName = newName, ... }
func (p *Parser) parseModuleInstanceDecl(pos ast.Position, description string, instanceName string) ast.Decl {
	// We've already consumed "module" and instance name, curToken is on "="
	p.nextToken() // consume "="

	// Parse module name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected module name after =, got %s", p.curToken.Type)
		return nil
	}
	moduleName := p.curToken.Literal
	p.nextToken() // consume module name

	// Parse "with" keyword
	if !p.curTokenIs(lexer.WITH) {
		p.addErrorf("expected with after module name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "with"

	// Parse substitution block
	if !p.curTokenIs(lexer.LBRACE) {
		p.addErrorf("expected { for substitution block, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "{"

	// Parse substitutions: originalName = newName
	substitutions := make(map[string]string)
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Parse original name
		if !p.curTokenIs(lexer.IDENT) {
			p.addErrorf("expected identifier in substitution, got %s", p.curToken.Type)
			return nil
		}
		originalName := p.curToken.Literal
		p.nextToken() // consume original name

		// Parse "="
		if !p.curTokenIs(lexer.ASSIGN) {
			p.addErrorf("expected = in substitution, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken() // consume "="

		// Parse new name
		if !p.curTokenIs(lexer.IDENT) {
			p.addErrorf("expected identifier after = in substitution, got %s", p.curToken.Type)
			return nil
		}
		newName := p.curToken.Literal
		p.nextToken() // consume new name

		substitutions[originalName] = newName

		// Check for comma or closing brace
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
		} else if !p.curTokenIs(lexer.RBRACE) {
			p.addErrorf("expected , or } after substitution, got %s", p.curToken.Type)
			return nil
		}
	}

	// Parse closing brace
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected } to close substitution block, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"

	return &ast.ModuleInstanceDecl{
		Position:      pos,
		Name:          instanceName,
		Module:        moduleName,
		Substitutions: substitutions,
	}
}

