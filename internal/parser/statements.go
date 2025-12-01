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
	case "if", "while", "do", "for", "switch", "return":
		return p.parseControlStmt()
	default:
		if isTypeKeyword(p.cur.Lexeme) {
			if p.nxt.Lexeme == "function" {
				return p.parseFunctionDecl(false)
			} else if p.nxt.Type == lexer.IDENT || p.nxt.Lexeme == ";" || p.nxt.Lexeme == "=" {
				// Pode ser declaração de variável com ou sem inicialização
				stmt = p.parseTypedVarDecl()
			} else {
				p.errorf("unexpected type keyword: %s", p.cur.Lexeme)
				return nil
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

func (p *Parser) syncToNextStmt() {
	for !p.isAtStmtStart() && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}
}

func (p *Parser) isAtStmtStart() bool {
	return p.cur.Lexeme == ";" || p.cur.Lexeme == "}" ||
		p.cur.Type == lexer.KEYWORD || isTypeKeyword(p.cur.Lexeme) ||
		p.cur.Lexeme == "{"
}
