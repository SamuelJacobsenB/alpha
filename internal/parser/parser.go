package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alpha/internal/lexer"
)

// ============================
// Estrutura do Parser
// ============================

// Parser representa o analisador sintático
type Parser struct {
	sc     *lexer.Scanner
	cur    lexer.Token
	nxt    lexer.Token
	Errors []string
}

// ============================
// Inicialização e Configuração
// ============================

// New cria uma nova instância do parser
func New(sc *lexer.Scanner) *Parser {
	p := &Parser{sc: sc}
	p.advanceToken() // Carrega primeiro token em cur
	p.advanceToken() // Carrega segundo token em nxt
	return p
}

// ============================
// Funções de Avanço de Tokens
// ============================

// advanceToken avança para o próximo token
func (p *Parser) advanceToken() {
	p.cur = p.nxt
	p.nxt = p.sc.NextToken()
}

// ============================
// Funções de Parsing Principal
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
			// Tenta sincronizar avançando um token
			p.advanceToken()
		}
	}

	return &Program{Body: body}
}

// ============================
// Funções de Parsing de Literais
// ============================

// parseNumberToken analisa um token numérico (inteiro ou float)
func (p *Parser) parseNumberToken(tok lexer.Token) Expr {
	var expr Expr

	if tok.Type == lexer.INT {
		val, _ := strconv.ParseInt(tok.Value, 10, 64)
		expr = &IntLiteral{Value: val}
	} else {
		val, _ := strconv.ParseFloat(tok.Value, 64)
		expr = &FloatLiteral{Value: val}
	}

	p.advanceToken()
	return expr
}

// ============================
// Funções de Verificação e Validação
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

// isAtEndOfStatement verifica se estamos no fim de um statement
func (p *Parser) isAtEndOfStatement() bool {
	return p.cur.Lexeme == ";" || p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF
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
// Funções de Sincronização
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
// Funções de Controle de Erros
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
// Funções Auxiliares
// ============================

// consumeOptionalSemicolon consome ponto-e-vírgula opcional
func (p *Parser) consumeOptionalSemicolon() {
	if p.cur.Lexeme == ";" {
		p.advanceToken()
	}
}
