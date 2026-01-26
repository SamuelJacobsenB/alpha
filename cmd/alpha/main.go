package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"os"

	"github.com/alpha/internal/ir"
	"github.com/alpha/internal/lexer"
	"github.com/alpha/internal/parser"
	"github.com/alpha/internal/semantic"
)

// ==========================================
// CONSTANTES PARA FORMATAÇÃO
// ==========================================

const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorCyan    = "\033[36m"
	ColorMagenta = "\033[35m"
	ColorBold    = "\033[1m"
	ColorGray    = "\033[90m"
	ColorWhite   = "\033[97m"
)

// ==========================================
// ESTRUTURAS DE ANÁLISE
// ==========================================

type AnalysisResult struct {
	Success        bool
	Message        string
	TokenCount     int
	Tokens         []lexer.Token
	LexerErrors    []string
	ParserErrors   []parserError
	SemanticErrors []semanticError
	ASTStructure   string
	IRModule       *ir.Module
	Duration       time.Duration
	Lines          []string // Armazena linhas do código para contexto
}

type parserError struct {
	Line    int
	Col     int
	Message string
}

type semanticError struct {
	Line    int
	Col     int
	Message string
}

// ==========================================
// FUNÇÃO PRINCIPAL
// ==========================================

func main() {
	printBanner("🧪 ANÁLISE DO ARQUIVO main.alpha")

	// Ler o arquivo main.alpha
	code, err := os.ReadFile("main.alpha")
	if err != nil {
		printError("Erro ao ler o arquivo main.alpha:" + err.Error())
		return
	}

	codeStr := string(code)
	lines := strings.Split(codeStr, "\n")

	printSection("📄 CONTEÚDO DO ARQUIVO", ColorWhite)
	fmt.Println(ColorGray + strings.Repeat("─", 80) + ColorReset)

	// Mostrar código com numeração de linhas
	for i, line := range lines {
		fmt.Printf("%s%3d │ %s%s\n", ColorGray, i+1, ColorReset, line)
	}

	fmt.Println(ColorGray + strings.Repeat("─", 80) + ColorReset)

	// Executar análise completa
	result := analyzeFile(codeStr, lines)
	printAnalysisResult(result)

	// Se a análise foi bem-sucedida, mostrar o IR gerado
	if result.Success && result.IRModule != nil {
		printIR(result.IRModule)
	}
}

// ==========================================
// ANÁLISE COMPLETA
// ==========================================

func analyzeFile(code string, lines []string) AnalysisResult {
	startTime := time.Now()
	result := AnalysisResult{
		Lines: lines,
	}

	printSection("🧪 ETAPA 1: ANÁLISE LÉXICA", ColorBlue)

	// ========== ETAPA 1: LEXER ==========
	printStep("Analisando tokens...", 1, 4)
	scanner := lexer.NewScanner(code)

	tokens := []lexer.Token{}
	lexerErrors := []string{}

	for {
		tok := scanner.NextToken()
		tokens = append(tokens, tok)

		if tok.Type == lexer.EOF {
			break
		}

		if tok.Type == lexer.ERROR {
			lexerErrors = append(lexerErrors,
				fmt.Sprintf("Linha %d:%d - Token ilegal: %s",
					tok.Line, tok.Col, tok.Lexeme))
		}
	}

	result.TokenCount = len(tokens) - 1 // Excluir EOF
	result.Tokens = tokens
	result.LexerErrors = lexerErrors

	if len(lexerErrors) > 0 {
		printStepResult(fmt.Sprintf("❌ (%d erros)", len(lexerErrors)), false)
	} else {
		printStepResult(fmt.Sprintf("✅ (%d tokens)", result.TokenCount), true)
	}

	// Mostrar tokens detalhados
	if len(tokens) > 0 {
		printSubsection("📋 TOKENS ENCONTRADOS")
		printTokens(tokens)
	}

	// Se houver erros léxicos, parar aqui
	if len(lexerErrors) > 0 {
		result.Success = false
		result.Message = "Erros léxicos encontrados"
		result.Duration = time.Since(startTime)
		return result
	}

	// ========== ETAPA 2: PARSER ==========
	printSection("🧪 ETAPA 2: ANÁLISE SINTÁTICA", ColorYellow)

	printStep("Analisando estrutura sintática...", 2, 4)

	// Criar novo scanner para o parser
	scanner = lexer.NewScanner(code)
	p := parser.New(scanner)
	program := p.ParseProgram()

	// Processar erros do parser
	parserErrors := []parserError{}
	for _, errMsg := range p.Errors {
		// Tentar extrair linha e coluna da mensagem de erro
		line, col, message := parseErrorPosition(errMsg)
		parserErrors = append(parserErrors, parserError{
			Line:    line,
			Col:     col,
			Message: message,
		})
	}

	result.ParserErrors = parserErrors

	if p.HasErrors() {
		printStepResult(fmt.Sprintf("❌ (%d erros)", len(p.Errors)), false)
		result.Success = false
		result.Message = "Erros sintáticos encontrados"
		result.Duration = time.Since(startTime)
		return result
	}

	printStepResult("✅", true)

	// Mostrar estrutura da AST
	printSubsection("📊 ESTRUTURA DA AST")
	astStr := printASTStructure(program, 0)
	result.ASTStructure = astStr

	// ========== ETAPA 3: SEMANTIC ==========
	printSection("🧪 ETAPA 3: ANÁLISE SEMÂNTICA", ColorMagenta)

	printStep("Analisando semântica...", 3, 4)

	checker := semantic.NewChecker()
	checker.CheckProgram(program)

	// Processar erros semânticos
	semanticErrors := []semanticError{}
	for _, err := range checker.Errors {
		// Tentar extrair linha e coluna da mensagem de erro semântico
		line, col, message := parseSemanticError(err.Error())
		semanticErrors = append(semanticErrors, semanticError{
			Line:    line,
			Col:     col,
			Message: message,
		})
	}

	result.SemanticErrors = semanticErrors

	if len(checker.Errors) > 0 {
		printStepResult(fmt.Sprintf("❌ (%d erros)", len(checker.Errors)), false)
		result.Success = false
		result.Message = "Erros semânticos encontrados"
		result.Duration = time.Since(startTime)
		return result
	}

	printStepResult("✅", true)

	// ========== ETAPA 4: GERAÇÃO DE IR ==========
	printSection("🧪 ETAPA 4: GERAÇÃO DE IR", ColorCyan)

	printStep("Gerando IR (Representação Intermediária)...", 4, 4)

	// Gerar o IR
	generator := ir.NewGenerator(checker)
	irModule := generator.Generate(program)
	result.IRModule = irModule

	printStepResult("✅", true)

	result.Success = true
	result.Message = "Análise completa bem-sucedida"
	result.Duration = time.Since(startTime)

	return result
}

// ==========================================
// FUNÇÕES AUXILIARES DE PARSING DE ERROS
// ==========================================

func parseErrorPosition(errMsg string) (int, int, string) {
	// Tenta extrair linha e coluna de mensagens como "expected ';', got 'EOF'"
	// ou "line X:col Y: message"
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 {
		errMsg = lines[0]
	}

	// Remove prefixos comuns
	errMsg = strings.TrimPrefix(errMsg, "error: ")
	errMsg = strings.TrimPrefix(errMsg, "parser error: ")

	// Tenta padrões como "line X:col Y: message"
	if strings.Contains(errMsg, "line") && strings.Contains(errMsg, "col") {
		parts := strings.Split(errMsg, ":")
		if len(parts) >= 3 {
			// Tenta extrair números
			for _, part := range parts {
				if strings.Contains(part, "line") {
					// Extrai número após "line"
					lineStr := strings.TrimSpace(strings.TrimPrefix(part, "line"))
					if line, err := strconv.Atoi(lineStr); err == nil {
						// Procura col na próxima parte
						for j, p := range parts {
							if strings.Contains(p, "col") && j > 0 {
								colStr := strings.TrimSpace(strings.TrimPrefix(p, "col"))
								if col, err := strconv.Atoi(colStr); err == nil {
									message := strings.Join(parts[2:], ":")
									return line, col, strings.TrimSpace(message)
								}
							}
						}
					}
				}
			}
		}
	}

	// Retorno padrão
	return 0, 0, errMsg
}

func parseSemanticError(errMsg string) (int, int, string) {
	// Formato: "[Semantic Error] @ X:Y: message"
	parts := strings.Split(errMsg, "@")
	if len(parts) == 2 {
		posAndMsg := strings.SplitN(parts[1], ":", 3)
		if len(posAndMsg) == 3 {
			if line, err := strconv.Atoi(strings.TrimSpace(posAndMsg[0])); err == nil {
				if col, err := strconv.Atoi(strings.TrimSpace(posAndMsg[1])); err == nil {
					return line, col, strings.TrimSpace(posAndMsg[2])
				}
			}
		}
	}
	return 0, 0, errMsg
}

// ==========================================
// FUNÇÕES DE IMPRESSÃO
// ==========================================

func printBanner(title string) {
	fmt.Println(ColorBold + ColorCyan + "╔══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║     %-55s ║\n", title)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝" + ColorReset)
}

func printSection(title string, color string) {
	fmt.Printf("\n%s%s%s\n", ColorBold+color, title, ColorReset)
	fmt.Println(ColorGray + strings.Repeat("─", 80) + ColorReset)
}

func printSubsection(title string) {
	fmt.Printf("\n%s%s%s\n", ColorBold, title, ColorReset)
}

func printStep(step string, current, total int) {
	fmt.Printf("%s[%d/%d] %s%s", ColorBlue, current, total, step, ColorReset)
}

func printStepResult(result string, success bool) {
	if success {
		fmt.Printf(" %s%s%s\n", ColorGreen, result, ColorReset)
	} else {
		fmt.Printf(" %s%s%s\n", ColorRed, result, ColorReset)
	}
}

func printTokens(tokens []lexer.Token) {
	tokenTypeNames := map[lexer.TokenType]string{
		lexer.EOF:     "EOF",
		lexer.ERROR:   "ERROR",
		lexer.KEYWORD: "KEYWORD",
		lexer.IDENT:   "IDENT",
		lexer.INT:     "INT",
		lexer.FLOAT:   "FLOAT",
		lexer.STRING:  "STRING",
		lexer.OP:      "OP",
		lexer.GENERIC: "GENERIC",
	}

	fmt.Printf("%-10s %-12s %-25s %s\n",
		ColorBold+"Posição"+ColorReset,
		ColorBold+"Tipo"+ColorReset,
		ColorBold+"Lexema"+ColorReset,
		ColorBold+"Valor"+ColorReset)
	fmt.Println(ColorGray + strings.Repeat("─", 80) + ColorReset)

	for i, tok := range tokens {
		if i >= 50 && i < len(tokens)-1 { // Limitar para não ficar muito grande
			fmt.Printf("%s... e mais %d tokens%s\n", ColorGray, len(tokens)-i-1, ColorReset)
			break
		}

		color := ColorReset
		switch tok.Type {
		case lexer.KEYWORD:
			color = ColorBlue
		case lexer.IDENT:
			color = ColorCyan
		case lexer.INT, lexer.FLOAT:
			color = ColorYellow
		case lexer.STRING:
			color = ColorGreen
		case lexer.OP:
			color = ColorMagenta
		case lexer.ERROR:
			color = ColorRed + ColorBold
		case lexer.GENERIC:
			color = ColorMagenta + ColorBold
		}

		fmt.Printf("%-10s %-12s %-25s %s\n",
			fmt.Sprintf("%d:%d", tok.Line, tok.Col),
			color+tokenTypeNames[tok.Type]+ColorReset,
			color+limitString(tok.Lexeme, 23)+ColorReset,
			color+limitString(tok.Value, 30)+ColorReset)
	}
}

func printASTStructure(node interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	var result string

	switch n := node.(type) {
	case *parser.Program:
		fmt.Printf("%s%sProgram%s\n", indentStr, ColorCyan+ColorBold, ColorReset)
		result = "Program"
		for i, stmt := range n.Body {
			fmt.Printf("%s  %s[%d]%s ", indentStr, ColorGray, i+1, ColorReset)
			printASTStructure(stmt, indent+1)
		}

	case *parser.PackageDecl:
		fmt.Printf("%s%sPackage: %s%s\n", indentStr, ColorGreen, n.Name, ColorReset)
		result = "PackageDecl"

	case *parser.ImportDecl:
		fmt.Printf("%s%sImport from: %s%s\n", indentStr, ColorGreen, n.Path, ColorReset)
		if n.Imports != nil {
			for _, imp := range n.Imports {
				if imp.Alias != "" {
					fmt.Printf("%s    %s as %s\n", indentStr, imp.Name, imp.Alias)
				} else {
					fmt.Printf("%s    %s\n", indentStr, imp.Name)
				}
			}
		}
		result = "ImportDecl"

	case *parser.VarDecl:
		fmt.Printf("%s%sVar: %s%s\n", indentStr, ColorYellow, n.Name, ColorReset)
		if n.Type != nil {
			fmt.Printf("%s  Type: ", indentStr)
			printASTStructure(n.Type, 0)
		}
		if n.Init != nil {
			fmt.Printf("%s  Init: ", indentStr)
			printASTStructure(n.Init, 0)
		}
		result = "VarDecl"

	case *parser.FunctionDecl:
		fmt.Printf("%s%sFunction: %s%s\n", indentStr, ColorBlue, n.Name, ColorReset)
		if n.ReturnTypes != nil {
			fmt.Printf("%s  ReturnType: ", indentStr)
			printASTStructure(n.ReturnTypes, 0)
		}
		if len(n.Params) > 0 {
			fmt.Printf("%s  Params:\n", indentStr)
			for _, param := range n.Params {
				fmt.Printf("%s    %s: ", indentStr, param.Name)
				printASTStructure(param.Type, 0)
			}
		}
		if len(n.Generics) > 0 {
			fmt.Printf("%s  Generics: ", indentStr)
			for _, g := range n.Generics {
				fmt.Printf("%s ", g.Name)
			}
			fmt.Println()
		}
		fmt.Printf("%s  Body (%d statements)\n", indentStr, len(n.Body))
		result = "FunctionDecl"

	case *parser.StructDecl:
		fmt.Printf("%s%sStruct: %s%s\n", indentStr, ColorMagenta, n.Name, ColorReset)
		if len(n.Generics) > 0 {
			fmt.Printf("%s  Generics: ", indentStr)
			for _, g := range n.Generics {
				fmt.Printf("%s ", g.Name)
			}
			fmt.Println()
		}
		if len(n.Fields) > 0 {
			fmt.Printf("%s  Fields:\n", indentStr)
			for _, field := range n.Fields {
				visibility := "public"
				if field.IsPrivate {
					visibility = "private"
				}
				fmt.Printf("%s    %s %s: ", indentStr, visibility, field.Name)
				printASTStructure(field.Type, 0)
			}
		}
		result = "StructDecl"

	case *parser.IfStmt:
		fmt.Printf("%s%sIf Statement%s\n", indentStr, ColorCyan, ColorReset)
		if n.Cond != nil {
			fmt.Printf("%s  Condition: ", indentStr)
			printASTStructure(n.Cond, 0)
		}
		fmt.Printf("%s  Then (%d statements)\n", indentStr, len(n.Then))
		if len(n.Else) > 0 {
			fmt.Printf("%s  Else (%d statements)\n", indentStr, len(n.Else))
		}
		result = "IfStmt"

	case *parser.WhileStmt:
		fmt.Printf("%s%sWhile Loop%s\n", indentStr, ColorCyan, ColorReset)
		if n.Cond != nil {
			fmt.Printf("%s  Condition: ", indentStr)
			printASTStructure(n.Cond, 0)
		}
		fmt.Printf("%s  Body (%d statements)\n", indentStr, len(n.Body))
		result = "WhileStmt"

	case *parser.ForStmt:
		fmt.Printf("%s%sFor Loop%s\n", indentStr, ColorCyan, ColorReset)
		if n.Init != nil {
			fmt.Printf("%s  Init: ", indentStr)
			printASTStructure(n.Init, 0)
		}
		if n.Cond != nil {
			fmt.Printf("%s  Condition: ", indentStr)
			printASTStructure(n.Cond, 0)
		}
		if n.Post != nil {
			fmt.Printf("%s  Post: ", indentStr)
			printASTStructure(n.Post, 0)
		}
		fmt.Printf("%s  Body (%d statements)\n", indentStr, len(n.Body))
		result = "ForStmt"

	case *parser.Identifier:
		fmt.Printf("%s%sIdentifier: %s%s\n", indentStr, ColorGreen, n.Name, ColorReset)
		result = "Identifier"

	case *parser.IntLiteral:
		fmt.Printf("%s%sInt: %d%s\n", indentStr, ColorYellow, n.Value, ColorReset)
		result = "IntLiteral"

	case *parser.StringLiteral:
		fmt.Printf("%s%sString: \"%s\"%s\n", indentStr, ColorGreen, limitString(n.Value, 40), ColorReset)
		result = "StringLiteral"

	case *parser.BinaryExpr:
		fmt.Printf("%s%sBinary: %s%s\n", indentStr, ColorMagenta, n.Op, ColorReset)
		fmt.Printf("%s  Left: ", indentStr)
		printASTStructure(n.Left, 0)
		fmt.Printf("%s  Right: ", indentStr)
		printASTStructure(n.Right, 0)
		result = "BinaryExpr"

	case *parser.CallExpr:
		fmt.Printf("%s%sFunction Call%s\n", indentStr, ColorBlue, ColorReset)
		if n.Callee != nil {
			fmt.Printf("%s  Callee: ", indentStr)
			printASTStructure(n.Callee, 0)
		}
		fmt.Printf("%s  Args (%d):\n", indentStr, len(n.Args))
		for i, arg := range n.Args {
			fmt.Printf("%s    [%d] ", indentStr, i+1)
			printASTStructure(arg, 0)
		}
		result = "CallExpr"

	case *parser.PrimitiveType:
		fmt.Printf("%s%sType: %s%s\n", indentStr, ColorCyan, n.Name, ColorReset)
		result = "PrimitiveType"

	case *parser.IdentifierType:
		fmt.Printf("%s%sType: %s%s\n", indentStr, ColorCyan, n.Name, ColorReset)
		result = "IdentifierType"

	case *parser.GenericType:
		fmt.Printf("%s%sGeneric Type: %s%s\n", indentStr, ColorCyan, n.Name, ColorReset)
		if len(n.TypeArgs) > 0 {
			fmt.Printf("%s  Type Args (%d):\n", indentStr, len(n.TypeArgs))
			for i, arg := range n.TypeArgs {
				fmt.Printf("%s    [%d] ", indentStr, i+1)
				printASTStructure(arg, 0)
			}
		}
		result = "GenericType"

	default:
		typeName := fmt.Sprintf("%T", n)
		simpleName := strings.TrimPrefix(typeName, "*parser.")
		fmt.Printf("%s%s%s%s\n", indentStr, ColorGray, simpleName, ColorReset)
		result = simpleName
	}

	return result
}

func printAnalysisResult(result AnalysisResult) {
	printBanner("📊 RESUMO DA ANÁLISE")

	fmt.Printf("\n%sESTATÍSTICAS:%s\n", ColorBold+ColorWhite, ColorReset)
	fmt.Printf("   Status:           %s%s%s\n",
		colorIf(result.Success, ColorGreen+"✅ ", ColorRed+"❌ "),
		result.Message,
		ColorReset)
	fmt.Printf("   Tokens:           %s%d%s\n", ColorBold, result.TokenCount, ColorReset)
	fmt.Printf("   Tempo Total:      %s%.3f segundos%s\n", ColorBold, result.Duration.Seconds(), ColorReset)

	if result.IRModule != nil {
		fmt.Printf("   Módulo IR:        %s%s (funções: %d)%s\n", ColorBold, result.IRModule.Name, len(result.IRModule.Functions), ColorReset)
	}

	// Mostrar erros por etapa com contexto
	if len(result.LexerErrors) > 0 {
		printErrorSection("ERROS LÉXICOS", len(result.LexerErrors))
		for i, err := range result.LexerErrors {
			fmt.Printf("   %s%d.%s %s\n", ColorRed, i+1, ColorReset, err)
		}
	}

	if len(result.ParserErrors) > 0 {
		printErrorSection("ERROS SINTÁTICOS", len(result.ParserErrors))
		for i, err := range result.ParserErrors {
			fmt.Printf("\n   %s%d.%s %s\n", ColorRed, i+1, ColorReset, err.Message)

			// Mostrar contexto do erro se tiver linha/coluna
			if err.Line > 0 && err.Line <= len(result.Lines) {
				lineIndex := err.Line - 1
				fmt.Printf("   %s%s%d │ %s%s\n", ColorGray, ColorBold, err.Line, ColorReset, result.Lines[lineIndex])

				// Mostrar ponteiro para a coluna
				if err.Col > 0 {
					spaces := strings.Repeat(" ", 6+len(fmt.Sprintf("%d", err.Line))) // Ajuste para alinhamento
					pointer := strings.Repeat(" ", err.Col-1) + "^"
					fmt.Printf("   %s%s%s%s\n", ColorGray, spaces, ColorRed+ColorBold, pointer+ColorReset)
				}
			}
		}
	}

	if len(result.SemanticErrors) > 0 {
		printErrorSection("ERROS SEMÂNTICOS", len(result.SemanticErrors))
		for i, err := range result.SemanticErrors {
			fmt.Printf("\n   %s%d.%s %s\n", ColorRed, i+1, ColorReset, err.Message)

			// Mostrar contexto do erro se tiver linha/coluna
			if err.Line > 0 && err.Line <= len(result.Lines) {
				lineIndex := err.Line - 1
				fmt.Printf("   %s%s%d │ %s%s\n", ColorGray, ColorBold, err.Line, ColorReset, result.Lines[lineIndex])

				// Mostrar ponteiro para a coluna
				if err.Col > 0 {
					spaces := strings.Repeat(" ", 6+len(fmt.Sprintf("%d", err.Line))) // Ajuste para alinhamento
					pointer := strings.Repeat(" ", err.Col-1) + "^"
					fmt.Printf("   %s%s%s%s\n", ColorGray, spaces, ColorRed+ColorBold, pointer+ColorReset)
				}
			}
		}
	}

	// Mensagem final
	fmt.Print("\n" + ColorBold)
	if result.Success {
		fmt.Println(ColorGreen + "✨ ANÁLISE COMPLETA BEM-SUCEDIDA! ✨" + ColorReset)
		fmt.Println(ColorCyan + "📋 O IR foi gerado com sucesso e será exibido abaixo." + ColorReset)
	} else {
		fmt.Println(ColorRed + "⚠️  ANÁLISE ENCONTROU ERROS" + ColorReset)
		fmt.Println(ColorYellow + "💡 Dica: Verifique a sintaxe e os tipos mencionados nos erros acima." + ColorReset)
	}
}

// ==========================================
// FUNÇÃO PARA EXIBIR IR
// ==========================================

func printIR(module *ir.Module) {
	printSection("🧬 REPRESENTAÇÃO INTERMEDIÁRIA (IR)", ColorCyan)

	fmt.Printf("%sMódulo: %s%s\n\n", ColorBold, module.Name, ColorReset)

	if len(module.Structs) > 0 {
		fmt.Printf("%sStructs: (%d)%s\n", ColorYellow+ColorBold, len(module.Structs), ColorReset)
		for i, s := range module.Structs {
			fmt.Printf("  %s%d.%s %s\n", ColorCyan, i+1, ColorReset, s.Name)
		}
		fmt.Println()
	}

	if len(module.Globals) > 0 {
		fmt.Printf("%sVariáveis Globais: (%d)%s\n", ColorYellow+ColorBold, len(module.Globals), ColorReset)
		for i, instr := range module.Globals {
			fmt.Printf("  %s%d.%s %s\n", ColorGray, i+1, ColorReset, instr.String())
		}
		fmt.Println()
	}

	if len(module.Functions) > 0 {
		fmt.Printf("%sFunções: (%d)%s\n", ColorYellow+ColorBold, len(module.Functions), ColorReset)
		for _, fn := range module.Functions {
			printFunction(fn)
		}
	}
}

func printFunction(fn *ir.Function) {
	fmt.Printf("\n%s%s%s ", ColorBold+ColorBlue, fn.Name, ColorReset)
	if fn.IsExported {
		fmt.Printf("%s(exported)%s ", ColorGreen, ColorReset)
	}
	fmt.Printf("%s{\n", ColorGray)

	// Parâmetros
	if len(fn.Params) > 0 {
		fmt.Printf("  %sParams:%s ", ColorCyan, ColorReset)
		for i, param := range fn.Params {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", param.String())
		}
		fmt.Printf("\n")
	}

	// Instruções
	fmt.Printf("  %sInstructions:%s\n", ColorCyan, ColorReset)
	for i, instr := range fn.Instructions {
		lineNum := i + 1
		fmt.Printf("  %s%3d%s│ %s\n", ColorGray, lineNum, ColorReset, instr.String())
	}

	fmt.Printf("%s}\n\n", ColorGray)
}

// ==========================================
// FUNÇÕES AUXILIARES
// ==========================================

func limitString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func colorIf(condition bool, trueColor, falseColor string) string {
	if condition {
		return trueColor
	}
	return falseColor
}

func printErrorSection(title string, count int) {
	fmt.Printf("\n%s%s (%d):%s\n", ColorRed+ColorBold, title, count, ColorReset)
}

func printError(message string) {
	fmt.Printf("%s❌ %s%s\n", ColorRed, message, ColorReset)
}
