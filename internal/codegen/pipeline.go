package codegen

import (
	"strings"

	"github.com/alpha/internal/ir"
)

type CompilationPipeline struct {
	module        *ir.Module
	optimizer     *ir.Optimizer
	registerAlloc *RegisterAllocator
}

func NewPipeline(module *ir.Module) *CompilationPipeline {
	return &CompilationPipeline{
		module:        module,
		optimizer:     ir.NewOptimizer(module),
		registerAlloc: NewRegisterAllocator(),
	}
}

func (p *CompilationPipeline) Compile() string {
	// Fase 1: Otimizações no IR
	p.optimizer.Optimize()

	// Fase 2: Alocação de registros
	for _, fn := range p.module.Functions {
		p.registerAlloc.Allocate(fn)
	}

	// Fase 3: Geração de código
	emitter := NewOptimizedEmitter(p.module)
	code := emitter.Emit()

	// Fase 4: Pós-processamento
	code = p.postProcess(code)

	return code
}

func (p *CompilationPipeline) postProcess(code string) string {
	// Remove código morto após goto
	lines := strings.Split(code, "\n")
	var result []string
	unreachable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detecta goto incondicional
		if strings.HasPrefix(trimmed, "goto") && !strings.Contains(trimmed, "if") {
			unreachable = true
			result = append(result, line)
		} else if strings.HasSuffix(trimmed, ":") {
			unreachable = false
			result = append(result, line)
		} else if !unreachable || strings.HasPrefix(trimmed, "//") {
			result = append(result, line)
		}
	}

	// Remove variáveis não utilizadas
	result = p.removeUnusedVars(result)

	return strings.Join(result, "\n")
}

func (p *CompilationPipeline) removeUnusedVars(lines []string) []string {
	usedVars := make(map[string]bool)

	// Primeira passagem: coleta variáveis usadas
	for _, line := range lines {
		if strings.Contains(line, "var ") {
			continue
		}

		// Extrai identificadores
		words := strings.FieldsFunc(line, func(r rune) bool {
			return r == ' ' || r == '=' || r == '+' || r == '-' || r == '*' ||
				r == '/' || r == '[' || r == ']' || r == '(' || r == ')' ||
				r == ',' || r == ';' || r == ':' || r == '{' || r == '}'
		})

		for _, word := range words {
			if strings.HasPrefix(word, "t") && len(word) > 1 {
				// Verifica se é um número após o 't'
				allDigits := true
				for _, ch := range word[1:] {
					if ch < '0' || ch > '9' {
						allDigits = false
						break
					}
				}
				if allDigits {
					usedVars[word] = true
				}
			}
		}
	}

	// Segunda passagem: remove declarações não usadas
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "var t") {
			varName := strings.Fields(trimmed)[1]
			if used, exists := usedVars[varName]; !exists || !used {
				continue // Remove declaração não usada
			}
		}
		result = append(result, line)
	}

	return result
}

type OptimizationPass interface {
	Apply(*ir.Module)
}

type ConstantPropagation struct{}
type DeadCodeElimination struct{}
type InlineSmallFunctions struct{}
type LoopUnrolling struct{}
type RegisterAllocation struct{}

func (cp *ConstantPropagation) Apply(m *ir.Module) {
	for _, fn := range m.Functions {
		cp.applyToFunction(fn)
	}
}

func (cp *ConstantPropagation) applyToFunction(fn *ir.Function) {
	// Implementação simplificada
	constants := make(map[string]*ir.Operand)

	for i := 0; i < len(fn.Instructions); i++ {
		instr := fn.Instructions[i]

		// Propaga constantes
		if instr.Op == ir.MOV && instr.Arg1 != nil && instr.Arg1.Kind == ir.OpLiteral {
			constants[instr.Result.Value] = instr.Arg1
		}

		// Substitui usos de constantes
		if instr.Arg1 != nil && constants[instr.Arg1.Value] != nil {
			instr.Arg1 = constants[instr.Arg1.Value]
		}
		if instr.Arg2 != nil && constants[instr.Arg2.Value] != nil {
			instr.Arg2 = constants[instr.Arg2.Value]
		}
	}
}
