package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// Constantes e Configurações
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

var infixOperators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true, "%": true,
	">=": true, "<=": true, ">": true, "<": true,
	"==": true, "!=": true, "&&": true, "||": true,
	"=": true, "+=": true, "-=": true, "*=": true, "/=": true,
}

var postfixOperators = map[string]bool{
	"++": true,
	"--": true,
}

// ============================
// Funções Auxiliares de Verificação
// ============================

func (p *Parser) isInfixOperator(token lexer.Token) bool {
	return token.Type == lexer.OP && infixOperators[token.Lexeme]
}

func (p *Parser) isPostfixOperator(token lexer.Token) bool {
	return token.Type == lexer.OP && postfixOperators[token.Lexeme]
}

func (p *Parser) precedenceOf(op string) int {
	if prec, exists := precedences[op]; exists {
		return prec
	}
	return LOWEST
}

// ============================
// Parsing de Expressões (Principal)
// ============================

func (p *Parser) parseExpression(precedence int) Expr {
	left := p.parsePrimary()
	if left == nil {
		return nil
	}

	for {
		curOp := p.cur.Lexeme

		// Caso especial para operador ternário
		if curOp == "?" {
			if TERNARY < precedence {
				return left
			}
			left = p.parseTernary(left)
			if left == nil {
				return nil
			}
			continue
		}

		// Se encontramos ":" fora de um ternário, paramos
		if curOp == ":" {
			return left
		}

		curPrec := p.precedenceOf(curOp)

		// Verifica se é um operador infixo válido
		if !p.isInfixOperator(p.cur) && !p.isPostfixOperator(p.cur) &&
			curOp != "(" && curOp != "[" && curOp != "." {
			return left
		}

		// Verifica precedência
		if curPrec < precedence {
			return left
		}

		// Processa o operador atual
		switch {
		case curOp == "(":
			left = p.parseCall(left)
		case curOp == "[":
			left = p.parseIndex(left)
		case curOp == ".":
			left = p.parseMemberAccess(left)
		case p.isInfixOperator(p.cur):
			left = p.parseInfix(left, curPrec)
		case p.isPostfixOperator(p.cur):
			left = p.parsePostfix(left)
		default:
			return left
		}

		if left == nil {
			return nil
		}
	}
}

// ============================
// Parsing de Expressões Primárias
// ============================

func (p *Parser) parsePrimary() Expr {
	switch p.cur.Type {
	case lexer.IDENT:
		// Tratamento explícito para Coleções (Set/Map)
		if (p.cur.Lexeme == "set" || p.cur.Lexeme == "map") && (p.nxt.Lexeme == "<" || p.nxt.Lexeme == "{") {
			return p.parseExplicitCollectionLiteral()
		}

		// Se temos um Identificador seguido de '{', é um literal de struct.
		if p.nxt.Lexeme == "{" {
			return p.parseTypedStructLiteral()
		}

		return p.parseIdentifier()
	case lexer.INT, lexer.FLOAT:
		return p.parseNumberToken(p.cur)
	case lexer.STRING:
		return p.parseStringLiteral()
	case lexer.KEYWORD:
		return p.parseKeywordExpr()
	case lexer.OP:
		if p.cur.Lexeme == ":" {
			return nil
		}
		return p.parseOperatorExpr()
	default:
		p.advanceToken()
		return nil
	}
}

func (p *Parser) parseIdentifier() Expr {
	ident := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()
	return ident
}

func (p *Parser) parseStringLiteral() Expr {
	str := &StringLiteral{Value: p.cur.Value}
	p.advanceToken()
	return str
}

func (p *Parser) parseKeywordExpr() Expr {
	switch p.cur.Lexeme {
	case "true", "false":
		return p.parseBoolLiteral()
	case "null":
		p.advanceToken()
		return &NullLiteral{}
	case "new":
		return p.parseNewExpr()
	case "this":
		p.advanceToken()
		return &ThisExpr{}
	case "generic":
		return p.parseGenericCallOrExpr()
	case "set", "map":
		return p.parseExplicitCollectionLiteral()
	default:
		if isTypeKeyword(p.cur.Lexeme) {
			return nil
		}
		p.errorf("unexpected keyword: %s", p.cur.Lexeme)
		return nil
	}
}

func (p *Parser) parseBoolLiteral() Expr {
	val := p.cur.Lexeme == "true"
	p.advanceToken()
	return &BoolLiteral{Value: val}
}

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
	case "-", "!", "+", "++", "--":
		return p.parsePrefixExpr()
	default:
		p.errorf("unexpected operator: %s", p.cur.Lexeme)
		return nil
	}
}

// ============================
// Parsing de Operadores
// ============================

func (p *Parser) parsePrefixExpr() Expr {
	op := p.cur.Lexeme
	p.advanceToken()

	expr := p.parseExpression(PREFIX)
	if expr == nil {
		return nil
	}

	return &UnaryExpr{Op: op, Expr: expr, Postfix: false}
}

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

func (p *Parser) parsePostfix(left Expr) Expr {
	op := p.cur.Lexeme
	p.advanceToken()
	return &UnaryExpr{Op: op, Expr: left, Postfix: true}
}

func (p *Parser) parseTernary(cond Expr) Expr {
	p.advanceToken() // consume '?'

	trueExpr := p.parseExpression(TERNARY + 1)
	if trueExpr == nil {
		p.errorf("expected expression after '?'")
		return nil
	}

	if !p.expectAndConsume(":") {
		p.errorf("expected ':' in ternary expression, got '%s'", p.cur.Lexeme)
		return nil
	}

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
// Parsing de Chamadas e Acessos
// ============================

func (p *Parser) parseCall(left Expr) Expr {
	// Trata GenericCallExpr sem argumentos
	if gce, ok := left.(*GenericCallExpr); ok && gce.Args == nil {
		if !p.expectAndConsume("(") {
			p.errorf("expected '(' for function call")
			return nil
		}

		var args []Expr
		if p.cur.Lexeme != ")" {
			args = p.parseArgumentList()
			if args == nil {
				return nil
			}
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

	// Chamada normal
	if !p.expectAndConsume("(") {
		return left
	}

	var args []Expr
	if p.cur.Lexeme != ")" {
		args = p.parseArgumentList()
		if args == nil {
			return nil
		}
	}

	if !p.expectAndConsume(")") {
		return nil
	}

	return &CallExpr{
		Callee: left,
		Args:   args,
	}
}

func (p *Parser) parseArgumentList() []Expr {
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

func (p *Parser) match(lexeme string) bool {
	if p.cur.Lexeme == lexeme {
		p.advanceToken()
		return true
	}
	return false
}

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
// Parsing de Literais
// ============================

func (p *Parser) parseBraceLiteral() Expr {
	p.advanceToken() // consume '{'

	// Caso vazio: assume Set (ou Map vazio, dependendo da semântica, aqui SetLiteral)
	if p.cur.Lexeme == "}" {
		p.advanceToken()
		return &SetLiteral{Elements: []Expr{}}
	}

	// Parseia o primeiro elemento/chave para decidir o tipo
	firstExpr := p.parseExpression(LOWEST)
	if firstExpr == nil {
		return nil
	}

	// Se o próximo token é ':', temos um Map ou Struct (Chave: Valor)
	if p.cur.Lexeme == ":" {
		p.advanceToken() // consome ':'

		firstValue := p.parseExpression(LOWEST)
		if firstValue == nil {
			return nil
		}

		// Decisão: Se a chave for um Identificador simples, tratamos como StructLiteral
		// Caso contrário, tratamos como MapLiteral
		if ident, ok := firstExpr.(*Identifier); ok {
			return p.continueStructLiteral(ident, firstValue)
		}
		return p.continueMapLiteral(firstExpr, firstValue)
	}

	// Se não for ':', é um Set (Lista de valores)
	return p.continueSetLiteral(firstExpr)
}

// CORREÇÃO: Função dedicada para literais de struct com tipo explícito (ex: Point { ... })
func (p *Parser) parseTypedStructLiteral() Expr {
	// O nome do tipo (IDENT) já foi verificado em parsePrimary
	// A AST atual não possui campo 'Name' em StructLiteral, então apenas consumimos o token.
	p.advanceToken() // consome o nome do tipo

	// Delega para o parser de literais com chaves
	return p.parseBraceLiteral()
}

func (p *Parser) continueStructLiteral(firstKey *Identifier, firstValue Expr) Expr {
	fields := make([]*StructField, 0, 3)
	fields = append(fields, &StructField{Name: firstKey.Name, Value: firstValue})

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

func (p *Parser) continueMapLiteral(firstKey, firstValue Expr) Expr {
	entries := make([]*MapEntry, 0, 3)
	entries = append(entries, &MapEntry{Key: firstKey, Value: firstValue})

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

func (p *Parser) continueSetLiteral(firstElem Expr) Expr {
	elements := make([]Expr, 0, 3)
	elements = append(elements, firstElem)

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

func (p *Parser) parseExplicitCollectionLiteral() Expr {
	typeName := p.cur.Lexeme
	p.advanceToken() // consome 'map' ou 'set'

	// Se houver parâmetros genéricos <...>, consuma-os
	if p.cur.Lexeme == "<" {
		if p.parseGenericType(typeName) == nil {
			return nil
		}
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
	if elements == nil || !p.expectAndConsume("]") {
		return nil
	}

	return &ArrayLiteral{Elements: elements}
}

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

func (p *Parser) parseCollectionLiteral() Expr {
	// Salvar estado atual
	savedCur, savedNxt := p.cur, p.nxt

	p.advanceToken() // consume '{'

	// Se estiver vazio
	if p.cur.Lexeme == "}" {
		p.cur, p.nxt = savedCur, savedNxt
		return p.parseSetLiteral()
	}

	tempParser := *p // Cópia rasa do parser

	// Tentar parsear uma expressão
	expr := tempParser.parseExpression(LOWEST)
	if expr == nil {
		// Se não conseguir parsear, restaurar e tentar como set
		p.cur, p.nxt = savedCur, savedNxt
		return p.parseSetLiteral()
	}

	// Verificar o token após a expressão
	if tempParser.cur.Lexeme == ":" {
		// É um map
		p.cur, p.nxt = savedCur, savedNxt
		return p.parseMapLiteral()
	}

	// Caso contrário, é um set
	p.cur, p.nxt = savedCur, savedNxt
	return p.parseSetLiteral()
}

func (p *Parser) parseSetLiteral() Expr {
	p.advanceToken() // consume '{'

	elements := make([]Expr, 0, 3)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		elements = append(elements, elem)

		if p.cur.Lexeme == "," {
			p.advanceToken()
			// Permitir trailing comma
			if p.cur.Lexeme == "}" {
				break
			}
		} else if p.cur.Lexeme != "}" {
			if p.cur.Type == lexer.EOF {
				p.errorf("unexpected end of file in set literal")
				return nil
			}
			p.errorf("expected ',' or '}' in set literal, got '%s'", p.cur.Lexeme)
			return nil
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &SetLiteral{Elements: elements}
}

func (p *Parser) parseMapLiteral() Expr {
	p.advanceToken() // consume '{'

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
			// Permitir trailing comma
			if p.cur.Lexeme == "}" {
				break
			}
		} else if p.cur.Lexeme != "}" {
			if p.cur.Type == lexer.EOF {
				p.errorf("unexpected end of file in map literal")
				return nil
			}
			p.errorf("expected ',' or '}' in map literal, got '%s'", p.cur.Lexeme)
			return nil
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &MapLiteral{Entries: entries}
}

// ============================
// Parsing de Expressões Especiais
// ============================

func (p *Parser) parseGenericCallOrExpr() Expr {
	p.advanceToken() // consume 'generic'

	// Usar nova lógica unificada
	typeArgs := p.parseTypeArgumentsList()
	if typeArgs == nil {
		return nil
	}

	// Adicionado suporte para 'new' após generic<...> (ex: generic<int> new Container())
	if p.cur.Lexeme == "new" {
		p.advanceToken() // consume 'new'

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected type name after 'new'")
			return nil
		}

		typeName := p.cur.Lexeme
		p.advanceToken()

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

		return &NewExpr{
			TypeName: typeName,
			TypeArgs: typeArgs,
			Args:     args,
		}
	}

	// Verifica se é uma chamada de função ou referência
	if p.cur.Type == lexer.IDENT {
		ident := &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()

		// Se tiver parênteses, é uma chamada de função
		if p.cur.Lexeme == "(" {
			p.advanceToken() // consume '('

			var args []Expr
			if p.cur.Lexeme != ")" {
				args = p.parseArgumentList()
				if args == nil {
					return nil
				}
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

		// Se não tem parênteses, é apenas uma referência
		return &GenericCallExpr{
			Callee:   ident,
			TypeArgs: typeArgs,
			Args:     nil,
		}
	}

	// Verifica se é um array literal
	if p.cur.Lexeme == "[" {
		arrayLit := p.parseArrayLiteral()
		if arrayLit == nil {
			return nil
		}

		return &GenericSpecialization{
			Callee:   arrayLit,
			TypeArgs: typeArgs,
		}
	}

	p.errorf("expected identifier, 'new' or array literal after generic type arguments, got %s", p.cur.Lexeme)
	return nil
}

func (p *Parser) parseNewExpr() Expr {
	p.advanceToken() // consume 'new'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name after 'new'")
		return nil
	}

	typeName := p.cur.Lexeme
	p.advanceToken()

	var typeArgs []Type
	if p.cur.Lexeme == "generic" {
		p.advanceToken() // consume 'generic'
		// Usar nova lógica unificada
		typeArgs = p.parseTypeArgumentsList()
		if typeArgs == nil {
			return nil
		}
	}

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

	return &NewExpr{
		TypeName: typeName,
		TypeArgs: typeArgs,
		Args:     args,
	}
}
