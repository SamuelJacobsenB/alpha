package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseExpression(precedence int) Expr {
	switch p.cur.Type {
	case lexer.IDENT, lexer.INT, lexer.FLOAT, lexer.STRING, lexer.KEYWORD, lexer.OP:
	default:
		if p.cur.Type != lexer.EOF && p.cur.Lexeme != "" {
			p.errorf("unexpected token %q in expression at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
		}
		return nil
	}

	var left Expr

	switch p.cur.Type {
	case lexer.IDENT:
		left = &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()
	case lexer.INT, lexer.FLOAT:
		left = p.parseNumberToken(p.cur)
		p.advanceToken()
	case lexer.STRING:
		left = &StringLiteral{Value: p.cur.Value}
		p.advanceToken()
	case lexer.KEYWORD:
		switch p.cur.Lexeme {
		case "true":
			left = &BoolLiteral{Value: true}
		case "false":
			left = &BoolLiteral{Value: false}
		case "null":
			left = &NullLiteral{}
		default:
			// Verificar se é uma função anônima
			if isTypeKeyword(p.cur.Lexeme) && p.nxt.Lexeme == "function" {
				left = p.parseFunctionExpr()
			} else {
				p.errorf("unexpected keyword %q at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
				return nil
			}
		}
		if left != nil && !isFunctionExpr(left) {
			p.advanceToken()
		}
	case lexer.OP:
		switch p.cur.Lexeme {
		case "-", "!", "+", "++", "--": // operadores prefixo
			op := p.cur.Lexeme
			p.advanceToken()
			right := p.parseExpression(PREFIX)
			if right == nil {
				return nil
			}
			left = &UnaryExpr{Op: op, Expr: right, Postfix: false}
		case "(":
			p.advanceToken()
			left = p.parseExpression(LOWEST)
			if left == nil {
				return nil
			}
			if p.cur.Lexeme == ")" {
				p.advanceToken()
			} else {
				p.errorf("expected ')' after subexpression at %d:%d", p.cur.Line, p.cur.Col)
				return nil
			}
		case "{":
			left = p.parseArrayLiteral()
		default:
			if p.cur.Lexeme == "{" {
				return nil
			}
			p.errorf("unexpected operator %q at start of expression at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
			return nil
		}
	}

	if left == nil {
		return nil
	}

	if p.cur.Type == lexer.EOF {
		return left
	}

	return p.parsePostfixExpression(left, precedence)
}

func (p *Parser) parsePostfixExpression(left Expr, precedence int) Expr {
	for {
		// Chamada de função
		if p.cur.Lexeme == "(" {
			left = p.parseCallExpression(left)
			continue
		}

		// Indexação de array
		if p.cur.Lexeme == "[" {
			left = p.parseIndexExpression(left)
			continue
		}

		// Operadores pós-fixos
		if p.cur.Lexeme == "++" || p.cur.Lexeme == "--" {
			op := p.cur.Lexeme
			p.advanceToken()
			left = &UnaryExpr{Op: op, Expr: left, Postfix: true}
			continue
		}

		// Operadores infix
		if isInfixOperator(p.cur) {
			op := p.cur.Lexeme
			pPrec := precedenceOf(op)
			if pPrec < precedence {
				break
			}

			p.advanceToken()
			right := p.parseExpression(pPrec)
			if right == nil {
				return nil
			}

			if op == "=" {
				left = &AssignExpr{Left: left, Right: right}
			} else {
				left = &BinaryExpr{Left: left, Op: op, Right: right}
			}
			continue
		}

		break
	}

	return left
}

// parseCallExpression parseia chamadas de função
func (p *Parser) parseCallExpression(left Expr) Expr {
	fmt.Printf("parseCallExpression: starting, left=%T\n", left)
	p.advanceToken() // consume '('

	args := []Expr{}
	if p.cur.Lexeme != ")" {
		for {
			arg := p.parseExpression(LOWEST)
			if arg == nil {
				return nil
			}
			args = append(args, arg)

			if p.cur.Lexeme == "," {
				p.advanceToken()
				continue
			}
			break
		}
	}

	if p.cur.Lexeme == ")" {
		p.advanceToken()
	} else {
		p.errorf("expected ')' to close function call at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	return &CallExpr{Callee: left, Args: args}
}

// Função auxiliar para verificar se é expressão de função
func isFunctionExpr(expr Expr) bool {
	_, ok := expr.(*FunctionExpr)
	return ok
}

// Novas funções auxiliares
func isPrefixOperator(op string) bool {
	return op == "-" || op == "!" || op == "++" || op == "--"
}

func isPostfixOperator(op string) bool {
	return op == "++" || op == "--"
}

func (p *Parser) isAtEnd() bool {
	return p.cur.Type == lexer.EOF
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
		"++": true, "--": true, // ADICIONAR ESTAS LINHAS
	}
	return infixOps[token.Lexeme]
}

func isDelimiter(lexeme string) bool {
	switch lexeme {
	case ")", "}", ";", ",", "{", "", "(", "[", "]":
		return true
	default:
		return false
	}
}

func precedenceOf(op string) int {
	if p, ok := precedences[op]; ok {
		return p
	}
	if op == "=" {
		return LOWEST
	}
	return LOWEST
}
