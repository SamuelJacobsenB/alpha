package parser

import (
	"github.com/alpha/internal/lexer"
)

var typeKeywords = map[string]bool{
	"int": true, "string": true, "float": true, "bool": true,
	"void": true, "byte": true, "char": true, "error": true,
	"set": true, "map": true,
}

func isTypeKeyword(lex string) bool {
	return typeKeywords[lex]
}

func (p *Parser) parseType() Type {
	left := p.parseSingleType()
	if left == nil {
		return nil
	}

	if p.cur.Lexeme == "|" {
		return p.parseUnionType(left)
	}

	return left
}

func (p *Parser) parseUnionType(firstType Type) Type {
	types := []Type{firstType}

	for p.cur.Lexeme == "|" {
		p.advanceToken() // consume "|"

		nextType := p.parseSingleType()
		if nextType == nil {
			return nil
		}
		types = append(types, nextType)
	}

	return &UnionType{Types: types}
}

func (p *Parser) parseSingleType() Type {
	if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT {
		return nil
	}

	var typ Type

	switch p.cur.Lexeme {
	case "set":
		typ = p.parseSetType()
	case "map":
		typ = p.parseMapType()
	default:
		typ = p.parseBaseType()
	}

	return p.parseTypeModifiers(typ)
}

func (p *Parser) parseBaseType() Type {
	name := p.cur.Lexeme
	p.advanceToken()

	if p.cur.Lexeme == "<" {
		if isGenericTypeStart(name) {
			return p.parseGenericType(name)
		}
	}

	if isTypeKeyword(name) {
		return &PrimitiveType{Name: name}
	}

	return &IdentifierType{Name: name}
}

func (p *Parser) parseGenericType(base string) Type {
	p.advanceToken()

	switch base {
	case "set":
		elemType := p.parseType()
		if !p.expectAndConsume(">") {
			return nil
		}
		return &SetType{ElementType: elemType}

	case "map":
		keyType := p.parseType()
		if !p.expectAndConsume(",") {
			return nil
		}
		valueType := p.parseType()
		if !p.expectAndConsume(">") {
			return nil
		}
		return &MapType{KeyType: keyType, ValueType: valueType}

	default:
		typ := &IdentifierType{Name: base}
		if !p.expectAndConsume(">") {
			return nil
		}
		return typ
	}
}

func isGenericTypeStart(name string) bool {
	return name == "set" || name == "map" || (name[0] >= 'A' && name[0] <= 'Z')
}

func (p *Parser) parseSetType() Type {
	p.advanceToken()
	if !p.expectAndConsume("<") {
		return nil
	}
	elemType := p.parseType()
	if !p.expectAndConsume(">") {
		return nil
	}
	return &SetType{ElementType: elemType}
}

func (p *Parser) parseMapType() Type {
	p.advanceToken()
	if !p.expectAndConsume("<") {
		return nil
	}
	keyType := p.parseType()
	if !p.expectAndConsume(",") {
		return nil
	}
	valueType := p.parseType()
	if !p.expectAndConsume(">") {
		return nil
	}
	return &MapType{KeyType: keyType, ValueType: valueType}
}

func (p *Parser) parseTypeModifiers(base Type) Type {
	current := base

	for {
		switch p.cur.Lexeme {
		case "?":
			current = &NullableType{BaseType: current}
			p.advanceToken()
		case "*":
			current = &PointerType{BaseType: current}
			p.advanceToken()
		case "[":
			current = p.parseArrayType(current)
		default:
			return current
		}
	}
}

func (p *Parser) parseArrayType(elementType Type) Type {
	p.advanceToken()
	var size Expr
	if p.cur.Lexeme != "]" {
		size = p.parseExpression(LOWEST)
	}
	if !p.expectAndConsume("]") {
		return nil
	}
	return &ArrayType{ElementType: elementType, Size: size}
}
