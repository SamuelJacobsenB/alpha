package main

import (
	"fmt"

	"github.com/alpha/internal/lexer"
	"github.com/alpha/internal/parser"
)

func testCase(name, src string) {
	fmt.Printf("\n=== TEST: %s ===\n", name)
	fmt.Println("Código:")
	fmt.Println(src)

	sc := lexer.NewScanner(src)
	p := parser.New(sc)
	prog := p.ParseProgram()

	if p.HasErrors() {
		fmt.Println("❌ Erros:")
		fmt.Println(p.ErrorsText())
	} else {
		fmt.Println("✅ Parsing bem-sucedido!")
		fmt.Printf("   Statements: %d\n", len(prog.Body))
	}
}

func main() {
	// Teste 1: For loop com declaração
	testCase("For Loop com Declaração", `
		int a;
		for(var i = 5; i < 10; i++) {
			a += i;
		}
	`)

	// Teste 2: Switch statement
	testCase("Switch Statement", `
		int sum = 5
		switch(sum) {
			case 1:
				sum++
			case 2:
				sum--
			default:
				sum = 0
		}
	`)

	// Teste 3: Operador ternário
	testCase("Operador Ternário", `
		int b = 10
		int a = b == 10 ? 1000 : -1000
	`)

	// Teste 4: Classe completa
	testCase("Classe Completa", `
		class User {
			string name
			int age
			string cpf

			constructor(string name, int age, string cpf) {
				this.name = name
				this.age = age
				this.cpf = cpf
			}

			string method cpf() {
				return this.cpf
			}

			<T> string method values(T generic) {
				return this.name
			}
		}

		User user = new User("Samuel", 15, "111-111-111.11")
		var cpf = user.cpf()
		var name = user.name
	`)

	// Teste 5: Type declarations
	testCase("Type Declarations", `
		type Number int | float
		Number num = 5.5

		type Message {
			string text
			string sender
		}
		Message message = {
			text: "This is a message",
			sender: "Samuel",
		}

		<T> type Car {
			T motor
			int year
		}
	`)

	// Teste 6: Múltiplas features juntas
	testCase("Features Combinadas", `
		class Calculator {
			int result

			constructor() {
				this.result = 0
			}

			int method add(int a, int b) {
				return a + b
			}
		}

		Calculator calc = new Calculator()
		int x = 10
		int y = 20
		int result = calc.add(x, y)

		switch(result) {
			case 30:
				result++
			default:
				result = 0
		}

		int final = result > 0 ? result : -1
	`)

	// Teste 7: Nested structures
	testCase("Estruturas Aninhadas", `
		for(int i = 0; i < 10; i++) {
			if(i % 2 == 0) {
				switch(i) {
					case 2:
						i++
					case 4:
						i--
					default:
						i = 0
				}
			} else {
				int val = i > 5 ? 100 : 50
			}
		}
	`)

	// Teste 8: Declarações sem inicialização
	testCase("Declarações Sem Inicialização", `
		int a;
		string b;
		float c;
		int[] arr;
		int? nullable;
		
		for(int i; i < 10; i++) {
			a += i
		}
	`)

	fmt.Println("\n=== TODOS OS TESTES CONCLUÍDOS ===")
}
