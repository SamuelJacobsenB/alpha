package parser

import "github.com/alpha/internal/lexer"

// ============================
// TIPOS E CONSTANTES
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
// PARSING DE PARÂMETROS GENÉRICOS
// ============================

// parseGenericParamsWithPrefix parseia generics com prefixo opcional "generic<T>"
func (p *Parser) parseGenericParamsWithPrefix() []*GenericParam {
	// Verifica se tem prefixo "generic"
	hasPrefix := p.cur.Lexeme == "generic" && p.nxt.Lexeme == "<"
	if hasPrefix {
		p.advanceToken() // consome "generic"
	}

	// Se não tem "<", retorna nil
	if p.cur.Lexeme != "<" {
		if hasPrefix {
			p.errorf("expected '<' after 'generic'")
		}
		return nil
	}

	return p.parseGenericParamsList()
}

// parseGenericParamsList parseia uma lista de parâmetros genéricos: <T, U, V>
func (p *Parser) parseGenericParamsList() []*GenericParam {
	if !p.expectAndConsume("<") {
		return nil
	}

	params := p.parseGenericParamListItems()
	if !p.expectAndConsume(">") {
		return nil
	}

	return params
}

// parseGenericParamListItems parseia itens da lista de parâmetros genéricos
func (p *Parser) parseGenericParamListItems() []*GenericParam {
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
		p.advanceToken() // consome ","

		if !p.isValidGenericParam() {
			p.errorf("expected generic parameter name after ',', got %s", p.cur.Lexeme)
			return nil
		}

		params = append(params, &GenericParam{Name: p.cur.Lexeme})
		p.advanceToken()
	}

	return params
}

// ============================
// PARSING DE ARGUMENTOS DE TIPO
// ============================

// parseTypeArgumentsList parseia uma lista de argumentos de tipo: <int, string>
func (p *Parser) parseTypeArgumentsList() []Type {
	if !p.expectAndConsume("<") {
		return nil
	}

	typeArgs := p.parseTypeArgumentListItems()
	if !p.expectAndConsume(">") {
		return nil
	}

	return typeArgs
}

// parseTypeArgumentListItems parseia itens da lista de argumentos de tipo
func (p *Parser) parseTypeArgumentListItems() []Type {
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
		p.advanceToken() // consome ","

		typ = p.parseType()
		if typ == nil {
			p.errorf("expected type after ',' in generic arguments")
			return nil
		}
		typeArgs = append(typeArgs, typ)
	}

	return typeArgs
}

// ============================
// FUNÇÕES AUXILIARES
// ============================

// isValidGenericParam verifica se o token atual é um nome válido para parâmetro genérico
func (p *Parser) isValidGenericParam() bool {
	// Aceita identificadores
	if p.cur.Type == lexer.IDENT {
		return true
	}

	// Aceita tokens genéricos
	if p.cur.Type == lexer.GENERIC {
		return true
	}

	// Aceita keywords que são letras maiúsculas únicas (como T, U, V)
	if p.cur.Type == lexer.KEYWORD {
		lexeme := p.cur.Lexeme
		return len(lexeme) == 1 && lexeme[0] >= 'A' && lexeme[0] <= 'Z'
	}

	return false
}
