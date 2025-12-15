package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseTopLevel() Stmt {
	// Ignorar semicolons soltos no início
	for p.cur.Lexeme == ";" {
		p.advanceToken()
	}

	if p.cur.Type == lexer.EOF {
		return nil
	}

	var stmt Stmt

	switch p.cur.Lexeme {
	case "var":
		stmt = p.parseVarDecl()
	case "const":
		stmt = p.parseConstDecl()
	case "class":
		return p.parseClass()
	case "type":
		return p.parseTypeDecl()
	case "if":
		return p.parseIf()
	case "while", "do", "for", "switch", "return", "break", "continue":
		return p.parseControlStmt()
	case "<":
		// Pode ser uma função genérica: <T> T function identity(...)
		return p.parseGenericFunctionDecl()
	default:
		if isTypeKeyword(p.cur.Lexeme) {
			if p.nxt.Lexeme == "function" {
				return p.parseFunctionDecl(false)
			} else {
				stmt = p.parseTypedVarDecl()
				if stmt == nil {
					return nil
				}
			}
		} else {
			stmt = p.parseExprStmt()
		}
	}

	if p.cur.Lexeme == ";" {
		p.advanceToken()
	}
	return stmt
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
	default:
		p.advanceToken()
		return nil
	}
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
	p.advanceToken() // consume 'do'

	body := p.parseBlockLike()
	if body == nil {
		return nil
	}

	if !p.expectAndConsume("while") {
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
	p.advanceToken() // consume 'for'

	if p.isForInLoop() {
		return p.parseForIn()
	}

	return p.parseForTraditional()
}

func (p *Parser) parseForTraditional() Stmt {
	if !p.expectAndConsume("(") {
		return nil
	}

	// Parse initializer (pode ser nil se começar com ';')
	var init Stmt
	if p.cur.Lexeme != ";" {
		init = p.parseForLoopInitializer()
		// Se não conseguir parsear o initializer, tenta sincronizar
		if init == nil && p.cur.Lexeme != ";" {
			// Avança até encontrar ';' ou ')'
			for p.cur.Lexeme != ";" && p.cur.Lexeme != ")" && p.cur.Type != lexer.EOF {
				p.advanceToken()
			}
		}
	}

	// Consumir o ';' após o initializer
	if !p.expectAndConsume(";") {
		// Se não encontrar ';', tenta continuar de qualquer forma
		if p.cur.Lexeme != ")" && p.cur.Type != lexer.EOF {
			p.advanceToken() // Avança e tenta continuar
		}
	}

	// Parse condition (pode ser nil se começar com ';')
	var cond Expr
	if p.cur.Lexeme != ";" {
		cond = p.parseExpression(LOWEST)
		// Se falhar, cond será nil, o que é aceitável
	}

	if !p.expectAndConsume(";") {
		if p.cur.Lexeme != ")" && p.cur.Type != lexer.EOF {
			p.advanceToken()
		}
	}

	// Parse post statement (pode ser nil se começar com ')')
	var post Stmt
	if p.cur.Lexeme != ")" {
		postExpr := p.parseExpression(LOWEST)
		if postExpr != nil {
			post = &ExprStmt{Expr: postExpr}
		}
	}

	if !p.expectAndConsume(")") {
		// Tenta recuperar procurando o ')'
		for p.cur.Lexeme != ")" && p.cur.Type != lexer.EOF {
			p.advanceToken()
		}
		if p.cur.Lexeme == ")" {
			p.advanceToken()
		}
	}

	body := p.parseBlockLike()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseForIn() Stmt {
	if !p.expectAndConsume("(") {
		return nil
	}

	// Parse dos identificadores
	var index *Identifier
	var item *Identifier

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier in for-in loop")
		return nil
	}

	// Primeiro identificador
	firstIdent := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()

	if p.cur.Lexeme == "," {
		// Temos dois identificadores: índice e item
		p.advanceToken() // consume ','

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected second identifier in for-in loop")
			return nil
		}

		index = firstIdent
		item = &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()
	} else {
		// Temos apenas um identificador: item
		index = nil
		item = firstIdent
	}

	if !p.expectAndConsume("in") {
		return nil
	}

	iterable := p.parseExpression(LOWEST)
	if iterable == nil {
		return nil
	}

	if !p.expectAndConsume(")") {
		return nil
	}

	body := p.parseBlockLike()
	return &ForInStmt{Index: index, Item: item, Iterable: iterable, Body: body}
}

func (p *Parser) parseForLoopInitializer() Stmt {
	// Proteger contra tentar parsear ';' como statement
	if p.cur.Lexeme == ";" {
		return nil
	}

	// Tenta como declaração com 'var'
	if p.cur.Lexeme == "var" {
		return p.parseVarDecl()
	}

	// Se for uma palavra-chave de tipo, tenta parsear como declaração tipada
	if isTypeKeyword(p.cur.Lexeme) {
		// Salva o estado atual, porque parseTypedVarDecl pode falhar e consumir tokens
		savedCur := p.cur
		savedNxt := p.nxt

		decl := p.parseTypedVarDecl()
		if decl != nil {
			return decl
		}

		// Se falhou, restaura e tenta como expressão
		p.cur = savedCur
		p.nxt = savedNxt
		return p.parseExprStmt()
	}

	// Por fim, tenta como expressão
	return p.parseExprStmt()
}

func (p *Parser) isForInLoop() bool {
	if p.cur.Lexeme != "(" {
		return false
	}

	// Se o próximo token for uma palavra-chave de tipo, não é for-in.
	if isTypeKeyword(p.nxt.Lexeme) {
		return false
	}

	// Se o próximo token for "var", também não é for-in.
	if p.nxt.Lexeme == "var" {
		return false
	}

	return p.nxt.Type == lexer.IDENT
}

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
	p.advanceToken() // consume 'break'
	p.consumeOptionalSemicolon()
	return &BreakStmt{}
}

func (p *Parser) parseContinue() Stmt {
	p.advanceToken() // consume 'continue'
	p.consumeOptionalSemicolon()
	return &ContinueStmt{}
}

func (p *Parser) parseBlockLike() []Stmt {
	if p.cur.Lexeme == "{" {
		return p.parseBlock()
	}

	// Single statement
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
		if stmt := p.parseTopLevel(); stmt != nil {
			stmts = append(stmts, stmt)
		} else {
			p.advanceToken()
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return stmts
}
