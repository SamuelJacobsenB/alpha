package main

import (
	"fmt"
	"log"

	"github.com/alpha/internal/codegen"
	"github.com/alpha/internal/ir"
	"github.com/alpha/internal/lexer"
	"github.com/alpha/internal/parser"
)

func main() {
	source := `
		// exemplo de código Alpha
		var idade = 25
		var nome = "Samuel"
		if (idade >= 18) {
			return "maior de idade"
		}
	`

	// === Etapa 1: Lexing (Tokens) ===
	fmt.Println("=== TOKENS ===")
	scTokens := lexer.NewScanner(source)
	for {
		tok := scTokens.NextToken()
		fmt.Printf("%s\t%q\tat %d:%d\n", tok.Type, tok.Lexeme, tok.Line, tok.Col)
		if tok.Type == lexer.EOF {
			break
		}
	}

	// === Etapa 2: Parsing (AST) ===
	fmt.Println("\n=== AST ===")
	sc := lexer.NewScanner(source)
	pr := parser.New(sc)
	ast := pr.ParseProgram()

	if pr.HasErrors() {
		fmt.Println("Parse errors:")
		fmt.Println(pr.ErrorsText())
		return
	}

	for _, stmt := range ast.Body {
		fmt.Printf("%#v\n", stmt)
	}

	// === Etapa 3: IR ===
	fmt.Println("\n=== IR ===")
	b := ir.NewBuilder()
	mod, err := b.BuildModule(ast)
	if err != nil {
		log.Fatalf("IR build error: %v", err)
	}
	ir.DumpModule(mod)

	// === Etapa 4: Otimização ===
	ir.ConstFold(mod)
	fmt.Println("\n=== IR After ConstFold ===")
	ir.DumpModule(mod)

	// === Etapa 5: Execução na VM ===
	fmt.Println("\n=== Executando VM ===")
	vm := codegen.NewVM(mod)
	val, err := vm.RunMain()
	if err != nil {
		log.Fatalf("runtime error: %v", err)
	}
	fmt.Println("Valor retornado:", val)
}
