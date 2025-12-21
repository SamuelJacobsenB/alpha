package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseTopLevel() Stmt {
	// Ignorar semicolons soltos
	for p.cur.Lexeme == ";" {
		p.advanceToken()
	}

	if p.cur.Type == lexer.EOF {
		return nil
	}

	switch p.cur.Lexeme {
	case "var":
		return p.parseAndConsume(p.parseVarDecl)
	case "const":
		return p.parseAndConsume(p.parseConstDecl)
	case "class":
		return p.parseClass()
	case "type":
		return p.parseTypeDecl()
	case "if", "while", "do", "for", "switch", "return", "break", "continue":
		return p.parseControlStmt()
	case "generic":
		return p.parseGenericFunctionDecl() // NOVO CASO
	default:
		return p.parseDefaultStmt()
	}
}

// ... resto do código permanece igual

func (p *Parser) parseAndConsume(fn func() Stmt) Stmt {
	stmt := fn()
	if stmt != nil && p.cur.Lexeme == ";" {
		p.advanceToken()
	}
	return stmt
}

func (p *Parser) parseDefaultStmt() Stmt {
	if isTypeKeyword(p.cur.Lexeme) {
		if p.nxt.Lexeme == "function" {
			return p.parseFunctionDecl(false)
		}
		return p.parseAndConsume(p.parseTypedVarDecl)
	}
	return p.parseAndConsume(p.parseExprStmt)
}

func (p *Parser) parseExprStmt() Stmt {
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}
	return &ExprStmt{Expr: expr}
}

func (p *Parser) parseControlStmt() Stmt {
	switch p.cur.Lexeme {
	case "if":
		return p.parseIf()
	case "while":
		return p.parseWhile()
	case "do":
		return p.parseDoWhile()
	case "for":
		return p.parseFor()
	case "switch":
		return p.parseSwitch()
	case "return":
		return p.parseReturn()
	case "break":
		return p.parseBreak()
	case "continue":
		return p.parseContinue()
	}
	return nil
}

func (p *Parser) parseIf() Stmt {
	p.advanceToken()

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	thenBlock := p.parseBlockLike()
	elseBlock := p.parseOptionalElse()

	return &IfStmt{Cond: cond, Then: thenBlock, Else: elseBlock}
}

func (p *Parser) parseWhile() Stmt {
	p.advanceToken()

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	return &WhileStmt{Cond: cond, Body: p.parseBlockLike()}
}

func (p *Parser) parseDoWhile() Stmt {
	p.advanceToken()

	body := p.parseBlockLike()
	if body == nil || !p.expectAndConsume("while") {
		p.errorf("expected 'while' after do block")
		return nil
	}

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	return &DoWhileStmt{Body: body, Cond: cond}
}

func (p *Parser) parseFor() Stmt {
	p.advanceToken()

	if p.isForInLoop() {
		return p.parseForIn()
	}
	return p.parseForTraditional()
}

func (p *Parser) parseForTraditional() Stmt {
	if !p.expectAndConsume("(") {
		return nil
	}

	var init Stmt
	if p.cur.Lexeme != ";" {
		init = p.parseForLoopInitializer()
	}

	if !p.expectAndConsume(";") && p.cur.Lexeme != ";" {
		p.syncTo(";")
	}

	var cond Expr
	if p.cur.Lexeme != ";" {
		cond = p.parseExpression(LOWEST)
	}

	if !p.expectAndConsume(";") && p.cur.Lexeme != ";" {
		p.syncTo(";")
	}

	var post Stmt
	if p.cur.Lexeme != ")" {
		postExpr := p.parseExpression(LOWEST)
		if postExpr != nil {
			post = &ExprStmt{Expr: postExpr}
		}
	}

	if !p.expectAndConsume(")") {
		p.syncTo(")")
	}

	body := p.parseBlockLike()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseForIn() Stmt {
	if !p.expectAndConsume("(") {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier in for-in loop")
		return nil
	}

	firstIdent := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()

	var index, item *Identifier
	if p.cur.Lexeme == "," {
		p.advanceToken()

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected second identifier in for-in loop")
			return nil
		}

		index = firstIdent
		item = &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()
	} else {
		index = nil
		item = firstIdent
	}

	if !p.expectAndConsume("in") {
		return nil
	}

	iterable := p.parseExpression(LOWEST)
	if iterable == nil || !p.expectAndConsume(")") {
		return nil
	}

	body := p.parseBlockLike()
	return &ForInStmt{Index: index, Item: item, Iterable: iterable, Body: body}
}

func (p *Parser) parseForLoopInitializer() Stmt {
	if p.cur.Lexeme == ";" {
		return nil
	}

	if p.cur.Lexeme == "var" {
		return p.parseVarDecl()
	}

	if isTypeKeyword(p.cur.Lexeme) {
		savedCur, savedNxt := p.cur, p.nxt
		decl := p.parseTypedVarDecl()
		if decl != nil {
			return decl
		}
		p.cur, p.nxt = savedCur, savedNxt
	}

	return p.parseExprStmt()
}

func (p *Parser) isForInLoop() bool {
	return p.cur.Lexeme == "(" &&
		p.nxt.Type == lexer.IDENT &&
		!isTypeKeyword(p.nxt.Lexeme) &&
		p.nxt.Lexeme != "var"
}

func (p *Parser) parseSwitch() Stmt {
	p.advanceToken()

	cond := p.parseCondition()
	if cond == nil || !p.expectAndConsume("{") {
		p.errorf("expected '{' after switch condition")
		return nil
	}

	cases := make([]*CaseClause, 0, 3)
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		caseClause := p.parseCaseClause()
		if caseClause != nil {
			cases = append(cases, caseClause)
		}
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
		p.advanceToken()
		value = p.parseExpression(LOWEST)
		if value == nil {
			p.errorf("expected expression after 'case'")
			return nil
		}
	case "default":
		p.advanceToken()
		value = nil
	default:
		p.errorf("expected 'case' or 'default', got '%s'", p.cur.Lexeme)
		return nil
	}

	if !p.expectAndConsume(":") {
		return nil
	}

	body := make([]Stmt, 0, 3)
	for !p.isCaseOrBlockEnd() {
		stmt := p.parseTopLevel()
		if stmt != nil {
			body = append(body, stmt)
		}
	}

	return &CaseClause{Value: value, Body: body}
}

func (p *Parser) isCaseOrBlockEnd() bool {
	return p.cur.Lexeme == "case" ||
		p.cur.Lexeme == "default" ||
		p.cur.Lexeme == "}" ||
		p.cur.Type == lexer.EOF
}

func (p *Parser) parseReturn() Stmt {
	p.advanceToken()

	if p.isAtEndOfStatement() {
		p.consumeOptionalSemicolon()
		return &ReturnStmt{Value: nil}
	}

	stmt := &ReturnStmt{Value: p.parseExpression(LOWEST)}
	p.consumeOptionalSemicolon()
	return stmt
}

func (p *Parser) parseCondition() Expr {
	if !p.expectAndConsume("(") {
		p.errorf("expected '(' after %s", p.cur.Lexeme)
		return nil
	}

	cond := p.parseExpression(LOWEST)
	if cond == nil {
		p.errorf("expected condition expression")
		p.syncToNextStmt()
		return nil
	}

	if !p.expectAndConsume(")") {
		p.errorf("expected ')' after condition")
		return nil
	}

	return cond
}

func (p *Parser) parseOptionalElse() []Stmt {
	if p.cur.Lexeme == "else" {
		p.advanceToken()
		return p.parseBlockLike()
	}
	return nil
}

func (p *Parser) parseBreak() Stmt {
	p.advanceToken()
	p.consumeOptionalSemicolon()
	return &BreakStmt{}
}

func (p *Parser) parseContinue() Stmt {
	p.advanceToken()
	p.consumeOptionalSemicolon()
	return &ContinueStmt{}
}

func (p *Parser) parseBlockLike() []Stmt {
	if p.cur.Lexeme == "{" {
		return p.parseBlock()
	}

	if stmt := p.parseTopLevel(); stmt != nil {
		return []Stmt{stmt}
	}

	return nil
}

func (p *Parser) parseBlock() []Stmt {
	if !p.expectAndConsume("{") {
		return nil
	}

	stmts := make([]Stmt, 0, 5)
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		stmt := p.parseTopLevel()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return stmts
}
