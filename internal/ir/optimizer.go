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
	optimized := make([]*Instruction, 0)
	unreachable := false

	for _, instr := range fn.Instructions {
		// Se encontrarmos um LABEL, o fluxo pode voltar a este ponto [cite: 18]
		if instr.Op == LABEL {
			unreachable = false
		}

		if !unreachable {
			optimized = append(optimized, instr)
		}

		// JMP e RET encerram o fluxo linear do bloco atual [cite: 18]
		if instr.Op == JMP || instr.Op == RET {
			unreachable = true
		}
	}
	fn.Instructions = optimized
}
