package parser

func (p *Parser) parseOperatorExpression() Expr {
	switch p.cur.Lexeme {
	case "-", "!", "+", "++", "--":
		return p.parsePrefixExpression()
	case "(":
		return p.parseParenthesizedExpression()
	case "{":
		return p.parseArrayLiteral()
	default:
		return p.handleUnknownOperator()
	}
}

func (p *Parser) parsePrefixExpression() Expr {
	op := p.cur.Lexeme
	p.advanceToken()

	right := p.parseExpression(PREFIX)
	if right == nil {
		return nil
	}

	return &UnaryExpr{Op: op, Expr: right, Postfix: false}
}

func (p *Parser) parseParenthesizedExpression() Expr {
	p.advanceToken() // consume '('

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	if !p.expectAndConsume(")") {
		p.errorf("expected ')' after subexpression at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	return expr
}

func (p *Parser) handleUnknownOperator() Expr {
	if p.cur.Lexeme == "{" {
		return nil
	}

	p.errorf("unexpected operator %q at start of expression at %d:%d",
		p.cur.Lexeme, p.cur.Line, p.cur.Col)
	return nil
}
