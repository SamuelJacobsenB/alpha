package parser

import (
	"github.com/alpha/internal/lexer"
)

var (
	typeKeywords = map[string]bool{
		"int": true, "string": true, "float": true, "bool": true,
		"void": true, "byte": true, "char": true, "error": true,
		"set": true, "map": true,
		// Aceita letras maiúsculas como tipos genéricos
	}
)

func isTypeKeyword(lex string) bool {
	// Aceita keywords de tipo e letras maiúsculas únicas (tipos genéricos)
	return typeKeywords[lex] || (len(lex) == 1 && lex[0] >= 'A' && lex[0] <= 'Z')
}

func (p *Parser) parseType() Type {
	left := p.parseSingleType()
	if left == nil {
		return nil
	}

	if p.cur.Lexeme != "|" {
		return left
	}

	return p.parseUnionType(left)
}

func (p *Parser) parseUnionType(firstType Type) Type {
	types := []Type{firstType}

	for p.cur.Lexeme == "|" {
		p.advanceToken()

		nextType := p.parseSingleType()
		if nextType == nil {
			return nil
		}
		types = append(types, nextType)
	}

	return &UnionType{Types: types}
}

func (p *Parser) parseSingleType() Type {
	if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT && p.cur.Type != lexer.GENERIC {
		return nil
	}

	var typ Type
	switch p.cur.Lexeme {
	case "set":
		typ = p.parseGenericType("set")
	case "map":
		typ = p.parseGenericType("map")
	default:
		typ = p.parseBaseType()
	}

	return p.parseTypeModifiers(typ)
}

func (p *Parser) parseBaseType() Type {
	name := p.cur.Lexeme
	p.advanceToken()

	// Verificar se é um tipo genérico com parâmetros: List<T>
	if p.cur.Lexeme == "<" {
		return p.parseGenericType(name)
	}

	// Verificar se é uma letra maiúscula única (tipo genérico)
	if len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' {
		return &IdentifierType{Name: name}
	}

	// Tipo primitivo ou identificador de tipo
	return &IdentifierType{Name: name}
}

// ... resto do código permanece igual

func (p *Parser) parseGenericType(name string) Type {
	if p.cur.Lexeme != "<" {
		return nil
	}

	p.advanceToken()

	switch name {
	case "set":
		elemType := p.parseType()
		if elemType == nil || !p.expectAndConsume(">") {
			return nil
		}
		return &SetType{ElementType: elemType}

	case "map":
		keyType := p.parseType()
		if keyType == nil || !p.expectAndConsume(",") {
			return nil
		}

		valueType := p.parseType()
		if valueType == nil || !p.expectAndConsume(">") {
			return nil
		}
		return &MapType{KeyType: keyType, ValueType: valueType}

	default:
		// Para tipos genéricos definidos pelo usuário (como List<T>)
		// Atualmente não armazenamos os parâmetros genéricos
		// TODO: Implementar GenericType struct
		_ = p.parseType()

		for p.cur.Lexeme == "," {
			p.advanceToken()
			_ = p.parseType()
		}

		if !p.expectAndConsume(">") {
			return nil
		}
		return &IdentifierType{Name: name}
	}
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
