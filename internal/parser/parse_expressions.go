package parser

import (
	"github.com/alpha/internal/lexer"
)

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

func (p *Parser) parsePostfix(left Expr, precedence int) Expr {
	op := p.cur.Lexeme
	p.advanceToken()
	return &UnaryExpr{Op: op, Expr: left, Postfix: true}
}

func (p *Parser) parseCall(left Expr) Expr {
	p.advanceToken()
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
	return &CallExpr{Callee: left, Args: args}
}

func (p *Parser) parseArgumentList() []Expr {
	var args []Expr
	for {
		arg := p.parseExpression(LOWEST)
		if arg == nil {
			return nil
		}
		args = append(args, arg)
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else {
			break
		}
	}
	return args
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

func (p *Parser) parseTernary(cond Expr) Expr {
	p.advanceToken() // consume '?'

	trueExpr := p.parseExpression(LOWEST)
	if trueExpr == nil {
		p.errorf("expected expression after '?'")
		return nil
	}

	if p.cur.Lexeme != ":" {
		p.errorf("expected ':' in ternary expression, got '%s'", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume ':'

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

func (p *Parser) parseNewExpr() Expr {
	p.advanceToken() // consume 'new'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name after 'new'")
		return nil
	}

	typeName := p.cur.Lexeme
	p.advanceToken()

	var typeArgs []Type
	if p.cur.Lexeme == "<" {
		p.advanceToken()
		typeArgs = p.parseTypeArguments()
		if !p.expectAndConsume(">") {
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

func (p *Parser) parseGenericCall(left Expr) Expr {
	p.advanceToken()
	typeArgs := p.parseTypeArguments()
	if typeArgs == nil {
		return nil
	}
	if !p.expectAndConsume(">") {
		return nil
	}
	if p.cur.Lexeme != "(" {
		p.errorf("expected '(' after generic arguments")
		return nil
	}
	call := p.parseCall(left)
	if call == nil {
		return nil
	}
	return &GenericSpecialization{
		Callee:   left,
		TypeArgs: typeArgs,
	}
}

func (p *Parser) parseTypeArguments() []Type {
	var typeArgs []Type
	for {
		typ := p.parseType()
		if typ == nil {
			return nil
		}
		typeArgs = append(typeArgs, typ)
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else {
			break
		}
	}
	return typeArgs
}

func (p *Parser) isGenericCall() bool {
	save := p.cur
	defer func() { p.cur = save }()

	if p.cur.Lexeme != "<" {
		return false
	}
	p.advanceToken()

	if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT {
		return false
	}

	for p.cur.Lexeme != ">" && p.cur.Type != lexer.EOF {
		if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT && p.cur.Lexeme != "," {
			return false
		}
		p.advanceToken()
	}

	if p.cur.Lexeme != ">" {
		return false
	}
	p.advanceToken()

	return p.cur.Lexeme == "("
}

func (p *Parser) isStructLiteral() bool {
	if p.cur.Lexeme != "{" {
		return false
	}

	// Salvar estado para não afetar o parsing real
	savedCur := p.cur
	savedNxt := p.nxt
	defer func() {
		p.cur = savedCur
		p.nxt = savedNxt
	}()

	p.advanceToken() // consume '{'

	// Se estiver vazio ou o próximo token for '}', não é struct literal
	if p.cur.Lexeme == "}" {
		return false
	}

	// Struct literal tem formato: { field: value, ... }
	// Verifica se o próximo token é um identificador seguido de ':'
	if p.cur.Type != lexer.IDENT {
		return false
	}

	p.advanceToken() // consume identifier

	// Verifica se tem ':'
	return p.cur.Lexeme == ":"
}

func (p *Parser) parseStructLiteral() Expr {
	p.advanceToken() // consume '{'

	var fields []*StructField

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected field name in struct literal")
			return nil
		}

		fieldName := p.cur.Lexeme
		p.advanceToken()

		if !p.expectAndConsume(":") {
			return nil
		}

		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}

		fields = append(fields, &StructField{Name: fieldName, Value: value})

		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != "}" {
			p.errorf("expected ',' or '}' in struct literal")
			return nil
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &StructLiteral{Fields: fields}
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

func (p *Parser) parseFunctionExpr() Expr {
	returnType := p.parseType()
	if !p.expectAndConsume("function") {
		return nil
	}
	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()
	return &FunctionExpr{
		ReturnType: returnType,
		Params:     params,
		Body:       body,
	}
}
