package parser

import "fmt"

func (p *Parser) parseCallExpression(left Expr) Expr {
	fmt.Printf("parseCallExpression: starting, left=%T\n", left)
	p.advanceToken() // consume '('

	args := p.parseArgumentList()
	if args == nil {
		return nil
	}

	if !p.expectAndConsume(")") {
		p.errorf("expected ')' to close function call at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	return &CallExpr{Callee: left, Args: args}
}

func (p *Parser) parseArgumentList() []Expr {
	var args []Expr

	if p.cur.Lexeme == ")" {
		return args // empty argument list
	}

	for {
		arg := p.parseExpression(LOWEST)
		if arg == nil {
			return nil
		}
		args = append(args, arg)

		if p.cur.Lexeme != "," {
			break
		}
		p.advanceToken() // consume ','
	}

	return args
}

func (p *Parser) parseIndexExpression(left Expr) Expr {
	fmt.Printf("parseIndexExpression: starting\n")
	p.advanceToken() // consume '['

	index := p.parseExpression(LOWEST)
	if index == nil {
		return nil
	}

	if !p.expectAndConsume("]") {
		p.errorf("expected ']' after index expression")
		return nil
	}

	return &IndexExpr{Array: left, Index: index}
}
