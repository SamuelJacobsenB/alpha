package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseFunctionName() string {
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected function name, got %q", p.cur.Lexeme)
		return ""
	}
	name := p.cur.Lexeme
	p.advanceToken()
	return name
}

func (p *Parser) parseOptionalGenerics() []*GenericParam {
	if p.cur.Lexeme == "[" {
		return p.parseGenericParams()
	}
	return nil
}

func (p *Parser) parseFunctionParameters() []*Param {
	fmt.Printf("parseFunctionParameters: starting, cur=%q\n", p.cur.Lexeme)

	if !p.expectAndConsume("(") {
		p.errorf("expected '(' after function name")
		return nil
	}

	params := p.parseParameterList()

	if !p.expectAndConsume(")") {
		p.errorf("expected ')' after function parameters")
		return nil
	}

	fmt.Printf("parseFunctionParameters: completed with %d params\n", len(params))
	return params
}

func (p *Parser) parseParameterList() []*Param {
	var params []*Param

	for p.cur.Lexeme != ")" && p.cur.Type != lexer.EOF {
		param := p.parseParameter()
		if param == nil {
			return nil
		}
		params = append(params, param)

		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != ")" {
			p.errorf("expected ',' or ')' after parameter")
			return nil
		}
	}

	return params
}

func (p *Parser) parseParameter() *Param {
	paramType := p.parseType()
	if paramType == nil {
		p.errorf("expected parameter type")
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected parameter name")
		return nil
	}

	paramName := p.cur.Lexeme
	p.advanceToken()

	return &Param{
		Name: paramName,
		Type: paramType,
	}
}

func (p *Parser) parseFunctionBody() []Stmt {
	body := p.parseBlockLike()
	if body == nil {
		p.errorf("expected function body")
	}
	return body
}
