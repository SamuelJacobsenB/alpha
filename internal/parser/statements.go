package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

// parseTopLevel é o ponto de entrada principal para parsing de statements
func (p *Parser) parseTopLevel() Stmt {
	fmt.Printf("parseTopLevel: cur=%q (type=%v) nxt=%q\n", p.cur.Lexeme, p.cur.Type, p.nxt.Lexeme)

	if p.cur.Type == lexer.EOF {
		return nil
	}

	switch {
	case p.isFunctionDeclaration():
		return p.parseFunctionDecl()
	case p.isTypedVariableDeclaration():
		return p.parseTypedVarDecl()
	case p.isKeywordStatement():
		return p.parseKeywordStatement()
	case p.isBlockStart():
		return &BlockStmt{Body: p.parseBlockLike()}
	case p.isBlockEnd():
		p.advanceToken()
		return nil
	case p.canStartExpression():
		return p.parseExprStmt()
	default:
		return p.parseUnknownToken()
	}
}

// Helpers para detecção de tipo de statement
func (p *Parser) isFunctionDeclaration() bool {
	return p.cur.Type == lexer.KEYWORD && isTypeKeyword(p.cur.Lexeme) &&
		(p.nxt.Lexeme == "function" || (p.nxt.Lexeme == "[" && p.peekNextAfterBrackets() == "function"))
}

func (p *Parser) isTypedVariableDeclaration() bool {
	return p.cur.Type == lexer.KEYWORD && isTypeKeyword(p.cur.Lexeme)
}

func (p *Parser) isKeywordStatement() bool {
	return p.cur.Type == lexer.KEYWORD && p.isKnownKeyword()
}

func (p *Parser) isKnownKeyword() bool {
	keywords := map[string]bool{
		"var": true, "const": true, "function": true,
		"if": true, "while": true, "for": true, "return": true,
	}
	return keywords[p.cur.Lexeme]
}

func (p *Parser) isBlockStart() bool {
	return p.cur.Lexeme == "{"
}

func (p *Parser) isBlockEnd() bool {
	return p.cur.Lexeme == "}"
}

func (p *Parser) parseUnknownToken() Stmt {
	fmt.Printf("parseTopLevel: unrecognized token %q, advancing\n", p.cur.Lexeme)
	p.advanceToken()
	return nil
}
