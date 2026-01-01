package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

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
	ParserErrors   []string
	SemanticErrors []string
	ASTStructure   string
	Duration       time.Duration
}

// ==========================================
// FUNÇÃO PRINCIPAL
// ==========================================

func main() {
	fmt.Println(ColorBold + ColorCyan + "╔════════════════════════════════════════════════════╗")
	fmt.Println("║     🧪 ANÁLISE DO ARQUIVO main.alpha     ║")
	fmt.Println("╚════════════════════════════════════════════════════╝" + ColorReset)

	// Ler o arquivo main.alpha
	code, err := ioutil.ReadFile("main.alpha")
	if err != nil {
		fmt.Println(ColorRed + "❌ Erro ao ler o arquivo main.alpha:" + ColorReset)
		fmt.Println(ColorRed + "   " + err.Error() + ColorReset)
		return
	}

	codeStr := string(code)

	fmt.Printf("\n%s📄 CONTEÚDO DO ARQUIVO:%s\n", ColorBold, ColorReset)
	fmt.Println(ColorGray + strings.Repeat("─", 60) + ColorReset)
	fmt.Println(codeStr)
	fmt.Println(ColorGray + strings.Repeat("─", 60) + ColorReset)

	// Executar análise completa
	result := analyzeFile(codeStr)
	printAnalysisResult(result)
}

// ==========================================
// ANÁLISE COMPLETA
// ==========================================

func analyzeFile(code string) AnalysisResult {
	startTime := time.Now()
	result := AnalysisResult{}

	fmt.Printf("\n%s🧪 ETAPA 1: ANÁLISE LÉXICA%s\n", ColorBold, ColorBlue)
	fmt.Println(ColorGray + strings.Repeat("─", 60) + ColorReset)

	// ========== ETAPA 1: LEXER ==========
	fmt.Printf("%s[1/3] 🧪 Analisando tokens...%s", ColorBlue, ColorReset)
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
		fmt.Printf(" %s❌ (%d erros)%s\n", ColorRed, len(lexerErrors), ColorReset)
	} else {
		fmt.Printf(" %s✅ (%d tokens)%s\n", ColorGreen, result.TokenCount, ColorReset)
	}

	// Mostrar tokens detalhados
	fmt.Printf("\n%s📋 TOKENS ENCONTRADOS:%s\n", ColorBold, ColorReset)
	printTokens(tokens)

	// Se houver erros léxicos, parar aqui
	if len(lexerErrors) > 0 {
		result.Success = false
		result.Message = "Erros léxicos encontrados"
		result.Duration = time.Since(startTime)
		return result
	}

	// ========== ETAPA 2: PARSER ==========
	fmt.Printf("\n%s🧪 ETAPA 2: ANÁLISE SINTÁTICA%s\n", ColorBold, ColorYellow)
	fmt.Println(ColorGray + strings.Repeat("─", 60) + ColorReset)

	fmt.Printf("%s[2/3] 📐 Analisando estrutura sintática...%s", ColorYellow, ColorReset)

	// Criar novo scanner para o parser
	scanner = lexer.NewScanner(code)
	p := parser.New(scanner)
	program := p.ParseProgram()

	if p.HasErrors() {
		result.ParserErrors = p.Errors
		fmt.Printf(" %s❌ (%d erros)%s\n", ColorRed, len(p.Errors), ColorReset)

		result.Success = false
		result.Message = "Erros sintáticos encontrados"
		result.Duration = time.Since(startTime)
		return result
	}

	fmt.Printf(" %s✅%s\n", ColorGreen, ColorReset)

	// Mostrar estrutura da AST
	fmt.Printf("\n%s📊 ESTRUTURA DA AST:%s\n", ColorBold, ColorReset)
	astStr := printASTStructure(program, 0)
	result.ASTStructure = astStr

	// ========== ETAPA 3: SEMANTIC ==========
	fmt.Printf("\n%s🧪 ETAPA 3: ANÁLISE SEMÂNTICA%s\n", ColorBold, ColorMagenta)
	fmt.Println(ColorGray + strings.Repeat("─", 60) + ColorReset)

	fmt.Printf("%s[3/3] 🎯 Analisando semântica...%s", ColorMagenta, ColorReset)

	checker := semantic.NewChecker()
	checker.CheckProgram(program)

	if len(checker.Errors) > 0 {
		semanticErrorMsgs := make([]string, len(checker.Errors))
		for i, err := range checker.Errors {
			semanticErrorMsgs[i] = err.Error()
		}
		result.SemanticErrors = semanticErrorMsgs
		fmt.Printf(" %s❌ (%d erros)%s\n", ColorRed, len(checker.Errors), ColorReset)

		result.Success = false
		result.Message = "Erros semânticos encontrados"
		result.Duration = time.Since(startTime)
		return result
	}

	fmt.Printf(" %s✅%s\n", ColorGreen, ColorReset)

	result.Success = true
	result.Message = "Análise completa bem-sucedida"
	result.Duration = time.Since(startTime)

	return result
}

// ==========================================
// FUNÇÕES DE IMPRESSÃO
// ==========================================

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

	fmt.Printf("%-8s %-12s %-20s %s\n",
		ColorBold+"Linha:Col"+ColorReset,
		ColorBold+"Tipo"+ColorReset,
		ColorBold+"Lexema"+ColorReset,
		ColorBold+"Valor"+ColorReset)
	fmt.Println(ColorGray + strings.Repeat("─", 80) + ColorReset)

	for i, tok := range tokens {
		if i >= 50 && i < len(tokens)-1 { // Limitar para não ficar muito grande
			fmt.Printf("... e mais %d tokens\n", len(tokens)-i-1)
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
			color = ColorRed
		}

		fmt.Printf("%-8s %-12s %-20s %s\n",
			fmt.Sprintf("%d:%d", tok.Line, tok.Col),
			color+tokenTypeNames[tok.Type]+ColorReset,
			color+limitString(tok.Lexeme, 18)+ColorReset,
			color+limitString(tok.Value, 30)+ColorReset)
	}
}

func printASTStructure(node interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)

	switch n := node.(type) {
	case *parser.Program:
		fmt.Printf("%s%sProgram%s\n", indentStr, ColorCyan, ColorReset)
		for i, stmt := range n.Body {
			fmt.Printf("%s  %s[%d]%s ", indentStr, ColorGray, i+1, ColorReset)
			printASTStructure(stmt, indent+1)
		}
		return "Program"

	case *parser.PackageDecl:
		fmt.Printf("%s%sPackage: %s%s\n", indentStr, ColorGreen, n.Name, ColorReset)
		return "PackageDecl"

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
		return "ImportDecl"

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
		return "VarDecl"

	case *parser.FunctionDecl:
		fmt.Printf("%s%sFunction: %s%s\n", indentStr, ColorBlue, n.Name, ColorReset)
		fmt.Printf("%s  ReturnType: ", indentStr)
		printASTStructure(n.ReturnType, 0)
		if len(n.Params) > 0 {
			fmt.Printf("%s  Params:\n", indentStr)
			for _, param := range n.Params {
				fmt.Printf("%s    %s: ", indentStr, param.Name)
				printASTStructure(param.Type, 0)
			}
		}
		fmt.Printf("%s  Body (%d statements)\n", indentStr, len(n.Body))
		return "FunctionDecl"

	case *parser.StructDecl:
		fmt.Printf("%s%sStruct: %s%s\n", indentStr, ColorMagenta, n.Name, ColorReset)
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
		return "StructDecl"

	case *parser.IfStmt:
		fmt.Printf("%s%sIf Statement%s\n", indentStr, ColorCyan, ColorReset)
		fmt.Printf("%s  Condition: ", indentStr)
		printASTStructure(n.Cond, 0)
		fmt.Printf("%s  Then (%d statements)\n", indentStr, len(n.Then))
		if len(n.Else) > 0 {
			fmt.Printf("%s  Else (%d statements)\n", indentStr, len(n.Else))
		}
		return "IfStmt"

	case *parser.WhileStmt:
		fmt.Printf("%s%sWhile Loop%s\n", indentStr, ColorCyan, ColorReset)
		fmt.Printf("%s  Condition: ", indentStr)
		printASTStructure(n.Cond, 0)
		fmt.Printf("%s  Body (%d statements)\n", indentStr, len(n.Body))
		return "WhileStmt"

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
		return "ForStmt"

	case *parser.Identifier:
		fmt.Printf("%s%sIdentifier: %s%s\n", indentStr, ColorGreen, n.Name, ColorReset)
		return "Identifier"

	case *parser.IntLiteral:
		fmt.Printf("%s%sInt: %d%s\n", indentStr, ColorYellow, n.Value, ColorReset)
		return "IntLiteral"

	case *parser.StringLiteral:
		fmt.Printf("%s%sString: \"%s\"%s\n", indentStr, ColorGreen, n.Value, ColorReset)
		return "StringLiteral"

	case *parser.BinaryExpr:
		fmt.Printf("%s%sBinary: %s%s\n", indentStr, ColorMagenta, n.Op, ColorReset)
		fmt.Printf("%s  Left: ", indentStr)
		printASTStructure(n.Left, 0)
		fmt.Printf("%s  Right: ", indentStr)
		printASTStructure(n.Right, 0)
		return "BinaryExpr"

	case *parser.CallExpr:
		fmt.Printf("%s%sFunction Call%s\n", indentStr, ColorBlue, ColorReset)
		fmt.Printf("%s  Callee: ", indentStr)
		printASTStructure(n.Callee, 0)
		fmt.Printf("%s  Args (%d):\n", indentStr, len(n.Args))
		for i, arg := range n.Args {
			fmt.Printf("%s    [%d] ", indentStr, i+1)
			printASTStructure(arg, 0)
		}
		return "CallExpr"

	case *parser.PrimitiveType:
		fmt.Printf("%s%sType: %s%s\n", indentStr, ColorCyan, n.Name, ColorReset)
		return "PrimitiveType"

	case *parser.ArrayType:
		fmt.Printf("%s%sArray Type%s\n", indentStr, ColorCyan, ColorReset)
		fmt.Printf("%s  Element: ", indentStr)
		printASTStructure(n.ElementType, 0)
		if n.Size != nil {
			fmt.Printf("%s  Size: ", indentStr)
			printASTStructure(n.Size, 0)
		}
		return "ArrayType"

	case *parser.MapType:
		fmt.Printf("%s%sMap Type%s\n", indentStr, ColorCyan, ColorReset)
		fmt.Printf("%s  Key: ", indentStr)
		printASTStructure(n.KeyType, 0)
		fmt.Printf("%s  Value: ", indentStr)
		printASTStructure(n.ValueType, 0)
		return "MapType"

	default:
		typeName := fmt.Sprintf("%T", n)
		simpleName := strings.TrimPrefix(typeName, "*parser.")
		fmt.Printf("%s%s%s%s\n", indentStr, ColorGray, simpleName, ColorReset)
		return simpleName
	}
}

func printAnalysisResult(result AnalysisResult) {
	fmt.Println("\n" + ColorBold + ColorCyan + "╔════════════════════════════════════════════════════╗")
	fmt.Println("║                 📊 RESUMO DA ANÁLISE               ║")
	fmt.Println("╚════════════════════════════════════════════════════╝" + ColorReset)

	fmt.Printf("\n%sESTATÍSTICAS:%s\n", ColorBold, ColorReset)
	fmt.Printf("   Status:           %s%s%s\n",
		colorIf(result.Success, ColorGreen, ColorRed),
		result.Message,
		ColorReset)
	fmt.Printf("   Tokens:           %s%d%s\n", ColorBold, result.TokenCount, ColorReset)
	fmt.Printf("   Tempo Total:      %s%.2f segundos%s\n", ColorBold, result.Duration.Seconds(), ColorReset)

	// Mostrar erros por etapa
	if len(result.LexerErrors) > 0 {
		fmt.Printf("\n%sERROS LÉXICOS (%d):%s\n", ColorRed, len(result.LexerErrors), ColorReset)
		for i, err := range result.LexerErrors {
			fmt.Printf("   %d. %s\n", i+1, err)
		}
	}

	if len(result.ParserErrors) > 0 {
		fmt.Printf("\n%sERROS SINTÁTICOS (%d):%s\n", ColorRed, len(result.ParserErrors), ColorReset)
		for i, err := range result.ParserErrors {
			fmt.Printf("   %d. %s\n", i+1, err)
		}
	}

	if len(result.SemanticErrors) > 0 {
		fmt.Printf("\n%sERROS SEMÂNTICOS (%d):%s\n", ColorRed, len(result.SemanticErrors), ColorReset)
		for i, err := range result.SemanticErrors {
			fmt.Printf("   %d. %s\n", i+1, err)
		}
	}

	// Mensagem final
	fmt.Print("\n" + ColorBold)
	if result.Success {
		fmt.Println(ColorGreen + "✨ ANÁLISE COMPLETA BEM-SUCEDIDA! ✨" + ColorReset)
	} else {
		fmt.Println(ColorRed + "⚠️  ANÁLISE ENCONTROU ERROS" + ColorReset)
	}
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
