package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseType() Type {
	fmt.Printf("parseType: cur=%q\n", p.cur.Lexeme)

	// Verificar se é um tipo primitivo
	if p.cur.Type == lexer.KEYWORD && isTypeKeyword(p.cur.Lexeme) {
		return p.parsePrimitiveOrArrayType()
	}

	// CORREÇÃO: Não tentar parsear 'function' como tipo
	if p.cur.Lexeme == "function" {
		p.errorf("unexpected 'function' keyword in type position")
		return nil
	}

	p.errorf("expected type, got %q", p.cur.Lexeme)
	return nil
}

func (p *Parser) parsePrimitiveOrArrayType() Type {
	typeName := p.cur.Lexeme
	p.advanceToken()

	// Verificar se é array type
	if p.cur.Lexeme == "[" {
		return p.parseArrayType(&PrimitiveType{Name: typeName})
	}

	return &PrimitiveType{Name: typeName}
}

func (p *Parser) parseArrayType(elementType Type) Type {
	fmt.Printf("parseArrayType: starting\n")
	p.advanceToken() // consume '['

	var size Expr
	if p.cur.Lexeme != "]" {
		size = p.parseExpression(LOWEST)
	}

	if !p.expectAndConsume("]") {
		p.errorf("expected ']' in array type")
		return nil
	}

	return &ArrayType{
		ElementType: elementType,
		Size:        size,
	}
}

func (p *Parser) parseArrayLiteral() Expr {
	fmt.Printf("parseArrayLiteral: starting\n")
	p.advanceToken() // consume '{'

	elements := p.parseArrayElements()
	if elements == nil {
		return nil
	}

	if !p.expectAndConsume("}") {
		p.errorf("expected '}' after array literal")
		return nil
	}

	return &ArrayLiteral{Elements: elements}
}

func (p *Parser) parseArrayElements() []Expr {
	var elements []Expr

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		elements = append(elements, elem)

		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != "}" {
			p.errorf("expected ',' or '}' in array literal")
			return nil
		}
	}

	return elements
}

var TypeKeywords = map[string]bool{
	"int":       true,
	"string":    true,
	"float":     true,
	"bool":      true,
	"void":      true,
	"byte":      true,
	"char":      true,
	"double":    true,
	"boolean":   true,
	"error":     true,
	"component": true,
}

func isTypeKeyword(lex string) bool {
	typeKeywords := map[string]bool{
		"int":       true,
		"string":    true,
		"float":     true,
		"bool":      true,
		"void":      true,
		"byte":      true,
		"char":      true,
		"double":    true,
		"boolean":   true,
		"error":     true,
		"component": true,
	}
	return typeKeywords[lex]
}
