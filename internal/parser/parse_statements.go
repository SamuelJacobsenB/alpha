package parser

import "github.com/alpha/internal/lexer"

// ============================
// PARSING DE NÍVEL SUPERIOR
// ============================

// parseTopLevel analisa declarações e statements de nível superior
func (p *Parser) parseTopLevel() Stmt {
	p.skipSemicolons()

	if p.cur.Type == lexer.EOF {
		return nil
	}

	switch p.cur.Lexeme {
	case "package":
		return p.parsePackageDecl()
	case "import":
		return p.parseImportDecl()
	case "export":
		return p.parseExportDecl()
	case "var":
		return p.parseAndConsume(p.parseVarDecl)
	case "const":
		return p.parseAndConsume(p.parseConstDecl)
	case "struct":
		return p.parseStructDecl(nil)
	case "implement":
		return p.parseImplementDecl()
	case "type":
		return p.parseTypeDecl()
	case "generic":
		return p.parseGenericDeclaration()
	case "<":
		return p.parseGenericTopLevel()
	default:
		return p.parseControlOrDefaultStmt()
	}
}

// parseControlOrDefaultStmt decide entre statement de controle ou padrão
func (p *Parser) parseControlOrDefaultStmt() Stmt {
	switch p.cur.Lexeme {
	case "if", "while", "do", "for", "switch", "return", "break", "continue":
		return p.parseControlStmt()
	default:
		return p.parseDefaultStmt()
	}
}

// parseAndConsume executa função de parsing e consome ponto-e-vírgula opcional
func (p *Parser) parseAndConsume(fn func() Stmt) Stmt {
	stmt := fn()
	if stmt != nil && p.cur.Lexeme == ";" {
		p.advanceToken()
	}
	return stmt
}

// skipSemicolons consome pontos e vírgulas consecutivos
func (p *Parser) skipSemicolons() {
	for p.cur.Lexeme == ";" {
		p.advanceToken()
	}
}

// parseDefaultStmt lida com declarações de função e statements de expressão
func (p *Parser) parseDefaultStmt() Stmt {
	// Caso 1: Declaração de função (tipo seguido de "function")
	if isTypeKeyword(p.cur.Lexeme) && p.nxt.Lexeme == "function" {
		return p.parseFunctionDecl(false)
	}

	// Caso 2: Declaração de variável tipada
	if p.seemsLikeTypeDeclaration() {
		return p.parseAndConsume(p.parseTypedVarDecl)
	}

	// Caso 3: Statement de expressão
	return p.parseAndConsume(p.parseExprStmt)
}

// seemsLikeTypeDeclaration verifica se parece com declaração de tipo
func (p *Parser) seemsLikeTypeDeclaration() bool {
	return isTypeKeyword(p.cur.Lexeme) ||
		(p.cur.Type == lexer.IDENT && p.nxt.Type == lexer.IDENT) ||
		(p.cur.Type == lexer.IDENT && p.nxt.Lexeme == "<") ||
		(p.cur.Type == lexer.IDENT && p.nxt.Lexeme == "[")
}

// ============================
// STATEMENTS DE EXPRESSÃO
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
// STATEMENTS DE CONTROLE
// ============================

// parseControlStmt roteia para parser específico de controle
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
		return nil
	}
}

// ============================
// CONDICIONAIS (IF/SWITCH)
// ============================

// parseIf analisa uma declaração if-else
func (p *Parser) parseIf() Stmt {
	p.advanceToken() // consome 'if'

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

// ============================
// SWITCH STATEMENT
// ============================

// parseSwitch analisa uma declaração switch
func (p *Parser) parseSwitch() Stmt {
	p.advanceToken() // consome 'switch'

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
// LOOPS (FOR/WHILE/DO-WHILE)
// ============================

// parseWhile analisa um loop while
func (p *Parser) parseWhile() Stmt {
	p.advanceToken() // consome 'while'

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	return &WhileStmt{Cond: cond, Body: p.parseBlockLike()}
}

// parseDoWhile analisa um loop do-while
func (p *Parser) parseDoWhile() Stmt {
	p.advanceToken() // consome 'do'

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
	p.advanceToken() // consome 'for'

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

	init := p.parseForLoopInitializer()

	if !p.expectAndConsume(";") {
		p.syncTo(";")
		return nil
	}

	cond := p.parseForLoopCondition()

	if !p.expectAndConsume(";") {
		p.syncTo(";")
		return nil
	}

	post := p.parseForLoopPost()

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

// parseForLoopInitializer analisa inicializador do for loop
func (p *Parser) parseForLoopInitializer() Stmt {
	if p.cur.Lexeme == ";" {
		return nil
	}

	if p.cur.Lexeme == "var" {
		return p.parseVarDecl()
	}

	savedCur, savedNxt := p.cur, p.nxt
	decl := p.parseTypedVarDecl()
	if decl != nil {
		return decl
	}

	p.cur, p.nxt = savedCur, savedNxt
	return p.parseExprStmt()
}

// parseForLoopCondition analisa condição do for loop
func (p *Parser) parseForLoopCondition() Expr {
	if p.cur.Lexeme == ";" {
		return nil
	}
	return p.parseExpression(LOWEST)
}

// parseForLoopPost analisa expressão de pós-iteração
func (p *Parser) parseForLoopPost() Stmt {
	if p.cur.Lexeme == ")" {
		return nil
	}

	expr := p.parseExpression(LOWEST)
	if expr != nil {
		return &ExprStmt{Expr: expr}
	}
	return nil
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

	index, item := p.parseForInIdentifiers()

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

// parseForInIdentifiers parseia identificadores do for-in
func (p *Parser) parseForInIdentifiers() (*Identifier, *Identifier) {
	firstIdent := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()

	if p.cur.Lexeme == "," {
		p.advanceToken()

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected second identifier in for-in loop")
			return nil, nil
		}

		secondIdent := &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()
		return firstIdent, secondIdent
	}

	return nil, firstIdent
}

// isForInLoop verifica se é um for-in loop
func (p *Parser) isForInLoop() bool {
	return p.cur.Lexeme == "(" &&
		p.nxt.Type == lexer.IDENT &&
		!isTypeKeyword(p.nxt.Lexeme) &&
		p.nxt.Lexeme != "var"
}

// ============================
// STATEMENTS DE FLUXO
// ============================

// parseReturn analisa um statement de retorno
func (p *Parser) parseReturn() Stmt {
	p.advanceToken() // consome 'return'

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

// isAtEndOfStatement verifica fim de statement
func (p *Parser) isAtEndOfStatement() bool {
	return p.cur.Lexeme == ";" || p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF
}

// ============================
// FUNÇÕES AUXILIARES
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
	for !p.isBlockEnd() {
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
// FUNÇÕES DE VERIFICAÇÃO
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

// isBlockEnd verifica se chegou ao fim de um bloco
func (p *Parser) isBlockEnd() bool {
	return p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF
}
