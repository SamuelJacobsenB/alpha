package parser

import "github.com/alpha/internal/lexer"

var infixOperators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true, "%": true,
	">=": true, "<=": true, ">": true, "<": true,
	"==": true, "!=": true, "&&": true, "||": true,
	"=": true, "+=": true, "-=": true, "*=": true, "/=": true,
}

func (p *Parser) isInfixOperator(token lexer.Token) bool {
	return token.Type == lexer.OP && infixOperators[token.Lexeme]
}

func (p *Parser) precedenceOf(op string) int {
	if prec, exists := precedences[op]; exists {
		return prec
	}
	return LOWEST
}

func (p *Parser) parseInfix(left Expr, precedence int) Expr {
	op := p.cur.Lexeme
	p.advanceToken()

	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}

	if op == "=" {
		return &AssignExpr{Left: left, Right: right}
	}
	return &BinaryExpr{Left: left, Op: op, Right: right}
}

func (p *Parser) parseCall(left Expr) Expr {
	p.advanceToken()
	var args []Expr
	if p.cur.Lexeme != ")" {
		args = p.parseArgumentList()
		if args == nil {
			return nil
		}
	}
	if !p.expectAndConsume(")") {
		return nil
	}
	return &CallExpr{Callee: left, Args: args}
}

func (p *Parser) parseArgumentList() []Expr {
	var args []Expr
	for {
		arg := p.parseExpression(LOWEST)
		if arg == nil {
			return nil
		}
		args = append(args, arg)
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else {
			break
		}
	}
	return args
}

func (p *Parser) parseIndex(left Expr) Expr {
	p.advanceToken()
	index := p.parseExpression(LOWEST)
	if index == nil {
		return nil
	}
	if !p.expectAndConsume("]") {
		return nil
	}
	return &IndexExpr{Array: left, Index: index}
}
