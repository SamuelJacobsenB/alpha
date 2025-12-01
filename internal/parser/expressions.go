package parser

import (
	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseExpression(precedence int) Expr {
	left := p.parsePrimary()
	if left == nil {
		return nil
	}

	for {
		curOp := p.cur.Lexeme

		if curOp == "?" {
			if TERNARY < precedence {
				return left
			}
			left = p.parseTernary(left)
			continue
		}

		curPrec := p.precedenceOf(curOp)

		switch {
		case curOp == "(":
			left = p.parseCall(left)
		case curOp == "[":
			left = p.parseIndex(left)
		case curOp == ".":
			left = p.parseMemberAccess(left)
		case curOp == "?":
			left = p.parseTernary(left)
		case p.isInfixOperator(p.cur):
			left = p.parseInfix(left, curPrec)
		case curOp == "<" && p.isGenericCall():
			left = p.parseGenericCall(left)
		case p.isPostfixOperator(p.cur):
			left = p.parsePostfix(left, curPrec)
		default:
			return left
		}
	}
}

func (p *Parser) parsePrimary() Expr {
	switch p.cur.Type {
	case lexer.IDENT:
		ident := &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()
		return ident

	case lexer.INT, lexer.FLOAT:
		return p.parseNumberToken(p.cur)

	case lexer.STRING:
		str := &StringLiteral{Value: p.cur.Value}
		p.advanceToken()
		return str

	case lexer.KEYWORD:
		return p.parseKeywordExpr()

	case lexer.OP:
		return p.parseOperatorExpr()

	default:
		p.advanceToken()
		return nil
	}
}

func (p *Parser) parseKeywordExpr() Expr {
	switch p.cur.Lexeme {
	case "true", "false":
		val := p.cur.Lexeme == "true"
		p.advanceToken()
		return &BoolLiteral{Value: val}
	case "null":
		p.advanceToken()
		return &NullLiteral{}
	case "new":
		return p.parseNewExpr()
	case "this":
		p.advanceToken()
		return &ThisExpr{}
	default:
		if isTypeKeyword(p.cur.Lexeme) && p.nxt.Lexeme == "function" {
			return p.parseFunctionExpr()
		}
		p.errorf("unexpected keyword: %s", p.cur.Lexeme)
		return nil
	}
}

func (p *Parser) parseOperatorExpr() Expr {
	switch p.cur.Lexeme {
	case "(":
		return p.parseParenthesizedExpr()
	case "&":
		return p.parseReferenceExpr()
	case "{":
		// Pode ser map literal ou struct literal
		if p.isStructLiteral() {
			return p.parseStructLiteral()
		}
		return p.parseMapLiteral()
	case "[":
		return p.parseArrayLiteral()
	case "-", "!", "+", "++", "--":
		return p.parsePrefixExpr()
	default:
		p.errorf("unexpected operator: %s", p.cur.Lexeme)
		return nil
	}
}

func (p *Parser) parseParenthesizedExpr() Expr {
	p.advanceToken()
	expr := p.parseExpression(LOWEST)
	if !p.expectAndConsume(")") {
		return nil
	}
	return expr
}

func (p *Parser) parseReferenceExpr() Expr {
	p.advanceToken()
	expr := p.parseExpression(PREFIX)
	if expr == nil {
		return nil
	}
	return &ReferenceExpr{Expr: expr}
}

func (p *Parser) parseArrayLiteral() Expr {
	p.advanceToken()
	elements := p.parseArrayElements()
	if !p.expectAndConsume("]") {
		return nil
	}
	return &ArrayLiteral{Elements: elements}
}

func (p *Parser) parseArrayElements() []Expr {
	var elements []Expr
	for p.cur.Lexeme != "]" && p.cur.Type != lexer.EOF {
		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		elements = append(elements, elem)
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != "]" {
			p.errorf("expected ',' or ']'")
			return nil
		}
	}
	return elements
}

func (p *Parser) parseMapLiteral() Expr {
	p.advanceToken()
	var entries []*MapEntry
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		key := p.parseExpression(LOWEST)
		if key == nil {
			return nil
		}
		if !p.expectAndConsume(":") {
			return nil
		}
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
		entries = append(entries, &MapEntry{Key: key, Value: value})
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != "}" {
			p.errorf("expected ',' or '}'")
			return nil
		}
	}
	if !p.expectAndConsume("}") {
		return nil
	}
	return &MapLiteral{Entries: entries}
}

func (p *Parser) parsePrefixExpr() Expr {
	op := p.cur.Lexeme
	p.advanceToken()
	expr := p.parseExpression(PREFIX)
	if expr == nil {
		return nil
	}
	return &UnaryExpr{Op: op, Expr: expr, Postfix: false}
}
