package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// DECLARAÇÕES DE PACOTE
// ============================

// parsePackageDecl analisa uma declaração de pacote
func (p *Parser) parsePackageDecl() Stmt {
	p.advanceToken() // consome 'package'

	if p.cur.Type != lexer.IDENT && p.cur.Lexeme != "." {
		p.errorf("expected package name after 'package'")
		return nil
	}

	name := p.parseQualifiedName()
	if name == "" {
		return nil
	}

	p.consumeOptionalSemicolon()
	return &PackageDecl{Name: name}
}

// parseQualifiedName parseia um nome qualificado (com pontos)
func (p *Parser) parseQualifiedName() string {
	name := p.cur.Lexeme
	p.advanceToken()

	for p.cur.Lexeme == "." {
		p.advanceToken() // consome '.'
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected identifier after '.' in package name")
			return ""
		}
		name += "." + p.cur.Lexeme
		p.advanceToken()
	}

	return name
}

// ============================
// DECLARAÇÕES DE IMPORTAÇÃO
// ============================

// parseImportDecl analisa uma declaração de importação
func (p *Parser) parseImportDecl() Stmt {
	p.advanceToken() // consome 'import'

	// CASO 1: Importação seletiva com chaves
	if p.cur.Lexeme == "{" {
		return p.parseSelectiveImport()
	}

	// CASO 2: Caminho com pontos (importação de módulo inteiro)
	if p.isQualifiedPath() {
		return p.parseModuleImport()
	}

	// CASO 3: Importação de lista sem chaves ou módulo simples
	return p.parseImportListOrModule()
}

// isQualifiedPath verifica se é um caminho qualificado (com pontos)
func (p *Parser) isQualifiedPath() bool {
	return p.cur.Type == lexer.IDENT && p.nxt.Lexeme == "."
}

// parseModuleImport parseia importação de módulo inteiro
func (p *Parser) parseModuleImport() Stmt {
	path := p.parseQualifiedName()
	if path == "" {
		return nil
	}

	p.consumeOptionalSemicolon()
	return &ImportDecl{Path: path, Imports: nil}
}

// parseImportListOrModule parseia importação de lista ou módulo simples
func (p *Parser) parseImportListOrModule() Stmt {
	specs := p.parseImportSpecListWithoutBraces()
	if specs == nil {
		return nil
	}

	// Se tem 'from', é importação de lista
	if p.cur.Lexeme == "from" {
		return p.parseImportListWithFrom(specs)
	}

	// Caso contrário, é importação de módulo simples
	return p.parseSimpleModuleImport(specs)
}

// parseImportListWithFrom parseia importação de lista com 'from'
func (p *Parser) parseImportListWithFrom(specs []*ImportSpec) Stmt {
	p.advanceToken() // consome 'from'

	path := p.parseModulePath()
	if path == "" {
		return nil
	}

	p.consumeOptionalSemicolon()
	return &ImportDecl{Path: path, Imports: specs}
}

// parseSimpleModuleImport parseia importação de módulo simples
func (p *Parser) parseSimpleModuleImport(specs []*ImportSpec) Stmt {
	if len(specs) != 1 {
		p.errorf("expected single module import or list with 'from'")
		return nil
	}

	if specs[0].Alias != "" {
		p.errorf("module import cannot have alias")
		return nil
	}

	path := specs[0].Name
	p.consumeOptionalSemicolon()
	return &ImportDecl{Path: path, Imports: nil}
}

// parseModulePath parseia caminho do módulo (identificador ou string)
func (p *Parser) parseModulePath() string {
	switch p.cur.Type {
	case lexer.IDENT:
		path := p.cur.Lexeme
		p.advanceToken()
		return path
	case lexer.STRING:
		path := p.cur.Value
		p.advanceToken()
		return path
	default:
		p.errorf("expected module name or path after 'from'")
		return ""
	}
}

// ============================
// IMPORT SELEITVA COM CHAVES
// ============================

// parseSelectiveImport analisa importação seletiva: import { PI, sqrt as sq } from math
func (p *Parser) parseSelectiveImport() Stmt {
	if !p.expectAndConsume("{") {
		return nil
	}

	imports := p.parseImportSpecList()
	if imports == nil {
		return nil
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	if !p.expectAndConsume("from") {
		p.errorf("expected 'from' after import list")
		return nil
	}

	path := p.parseModulePath()
	if path == "" {
		return nil
	}

	p.consumeOptionalSemicolon()
	return &ImportDecl{Path: path, Imports: imports}
}

// ============================
// LISTAS DE ESPECIFICAÇÕES
// ============================

// parseImportSpecList parseia lista de especificações de importação
func (p *Parser) parseImportSpecList() []*ImportSpec {
	return p.parseImportSpecListInternal(true)
}

// parseImportSpecListWithoutBraces parseia lista sem chaves
func (p *Parser) parseImportSpecListWithoutBraces() []*ImportSpec {
	return p.parseImportSpecListInternal(false)
}

// parseImportSpecListInternal implementação comum para listas de importação
func (p *Parser) parseImportSpecListInternal(withBraces bool) []*ImportSpec {
	imports := make([]*ImportSpec, 0, 3)

	for {
		p.skipCommas()

		// Verificar se chegou ao fim
		if p.isImportListEnd(withBraces) {
			break
		}

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected identifier in import list, got %s", p.cur.Lexeme)
			return nil
		}

		spec := p.parseImportSpec()
		if spec == nil {
			return nil
		}
		imports = append(imports, spec)

		// Se o próximo token não for vírgula, terminamos
		if p.cur.Lexeme != "," {
			break
		}
	}

	return imports
}

// isImportListEnd verifica fim da lista de importação
func (p *Parser) isImportListEnd(withBraces bool) bool {
	if withBraces {
		return p.cur.Lexeme == "}" || p.cur.Lexeme == "from" ||
			p.cur.Lexeme == ";" || p.cur.Type == lexer.EOF
	}
	return p.cur.Lexeme == "from" || p.cur.Lexeme == ";" || p.cur.Type == lexer.EOF
}

// parseImportSpec parseia uma especificação de importação individual
func (p *Parser) parseImportSpec() *ImportSpec {
	name := p.cur.Lexeme
	p.advanceToken()

	alias := ""
	if p.cur.Lexeme == "as" {
		p.advanceToken() // consome 'as'

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected alias name after 'as'")
			return nil
		}

		alias = p.cur.Lexeme
		p.advanceToken()
	}

	return &ImportSpec{Name: name, Alias: alias}
}

// skipCommas consome vírgulas consecutivas
func (p *Parser) skipCommas() {
	for p.cur.Lexeme == "," {
		p.advanceToken()
	}
}

// ============================
// DECLARAÇÕES DE EXPORTAÇÃO
// ============================

// parseExportDecl analisa uma declaração de exportação
func (p *Parser) parseExportDecl() Stmt {
	p.advanceToken() // consome 'export'

	exports := p.parseExportSpecList()
	if exports == nil {
		return nil
	}

	p.consumeOptionalSemicolon()
	return &ExportDecl{Exports: exports}
}

// parseExportSpecList parseia lista de especificações de exportação
func (p *Parser) parseExportSpecList() []*ExportSpec {
	exports := make([]*ExportSpec, 0, 3)

	for {
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected identifier in export list, got %s", p.cur.Lexeme)
			return nil
		}

		name := p.cur.Lexeme
		p.advanceToken()

		alias := ""
		if p.cur.Lexeme == "as" {
			p.advanceToken() // consome 'as'

			if p.cur.Type != lexer.IDENT {
				p.errorf("expected alias name after 'as'")
				return nil
			}

			alias = p.cur.Lexeme
			p.advanceToken()
		}

		exports = append(exports, &ExportSpec{Name: name, Alias: alias})

		// Verificar se tem mais
		if p.cur.Lexeme != "," {
			break
		}

		p.advanceToken() // consome ','

		// Se após vírgula não tem nada, quebra
		if p.cur.Type == lexer.EOF {
			break
		}
	}

	return exports
}
