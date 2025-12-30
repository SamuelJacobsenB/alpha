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

		// === NOVOS TESTES ===
		testCase("Atribuição Composta e Expressões", `
        int a = 10
        a += 5
        a -= 2
        a *= 3
        a /= 2
        var b = (a + 10) * 2
        `)

		testCase("Concatenação de Strings", `
        string firstName = "John"
        string lastName = "Doe"
        string fullName = firstName + " " + lastName
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

		testCase("If Aninhado", `
            int x = 10
            int y = 20
            if (x > 5) {
                if (y > 15) {
                    x = y
                } else {
                    x = 0
                }
            }
        `)

		testCase("Lógica Booleana Complexa", `
            bool isActive = true
            bool isAdmin = false
            if ((isActive && !isAdmin) || (5 > 3)) {
                print("Access allowed")
            }
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

		// === NOVOS TESTES ===
		testCase("Loops Aninhados", `
            for(int i = 0; i < 5; i++) {
                for(int j = 0; j < 5; j++) {
                   int k = i * j
                }
            }
        `)

		testCase("For Loop Infinito (Sintaxe)", `
            for(;;) {
                break
            }

			while(true) {
				break
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

		// === NOVOS TESTES ===
		testCase("Recursividade", `
            int function factorial(int n) {
                if (n <= 1) return 1
                return n * factorial(n - 1)
            }
        `)

		testCase("Retornando Struct/Objeto", `
            struct Point {
				int x
				int y 
			}
            
            Point function createPoint(int a, int b) {
                return Point { x: a, y: b }
            }
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

		testCase("Generics Aninhados (Nested)", `
            map<string, set<int>> userGroups
            list<map<string, string>> dataList
        `)

	case "structs":
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

		testCase("Struct com Private", `
            struct Man {
                int age
                private string cpf
            }

            var man = Man {
                age: 24
            }
        `)

		testCase("Struct Implementation (init & self)", `
            struct User {
                string email
                private string password
                public int age
            }

            implement User {
                init(string email, string password, int age) {
                    self.email = email
                    self.password = password
                    self.age = age
                }

                generic<T> bool validatePassword() {
                    return typeof(password) == T
                } 
            }

            var user = User { email: "email@email.com", password: "Senha12345", age: 20 }
        `)

		testCase("Composição de Structs (Aninhadas)", `
            struct Address {
                string street
                string city
            }
            
            struct Person {
                string name
                Address addr
            }
            
            var p = Person {
                name: "John",
                addr: Address {
                    street: "Main St",
                    city: "NY",
                },
            }
        `)

		testCase("Struct com Array", `
            struct Group {
                string name
                string[] members
            }
            
            var g = Group {
                name: "Admins",
                members: ["Alice", "Bob"],
            }
        `)

	case "expressions":
		testCase("Precedência Matemática", `
            int x = 10 + 5 * 2
            int y = (10 + 5) * 2
            int z = 100 / 10 * 2 // deve ser 20, esquerda para direita
        `)

		testCase("Operadores Unários e Lógicos", `
            bool check = !true
            int i = 10
            i++
            int j = --i
            bool res = (i > 5) && (j < 20)
        `)

	default:
		panic("Give a correct case name")
	}

	fmt.Println("\n=== TODOS OS TESTES CONCLUÍDOS ===")
}
