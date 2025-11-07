package parser

import (
	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseExpression(precedence int) Expr {
	if !p.canStartExpression() {
		if p.cur.Type != lexer.EOF && p.cur.Lexeme != "" {
			p.errorf("unexpected token %q in expression at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
		}
		return nil
	}

	left := p.parsePrimaryExpression()
	if left == nil {
		return nil
	}

	if p.cur.Type == lexer.EOF {
		return left
	}

	return p.parsePostfixExpression(left, precedence)
}

func (p *Parser) parsePrimaryExpression() Expr {
	switch p.cur.Type {
	case lexer.IDENT:
		return p.parseIdentifier()
	case lexer.INT, lexer.FLOAT:
		return p.parseNumber()
	case lexer.STRING:
		return p.parseString()
	case lexer.KEYWORD:
		return p.parseKeywordExpression()
	case lexer.OP:
		return p.parseOperatorExpression()
	default:
		return nil
	}
}

func (p *Parser) parseIdentifier() Expr {
	ident := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()
	return ident
}

func (p *Parser) parseNumber() Expr {
	expr := p.parseNumberToken(p.cur)
	p.advanceToken()
	return expr
}

func (p *Parser) parseString() Expr {
	expr := &StringLiteral{Value: p.cur.Value}
	p.advanceToken()
	return expr
}
