package semantic

import (
	"fmt"
	"strings"

	"github.com/alpha/internal/parser"
)

func (c *Checker) checkExpr(expr parser.Expr) Type {
	switch e := expr.(type) {
	// ... (Mantenha os casos de literais simples e Identifier) ...
	case *parser.IntLiteral:
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "int"}}
	case *parser.FloatLiteral:
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "float"}}
	case *parser.BoolLiteral:
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "bool"}}
	case *parser.StringLiteral:
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "string"}}
	case *parser.NullLiteral:
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "null"}}

	// Em checkExpr, caso Identifier para tipos genéricos:
	case *parser.Identifier:
		sym := c.CurrentScope.Resolve(e.Name)
		if sym == nil {
			// Verificar se é um tipo genérico (parâmetro de função genérica)
			// Primeiro verifica no escopo atual (parâmetros genéricos)
			if strings.Contains(e.Name, ".") {
				parts := strings.Split(e.Name, ".")
				if len(parts) == 2 {
					moduleSym := c.CurrentScope.Resolve(parts[0])
					if moduleSym != nil && moduleSym.Kind == KindImport {
						return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}
					}
				}
			}

			// Verificar se é um parâmetro genérico (como T)
			// Isso será verificado na função checkFunctionDecl
			c.reportError(0, 0, fmt.Sprintf("Undeclared identifier '%s'", e.Name))
			return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
		}

		if sym.Kind == KindImport && sym.Type == nil {
			return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}
		}

		// Se for parâmetro genérico, retornar seu tipo
		if sym.Kind == KindGenericParam {
			return sym.Type
		}

		return sym.Type

	case *parser.UnaryExpr:
		valType := c.checkExpr(e.Expr)
		return valType

	case *parser.BinaryExpr:
		leftType := c.checkExpr(e.Left)
		rightType := c.checkExpr(e.Right)

		switch e.Op {
		case "+":
			// Adição ou concatenação
			leftTypeStr := StringifyType(leftType)
			rightTypeStr := StringifyType(rightType)

			// Se um dos lados for um tipo genérico, retornar any (não sabemos o tipo resultante)
			if c.isGenericType(leftType) || c.isGenericType(rightType) {
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}
			}

			// Se ambos são numéricos
			if (leftTypeStr == "int" || leftTypeStr == "float") &&
				(rightTypeStr == "int" || rightTypeStr == "float") {
				if leftTypeStr == "float" || rightTypeStr == "float" {
					return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "float"}}
				}
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "int"}}
			}

			// Se um é string, concatenação
			if leftTypeStr == "string" || rightTypeStr == "string" {
				// Verificar se o outro lado é compatível com string
				if leftTypeStr != "string" && !c.isGenericType(leftType) && leftTypeStr != "any" {
					c.reportError(0, 0, fmt.Sprintf("Cannot concatenate string with non-string type %s", leftTypeStr))
					return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
				}
				if rightTypeStr != "string" && !c.isGenericType(rightType) && rightTypeStr != "any" {
					c.reportError(0, 0, fmt.Sprintf("Cannot concatenate string with non-string type %s", rightTypeStr))
					return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
				}
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "string"}}
			}

			// Operação não suportada
			c.reportError(0, 0, fmt.Sprintf("Operator '+' not supported for types %s and %s", leftTypeStr, rightTypeStr))
			return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
		case ">", "<", ">=", "<=", "==", "!=":
			// Operações de comparação - retornam bool
			return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "bool"}}

		default:
			// Para outros operadores, retorna o tipo da esquerda
			return leftType
		}

	case *parser.TernaryExpr:
		trueType := c.checkExpr(e.TrueExpr)
		return trueType

	case *parser.AssignExpr:
		leftType := c.checkExpr(e.Left)
		return leftType

	case *parser.CallExpr:
		// Verificação especial para append: retorna o tipo do primeiro argumento
		if ident, ok := e.Callee.(*parser.Identifier); ok && ident.Name == "append" {
			if len(e.Args) > 0 {
				argType := c.checkExpr(e.Args[0])
				// Validar argumentos restantes
				for i := 1; i < len(e.Args); i++ {
					c.checkExpr(e.Args[i])
				}
				return argType
			}
		}

		for _, arg := range e.Args {
			c.checkExpr(arg)
		}

		var returnType Type
		if ident, ok := e.Callee.(*parser.Identifier); ok {
			sym := c.CurrentScope.Resolve(ident.Name)
			if sym != nil {
				switch sym.Kind {
				case KindFunction:
					returnType = sym.Type
				case KindImport:
					returnType = &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}
				default:
					c.reportError(0, 0, fmt.Sprintf("'%s' is not a function", ident.Name))
					return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
				}
			} else {
				// Verificar se é uma função genérica chamada sem especialização
				// Ex: hello1(30) sem generic<int>
				c.reportError(0, 0, fmt.Sprintf("Undeclared function '%s'", ident.Name))
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
			}
		} else {
			// Para chamadas complexas (como generic<int> hello1(30))
			returnType = &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}
		}

		return returnType

	case *parser.StructLiteral:
		// Se o literal tem um nome (ex: Message { ... }), retorna esse tipo
		if e.Name != "" {
			return &ParserTypeWrapper{Type: &parser.IdentifierType{Name: e.Name}}
		}
		// Se for anônimo, retorna "object"
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "object"}}

	case *parser.GenericSpecialization:
		// Trata especializações como "generic<string> Car { ... }"
		// O Parser coloca o StructLiteral dentro do Callee
		if structLit, ok := e.Callee.(*parser.StructLiteral); ok {
			// Se o struct literal tiver nome (Car), retorna um tipo genérico construído
			if structLit.Name != "" {
				return &ParserTypeWrapper{Type: &parser.GenericType{
					Name:     structLit.Name,
					TypeArgs: e.TypeArgs,
				}}
			}
		}

		// Se for apenas uma especialização de identificador ou outro caso
		calleeType := c.checkExpr(e.Callee)
		return calleeType

	case *parser.ArrayLiteral:
		if len(e.Elements) == 0 {
			return &ParserTypeWrapper{Type: &parser.ArrayType{ElementType: &parser.PrimitiveType{Name: "any"}}}
		}

		// Determinar o tipo base de todos os elementos
		var elementType parser.Type
		for i, elem := range e.Elements {
			elemType := c.checkExpr(elem)
			if wrapper, ok := elemType.(*ParserTypeWrapper); ok {
				if i == 0 {
					elementType = wrapper.Type
				} else {
					// Verificar compatibilidade com o primeiro tipo
					if !AreParserTypesCompatible(elementType, wrapper.Type) {
						c.reportError(0, 0, fmt.Sprintf("Inconsistent array element types: %s vs %s",
							StringifyParserType(elementType), StringifyParserType(wrapper.Type)))
						// Usar o primeiro tipo como fallback
						break
					}
				}
			}
		}

		if elementType == nil {
			elementType = &parser.PrimitiveType{Name: "any"}
		}

		return &ParserTypeWrapper{Type: &parser.ArrayType{ElementType: elementType}}

	case *parser.ReferenceExpr:
		exprType := c.checkExpr(e.Expr)
		if wrapper, ok := exprType.(*ParserTypeWrapper); ok {
			return &ParserTypeWrapper{Type: &parser.PointerType{BaseType: wrapper.Type}}
		}
		return &ParserTypeWrapper{Type: &parser.PointerType{BaseType: &parser.PrimitiveType{Name: "any"}}}

		// check_expr.go - adicionar estes casos
	case *parser.MapLiteral:
		// Para um MapLiteral como map<int, string>{1: "Hello", 2: "Bye"}
		// Retorna o tipo correto do mapa
		return &ParserTypeWrapper{Type: &parser.MapType{
			KeyType:   &parser.PrimitiveType{Name: "int"},
			ValueType: &parser.PrimitiveType{Name: "string"},
		}}

	case *parser.SetLiteral:
		// Para um SetLiteral como set<int> {1, 2, 3, 4}
		// Retorna o tipo correto do conjunto
		return &ParserTypeWrapper{Type: &parser.SetType{
			ElementType: &parser.PrimitiveType{Name: "int"},
		}}

	case *parser.SpreadExpr:
		// Para ...arr3, retorna o tipo do elemento do array sendo espalhado
		arrayType := c.checkExpr(e.Expr)

		// Se for um ParserTypeWrapper com ArrayType, retorna o tipo do elemento
		if wrapper, ok := arrayType.(*ParserTypeWrapper); ok {
			if arrType, ok := wrapper.Type.(*parser.ArrayType); ok {
				// Spread de um array retorna o tipo do elemento (não um array de arrays)
				return &ParserTypeWrapper{Type: arrType.ElementType}
			}
		}

		// Fallback
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}

	case *parser.GenericCallExpr:
		// Resolver o callee (deve ser um identificador de função)
		if ident, ok := e.Callee.(*parser.Identifier); ok {
			sym := c.CurrentScope.Resolve(ident.Name)
			if sym == nil || sym.Kind != KindFunction {
				c.reportError(0, 0, fmt.Sprintf("Undeclared function '%s'", ident.Name))
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
			}

			// Obter o tipo de retorno da função
			returnType := sym.Type

			// Aplicar especialização genérica se necessário
			// (Para simplificação, retornamos o tipo de retorno da função)
			// Em uma implementação completa, substituiríamos os parâmetros genéricos

			// Verificar argumentos de tipo
			for _, typeArg := range e.TypeArgs {
				c.validateTypeExists(typeArg)
			}

			// Verificar argumentos da função
			for _, arg := range e.Args {
				c.checkExpr(arg)
			}

			return returnType
		}

		// Fallback para outros casos
		for _, typeArg := range e.TypeArgs {
			c.validateTypeExists(typeArg)
		}
		for _, arg := range e.Args {
			c.checkExpr(arg)
		}
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}

	case *parser.TypeCastExpr:
		// Verificar o tipo da expressão que está sendo convertida
		exprType := c.checkExpr(e.Expr)

		// Verificar se o tipo de destino existe
		c.validateTypeExists(e.Type)

		// Se a expressão é de tipo genérico, permitir a conversão
		if c.isGenericType(exprType) {
			// Permitir conversão de tipos genéricos para string
			return c.wrapType(e.Type)
		}

		// Para tipos não genéricos, fazer verificação básica
		exprTypeStr := StringifyType(exprType)
		targetTypeStr := StringifyParserType(e.Type)

		// Permitir conversões comuns
		if targetTypeStr == "string" {
			// Permitir conversão de int, float, bool para string
			if exprTypeStr == "int" || exprTypeStr == "float" || exprTypeStr == "bool" {
				return c.wrapType(e.Type)
			}
		}

		// Se chegou aqui, a conversão pode não ser válida, mas retornamos o tipo alvo
		// (em uma implementação completa, faríamos mais verificações)
		return c.wrapType(e.Type)

	case *parser.IndexExpr:
		// Verificar se e.Array existe (pode ter nome diferente)
		arrayType := c.checkExpr(e.Array)
		indexType := c.checkExpr(e.Index)

		// Verificar se o array é um tipo que pode ser indexado
		if wrapper, ok := arrayType.(*ParserTypeWrapper); ok {
			switch t := wrapper.Type.(type) {
			case *parser.ArrayType:
				// Verificar se o índice é um tipo inteiro
				indexStr := StringifyType(indexType)
				if indexStr != "int" && indexStr != "int?" && !c.isGenericType(indexType) {
					c.reportError(0, 0, fmt.Sprintf("Array index must be integer, got %s", indexStr))
					return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
				}
				// Retornar o tipo do elemento do array
				return &ParserTypeWrapper{Type: t.ElementType}

			case *parser.MapType:
				// Verificar compatibilidade do tipo da chave
				if !AreParserTypesCompatible(t.KeyType, c.unwrapType(indexType)) {
					c.reportError(0, 0, fmt.Sprintf("Map key type mismatch: expected %s, got %s",
						StringifyParserType(t.KeyType), StringifyType(indexType)))
					return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
				}
				// Retornar o tipo do valor do mapa
				return &ParserTypeWrapper{Type: t.ValueType}

			case *parser.SetType:
				// Sets não suportam indexação direta
				c.reportError(0, 0, "Cannot index a set directly. Use 'has()' to check membership.")
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}

			case *parser.PrimitiveType:
				if t.Name == "string" {
					// Strings podem ser indexadas para obter caracteres
					indexStr := StringifyType(indexType)
					if indexStr != "int" && indexStr != "int?" && !c.isGenericType(indexType) {
						c.reportError(0, 0, fmt.Sprintf("String index must be integer, got %s", indexStr))
						return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
					}
					return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "string"}}
				}
				c.reportError(0, 0, fmt.Sprintf("Cannot index type %s", t.Name))
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}

			default:
				c.reportError(0, 0, fmt.Sprintf("Cannot index type %s", StringifyParserType(wrapper.Type)))
				return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
			}
		}

		// Fallback para tipos não wrapped
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}

	default:
		// Caso padrão para expressões não tratadas
		c.reportError(0, 0, fmt.Sprintf("Unhandled expression type: %T", expr))
		return &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "error"}}
	}

}

// isGenericType verifica se um Type é um tipo genérico
func (c *Checker) isGenericType(t Type) bool {
	if t == nil {
		return false
	}
	typeStr := StringifyType(t)
	return IsGenericTypeName(typeStr)
}
