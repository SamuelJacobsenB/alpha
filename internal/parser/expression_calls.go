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
