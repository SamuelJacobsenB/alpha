package semantic

import (
	"fmt"

	"github.com/alpha/internal/parser"
)

func (c *Checker) checkExpr(expr parser.Expr) parser.Type {
	switch e := expr.(type) {

	// ... [Literais Int, Float, Bool, String, Null, Identifier mantidos iguais] ...
	case *parser.IntLiteral:
		return &parser.PrimitiveType{Name: "int"}
	case *parser.FloatLiteral:
		return &parser.PrimitiveType{Name: "float"}
	case *parser.BoolLiteral:
		return &parser.PrimitiveType{Name: "bool"}
	case *parser.StringLiteral:
		return &parser.PrimitiveType{Name: "string"}
	case *parser.NullLiteral:
		return &parser.PrimitiveType{Name: "null"}
	case *parser.Identifier:
		sym := c.CurrentScope.Resolve(e.Name)
		if sym == nil {
			c.reportError(0, 0, fmt.Sprintf("Undeclared identifier '%s'", e.Name))
			return &parser.PrimitiveType{Name: "error"}
		}
		return sym.Type

	case *parser.UnaryExpr:
		valType := c.checkExpr(e.Expr)

		switch e.Op {
		case "!":
			if StringifyType(valType) != "bool" {
				c.reportError(0, 0, "Operator '!' requires boolean operand")
			}
			return &parser.PrimitiveType{Name: "bool"}
		case "-":
			if !isNumeric(valType) {
				c.reportError(0, 0, "Unary '-' requires numeric operand")
			}
			return valType
		// --- NOVO: Suporte a ++ e -- ---
		case "++", "--":
			if !isNumeric(valType) {
				c.reportError(0, 0, fmt.Sprintf("Operator '%s' requires numeric operand", e.Op))
			}
			if !isLValue(e.Expr) {
				c.reportError(0, 0, fmt.Sprintf("Operator '%s' requires a valid assignment target (variable)", e.Op))
			}
			// Verificar se não é constante
			if id, ok := e.Expr.(*parser.Identifier); ok {
				sym := c.CurrentScope.Resolve(id.Name)
				if sym != nil && sym.Kind == KindConst {
					c.reportError(0, 0, fmt.Sprintf("Cannot change value of constant '%s'", id.Name))
				}
			}
			return valType
		}
		return valType

	case *parser.BinaryExpr:
		leftType := c.checkExpr(e.Left)
		rightType := c.checkExpr(e.Right)

		if !AreTypesCompatible(leftType, rightType) {
			c.reportError(0, 0, fmt.Sprintf("Type mismatch in binary operation '%s': %s vs %s",
				e.Op, StringifyType(leftType), StringifyType(rightType)))
			return &parser.PrimitiveType{Name: "error"}
		}

		// --- NOVO: Verificação de Atribuição Composta (+=, -=, etc) ---
		if isAssignmentOp(e.Op) {
			if !isLValue(e.Left) {
				c.reportError(0, 0, "Left side of compound assignment must be a variable")
			}
			if id, ok := e.Left.(*parser.Identifier); ok {
				sym := c.CurrentScope.Resolve(id.Name)
				if sym != nil && sym.Kind == KindConst {
					c.reportError(0, 0, fmt.Sprintf("Cannot assign to constant '%s'", id.Name))
				}
			}
			return leftType
		}

		if isComparisonOp(e.Op) {
			return &parser.PrimitiveType{Name: "bool"}
		}
		return leftType

	case *parser.TernaryExpr:
		condType := c.checkExpr(e.Cond)
		// --- ATUALIZADO: Aceitar Nullables no ternário ---
		if !isConditionable(condType) {
			c.reportError(0, 0, "Ternary condition must be boolean or nullable")
		}
		trueType := c.checkExpr(e.TrueExpr)
		falseType := c.checkExpr(e.FalseExpr)

		if !AreTypesCompatible(trueType, falseType) {
			c.reportError(0, 0, fmt.Sprintf("Ternary branches mismatch: %s vs %s", StringifyType(trueType), StringifyType(falseType)))
		}
		return trueType

	case *parser.AssignExpr:
		leftType := c.checkExpr(e.Left)
		rightType := c.checkExpr(e.Right)

		if !isLValue(e.Left) {
			c.reportError(0, 0, "Invalid assignment target")
		}

		if id, ok := e.Left.(*parser.Identifier); ok {
			sym := c.CurrentScope.Resolve(id.Name)
			if sym != nil && sym.Kind == KindConst {
				c.reportError(0, 0, fmt.Sprintf("Cannot assign to constant '%s'", id.Name))
			}
		}

		if !AreTypesCompatible(leftType, rightType) {
			c.reportError(0, 0, fmt.Sprintf("Cannot assign type '%s' to '%s'",
				StringifyType(rightType), StringifyType(leftType)))
		}
		return leftType

	// ... [CallExpr, IndexExpr, MemberExpr, StructLiteral mantidos iguais] ...
	case *parser.CallExpr:
		// (Mantenha sua implementação existente aqui)
		calleeType := c.checkExpr(e.Callee)
		if StringifyType(calleeType) == "error" {
			return calleeType
		}
		return calleeType // Simplificado para brevidade

	case *parser.IndexExpr:
		// (Mantenha sua implementação existente aqui)
		return &parser.PrimitiveType{Name: "error"} // Placeholder

	case *parser.MemberExpr:
		// (Mantenha sua implementação existente aqui)
		return &parser.PrimitiveType{Name: "error"} // Placeholder

	case *parser.StructLiteral:
		// (Mantenha sua implementação existente aqui)
		return &parser.PrimitiveType{Name: "object"}

	case *parser.ArrayLiteral:
		if len(e.Elements) == 0 {
			return &parser.ArrayType{ElementType: &parser.PrimitiveType{Name: "any"}}
		}
		firstType := c.checkExpr(e.Elements[0])
		for i := 1; i < len(e.Elements); i++ {
			nextType := c.checkExpr(e.Elements[i])
			if !AreTypesCompatible(firstType, nextType) {
				c.reportError(0, 0, "Array elements must have the same type")
			}
		}
		return &parser.ArrayType{ElementType: firstType}

	// --- NOVO: Map Literal ---
	case *parser.MapLiteral:
		if len(e.Entries) == 0 {
			return &parser.MapType{
				KeyType:   &parser.PrimitiveType{Name: "any"},
				ValueType: &parser.PrimitiveType{Name: "any"},
			}
		}
		kType := c.checkExpr(e.Entries[0].Key)
		vType := c.checkExpr(e.Entries[0].Value)

		for _, entry := range e.Entries {
			curK := c.checkExpr(entry.Key)
			curV := c.checkExpr(entry.Value)
			if !AreTypesCompatible(kType, curK) {
				c.reportError(0, 0, "All map keys must have same type")
			}
			if !AreTypesCompatible(vType, curV) {
				c.reportError(0, 0, "All map values must have same type")
			}
		}
		return &parser.MapType{KeyType: kType, ValueType: vType}
	}

	return nil
}

// Helpers Adicionais
func isAssignmentOp(op string) bool {
	return op == "+=" || op == "-=" || op == "*=" || op == "/="
}

func isComparisonOp(op string) bool {
	return op == "==" || op == "!=" || op == "<" || op == ">" || op == "<=" || op == ">="
}

func isLValue(expr parser.Expr) bool {
	switch expr.(type) {
	case *parser.Identifier, *parser.MemberExpr, *parser.IndexExpr:
		return true
	}
	return false
}
