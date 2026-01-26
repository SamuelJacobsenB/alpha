package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// CONSTANTES E CONFIGURAÇÕES
// ============================

// typeKeywords define as palavras-chave de tipo reconhecidas
var typeKeywords = map[string]bool{
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
}

// isTypeKeyword verifica se uma string é uma palavra-chave de tipo
func isTypeKeyword(lex string) bool {
	return typeKeywords[lex] || (len(lex) > 0 && lex[0] >= 'A' && lex[0] <= 'Z')
}

// ============================
// PARSING DE TIPOS - FUNÇÃO PRINCIPAL
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
// TIPOS UNION
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
// TIPOS SIMPLES (BASE)
// ============================

// parseSingleType analisa um tipo simples (sem união)
func (p *Parser) parseSingleType() Type {
	if !p.isValidTypeStart() {
		return nil
	}

	var typ Type
	switch p.cur.Lexeme {
	case "set", "map":
		name := p.cur.Lexeme
		p.advanceToken()
		typ = p.parseGenericType(name)
	default:
		typ = p.parseBaseType()
	}

	return p.parseTypeModifiers(typ)
}

// isValidTypeStart verifica se o token atual pode iniciar um tipo
func (p *Parser) isValidTypeStart() bool {
	return isTypeKeyword(p.cur.Lexeme) || p.cur.Type == lexer.IDENT || p.cur.Type == lexer.GENERIC
}

// parseBaseType analisa um tipo base (primitivo ou identificador)
func (p *Parser) parseBaseType() Type {
	name := p.cur.Lexeme

	if !p.isValidBaseTypeName(name) {
		return nil
	}

	p.advanceToken()

	// Verificar se é um tipo genérico com parâmetros: List<T>
	if p.cur.Lexeme == "<" {
		return p.parseGenericType(name)
	}

	// Tipo primitivo ou identificador de tipo
	return &IdentifierType{Name: name}
}

// isValidBaseTypeName verifica se o nome é válido para tipo base
func (p *Parser) isValidBaseTypeName(name string) bool {
	return isTypeKeyword(name) || (len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z')
}

// ============================
// TIPOS GENÉRICOS
// ============================

// parseGenericType analisa um tipo genérico (List<T>, Set<T>, Map<K,V>)
func (p *Parser) parseGenericType(name string) Type {
	// O token atual deve ser '<' (o nome já foi consumido)
	if p.cur.Lexeme != "<" {
		return nil
	}

	p.advanceToken()

	// Processa baseado no tipo de coleção
	switch name {
	case "set":
		return p.parseSetType()
	case "map":
		return p.parseMapType()
	default:
		return p.parseUserDefinedGenericType(name)
	}
}

// parseSetType analisa tipo de conjunto (Set<T>)
func (p *Parser) parseSetType() Type {
	elemType := p.parseType()
	if elemType == nil || !p.expectAndConsume(">") {
		return nil
	}

	return &SetType{ElementType: elemType}
}

// parseMapType analisa tipo de mapa (Map<K,V>)
func (p *Parser) parseMapType() Type {
	keyType := p.parseType()
	if keyType == nil || !p.expectAndConsume(",") {
		return nil
	}

	valueType := p.parseType()
	if valueType == nil || !p.expectAndConsume(">") {
		return nil
	}

	return &MapType{KeyType: keyType, ValueType: valueType}
}

// parseUserDefinedGenericType analisa tipos genéricos definidos pelo usuário
func (p *Parser) parseUserDefinedGenericType(name string) Type {
	// Coletar todos os argumentos de tipo
	typeArgs := p.parseTypeArgumentList()
	if typeArgs == nil || !p.expectAndConsume(">") {
		return nil
	}

	return &GenericType{
		Name:     name,
		TypeArgs: typeArgs,
	}
}

// parseTypeArgumentList analisa lista de argumentos de tipo para genéricos
func (p *Parser) parseTypeArgumentList() []Type {
	typeArgs := make([]Type, 0, 2)

	firstTypeArg := p.parseType()
	if firstTypeArg == nil {
		return nil
	}
	typeArgs = append(typeArgs, firstTypeArg)

	for p.cur.Lexeme == "," {
		p.advanceToken()
		nextTypeArg := p.parseType()
		if nextTypeArg == nil {
			return nil
		}
		typeArgs = append(typeArgs, nextTypeArg)
	}

	return typeArgs
}

// ============================
// MODIFICADORES DE TIPO
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
	p.advanceToken() // consome '['

	var size Expr
	if p.cur.Lexeme != "]" {
		size = p.parseExpression(LOWEST)
	}

	if !p.expectAndConsume("]") {
		return nil
	}

	return &ArrayType{ElementType: elementType, Size: size}
}

// parseReturnTypeList analisa um ou mais tipos de retorno (T1, T2)
func (p *Parser) parseReturnTypeList() []Type {
	types := make([]Type, 0, 1)

	firstType := p.parseType()
	if firstType == nil {
		return nil
	}
	types = append(types, firstType)

	for p.cur.Lexeme == "," {
		p.advanceToken()
		nextType := p.parseType()
		if nextType == nil {
			p.errorf("expected type after ',' in return list")
			return nil
		}
		types = append(types, nextType)
	}

	return types
}
