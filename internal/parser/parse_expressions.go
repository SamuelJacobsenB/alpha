package parser

import (
	"strconv"

	"github.com/alpha/internal/lexer"
)

// ============================
// CONSTANTES E CONFIGURAÇÕES
// ============================

const (
	_ int = iota
	LOWEST
	TERNARY
	ASSIGNMENT
	LOGICALOR
	LOGICALAND
	EQUALITY
	COMPARISON
	SUM
	PRODUCT
	PREFIX
	CALL
	MEMBER
	INDEX
	POSTFIX
)

var precedences = map[string]int{
	"?":  TERNARY,
	"=":  ASSIGNMENT,
	"+=": ASSIGNMENT,
	"-=": ASSIGNMENT,
	"*=": ASSIGNMENT,
	"/=": ASSIGNMENT,
	"||": LOGICALOR,
	"&&": LOGICALAND,
	"==": EQUALITY,
	"!=": EQUALITY,
	"<":  COMPARISON,
	">":  COMPARISON,
	"<=": COMPARISON,
	">=": COMPARISON,
	"+":  SUM,
	"-":  SUM,
	"*":  PRODUCT,
	"/":  PRODUCT,
	"%":  PRODUCT,
	"(":  CALL,
	".":  MEMBER,
	"[":  INDEX,
	"++": POSTFIX,
	"--": POSTFIX,
}

// Conjuntos de operadores para verificação rápida
var (
	infixOperators = map[string]bool{
		"+": true, "-": true, "*": true, "/": true, "%": true,
		">=": true, "<=": true, ">": true, "<": true,
		"==": true, "!=": true, "&&": true, "||": true,
		"=": true, "+=": true, "-=": true, "*=": true, "/=": true,
	}

	postfixOperators = map[string]bool{
		"++": true,
		"--": true,
	}

	prefixOperators = map[string]bool{
		"-": true, "!": true, "+": true, "++": true, "--": true, "*": true, "&": true,
	}
)

// ============================
// FUNÇÕES AUXILIARES DE VERIFICAÇÃO
// ============================

// isInfixOperator verifica se o token é um operador infixo
func (p *Parser) isInfixOperator(token lexer.Token) bool {
	return token.Type == lexer.OP && infixOperators[token.Lexeme]
}

// isPostfixOperator verifica se o token é um operador pós-fixo
func (p *Parser) isPostfixOperator(token lexer.Token) bool {
	return token.Type == lexer.OP && postfixOperators[token.Lexeme]
}

// isPrefixOperator verifica se o token é um operador pré-fixo
func (p *Parser) isPrefixOperator(token lexer.Token) bool {
	return token.Type == lexer.OP && prefixOperators[token.Lexeme]
}

// precedenceOf retorna a precedência de um operador
func (p *Parser) precedenceOf(op string) int {
	if prec, exists := precedences[op]; exists {
		return prec
	}
	return LOWEST
}

// ============================
// PARSING DE EXPRESSÕES (PRINCIPAL)
// ============================

// parseExpression implementa o algoritmo de precedência de operadores (Pratt parsing)
func (p *Parser) parseExpression(precedence int) Expr {
	left := p.parsePrimary()
	if left == nil {
		return nil
	}

	for {
		curOp := p.cur.Lexeme

		// Isso permite que o parseTernary recupere o controle ao ver o delimitador
		if curOp == ";" || curOp == ")" || curOp == "}" || curOp == "]" || curOp == "," || curOp == ":" || p.cur.Type == lexer.EOF {
			return left
		}

		// Caso especial para operador ternário
		if curOp == "?" {
			// Se a precedência atual for maior ou igual ao Ternário,
			// paramos aqui para que a instância superior do parser lide com ele.
			if precedence >= TERNARY {
				return left
			}
			left = p.parseTernary(left)
			if left == nil {
				return nil
			}
			continue
		}

		// Verifica se é um operador válido para continuar
		if !p.isValidContinuationOperator() {
			return left
		}

		curPrec := p.precedenceOf(curOp)

		// Verifica precedência
		if curPrec < precedence {
			return left
		}

		// Processa o operador atual
		left = p.processOperator(left, curOp, curPrec)
		if left == nil {
			return nil
		}
	}
}

// isValidContinuationOperator verifica se o token atual permite continuar a expressão
func (p *Parser) isValidContinuationOperator() bool {
	curOp := p.cur.Lexeme
	return p.isInfixOperator(p.cur) || p.isPostfixOperator(p.cur) ||
		curOp == "(" || curOp == "[" || curOp == "."
}

// processOperator processa o operador atual baseado em seu tipo
func (p *Parser) processOperator(left Expr, curOp string, curPrec int) Expr {
	switch {
	case curOp == "(":
		return p.parseCall(left)
	case curOp == "[":
		return p.parseIndex(left)
	case curOp == ".":
		return p.parseMemberAccess(left)
	case p.isInfixOperator(p.cur):
		return p.parseInfix(left, curPrec)
	case p.isPostfixOperator(p.cur):
		return p.parsePostfix(left)
	default:
		return left
	}
}

// ============================
// PARSING DE EXPRESSÕES PRIMÁRIAS
// ============================

// parsePrimary processa expressões primárias (literais, identificadores, etc.)
func (p *Parser) parsePrimary() Expr {
	switch p.cur.Type {
	case lexer.IDENT:
		return p.parseIdentifierExpr()
	case lexer.INT:
		return p.parseIntLiteral()
	case lexer.FLOAT:
		return p.parseFloatLiteral()
	case lexer.STRING:
		return p.parseStringLiteral()
	case lexer.KEYWORD:
		return p.parseKeywordExpr()
	case lexer.OP:
		return p.parseOperatorExpr()
	default:
		p.advanceToken()
		return nil
	}
}

// parseIdentifierExpr processa identificadores com casos especiais
func (p *Parser) parseIdentifierExpr() Expr {
	// Tratamento especial para coleções explícitas (set, map)
	if (p.cur.Lexeme == "set" || p.cur.Lexeme == "map") && (p.nxt.Lexeme == "<" || p.nxt.Lexeme == "{") {
		return p.parseExplicitCollectionLiteral()
	}

	// Struct literal tipado (ex: Point { ... })
	if p.nxt.Lexeme == "{" {
		return p.parseTypedStructLiteral()
	}

	return p.parseIdentifier()
}

// parseIdentifier cria um nó de identificador
func (p *Parser) parseIdentifier() Expr {
	ident := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()
	return ident
}

// parseIntLiteral processa literais inteiros
func (p *Parser) parseIntLiteral() Expr {
	value, err := strconv.ParseInt(p.cur.Lexeme, 10, 64)
	if err != nil {
		p.errorf("invalid integer literal: %s", p.cur.Lexeme)
		return nil
	}

	p.advanceToken()
	return &IntLiteral{Value: value}
}

// parseFloatLiteral processa literais de ponto flutuante
func (p *Parser) parseFloatLiteral() Expr {
	value, err := strconv.ParseFloat(p.cur.Lexeme, 64)
	if err != nil {
		p.errorf("invalid float literal: %s", p.cur.Lexeme)
		return nil
	}

	p.advanceToken()
	return &FloatLiteral{Value: value}
}

// parseStringLiteral processa literais de string
func (p *Parser) parseStringLiteral() Expr {
	str := &StringLiteral{Value: p.cur.Value}
	p.advanceToken()
	return str
}

// parseKeywordExpr processa expressões iniciadas por keywords
func (p *Parser) parseKeywordExpr() Expr {
	switch p.cur.Lexeme {
	case "true", "false":
		return p.parseBoolLiteral()
	case "null":
		p.advanceToken()
		return &NullLiteral{}
	case "self":
		p.advanceToken()
		return &SelfExpr{}
	case "lenght", "append", "remove", "removeIndex":
		// Trata built-ins como identificadores especiais
		return p.parseBuiltinCall()
	default:
		if isTypeKeyword(p.cur.Lexeme) {
			return nil
		}
		p.errorf("unexpected keyword: %s", p.cur.Lexeme)
		return nil
	}
}

// parseBuiltinCall processa chamadas de funções built-in (len, append, delete)
func (p *Parser) parseBuiltinCall() Expr {
	// Salva o nome da built-in
	builtinName := p.cur.Lexeme
	p.advanceToken() // consome o nome da built-in

	// Verifica se é uma chamada de função
	if p.cur.Lexeme != "(" {
		// Se não tem parênteses, retorna como identificador
		return &Identifier{Name: builtinName}
	}

	// Parseia os argumentos
	if !p.expectAndConsume("(") {
		return nil
	}

	args := p.parseArgumentList()
	if !p.expectAndConsume(")") {
		return nil
	}

	// Cria um CallExpr especial (ou poderia ser um BuiltinCallExpr se quiser diferenciar)
	return &CallExpr{
		Callee: &Identifier{Name: builtinName},
		Args:   args,
	}
}

// parseBoolLiteral processa literais booleanos
func (p *Parser) parseBoolLiteral() Expr {
	val := p.cur.Lexeme == "true"
	p.advanceToken()
	return &BoolLiteral{Value: val}
}

// parseOperatorExpr processa expressões iniciadas por operadores
func (p *Parser) parseOperatorExpr() Expr {
	switch p.cur.Lexeme {
	case "(":
		return p.parseParenthesizedExpr()
	case "&":
		return p.parseReferenceExpr()
	case "{":
		return p.parseBraceLiteral()
	case "[":
		return p.parseArrayLiteral()
	case ":":
		return nil
	default:
		if p.isPrefixOperator(p.cur) {
			return p.parsePrefixExpr()
		}
		p.errorf("unexpected operator: %s", p.cur.Lexeme)
		return nil
	}
}

// ============================
// PARSING DE OPERADORES
// ============================

// parsePrefixExpr processa operadores prefixos
func (p *Parser) parsePrefixExpr() Expr {
	op := p.cur.Lexeme
	p.advanceToken()

	expr := p.parseExpression(PREFIX)
	if expr == nil {
		return nil
	}

	return &UnaryExpr{Op: op, Expr: expr, Postfix: false}
}

// parseInfix processa operadores infixos
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

// parsePostfix processa operadores pós-fixos
func (p *Parser) parsePostfix(left Expr) Expr {
	op := p.cur.Lexeme
	p.advanceToken()
	return &UnaryExpr{Op: op, Expr: left, Postfix: true}
}

// parseTernary processa operador ternário (cond ? true : false)
func (p *Parser) parseTernary(cond Expr) Expr {
	p.advanceToken() // consume '?'

	// Parse a expressão verdadeira
	trueExpr := p.parseExpression(LOWEST)
	if trueExpr == nil {
		p.errorf("expected expression after '?'")
		return nil
	}

	// CORREÇÃO: Verificar explicitamente se temos ':' antes de tentar consumir
	if p.cur.Lexeme != ":" {
		p.errorf("expected ':' in ternary expression, got '%s'", p.cur.Lexeme)
		return nil
	}

	// Deve ter ':' após a expressão verdadeira
	if !p.expectAndConsume(":") {
		return nil
	}

	// Parse a expressão falsa com precedência TERNARY
	falseExpr := p.parseExpression(TERNARY)
	if falseExpr == nil {
		p.errorf("expected expression after ':' in ternary")
		return nil
	}

	return &TernaryExpr{
		Cond:      cond,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}

// ============================
// PARSING DE CHAMADAS E ACESSOS
// ============================

// parseCall processa chamadas de função
func (p *Parser) parseCall(left Expr) Expr {
	// Trata GenericCallExpr sem argumentos
	if gce, ok := left.(*GenericCallExpr); ok && gce.Args == nil {
		return p.parseGenericFunctionCall(gce)
	}

	// Chamada normal
	if !p.expectAndConsume("(") {
		return left
	}

	args := p.parseArgumentList()
	if !p.expectAndConsume(")") {
		return nil
	}

	return &CallExpr{
		Callee: left,
		Args:   args,
	}
}

// parseGenericFunctionCall processa chamada de função genérica
func (p *Parser) parseGenericFunctionCall(gce *GenericCallExpr) *GenericCallExpr {
	if !p.expectAndConsume("(") {
		return nil
	}

	var args []Expr
	if p.cur.Lexeme != ")" {
		args = p.parseArgumentList()
	}

	if !p.expectAndConsume(")") {
		return nil
	}

	return &GenericCallExpr{
		Callee:   gce.Callee,
		TypeArgs: gce.TypeArgs,
		Args:     args,
	}
}

// parseArgumentList processa lista de argumentos
func (p *Parser) parseArgumentList() []Expr {
	// VERIFICAÇÃO ADICIONADA: lista de argumentos vazia
	if p.cur.Lexeme == ")" {
		return []Expr{} // Retorna lista vazia
	}

	args := make([]Expr, 0, 3)

	for {
		arg := p.parseExpression(LOWEST)
		if arg == nil {
			return nil
		}
		args = append(args, arg)

		if !p.match(",") {
			break
		}
	}

	return args
}

// match verifica e consome um token se corresponder
func (p *Parser) match(lexeme string) bool {
	if p.cur.Lexeme == lexeme {
		p.advanceToken()
		return true
	}
	return false
}

// parseIndex processa acesso por índice (array[expr])
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

// parseMemberAccess processa acesso a membro (object.member)
func (p *Parser) parseMemberAccess(left Expr) Expr {
	p.advanceToken() // consume '.'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected member name after '.'")
		return nil
	}

	member := p.cur.Lexeme
	p.advanceToken()

	return &MemberExpr{Object: left, Member: member}
}

// ============================
// PARSING DE LITERAIS
// ============================

// parseBraceLiteral processa literais com chaves { ... }
func (p *Parser) parseBraceLiteral() Expr {
	p.advanceToken() // consume '{'

	// Caso vazio
	if p.cur.Lexeme == "}" {
		p.advanceToken()
		return &SetLiteral{Elements: []Expr{}}
	}

	// Parse primeiro elemento para determinar tipo
	firstExpr := p.parseExpression(LOWEST)
	if firstExpr == nil {
		return nil
	}

	// Se tiver ':', é Map ou Struct
	if p.cur.Lexeme == ":" {
		p.advanceToken()
		firstValue := p.parseExpression(LOWEST)
		if firstValue == nil {
			return nil
		}

		// Struct se a chave for identificador simples
		if ident, ok := firstExpr.(*Identifier); ok {
			return p.continueStructLiteral(ident, firstValue)
		}
		return p.continueMapLiteral(firstExpr, firstValue)
	}

	// Caso contrário, é Set
	return p.continueSetLiteral(firstExpr)
}

// parseTypedStructLiteral processa struct literais tipados (ex: Point { ... })
func (p *Parser) parseTypedStructLiteral() Expr {
	p.advanceToken() // consome o nome do tipo
	return p.parseBraceLiteral()
}

// continueStructLiteral continua parsing de struct literal
func (p *Parser) continueStructLiteral(firstKey *Identifier, firstValue Expr) Expr {
	fields := []*StructField{{Name: firstKey.Name, Value: firstValue}}

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == "," {
			p.advanceToken()
			if p.cur.Lexeme == "}" {
				break
			}
		}

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected field name in struct literal, got %s", p.cur.Lexeme)
			return nil
		}

		fieldName := p.cur.Lexeme
		p.advanceToken()

		if !p.expectAndConsume(":") {
			return nil
		}

		val := p.parseExpression(LOWEST)
		if val == nil {
			return nil
		}

		fields = append(fields, &StructField{Name: fieldName, Value: val})
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return &StructLiteral{Fields: fields}
}

// continueMapLiteral continua parsing de map literal
func (p *Parser) continueMapLiteral(firstKey, firstValue Expr) Expr {
	entries := []*MapEntry{{Key: firstKey, Value: firstValue}}

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == "," {
			p.advanceToken()
			if p.cur.Lexeme == "}" {
				break
			}
		}

		key := p.parseExpression(LOWEST)
		if key == nil {
			return nil
		}

		if !p.expectAndConsume(":") {
			return nil
		}

		val := p.parseExpression(LOWEST)
		if val == nil {
			return nil
		}

		entries = append(entries, &MapEntry{Key: key, Value: val})
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return &MapLiteral{Entries: entries}
}

// continueSetLiteral continua parsing de set literal
func (p *Parser) continueSetLiteral(firstElem Expr) Expr {
	elements := []Expr{firstElem}

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == "," {
			p.advanceToken()
			if p.cur.Lexeme == "}" {
				break
			}
		}

		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		elements = append(elements, elem)
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return &SetLiteral{Elements: elements}
}

// parseExplicitCollectionLiteral processa literais explícitos de coleção (set/map)
func (p *Parser) parseExplicitCollectionLiteral() Expr {
	typeName := p.cur.Lexeme
	p.advanceToken() // consome 'map' ou 'set'

	// Consome parâmetros genéricos se existirem
	if p.cur.Lexeme == "<" {
		p.parseGenericType(typeName)
	}

	if p.cur.Lexeme != "{" {
		p.errorf("expected '{' after %s type definition", typeName)
		return nil
	}

	if typeName == "set" {
		return p.parseSetLiteral()
	}
	return p.parseMapLiteral()
}

// parseParenthesizedExpr processa expressões entre parênteses
func (p *Parser) parseParenthesizedExpr() Expr {
	p.advanceToken()

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	if !p.expectAndConsume(")") {
		return nil
	}

	return expr
}

// parseReferenceExpr processa expressões de referência (&expr)
func (p *Parser) parseReferenceExpr() Expr {
	p.advanceToken()

	expr := p.parseExpression(PREFIX)
	if expr == nil {
		return nil
	}

	return &ReferenceExpr{Expr: expr}
}

// parseArrayLiteral processa literais de array
func (p *Parser) parseArrayLiteral() Expr {
	p.advanceToken()

	elements := p.parseArrayElements()
	if elements == nil || !p.expectAndConsume("]") {
		return nil
	}

	return &ArrayLiteral{Elements: elements}
}

// parseArrayElements processa elementos de array
func (p *Parser) parseArrayElements() []Expr {
	elements := make([]Expr, 0, 3)

	for p.cur.Lexeme != "]" && p.cur.Type != lexer.EOF {
		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		elements = append(elements, elem)

		if !p.match(",") {
			if p.cur.Lexeme != "]" {
				p.errorf("expected ',' or ']'")
				return nil
			}
		}
	}

	return elements
}

// parseCollectionLiteral decide entre set e map literal (função de conveniência)
func (p *Parser) parseCollectionLiteral() Expr {
	savedCur, savedNxt := p.cur, p.nxt

	p.advanceToken()

	if p.cur.Lexeme == "}" {
		p.cur, p.nxt = savedCur, savedNxt
		return p.parseSetLiteral()
	}

	tempParser := *p
	expr := tempParser.parseExpression(LOWEST)
	if expr == nil {
		p.cur, p.nxt = savedCur, savedNxt
		return p.parseSetLiteral()
	}

	if tempParser.cur.Lexeme == ":" {
		p.cur, p.nxt = savedCur, savedNxt
		return p.parseMapLiteral()
	}

	p.cur, p.nxt = savedCur, savedNxt
	return p.parseSetLiteral()
}

// parseSetLiteral processa literais de set
func (p *Parser) parseSetLiteral() Expr {
	p.advanceToken()
	elements := make([]Expr, 0, 3)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		elements = append(elements, elem)

		if p.cur.Lexeme == "," {
			p.advanceToken()
			if p.cur.Lexeme == "}" {
				break
			}
		} else if p.cur.Lexeme != "}" {
			p.handleSetLiteralError()
			return nil
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return &SetLiteral{Elements: elements}
}

// handleSetLiteralError lida com erros em set literals
func (p *Parser) handleSetLiteralError() {
	if p.cur.Type == lexer.EOF {
		p.errorf("unexpected end of file in set literal")
	} else {
		p.errorf("expected ',' or '}' in set literal, got '%s'", p.cur.Lexeme)
	}
}

// parseMapLiteral processa literais de map
func (p *Parser) parseMapLiteral() Expr {
	p.advanceToken()
	entries := make([]*MapEntry, 0, 3)

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
			if p.cur.Lexeme == "}" {
				break
			}
		} else if p.cur.Lexeme != "}" {
			p.handleMapLiteralError()
			return nil
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return &MapLiteral{Entries: entries}
}

// handleMapLiteralError lida com erros em map literals
func (p *Parser) handleMapLiteralError() {
	if p.cur.Type == lexer.EOF {
		p.errorf("unexpected end of file in map literal")
	} else {
		p.errorf("expected ',' or '}' in map literal, got '%s'", p.cur.Lexeme)
	}
}

// ============================
// PARSING DE EXPRESSÕES ESPECIAIS
// ============================

// parseGenericCallOrExpr processa expressões genéricas
func (p *Parser) parseGenericCallOrExpr() Expr {
	p.advanceToken()
	typeArgs := p.parseTypeArgumentsList()
	if typeArgs == nil {
		return nil
	}

	// Função ou struct genérica
	if p.cur.Type == lexer.IDENT {
		return p.parseGenericIdentExpr(typeArgs)
	}

	// Array genérico
	if p.cur.Lexeme == "[" {
		return p.parseGenericArrayExpr(typeArgs)
	}

	p.errorf("expected identifier or array literal after generic type arguments, got %s", p.cur.Lexeme)
	return nil
}

// parseGenericIdentExpr processa identificadores genéricos
func (p *Parser) parseGenericIdentExpr(typeArgs []Type) Expr {
	ident := &Identifier{Name: p.cur.Lexeme}

	// Struct literal genérico
	if p.nxt.Lexeme == "{" {
		p.advanceToken()
		lit := p.parseBraceLiteral()
		if lit == nil {
			return nil
		}
		return &GenericSpecialization{Callee: lit, TypeArgs: typeArgs}
	}

	p.advanceToken()

	// Chamada de função genérica
	if p.cur.Lexeme == "(" {
		p.advanceToken()
		var args []Expr
		if p.cur.Lexeme != ")" {
			args = p.parseArgumentList()
		}

		if !p.expectAndConsume(")") {
			return nil
		}

		return &GenericCallExpr{
			Callee:   ident,
			TypeArgs: typeArgs,
			Args:     args,
		}
	}

	// Referência especializada
	return &GenericCallExpr{
		Callee:   ident,
		TypeArgs: typeArgs,
		Args:     nil,
	}
}

// parseGenericArrayExpr processa arrays genéricos
func (p *Parser) parseGenericArrayExpr(typeArgs []Type) Expr {
	arrayLit := p.parseArrayLiteral()
	if arrayLit == nil {
		return nil
	}
	return &GenericSpecialization{Callee: arrayLit, TypeArgs: typeArgs}
}
