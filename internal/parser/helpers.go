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

func (p *Parser) peekNextAfterBrackets() string {
	saveCur := p.cur
	saveNxt := p.nxt

	// Avançar até encontrar o fechamento de ]
	for p.cur.Lexeme != "]" && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}

	if p.cur.Lexeme == "]" {
		p.advanceToken() // consume ]
	}

	next := p.cur.Lexeme

	// Restaurar estado
	p.cur = saveCur
	p.nxt = saveNxt

	return next
}

func (p *Parser) isValidExpressionKeyword() bool {
	return p.cur.Lexeme == "true" || p.cur.Lexeme == "false" ||
		p.cur.Lexeme == "null" || (isTypeKeyword(p.cur.Lexeme) && p.nxt.Lexeme == "function")
}

func (p *Parser) isValidExpressionOperator() bool {
	return p.cur.Lexeme == "-" || p.cur.Lexeme == "!" || p.cur.Lexeme == "+" ||
		p.cur.Lexeme == "(" || p.cur.Lexeme == "{"
}

func isFunctionExpr(expr Expr) bool {
	_, ok := expr.(*FunctionExpr)
	return ok
}

func isInfixOperator(token lexer.Token) bool {
	if token.Type != lexer.OP {
		return false
	}

	infixOps := map[string]bool{
		"+": true, "-": true, "*": true, "/": true, "%": true,
		">=": true, "<=": true, ">": true, "<": true,
		"==": true, "!=": true, "&&": true, "||": true,
		"=": true, "+=": true, "-=": true, "*=": true, "/=": true,
		"++": true, "--": true,
	}
	return infixOps[token.Lexeme]
}

func precedenceOf(op string) int {
	if precedence, ok := precedences[op]; ok {
		return precedence
	}
	return LOWEST
}
