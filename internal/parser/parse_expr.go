package parser

import "github.com/alpha/internal/lexer"

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
			p.errorf("unexpected keyword %q at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
			return nil
		}

		p.advanceToken()
	case lexer.OP:
		switch p.cur.Lexeme {
		case "-", "!", "+":
			op := p.cur.Lexeme
			p.advanceToken()
			right := p.parseExpression(PREFIX)
			if right == nil {
				return nil
			}
			left = &UnaryExpr{Op: op, Expr: right}
		case "(":
			p.advanceToken() // consome '('
			left = p.parseExpression(LOWEST)
			if left == nil {
				return nil
			}

			if p.cur.Lexeme == ")" {
				p.advanceToken() // consome ')'
			} else {
				p.errorf("expected ')' after subexpression at %d:%d", p.cur.Line, p.cur.Col)
				return nil
			}
		default:
			if p.cur.Lexeme == "{" {
				return nil
			}
			p.errorf("unexpected operator %q at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
			return nil
		}
	}

	if p.cur.Type == lexer.EOF || isDelimiter(p.nxt.Lexeme) {
		return left
	}

	for p.cur.Type != lexer.EOF && isInfixOperator(p.cur) {
		op := p.cur.Lexeme
		pPrec := precedenceOf(op)
		if pPrec < precedence {
			break
		}

		// Processar chamada de função
		if op == "(" {
			if _, isIdent := left.(*Identifier); !isIdent {
				break
			}

			p.advanceToken()

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
				p.errorf("expected ')' to close call at %d:%d", p.cur.Line, p.cur.Col)
				return nil
			}

			left = &CallExpr{Callee: left, Args: args}
			continue
		}

		nextPrec := precedenceOf(op)
		if op == "=" {
			nextPrec = LOWEST // assignment tem baixa precedência
		}

		p.advanceToken()
		right := p.parseExpression(nextPrec)
		if right == nil {
			return nil
		}

		if op == "=" {
			left = &AssignExpr{Left: left, Right: right}
		} else {
			left = &BinaryExpr{Left: left, Op: op, Right: right}
		}
	}

	return left
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
	}

	return infixOps[token.Lexeme]
}

func isDelimiter(lexeme string) bool {
	switch lexeme {
	case ")", "}", ";", ",", "{", "":
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
