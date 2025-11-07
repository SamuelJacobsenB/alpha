package main

import (
	"fmt"

	"github.com/alpha/internal/lexer"
	"github.com/alpha/internal/parser"
)

func main() {
	source := `
        // Função simples
        int function soma(int a, int b) {
            return a + b
        }
        
        // Função void
        void function dizerOla(string nome) {}
        
        // Função com genéricos
        void [T] function processar(T item) {}
        
        // Variável com função anônima
        var funcao = void function() {}
        
        // Chamadas de função
        int resultado = soma(5, 3)
        dizerOla("Mundo")
        processar<int>(42)
        
        // Função recursiva
        int function fatorial(int n) {
            if (n <= 1) {
                return 1
            }
            return n * fatorial(n - 1)
        }
    `

	fmt.Println("=== TESTE SEM LIMITES DA LINGUAGEM ALPHA ===")

	sc := lexer.NewScanner(source)
	pr := parser.New(sc)
	ast := pr.ParseProgram()

	if pr.HasErrors() {
		fmt.Println("ERROS ENCONTRADOS:")
		fmt.Println(pr.ErrorsText())
	} else {
		fmt.Printf("✅ Análise concluída com sucesso! %d statements\n\n", len(ast.Body))

		for i, stmt := range ast.Body {
			fmt.Printf("%d. %T\n", i+1, stmt)

			switch s := stmt.(type) {
			case *parser.FunctionDecl:
				fmt.Printf("   FUNÇÃO: %s", s.Name)
				if len(s.Generics) > 0 {
					fmt.Printf(" [%d genéricos]", len(s.Generics))
				}
				fmt.Printf(" -> %d parâmetros, %d statements\n", len(s.Params), len(s.Body))

			case *parser.VarDecl:
				fmt.Printf("   VAR: %s", s.Name)
				if s.Init != nil {
					if _, ok := s.Init.(*parser.FunctionExpr); ok {
						fmt.Printf(" = função anônima")
					}
				}
				fmt.Println()

			case *parser.ExprStmt:
				if call, ok := s.Expr.(*parser.CallExpr); ok {
					fmt.Printf("   CHAMADA: ")
					if ident, ok := call.Callee.(*parser.Identifier); ok {
						fmt.Printf("%s", ident.Name)
					}
					fmt.Printf("(%d args)\n", len(call.Args))
				}
			}
		}
	}
}
