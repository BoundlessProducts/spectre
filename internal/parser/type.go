package parser

import (
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// parseType parses a type expression
func (p *Parser) parseType() ast.Type {
	var typ ast.Type
	
	switch p.curToken.Type {
	case lexer.IDENT:
		// Could be primitive type or named type
		pos := tokenPos(p.curToken)
		name := p.curToken.Literal
		p.nextToken()
		
		// Check if it's a primitive type
		switch name {
		case "int", "bool", "str", "float":
			typ = &ast.PrimitiveType{
				Position: pos,
				Name:     name,
			}
		default:
			// Named type
			typ = &ast.NamedType{
				Position: pos,
				Name:     name,
			}
		}
	case lexer.SET:
		typ = p.parseSetType()
	case lexer.MAP:
		typ = p.parseMapType()
	case lexer.LIST:
		typ = p.parseListType()
	case lexer.OPTION:
		typ = p.parseOptionType()
	case lexer.ENUM:
		typ = p.parseEnumType()
	case lexer.LBRACE:
		typ = p.parseRecordType()
	default:
		p.addErrorf("unexpected token in type: %s", p.curToken.Type)
		return nil
	}
	
	return typ
}

// parseSetType parses a Set type (e.g., Set<int>)
func (p *Parser) parseSetType() ast.Type {
	pos := tokenPos(p.curToken)
	
	if !p.expectPeek(lexer.LT) {
		return nil
	}
	
	p.nextToken() // consume "<"
	elementType := p.parseType()
	if elementType == nil {
		return nil
	}
	
	// parseType() already advanced, so curToken is now on ">"
	if !p.curTokenIs(lexer.GT) {
		p.addErrorf("expected >, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ">"
	
	return &ast.SetType{
		Position: pos,
		Element:  elementType,
	}
}

// parseMapType parses a Map type (e.g., Map<int, str>)
func (p *Parser) parseMapType() ast.Type {
	pos := tokenPos(p.curToken)
	
	if !p.expectPeek(lexer.LT) {
		return nil
	}
	
	p.nextToken() // consume "<"
	keyType := p.parseType()
	if keyType == nil {
		return nil
	}
	
	// parseType() already advanced, so curToken is now on ","
	if !p.curTokenIs(lexer.COMMA) {
		p.addErrorf("expected comma, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ","
	
	valueType := p.parseType()
	if valueType == nil {
		return nil
	}
	
	// parseType() already advanced, so curToken is now on ">"
	if !p.curTokenIs(lexer.GT) {
		p.addErrorf("expected >, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ">"
	
	return &ast.MapType{
		Position: pos,
		Key:      keyType,
		Value:    valueType,
	}
}

// parseListType parses a List type (e.g., List<int>)
func (p *Parser) parseListType() ast.Type {
	pos := tokenPos(p.curToken)
	
	if !p.expectPeek(lexer.LT) {
		return nil
	}
	
	p.nextToken() // consume "<"
	elementType := p.parseType()
	if elementType == nil {
		return nil
	}
	
	// parseType() already advanced, so curToken is now on ">"
	if !p.curTokenIs(lexer.GT) {
		p.addErrorf("expected >, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ">"
	
	return &ast.ListType{
		Position: pos,
		Element:  elementType,
	}
}

// parseOptionType parses an Option type (e.g., Option<int>)
func (p *Parser) parseOptionType() ast.Type {
	pos := tokenPos(p.curToken)
	
	if !p.expectPeek(lexer.LT) {
		return nil
	}
	
	p.nextToken() // consume "<"
	elementType := p.parseType()
	if elementType == nil {
		return nil
	}
	
	// parseType() already advanced, so curToken is now on ">"
	if !p.curTokenIs(lexer.GT) {
		p.addErrorf("expected >, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ">"
	
	return &ast.OptionType{
		Position: pos,
		Element:  elementType,
	}
}

// parseEnumType parses an enum type (e.g., enum Status { Pending, Active })
func (p *Parser) parseEnumType() ast.Type {
	pos := tokenPos(p.curToken)
	
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	
	// expectPeek already advanced, so curToken is now the IDENT
	name := p.curToken.Literal
	
	if !p.expectPeek(lexer.LBRACE) {
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
	
	return &ast.EnumType{
		Position: pos,
		Name:     name,
		Values:   values,
	}
}

// parseRecordType parses a record type (e.g., { name: str, age: int })
func (p *Parser) parseRecordType() ast.Type {
	pos := tokenPos(p.curToken)
	p.nextToken() // consume '{'
	var fields []ast.Field
	
	// Parse fields
	if !p.curTokenIs(lexer.RBRACE) {
		field := p.parseField()
		if field == nil {
			return nil
		}
		fields = append(fields, *field)
		
		for p.curTokenIs(lexer.COMMA) {
			p.nextToken() // consume ","
			field := p.parseField()
			if field == nil {
				return nil
			}
			fields = append(fields, *field)
		}
	}
	
	// parseField() already advanced, so curToken should be on "}" now
	if !p.curTokenIs(lexer.RBRACE) {
		p.addErrorf("expected }, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "}"
	
	return &ast.RecordType{
		Position: pos,
		Fields:   fields,
	}
}

// parseField parses a field in a record (e.g., name: str)
func (p *Parser) parseField() *ast.Field {
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected field name, got %s", p.curToken.Type)
		return nil
	}
	
	fieldPos := tokenPos(p.curToken)
	name := p.curToken.Literal
	p.nextToken() // consume field name
	
	// After nextToken(), curToken should be ":"
	if !p.curTokenIs(lexer.COLON) {
		p.addErrorf("expected :, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume ":"
	fieldType := p.parseType()
	if fieldType == nil {
		return nil
	}
	
	return &ast.Field{
		Position: fieldPos,
		Name:     name,
		Type:     fieldType,
	}
}

