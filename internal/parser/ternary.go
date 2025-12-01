package parser

func (p *Parser) parseTernary(cond Expr) Expr {
	p.advanceToken() // consume '?'

	// Parseia expressão true com LOWEST para capturar tudo até ':'
	trueExpr := p.parseExpression(LOWEST)
	if trueExpr == nil {
		p.errorf("expected expression after '?'")
		return nil
	}

	// ✅ Verifica e consome ':' manualmente (não via precedência)
	if p.cur.Lexeme != ":" {
		p.errorf("expected ':' in ternary expression, got '%s'", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume ':'

	// Parseia expressão false com precedência TERNARY
	falseExpr := p.parseExpression(TERNARY)
	if falseExpr == nil {
		p.errorf("expected expression after ':' in ternary")
		return nil
	}

	return &TernaryExpr{
		Cond:      cond,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}
