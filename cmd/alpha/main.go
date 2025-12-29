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
	var caseName string
	fmt.Print("Test Case: ")
	fmt.Scan(&caseName)
	fmt.Println("========================")

	switch caseName {
	case "variables":
		testCase("Variáveis Simples", `
        int num
		int num1 = 10
		var num2 = 20
		const num3 = 30
    `)

		testCase("Variáveis Nulas", `
		int? num4 // null
		int? num5 = 10
		int? num6 = null
    `)

		testCase("Arrays", `
		string[2] arr
		string[2] arr2 = ["Hello", "World"]
		var arr3 = ["Hello", "World"] // string[2] 
		string[] linked
	`)

		testCase("Matrixes", `
		int[2][2] matrix
		int[2][2] matrix2 = [
			[1, 2],
			[3, 4]
		]
		var matrix3 = [
			[1, 2],
			[3, 4]
		]
		int[][] matrix4
		int[][] matrix5 = [
			[1, 2],
			[3, 4]
		]
	`)

		testCase("Ponteiros", `
		int value = 42
		int* ptr = &value
		int** ptr2 = &ptr
		var val = *ptr
		`)

		testCase("Maps", `
		map<int, string> mapTest = map<int, string>{
			1: "Hello",
			2: "Bye",
		}

		var mapTest2 = map<int, string>{
			1: "Hello",
			2: "Bye",
		}

		map<int, string> mapTest3
		`)

		testCase("Union types", `
		string | float types2 // ""
		types2 = 3.1415

		string | float | int types3 // ""
		types3 = "Hello"
		`)

	case "conditions":
		testCase("If-else simples", `
			int x = 10
			if(x > 5) {
				x = 100
			} else {
				x = 0
			}
		`)

		testCase("If-else if", `
			int score = 85
			if(score >= 90) {
				string grade = "A"
			} else if(score >= 80) {
				string grade = "B"
			} else if(score >= 70) {
				string grade = "C"
			} else {
				string grade = "F"
			}
		`)

		testCase("If sem chaves", `
			int x = 5
			if(x > 0)
				x++
			else
				x--
		`)

		testCase("Switch statement", `
			int day = 3
			switch(day) {
				case 1:
					string name = "Monday"
				case 2:
					string name = "Tuesday"
				case 3:
					string name = "Wednesday"
				default:
					string name = "Unknown"
			}
		`)

		testCase("Switch com múltiplos cases", `
			int month = 2
			int days = 0
			switch(month) {
				case 1: case 3: case 5: case 7:
				case 8: case 10: case 12:
					days = 31
				case 4: case 6: case 9: case 11:
					days = 30
				case 2:
					days = 28
				default:
					days = -1
			}
		`)

		testCase("Operador ternário complexo", `
			int a = 10
			int b = 20
			int max = a > b ? a : b
			int min = a < b ? a : b
			string result = a == b ? "equal" : "different"
		`)

	case "loops":
		testCase("For loop tradicional", `
			for(int i = 0; i < 10; i++) {
				int x = i * 2
			}
		`)

		testCase("For-in loop", `
			int[] numbers = [1, 2, 3, 4, 5]
			for(num in numbers) {
				int squared = num * num
			}
		`)

		testCase("For-in loop com índice", `
			string[] names = ["Alice", "Bob", "Charlie"]
			for(i, name in names) {
				string greeting = "Hello " + name
			}
		`)

		testCase("While loop", `
			int counter = 0
			while(counter < 10) {
				counter++
			}
		`)

		testCase("Do-while loop", `
			int x = 0
			do {
				x++
			} while(x < 5)
		`)

		testCase("Loop com break e continue", `
			int i = 0
			while(true) {
				if(i >= 10) {
					break
				}
				if(i % 2 == 0) {
					i++
					continue
				}
				i++
			}
		`)

	case "functions":
		testCase("Função simples", `
			int function sum(int a, int b) {
				return a + b
			}
			
			int result = sum(5, 10)
		`)

		testCase("Função sem retorno", `
			void function printMessage(string msg) {
				// imprime mensagem
			}
			
			printMessage("Hello")
		`)

		testCase("Função com tipo genérico sem chamada", `
			generic<T> T function identity(T value) {
				return value
			}
		`)

		testCase("Função com tipo genérico", `
			generic<T> T function identity(T value) {
				return value
			}
			
			int num = generic<int> identity(5)
			string text = generic<string> identity("test")
		`)

		testCase("Função com múltiplos parâmetros genéricos", `
			generic<T, U> T function first(T a, U b) {
				return a
			}
			
			int result = generic<int, string> first(10, "hello")
		`)

		testCase("Função com array como parâmetro", `
			int function sumArray(int[] numbers) {
				int total = 0
				for(n in numbers) {
					total += n
				}
				return total
			}
			
			int[] nums = [1, 2, 3, 4, 5]
			int total = sumArray(nums)
		`)

	case "types":
		testCase("Type alias simples", `
			type Age int
			Age myAge = 25
		`)

		testCase("Union type", `
			type Number int | float
			Number num1 = 10
			Number num2 = 3.14
		`)

		testCase("Nullable types", `
			int? maybeNumber = null
			string? maybeString = "hello"
			float? maybeFloat = 3.14
		`)

		testCase("Pointer types", `
			int value = 10
			int* ptr = &value
			int** ptrToPtr = &ptr
		`)

		testCase("Map types", `
			map<string, int> scores = map<string, int>{
				"Alice": 95,
				"Bob": 87,
				"Charlie": 92
			}

			var scores1 = map<string, int>{
				"Alice": 95,
				"Bob": 87,
				"Charlie": 92
			}
		`)

		testCase("Set types", `
			set<int> numbers = set<int>{1, 2, 3, 4, 5}
			var names = set<string>{"Alice", "Bob"}
		`)

	case "structs":
		// CASO 1: Struct Simples e Instanciação
		testCase("Struct Simples (Dados)", `
			// simple definition
			struct Message {
				string text
				string sender
			}

			// instantiation
			Message message = Message {
				text: "This is a message",
				sender: "Samuel",
			}
		`)

		// CASO 2: Struct com Genéricos
		testCase("Struct Genérico", `
			generic<T> struct Car {
				T motor
				int year
			}

			// instantiation with generic type
			Car car = generic<string> Car {
				motor: "654MT4",
				year: 2001,
			}
		`)

		// CASO 3: Struct com Campos Privados
		testCase("Struct com Private", `
			struct Man {
				int age
				private string cpf
			}

			var man = Man {
				age: 24
				// cpf não deve ser acessível diretamente aqui se houver checagem semântica,
				// mas sintaticamente o lexer deve reconhecer 'private'
			}
		`)

		// CASO 4: Implementação de Métodos e Construtor (Novo Padrão)
		testCase("Struct Implementation (init & self)", `
			struct User {
				string email
				private string password
			}

			implement User {
				// Construtor
				init(string email, string password) {
					self.email = email
					self.password = password
				}

				// Método com genéricos
				generic<T> bool validatePassword() {
					// Exemplo do txt: typeof(password) ou self.password
					return typeof(self.password) == T
				} 
			}

			// Uso do construtor
			var user = User("email@email.com", "Senha12345")
		`)

	default:
		panic("Give a correct case name")
	}

	fmt.Println("\n=== TODOS OS TESTES CONCLUÍDOS ===")
}
