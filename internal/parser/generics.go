package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseGenericParams() []*GenericParam {
	fmt.Printf("parseGenericParams: starting at %q\n", p.cur.Lexeme)
	p.advanceToken() // consume '['

	var generics []*GenericParam

	for p.cur.Lexeme != "]" && p.cur.Type != lexer.EOF {
		generic := p.parseGenericParam()
		if generic == nil {
			break
		}
		generics = append(generics, generic)

		if !p.consumeOptionalComma() {
			break
		}
	}

	if !p.expectAndConsume("]") {
		p.errorf("expected ']' after generic parameters")
	}

	fmt.Printf("parseGenericParams: found %d generic parameters\n", len(generics))
	return generics
}

func (p *Parser) parseGenericParam() *GenericParam {
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected generic parameter name, got %q", p.cur.Lexeme)
		return nil
	}

	generic := &GenericParam{Name: p.cur.Lexeme}
	p.advanceToken()
	return generic
}

func (p *Parser) consumeOptionalComma() bool {
	if p.cur.Lexeme == "," {
		p.advanceToken()
		return true
	}
	return false
}

func (p *Parser) parseGenericCall(left Expr) Expr {
	fmt.Printf("parseGenericCall: starting with left=%T\n", left)
	p.advanceToken() // consume '<'

	typeArgs := p.parseTypeArguments()
	if typeArgs == nil {
		return nil
	}

	if !p.expectAndConsume(">") {
		p.errorf("expected '>' to close generic arguments")
		return nil
	}

	// Se houver parênteses, é uma chamada de função genérica
	if p.cur.Lexeme == "(" {
		return p.parseGenericFunctionCall(left, typeArgs)
	}

	// Caso contrário, é uma especialização genérica
	return &GenericSpecialization{
		Callee:   left,
		TypeArgs: typeArgs,
	}
}

func (p *Parser) parseTypeArguments() []Type {
	var typeArgs []Type

	for p.cur.Lexeme != ">" && p.cur.Type != lexer.EOF {
		typ := p.parseType()
		if typ != nil {
			typeArgs = append(typeArgs, typ)
		} else {
			p.errorf("expected type in generic arguments")
			break
		}

		if !p.consumeOptionalComma() {
			break
		}
	}

	return typeArgs
}

func (p *Parser) parseGenericFunctionCall(left Expr, typeArgs []Type) Expr {
	// Primeiro criar a especialização genérica
	specialization := &GenericSpecialization{
		Callee:   left,
		TypeArgs: typeArgs,
	}

	// Depois parsear a chamada de função normal
	return p.parseCallExpression(specialization)
}
