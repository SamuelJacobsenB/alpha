package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// Constantes e Configurações
// ============================

var (
	typeKeywords = map[string]bool{
		"int":    true,
		"string": true,
		"float":  true,
		"bool":   true,
		"void":   true,
		"byte":   true,
		"char":   true,
		"error":  true,
		"set":    true,
		"map":    true,
		// Letras maiúsculas são aceitas como tipos genéricos
	}
)

// isTypeKeyword verifica se uma string é uma palavra-chave de tipo
func isTypeKeyword(lex string) bool {
	return typeKeywords[lex] || (len(lex) == 1 && lex[0] >= 'A' && lex[0] <= 'Z')
}

// ============================
// Parsing de Tipos (Função Principal)
// ============================

// parseType analisa um tipo, incluindo tipos union (T1 | T2 | T3)
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

// ============================
// Tipos Union
// ============================

// parseUnionType analisa um tipo union (T1 | T2 | T3)
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

// ============================
// Tipos Simples
// ============================

// parseSingleType analisa um tipo simples (sem união)
func (p *Parser) parseSingleType() Type {
	if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT && p.cur.Type != lexer.GENERIC {
		return nil
	}

	var typ Type
	switch p.cur.Lexeme {
	case "set":
		p.advanceToken() // consume 'set'
		typ = p.parseGenericType("set")
	case "map":
		p.advanceToken() // consume 'map'
		typ = p.parseGenericType("map")
	default:
		typ = p.parseBaseType()
	}

	return p.parseTypeModifiers(typ)
}

// parseBaseType analisa um tipo base (primitivo ou identificador)
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

// ============================
// Tipos Genéricos
// ============================

// parseGenericType analisa um tipo genérico (List<T>, Set<T>, Map<K,V>)
// IMPORTANTE: Esta função espera que o nome do tipo já tenha sido consumido.
func (p *Parser) parseGenericType(name string) Type {
	// O token atual deve ser '<' (o nome já foi consumido)
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
		firstTypeArg := p.parseType()
		if firstTypeArg == nil {
			return nil
		}

		// Coletar todos os argumentos de tipo
		typeArgs := []Type{firstTypeArg}
		for p.cur.Lexeme == "," {
			p.advanceToken()
			nextTypeArg := p.parseType()
			if nextTypeArg == nil {
				return nil
			}
			typeArgs = append(typeArgs, nextTypeArg)
		}

		if !p.expectAndConsume(">") {
			return nil
		}

		return &GenericType{
			Name:     name,
			TypeArgs: typeArgs,
		}
	}
}

// ============================
// Modificadores de Tipo
// ============================

// parseTypeModifiers aplica modificadores a um tipo (? nullable, * pointer, [] array)
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

// parseArrayType analisa um tipo de array (elementType[])
func (p *Parser) parseArrayType(elementType Type) Type {
	p.advanceToken() // consume '['

	var size Expr
	if p.cur.Lexeme != "]" {
		size = p.parseExpression(LOWEST)
	}

	if !p.expectAndConsume("]") {
		return nil
	}

	return &ArrayType{ElementType: elementType, Size: size}
}
