package parser

import (
	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseFunctionParameters() []*Param {
	if !p.expectAndConsume("(") {
		return nil
	}

	if p.cur.Lexeme == ")" {
		p.advanceToken()
		return []*Param{}
	}

	var params []*Param

	for {
		paramType := p.parseType()
		if paramType == nil {
			return nil
		}

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected parameter name")
			return nil
		}

		params = append(params, &Param{
			Name: p.cur.Lexeme,
			Type: paramType,
		})
		p.advanceToken()

		if p.cur.Lexeme == ")" {
			p.advanceToken()
			break
		}

		if !p.expectAndConsume(",") {
			return nil
		}
	}

	return params
}

func (p *Parser) parseFunctionBody() []Stmt {
	if p.cur.Lexeme != "{" {
		p.errorf("expected '{' for function body")
		return nil
	}
	return p.parseBlock()
}
