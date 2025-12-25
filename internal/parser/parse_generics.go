package parser

import "github.com/alpha/internal/lexer"

// ============================
// Lógica Unificada de Generics
// ============================

// GenericContext define o contexto onde generics são parseados
type GenericContext int

const (
	GenericContextDeclaration GenericContext = iota // Para declarações: <T, U>
	GenericContextTypeArgs                          // Para argumentos de tipo: <int, string>
)

// ParsedGenerics contém o resultado de parsear generics
type ParsedGenerics struct {
	Params   []*GenericParam // Para declarações: <T, U>
	TypeArgs []Type          // Para usos: <int, string>
}

// ============================
// Funções Principais Unificadas
// ============================

// parseGenericParamsWithPrefix parseia generics com prefixo opcional "generic<T>"
func (p *Parser) parseGenericParamsWithPrefix() []*GenericParam {
	hasPrefix := false

	// Verifica se tem prefixo "generic"
	if p.cur.Lexeme == "generic" && p.nxt.Lexeme == "<" {
		p.advanceToken() // consume "generic"
		hasPrefix = true
	}

	// Se não tem "<", retorna nil
	if p.cur.Lexeme != "<" {
		return nil
	}

	// Se tinha prefixo mas não tem "<", é erro
	if hasPrefix && p.cur.Lexeme != "<" {
		p.errorf("expected '<' after 'generic'")
		return nil
	}

	return p.parseGenericParamsList()
}

// parseGenericParamsList parseia uma lista de parâmetros genéricos: <T, U, V>
func (p *Parser) parseGenericParamsList() []*GenericParam {
	if !p.expectAndConsume("<") {
		return nil
	}

	params := make([]*GenericParam, 0, 2)

	// Primeiro parâmetro
	if !p.isValidGenericParam() {
		p.errorf("expected generic parameter name, got %s", p.cur.Lexeme)
		return nil
	}

	params = append(params, &GenericParam{Name: p.cur.Lexeme})
	p.advanceToken()

	// Parâmetros adicionais
	for p.cur.Lexeme == "," {
		p.advanceToken() // consume ","

		if !p.isValidGenericParam() {
			p.errorf("expected generic parameter name after ',', got %s", p.cur.Lexeme)
			return nil
		}

		params = append(params, &GenericParam{Name: p.cur.Lexeme})
		p.advanceToken()
	}

	if !p.expectAndConsume(">") {
		return nil
	}

	return params
}

// parseTypeArgumentsList parseia uma lista de argumentos de tipo: <int, string>
func (p *Parser) parseTypeArgumentsList() []Type {
	if !p.expectAndConsume("<") {
		return nil
	}

	typeArgs := make([]Type, 0, 2)

	// Primeiro tipo
	typ := p.parseType()
	if typ == nil {
		p.errorf("expected type in generic arguments")
		return nil
	}
	typeArgs = append(typeArgs, typ)

	// Tipos adicionais
	for p.cur.Lexeme == "," {
		p.advanceToken() // consume ","

		typ = p.parseType()
		if typ == nil {
			p.errorf("expected type after ',' in generic arguments")
			return nil
		}
		typeArgs = append(typeArgs, typ)
	}

	if !p.expectAndConsume(">") {
		return nil
	}

	return typeArgs
}

// ============================
// Funções Auxiliares
// ============================

// isValidGenericParam verifica se o token atual é um nome válido para parâmetro genérico
func (p *Parser) isValidGenericParam() bool {
	// Aceita identificadores, keywords que podem ser usados como generics, e tipos genéricos
	return p.cur.Type == lexer.IDENT ||
		p.cur.Type == lexer.GENERIC ||
		(p.cur.Type == lexer.KEYWORD && len(p.cur.Lexeme) == 1 && p.cur.Lexeme[0] >= 'A' && p.cur.Lexeme[0] <= 'Z')
}
