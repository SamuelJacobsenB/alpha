package parser

func (p *Parser) parsePostfixExpression(left Expr, precedence int) Expr {
	for {
		switch {
		case p.cur.Lexeme == "(":
			left = p.parseCallExpression(left)
		case p.cur.Lexeme == "[":
			left = p.parseIndexExpression(left)
		case p.cur.Lexeme == "<":
			left = p.parseGenericCall(left)
		case p.cur.Lexeme == "++" || p.cur.Lexeme == "--":
			left = p.parsePostfixOperator(left)
		case isInfixOperator(p.cur):
			left = p.parseInfixExpression(left, precedence)
			if left == nil {
				return nil
			}
		default:
			return left
		}
	}
}

func (p *Parser) parsePostfixOperator(left Expr) Expr {
	op := p.cur.Lexeme
	p.advanceToken()
	return &UnaryExpr{Op: op, Expr: left, Postfix: true}
}

func (p *Parser) parseInfixExpression(left Expr, precedence int) Expr {
	op := p.cur.Lexeme
	opPrec := precedenceOf(op)

	if opPrec < precedence {
		return left
	}

	p.advanceToken()
	right := p.parseExpression(opPrec)
	if right == nil {
		return nil
	}

	return p.createInfixExpression(left, op, right)
}

func (p *Parser) createInfixExpression(left Expr, op string, right Expr) Expr {
	if op == "=" {
		return &AssignExpr{Left: left, Right: right}
	}
	return &BinaryExpr{Left: left, Op: op, Right: right}
}
