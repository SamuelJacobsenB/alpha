package main

import (
	"fmt"

	"github.com/alpha/internal/lexer"
	"github.com/alpha/internal/parser"
)

const source = `
	int a;
	for(int i; i < 10; i++) {
		a += i
	}
`

func main() {
	fmt.Print(source)

	// Análisis léxico
	fmt.Println("\n📋 TOKENS:")
	scanner := lexer.NewScanner(source)
	for {
		token := scanner.NextToken()
		fmt.Printf("%-10s %q\n", tokenTypeName(token.Type), token.Lexeme)

		if token.Type == lexer.EOF || token.Type == lexer.ERROR {
			break
		}
	}

	// Análisis sintáctico
	fmt.Println("\n🌳 ÁRBOL SINTÁCTICO:")
	parser := parser.New(lexer.NewScanner(source))
	ast := parser.ParseProgram()

	if parser.HasErrors() {
		fmt.Println("❌ Errores de parsing:")
		for _, err := range parser.Errors {
			fmt.Println(" -", err)
		}
	} else {
		fmt.Printf("✅ Programa analizado correctamente\n")
		fmt.Printf("   %d declaraciones encontradas\n", len(ast.Body))
	}
}

func tokenTypeName(t lexer.TokenType) string {
	switch t {
	case lexer.EOF:
		return "EOF"
	case lexer.ERROR:
		return "ERROR"
	case lexer.KEYWORD:
		return "KEYWORD"
	case lexer.IDENT:
		return "IDENT"
	case lexer.INT:
		return "INT"
	case lexer.FLOAT:
		return "FLOAT"
	case lexer.STRING:
		return "STRING"
	case lexer.OP:
		return "OPERADOR"
	case lexer.GENERIC:
		return "GENÉRICO"
	default:
		return "DESCONOCIDO"
	}
}
