package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseTopLevel() Stmt {
	switch p.cur.Lexeme {
	case "var":
		return p.parseVarDecl() // ⬅️ CHAMA APENAS VAR
	case "const":
		return p.parseConstDecl() // ⬅️ CHAMA APENAS CONST
	case "if", "while", "for", "return":
		return p.parseControlStmt()
	default:
		if isTypeKeyword(p.cur.Lexeme) {
			if p.nxt.Lexeme == "function" {
				return p.parseFunctionDecl(false)
			}
			return p.parseTypedVarDecl()
		}
		return p.parseExprStmt()
	}
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
		p.cur.Type == lexer.KEYWORD || isTypeKeyword(p.cur.Lexeme)
}
