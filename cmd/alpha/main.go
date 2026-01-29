package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alpha/internal/codegen"
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
	GeneratedCode  string
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
	printBanner("🧪 COMPILADOR ALPHA - FULL STACK")

	// Verificar argumentos
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		if len(os.Args) < 3 {
			printError("Uso: alpha analyze <arquivo.alpha>")
			return
		}
		analyzeFileCommand(os.Args[2])

	case "compile":
		if len(os.Args) < 3 {
			printError("Uso: alpha compile <arquivo.alpha> [output.go]")
			return
		}
		output := "output.go"
		if len(os.Args) > 3 {
			output = os.Args[3]
		}
		compileFileCommand(os.Args[2], output)

	case "run":
		if len(os.Args) < 3 {
			printError("Uso: alpha run <arquivo.alpha>")
			return
		}
		runFileCommand(os.Args[2])

	default:
		printError(fmt.Sprintf("Comando desconhecido: %s", command))
		printHelp()
	}
}

func printHelp() {
	fmt.Println(ColorBold + ColorCyan + "Uso: alpha <comando> [argumentos]" + ColorReset)
	fmt.Println()
	fmt.Println("Comandos disponíveis:")
	fmt.Println("  analyze <arquivo.alpha>  - Analisa o arquivo e mostra detalhes")
	fmt.Println("  compile <arquivo.alpha> [output.go] - Compila para Go")
	fmt.Println("  run <arquivo.alpha>      - Compila e executa")
	fmt.Println()
}

func analyzeFileCommand(filename string) {
	printBanner(fmt.Sprintf("🧪 ANÁLISE DO ARQUIVO %s", filename))

	code, err := os.ReadFile(filename)
	if err != nil {
		printError("Erro ao ler o arquivo " + filename + ": " + err.Error())
		return
	}

	codeStr := string(code)
	lines := strings.Split(codeStr, "\n")

	printSection("📄 CONTEÚDO DO ARQUIVO", ColorWhite)
	fmt.Println(ColorGray + strings.Repeat("─", 80) + ColorReset)

	for i, line := range lines {
		fmt.Printf("%s%3d │ %s%s\n", ColorGray, i+1, ColorReset, line)
	}

	fmt.Println(ColorGray + strings.Repeat("─", 80) + ColorReset)

	result := analyzeFile(codeStr, lines)
	printAnalysisResult(result)

	if result.Success && result.IRModule != nil {
		printIR(result.IRModule)
	}
}

func compileFileCommand(inputFile, outputFile string) {
	printBanner(fmt.Sprintf("🛠️  COMPILANDO %s → %s", inputFile, outputFile))

	code, err := os.ReadFile(inputFile)
	if err != nil {
		printError("Erro ao ler o arquivo " + inputFile + ": " + err.Error())
		return
	}

	codeStr := string(code)
	lines := strings.Split(codeStr, "\n")

	// Executar análise completa
	result := analyzeFile(codeStr, lines)
	if !result.Success {
		printError("Compilação abortada devido a erros na análise")
		return
	}

	// Gerar código Go
	if result.GeneratedCode == "" && result.IRModule != nil {
		printSection("🔧 GERANDO CÓDIGO GO", ColorGreen)
		generator := codegen.NewCodeGenerator(nil)
		result.GeneratedCode = generator.GenerateCode(result.IRModule)
	}

	if result.GeneratedCode != "" {
		// Salvar arquivo
		err = os.WriteFile(outputFile, []byte(result.GeneratedCode), 0644)
		if err != nil {
			printError("Erro ao salvar arquivo: " + err.Error())
			return
		}

		printSuccess(fmt.Sprintf("✅ Código Go gerado com sucesso em: %s", outputFile))

		// Mostrar estatísticas
		lineCount := strings.Count(result.GeneratedCode, "\n")
		fileInfo, _ := os.Stat(outputFile)
		fmt.Printf("%s📊 Estatísticas:%s\n", ColorBold+ColorWhite, ColorReset)
		fmt.Printf("   Linhas de código: %d\n", lineCount)
		fmt.Printf("   Tamanho do arquivo: %.2f KB\n", float64(fileInfo.Size())/1024)
	} else {
		printError("❌ Falha ao gerar código Go")
	}
}

func runFileCommand(filename string) {
	printBanner(fmt.Sprintf("🚀 EXECUTANDO %s", filename))

	// Primeiro compilar
	tempOutput := "temp_alpha_output.go"
	compileFileCommand(filename, tempOutput)

	// Verificar se a compilação foi bem-sucedida
	if _, err := os.Stat(tempOutput); err != nil {
		printError("Compilação falhou, não é possível executar")
		return
	}

	// Tentar executar o código Go
	printSection("▶️  EXECUTANDO PROGRAMA", ColorCyan)

	// Em um sistema real, você compilaria e executaria o Go
	// Por enquanto, apenas mostra que seria executado
	printSuccess("✅ Programa compilado com sucesso")
	fmt.Println(ColorYellow + "💡 Dica: Para executar, use: go run " + tempOutput + ColorReset)

	// Limpeza (opcional)
	// os.Remove(tempOutput)
}

func analyzeFile(code string, lines []string) AnalysisResult {
	startTime := time.Now()
	result := AnalysisResult{
		Lines: lines,
	}

	// ========== ETAPA 1: LEXER ==========
	printSection("🧪 ETAPA 1: ANÁLISE LÉXICA", ColorBlue)
	printStep("Analisando tokens...", 1, 6)
	scanner := lexer.NewScanner(code)

	// Coletar tokens
	tokens := []lexer.Token{}
	for {
		tok := scanner.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == lexer.EOF {
			break
		}
	}
	result.Tokens = tokens
	result.TokenCount = len(tokens)

	// Verificar erros léxicos
	hasLexerErrors := false
	for _, tok := range tokens {
		if tok.Type == lexer.ERROR {
			hasLexerErrors = true
			result.LexerErrors = append(result.LexerErrors, tok.Value)
		}
	}

	if hasLexerErrors {
		printStepResult("❌", false)
		result.Success = false
		result.Message = "Erros léxicos encontrados"
		return result
	}
	printStepResult("✅", true)

	// ========== ETAPA 2: PARSER ==========
	printSection("🧪 ETAPA 2: ANÁLISE SINTÁTICA", ColorYellow)
	printStep("Analisando estrutura sintática...", 2, 6)
	scanner = lexer.NewScanner(code)
	p := parser.New(scanner)
	program := p.ParseProgram()

	if p.HasErrors() {
		printStepResult("❌", false)
		result.Success = false
		result.Message = "Erros sintáticos encontrados"

		// Converter erros do parser
		for _, err := range p.Errors {
			line, col, msg := parseErrorPosition(err)
			result.ParserErrors = append(result.ParserErrors, parserError{
				Line:    line,
				Col:     col,
				Message: msg,
			})
		}
		return result
	}
	printStepResult("✅", true)

	// ========== ETAPA 3: SEMANTIC ==========
	printSection("🧪 ETAPA 3: ANÁLISE SEMÂNTICA", ColorMagenta)
	printStep("Analisando semântica...", 3, 6)
	checker := semantic.NewChecker()
	checker.CheckProgram(program)

	if len(checker.Errors) > 0 {
		printStepResult("❌", false)
		result.Success = false
		result.Message = "Erros semânticos encontrados"

		// Converter erros semânticos
		for _, err := range checker.Errors {
			result.SemanticErrors = append(result.SemanticErrors, semanticError{
				Line:    err.Line,
				Col:     err.Col,
				Message: err.Msg,
			})
		}
		return result
	}
	printStepResult("✅", true)

	// ========== ETAPA 4: GERAÇÃO DE IR ==========
	printSection("🧪 ETAPA 4: GERAÇÃO DE IR", ColorCyan)
	printStep("Gerando IR (Representação Intermediária)...", 4, 6)
	generator := ir.NewGenerator(checker)
	irModule := generator.Generate(program)
	printStepResult("✅", true)

	// ========== ETAPA 5: OTIMIZAÇÃO DE IR ==========
	printSection("🧪 ETAPA 5: OTIMIZAÇÃO DE IR", ColorGreen)
	printStep("Aplicando Constant Folding e Dead Code Elimination...", 5, 6)
	optimizer := ir.NewOptimizer(irModule)
	optimizer.Optimize()
	result.IRModule = irModule
	printStepResult("✅", true)

	// ========== ETAPA 6: GERAÇÃO DE CÓDIGO ==========
	printSection("🧪 ETAPA 6: GERAÇÃO DE CÓDIGO GO", ColorGreen)
	printStep("Gerando código Go otimizado...", 6, 6)
	codegen := codegen.NewCodeGenerator(checker)
	result.GeneratedCode = codegen.GenerateCode(irModule)
	printStepResult("✅", true)

	result.Success = true
	result.Message = "Análise, Otimização e Geração de Código completas com sucesso"
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

func printSuccess(message string) {
	fmt.Printf("%s✅ %s%s\n", ColorGreen, message, ColorReset)
}

func printError(message string) {
	fmt.Printf("%s❌ %s%s\n", ColorRed, message, ColorReset)
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
		if result.GeneratedCode != "" {
			fmt.Println(ColorCyan + "📝 Código Go gerado com sucesso." + ColorReset)
		}
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
