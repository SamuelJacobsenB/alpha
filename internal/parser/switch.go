package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseSwitch() Stmt {
	p.advanceToken() // consume 'switch'

	// Parse condition
	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	if !p.expectAndConsume("{") {
		p.errorf("expected '{' after switch condition")
		return nil
	}

	var cases []*CaseClause

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		caseClause := p.parseCaseClause()
		if caseClause == nil {
			// Skip to next case or }
			p.syncToNextCase()
			continue
		}
		cases = append(cases, caseClause)
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &SwitchStmt{Expr: cond, Cases: cases}
}

func (p *Parser) parseCaseClause() *CaseClause {
	var value Expr

	switch p.cur.Lexeme {
	case "case":
		p.advanceToken() // consume 'case'
		value = p.parseExpression(LOWEST)
		if value == nil {
			p.errorf("expected expression after 'case'")
			return nil
		}

	case "default":
		p.advanceToken() // consume 'default'
		value = nil      // default tem valor nil

	default:
		p.errorf("expected 'case' or 'default', got '%s'", p.cur.Lexeme)
		return nil
	}

	if !p.expectAndConsume(":") {
		return nil
	}

	// Parse body até encontrar outro case/default ou }
	var body []Stmt
	for p.cur.Lexeme != "case" && p.cur.Lexeme != "default" &&
		p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {

		stmt := p.parseTopLevel()
		if stmt != nil {
			body = append(body, stmt)
		} else {
			// Se não conseguiu parsear, avança para evitar loop infinito
			p.advanceToken()
		}
	}

	return &CaseClause{Value: value, Body: body}
}

func (p *Parser) syncToNextCase() {
	for p.cur.Lexeme != "case" && p.cur.Lexeme != "default" &&
		p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}
}
