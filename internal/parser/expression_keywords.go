package parser

func (p *Parser) parseKeywordExpression() Expr {
	switch p.cur.Lexeme {
	case "true":
		return p.parseBoolean(true)
	case "false":
		return p.parseBoolean(false)
	case "null":
		return p.parseNull()
	default:
		return p.parseFunctionExpressionOrError()
	}
}

func (p *Parser) parseBoolean(value bool) Expr {
	p.advanceToken()
	return &BoolLiteral{Value: value}
}

func (p *Parser) parseNull() Expr {
	p.advanceToken()
	return &NullLiteral{}
}

func (p *Parser) parseFunctionExpressionOrError() Expr {
	if isTypeKeyword(p.cur.Lexeme) && p.nxt.Lexeme == "function" {
		return p.parseFunctionExpr()
	}

	p.errorf("unexpected keyword %q at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
	return nil
}
