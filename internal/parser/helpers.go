package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) canStartExpression() bool {
	switch p.cur.Type {
	case lexer.IDENT, lexer.INT, lexer.FLOAT, lexer.STRING:
		return true
	case lexer.KEYWORD:
		return p.isValidExpressionKeyword()
	case lexer.OP:
		return p.isValidExpressionOperator()
	default:
		return false
	}
}

func (p *Parser) isValidExpressionKeyword() bool {
	return p.cur.Lexeme == "true" || p.cur.Lexeme == "false" ||
		p.cur.Lexeme == "null" || (isTypeKeyword(p.cur.Lexeme) && p.nxt.Lexeme == "function")
}

func (p *Parser) isValidExpressionOperator() bool {
	return p.cur.Lexeme == "-" || p.cur.Lexeme == "!" || p.cur.Lexeme == "+" ||
		p.cur.Lexeme == "(" || p.cur.Lexeme == "{" || p.cur.Lexeme == "[" ||
		p.cur.Lexeme == "&" // Referência
}

func (p *Parser) isKnownKeyword() bool {
	keywords := map[string]bool{
		"var": true, "const": true, "function": true,
		"if": true, "while": true, "do": true, "for": true, "return": true,
	}
	return keywords[p.cur.Lexeme]
}
