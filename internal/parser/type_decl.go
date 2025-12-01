package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseTypeDecl() Stmt {
	p.advanceToken() // consume 'type'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name after 'type'")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Generics opcionais: type Car<T>
	var generics []*GenericParam
	if p.cur.Lexeme == "<" {
		generics = p.parseGenericParams()
	}

	// Pode ser union type ou struct type
	typ := p.parseTypeBody()
	if typ == nil {
		return nil
	}

	return &TypeDecl{Name: name, Generics: generics, Type: typ}
}

func (p *Parser) parseTypeBody() Type {
	// Se começar com {, é um struct type
	if p.cur.Lexeme == "{" {
		return p.parseStructType()
	}

	// Caso contrário, é um union type ou tipo simples
	return p.parseType()
}

func (p *Parser) parseStructType() Type {
	p.advanceToken() // consume '{'

	var fields []*FieldDecl

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		field := p.parseField()
		if field != nil {
			fields = append(fields, field)
		} else {
			p.advanceToken()
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &StructType{Fields: fields}
}

// parseStructLiteral parseia { text: "hello", sender: "Sam" }
func (p *Parser) parseStructLiteral() Expr {
	p.advanceToken() // consume '{'

	var fields []*StructField

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected field name in struct literal")
			return nil
		}

		fieldName := p.cur.Lexeme
		p.advanceToken()

		if !p.expectAndConsume(":") {
			return nil
		}

		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}

		fields = append(fields, &StructField{Name: fieldName, Value: value})

		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != "}" {
			p.errorf("expected ',' or '}' in struct literal")
			return nil
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &StructLiteral{Fields: fields}
}

// isStructLiteral verifica se { começa um struct literal (não um mapa)
// Struct: { name: value, ... }
// Map: { key: value, ... } onde key é uma expressão
func (p *Parser) isStructLiteral() bool {
	if p.cur.Lexeme != "{" {
		return false
	}

	// Salvar estado
	saved := p.cur
	defer func() { p.cur = saved }()

	p.advanceToken() // consume '{'

	// Se vazio ou primeira coisa não é IDENT, não é struct
	if p.cur.Lexeme == "}" || p.cur.Type != lexer.IDENT {
		return false
	}

	p.advanceToken() // consume identifier

	// Se não tem ':', não é struct
	return p.cur.Lexeme == ":"
}
