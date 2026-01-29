package ir

import (
	"fmt"
	"strconv"
)

// Optimizer orquestra as transformações no IR
type Optimizer struct {
	Module *Module // Referência ao módulo definido em ir.txt
}

func NewOptimizer(mod *Module) *Optimizer {
	return &Optimizer{Module: mod}
}

// Optimize percorre todas as funções do módulo para aplicar melhorias
func (o *Optimizer) Optimize() {
	for _, fn := range o.Module.Functions {
		o.ConstantFolding(fn)
		o.EliminateUnreachableCode(fn)
	}
}

// ConstantFolding simplifica expressões matemáticas com literais
func (o *Optimizer) ConstantFolding(fn *Function) {
	for _, instr := range fn.Instructions {
		// Verifica se os argumentos são do tipo OpLiteral [cite: 20]
		if instr.Arg1 != nil && instr.Arg1.Kind == OpLiteral &&
			instr.Arg2 != nil && instr.Arg2.Kind == OpLiteral {

			val1, _ := strconv.ParseInt(instr.Arg1.Value, 10, 64)
			val2, _ := strconv.ParseInt(instr.Arg2.Value, 10, 64)
			var result int64

			switch instr.Op {
			case ADD: // Definido no seu iota de OpCode [cite: 18]
				result = val1 + val2
			case SUB:
				result = val1 - val2
			case MUL: // [cite: 23]
				result = val1 * val2
			case DIV:
				if val2 != 0 {
					result = val1 / val2
				} else {
					continue // Não dobrar divisão por zero
				}
			default:
				continue
			}

			// Transforma a instrução complexa em um simples MOV [cite: 18]
			instr.Op = MOV
			instr.Arg1 = &Operand{Kind: OpLiteral, Value: fmt.Sprintf("%d", result)}
			instr.Arg2 = nil
		}
	}
}

// EliminateUnreachableCode remove código após JMP ou RET que não tenha Label
func (o *Optimizer) EliminateUnreachableCode(fn *Function) {
	optimized := make([]*Instruction, 0, len(fn.Instructions))
	unreachable := false
	labels := make(map[string]bool)

	// Primeira passagem: coletar todos os labels
	for _, instr := range fn.Instructions {
		if instr.Op == LABEL && instr.Arg1 != nil {
			labels[instr.Arg1.String()] = true
		}
	}

	// Segunda passagem: eliminar código inalcançável
	for i, instr := range fn.Instructions {
		// Se encontrarmos um LABEL, o fluxo pode voltar a este ponto [cite: 18]
		if instr.Op == LABEL {
			unreachable = false
		}

		if !unreachable {
			optimized = append(optimized, instr)
		}

		// JMP e RET encerram o fluxo linear do bloco atual [cite: 18]
		if instr.Op == JMP || instr.Op == RET {
			// Verificar se o próximo bloco começa com um label
			// Se não, marca como inalcançável
			if i+1 < len(fn.Instructions) {
				nextInstr := fn.Instructions[i+1]
				if nextInstr.Op != LABEL {
					unreachable = true
				}
			} else {
				unreachable = true
			}
		}
	}
	fn.Instructions = optimized
}
