package parser

import "github.com/alpha/internal/lexer"

// parseExpression implementa Pratt parsing com precedência.
// Convenção: ao retornar, p.cur aponta para o último token da expressão construída.
func (p *Parser) parseExpression(precedence int) Expr {
	// pré-fixo: p.cur já contém o token inicial
	var left Expr

	switch p.cur.Type {
	case lexer.IDENT:
		left = &Identifier{Name: p.cur.Lexeme}
	case lexer.INT, lexer.FLOAT:
		left = p.parseNumberToken(p.cur)
	case lexer.STRING:
		left = &StringLiteral{Value: p.cur.Value}
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
	case lexer.OP:
		// prefix operators: - ! +
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
			// consome '(' e posiciona p.cur no primeiro token da subexpressão (ou ')')
			p.advanceToken() // p.cur == primeiro token dentro da subexpr (ou ')')

			// parse subexpressão. Ao retornar, p.cur deve apontar para o último token da subexpr
			left = p.parseExpression(LOWEST)
			if left == nil {
				return nil
			}

			// garantir que o ')' esteja em p.cur ou em p.nxt; deixar p.cur == ')' (último token da expressão)
			if p.cur.Lexeme == ")" {
				// ok: p.cur já é ')'
			} else if p.nxt.Lexeme == ")" {
				p.advanceToken() // posiciona p.cur == ')'
			} else {
				p.errorf("expected ')' after subexpression at %d:%d", p.cur.Line, p.cur.Col)
				return nil
			}
		default:
			p.errorf("unexpected operator %q at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
			return nil
		}
	default:
		p.errorf("unexpected token %q at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
		return nil
	}

	// infix loop: operacoes possíveis estão em p.nxt (ou instanceof keyword)
	for (p.nxt.Type == lexer.OP) || (p.nxt.Type == lexer.KEYWORD && p.nxt.Lexeme == "instanceof") {
		op := p.nxt.Lexeme
		pPrec := precedenceOf(op)
		if pPrec < precedence {
			break
		}

		// call: '(' after expression
		if op == "(" {
			// consume '(' into p.cur (posiciona no '(')
			p.advanceToken() // agora p.cur == "(", p.nxt == primeiro argumento ou ")"

			// se o próximo for ')' então chamada sem argumentos:
			if p.nxt.Lexeme == ")" {
				// posiciona p.cur == ')'
				p.advanceToken()
				// p.cur agora é ')', a chamada tem args vazios
				left = &CallExpr{Callee: left, Args: []Expr{}}
				continue
			}

			// avançar para o primeiro token dentro dos parênteses
			p.advanceToken() // agora p.cur == primeiro arg

			args := []Expr{}
			if p.cur.Lexeme != ")" {
				for {
					// parseExpression espera p.cur no início da expressão do argumento
					arg := p.parseExpression(LOWEST)
					if arg == nil {
						return nil
					}
					args = append(args, arg)

					// se o próximo token for ',', consumi-lo e avançar para o token que inicia o próximo argumento
					if p.nxt.Lexeme == "," {
						// consumir ',' para posicionar p.cur == ','
						p.advanceToken()
						// avançar para o token que inicia o próximo argumento
						p.advanceToken()
						continue
					}
					break
				}
			}

			// agora garantir que fechamento ')' esteja em p.cur ou p.nxt; deixar p.cur == ')'
			if p.cur.Lexeme == ")" {
				// ok
			} else if p.nxt.Lexeme == ")" {
				p.advanceToken() // posiciona p.cur == ')'
			} else {
				p.errorf("expected ')' to close call at %d:%d", p.cur.Line, p.cur.Col)
				return nil
			}

			// deixar p.cur == ')' (contrato: parseExpression termina com p.cur no último token)
			left = &CallExpr{Callee: left, Args: args}
			// continuar loop para permitir chamadas em cadeia ou operadores após a chamada
			continue
		}

		// binary / assignment
		// consume operator into p.cur
		p.advanceToken() // p.cur == operador
		opToken := p.cur.Lexeme
		nextPrec := precedenceOf(opToken)

		// advance to first token of right-hand side
		p.advanceToken() // p.cur == primeiro token do RHS
		right := p.parseExpression(nextPrec + 1)
		if right == nil {
			return nil
		}

		if opToken == "=" {
			left = &AssignExpr{Left: left, Right: right}
		} else {
			left = &BinaryExpr{Left: left, Op: opToken, Right: right}
		}
		// after parsing RHS, p.cur is the last token of RHS (contrato)
	}

	// retorno: p.cur aponta para o último token da expressão construída
	return left
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
