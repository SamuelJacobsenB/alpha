package parser

import "github.com/alpha/internal/lexer"

// ============================
// Parsing de Nível Superior
// ============================

// parseTopLevel analisa declarações e statements de nível superior
func (p *Parser) parseTopLevel() Stmt {
	// Ignorar pontos-e-vírgulas soltos
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
	case "generic":
		return p.parseGenericFunctionDecl()
	case "if", "while", "do", "for", "switch", "return", "break", "continue":
		return p.parseControlStmt()
	default:
		return p.parseDefaultStmt()
	}
}

// parseAndConsume executa uma função de parsing e consome ponto-e-vírgula opcional
func (p *Parser) parseAndConsume(fn func() Stmt) Stmt {
	stmt := fn()
	if stmt != nil && p.cur.Lexeme == ";" {
		p.advanceToken()
	}
	return stmt
}

// parseDefaultStmt lida com declarações de função e statements de expressão
func (p *Parser) parseDefaultStmt() Stmt {
	if isTypeKeyword(p.cur.Lexeme) {
		if p.nxt.Lexeme == "function" {
			return p.parseFunctionDecl(false)
		}
		return p.parseAndConsume(p.parseTypedVarDecl)
	}
	return p.parseAndConsume(p.parseExprStmt)
}

// ============================
// Statements de Expressão
// ============================

// parseExprStmt analisa um statement de expressão
func (p *Parser) parseExprStmt() Stmt {
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}
	return &ExprStmt{Expr: expr}
}

// ============================
// Statements de Controle
// ============================

// parseControlStmt roteia para o parser específico de controle
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

// ============================
// Condicionais (if/switch)
// ============================

// parseIf analisa uma declaração if-else
func (p *Parser) parseIf() Stmt {
	p.advanceToken() // consume 'if'

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	thenBlock := p.parseBlockLike()
	elseBlock := p.parseOptionalElse()

	return &IfStmt{Cond: cond, Then: thenBlock, Else: elseBlock}
}

// parseOptionalElse analisa um bloco else opcional
func (p *Parser) parseOptionalElse() []Stmt {
	if p.cur.Lexeme != "else" {
		return nil
	}

	p.advanceToken()
	return p.parseBlockLike()
}

// parseSwitch analisa uma declaração switch
func (p *Parser) parseSwitch() Stmt {
	p.advanceToken() // consume 'switch'

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	if !p.expectAndConsume("{") {
		p.errorf("expected '{' after switch condition")
		return nil
	}

	cases := p.parseSwitchCases()
	if cases == nil || !p.expectAndConsume("}") {
		return nil
	}

	return &SwitchStmt{Expr: cond, Cases: cases}
}

// parseSwitchCases analisa todos os casos de um switch
func (p *Parser) parseSwitchCases() []*CaseClause {
	cases := make([]*CaseClause, 0, 3)

	for !p.isAtSwitchEnd() {
		clause := p.parseCaseClause()
		if clause != nil {
			cases = append(cases, clause)
		}
	}

	return cases
}

// parseCaseClause analisa um único caso ou default
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

	return &CaseClause{
		Value: value,
		Body:  p.parseCaseBody(),
	}
}

// parseCaseBody analisa o corpo de um caso
func (p *Parser) parseCaseBody() []Stmt {
	body := make([]Stmt, 0, 3)

	for !p.isAtCaseEnd() {
		stmt := p.parseTopLevel()
		if stmt != nil {
			body = append(body, stmt)
		}
	}

	return body
}

// ============================
// Loops (for/while/do-while)
// ============================

// parseWhile analisa um loop while
func (p *Parser) parseWhile() Stmt {
	p.advanceToken() // consume 'while'

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	return &WhileStmt{Cond: cond, Body: p.parseBlockLike()}
}

// parseDoWhile analisa um loop do-while
func (p *Parser) parseDoWhile() Stmt {
	p.advanceToken() // consume 'do'

	body := p.parseBlockLike()
	if body == nil {
		p.errorf("expected block after 'do'")
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

// parseFor decide entre for tradicional e for-in
func (p *Parser) parseFor() Stmt {
	p.advanceToken() // consume 'for'

	if p.isForInLoop() {
		return p.parseForIn()
	}
	return p.parseForTraditional()
}

// parseForTraditional analisa um for loop tradicional
func (p *Parser) parseForTraditional() Stmt {
	if !p.expectAndConsume("(") {
		return nil
	}

	var init Stmt
	if p.cur.Lexeme != ";" {
		init = p.parseForLoopInitializer()
	}

	if !p.expectAndConsume(";") {
		p.syncTo(";")
		return nil
	}

	var cond Expr
	if p.cur.Lexeme != ";" {
		cond = p.parseExpression(LOWEST)
	}

	if !p.expectAndConsume(";") {
		p.syncTo(";")
		return nil
	}

	var post Stmt
	if p.cur.Lexeme != ")" {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			post = &ExprStmt{Expr: expr}
		}
	}

	if !p.expectAndConsume(")") {
		p.syncTo(")")
		return nil
	}

	return &ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: p.parseBlockLike(),
	}
}

// parseForIn analisa um for-in loop
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

	return &ForInStmt{
		Index:    index,
		Item:     item,
		Iterable: iterable,
		Body:     p.parseBlockLike(),
	}
}

// parseForLoopInitializer analisa o inicializador de um for loop
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

// isForInLoop verifica se o próximo token indica um for-in loop
func (p *Parser) isForInLoop() bool {
	return p.cur.Lexeme == "(" &&
		p.nxt.Type == lexer.IDENT &&
		!isTypeKeyword(p.nxt.Lexeme) &&
		p.nxt.Lexeme != "var"
}

// ============================
// Statements de Fluxo
// ============================

// parseReturn analisa um statement de retorno
func (p *Parser) parseReturn() Stmt {
	p.advanceToken() // consume 'return'

	if p.isAtEndOfStatement() {
		p.consumeOptionalSemicolon()
		return &ReturnStmt{Value: nil}
	}

	value := p.parseExpression(LOWEST)
	p.consumeOptionalSemicolon()

	return &ReturnStmt{Value: value}
}

// parseBreak analisa um statement break
func (p *Parser) parseBreak() Stmt {
	p.advanceToken()
	p.consumeOptionalSemicolon()
	return &BreakStmt{}
}

// parseContinue analisa um statement continue
func (p *Parser) parseContinue() Stmt {
	p.advanceToken()
	p.consumeOptionalSemicolon()
	return &ContinueStmt{}
}

// ============================
// Funções Auxiliares
// ============================

// parseCondition analisa uma condição entre parênteses
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

// parseBlockLike analisa um bloco ou statement único
func (p *Parser) parseBlockLike() []Stmt {
	if p.cur.Lexeme == "{" {
		return p.parseBlock()
	}

	if stmt := p.parseTopLevel(); stmt != nil {
		return []Stmt{stmt}
	}

	return nil
}

// parseBlock analisa um bloco de statements
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

// ============================
// Funções de Verificação
// ============================

// isAtSwitchEnd verifica se chegou ao fim do switch
func (p *Parser) isAtSwitchEnd() bool {
	return p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF
}

// isAtCaseEnd verifica se chegou ao fim de um caso
func (p *Parser) isAtCaseEnd() bool {
	return p.cur.Lexeme == "case" ||
		p.cur.Lexeme == "default" ||
		p.cur.Lexeme == "}" ||
		p.cur.Type == lexer.EOF
}
