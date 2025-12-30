package parser

import (
	"fmt"
	"strings"

	"github.com/alpha/internal/lexer"
)

// ============================
// ESTRUTURA DO PARSER
// ============================

// Parser representa o analisador sintático
type Parser struct {
	sc     *lexer.Scanner
	cur    lexer.Token
	nxt    lexer.Token
	Errors []string
}

// ============================
// INICIALIZAÇÃO E CONFIGURAÇÃO
// ============================

// New cria uma nova instância do parser
func New(sc *lexer.Scanner) *Parser {
	p := &Parser{sc: sc}
	p.advanceToken() // Carrega primeiro token em cur
	p.advanceToken() // Carrega segundo token em nxt
	return p
}

// ============================
// FUNÇÕES DE AVANÇO DE TOKENS
// ============================

// advanceToken avança para o próximo token
func (p *Parser) advanceToken() {
	p.cur = p.nxt
	p.nxt = p.sc.NextToken()
}

// ============================
// FUNÇÕES DE PARSING PRINCIPAL
// ============================

// ParseProgram analisa um programa completo
func (p *Parser) ParseProgram() *Program {
	body := make([]Stmt, 0, 10)

	for p.cur.Type != lexer.EOF {
		stmt := p.parseTopLevel()
		if stmt != nil {
			body = append(body, stmt)
		} else {
			// Evita loop infinito quando não consegue parsear
			if p.cur.Type == lexer.EOF {
				break
			}
			p.advanceToken() // Tenta sincronizar
		}
	}

	return &Program{Body: body}
}

// ============================
// FUNÇÕES DE VERIFICAÇÃO E VALIDAÇÃO
// ============================

// expectAndConsume verifica se o token atual é o esperado e consome
func (p *Parser) expectAndConsume(expected string) bool {
	if p.cur.Lexeme == expected {
		p.advanceToken()
		return true
	}
	p.errorf("expected '%s', got '%s'", expected, p.cur.Lexeme)
	return false
}

// isAtStmtStart verifica se estamos no início de um statement
func (p *Parser) isAtStmtStart() bool {
	return p.cur.Lexeme == ";" ||
		p.cur.Lexeme == "}" ||
		p.cur.Type == lexer.KEYWORD ||
		p.cur.Lexeme == "{" ||
		isTypeKeyword(p.cur.Lexeme)
}

// ============================
// FUNÇÕES DE SINCRONIZAÇÃO
// ============================

// syncToNextStmt sincroniza para o próximo statement após um erro
func (p *Parser) syncToNextStmt() {
	for !p.isAtStmtStart() && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}
}

// syncTo sincroniza até encontrar um token específico
func (p *Parser) syncTo(token string) {
	for p.cur.Lexeme != token && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}
	if p.cur.Lexeme == token {
		p.advanceToken()
	}
}

// ============================
// FUNÇÕES DE CONTROLE DE ERROS
// ============================

// errorf adiciona um erro à lista de erros
func (p *Parser) errorf(format string, args ...interface{}) {
	p.Errors = append(p.Errors, fmt.Sprintf(format, args...))
}

// HasErrors verifica se há erros no parser
func (p *Parser) HasErrors() bool {
	return len(p.Errors) > 0
}

// ErrorsText retorna todos os erros como uma única string
func (p *Parser) ErrorsText() string {
	return strings.Join(p.Errors, "\n")
}

// ============================
// FUNÇÕES AUXILIARES
// ============================

// consumeOptionalSemicolon consome ponto-e-vírgula opcional
func (p *Parser) consumeOptionalSemicolon() {
	if p.cur.Lexeme == ";" {
		p.advanceToken()
	}
}
