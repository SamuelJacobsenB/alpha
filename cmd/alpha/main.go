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
		testCase("Declarações tipadas", `
			int num
			int num1 = 10

			float flt
			float flt1 = 3.14

			string str
			string str1 = "Hello World!"

			bool bl
			bool bl1 = true
		`)

		testCase("Declarações autoinferidas", `
			var num = 10

			var flt = 3.14

			var str = "Hello World!"

			var bl = false
		`)

		testCase("Declarações constantes", `
			const num = 10

			const flt = 3.14

			const str = "Hello World!"

			const bl = false
		`)

		testCase("Declarações de arrays", `
			int[] num
			int[] num1 = [10, 20]

			float[] flt
			float[] flt1 = [3.14, 5]

			string[] str
			string[] str1 = ["Hello World!"]

			bool[] bl
			bool[] bl1 = [true, true, false]

			int[5] fixed = [1, 2, 3, 4, 5]
		`)

		testCase("Declarações de referências", `
			var num = 10
			int* num1 = &num
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

		testCase("Type alias com genéricos", `
			generic<T> type Pair [T, T]
			Pair coordinates = generic<int> [10, 20]
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
		testCase("Struct simples", `
			struct Point {
				int x
				int y
			}
			
			Point p = Point {
				x: 10,
				y: 20
			}
		`)

		testCase("Struct com métodos", `
			struct Rectangle {
				int width
				int height
			}
			
			Rectangle rect = Rectangle {
				width: 10,
				height: 20
			}
		`)

		testCase("Struct aninhado", `
			struct Address {
				string street
				string city
			}
			
			struct Person {
				string name
				int age
				Address address
			}
			
			Person person = Person {
				name: "Alice",
				age: 30,
				address: Address {
					street: "Main St",
					city: "New York"
				}
			}
		`)

		testCase("Struct com tipo genérico", `
			generic<T> struct Box {
				T content
			}
			
			Box intBox = generic<int> Box {
				content: 42
			}
			Box stringBox = generic<string> Box {
				content: "hello"
			}
		`)

	case "classes":
		testCase("Classe simples", `
			class Person {
				string name
				int age
				
				constructor(string name, int age) {
					this.name = name
					this.age = age
				}
				
				string method getName() {
					return this.name
				}
			}
			
			Person person = new Person("Alice", 30)
			string name = person.getName()
		`)

		testCase("Classe com herança implícita", `
			class Animal {
				string species
				
				constructor(string species) {
					this.species = species
				}
				
				string method getSpecies() {
					return this.species
				}
			}
			
			class Dog {
				string breed
				
				constructor(string breed) {
					this.breed = breed
				}
				
				string method bark() {
					return "Woof!"
				}
			}
		`)

		testCase("Classe com métodos estáticos", `
			class MathUtils {
				int method max(int a, int b) {
					return a > b ? a : b
				}
				
				float method pi() {
					return 3.14159
				}
			}
			
			int maximum = MathUtils.max(10, 20)
			float piValue = MathUtils.pi()
		`)

		testCase("Classe com tipo genérico", `
			generic<T> class Container {
				T[] items
				
				constructor() {
					this.items = []
				}
				
				void method add(T item) {
					// this will add
				}
				
				T method get(int index) {
					// this will get
				}
			}
			
			Container intContainer = genric<int> new Container()
			intContainer.add(10)
			intContainer.add(20)
			intContainer.get(0)
		`)

		testCase("Classe com múltiplos genéricos", `
			generic<K, V> class Pair {
				K key
				V value
				
				constructor(K key, V value) {
					this.key = key
					this.value = value
				}
				
				K method getKey() {
					return this.key
				}
				
				V method getValue() {
					return this.value
				}
			}
			
			Pair entry = generic<string, int> new Pair("age", 30)
			string key = entry.getKey()
			int value = entry.getValue()
		`)

	default:
		panic("Give a correct case name")
	}

	fmt.Println("\n=== TODOS OS TESTES CONCLUÍDOS ===")
}
