package codegen

import (
	"github.com/alpha/internal/ir"
	"github.com/alpha/internal/parser"
	"github.com/alpha/internal/semantic"
)

// CodeGenerator é o ponto de entrada principal para geração de código
type CodeGenerator struct {
	checker *semantic.Checker
	module  *ir.Module
}

func NewCodeGenerator(checker *semantic.Checker) *CodeGenerator {
	return &CodeGenerator{
		checker: checker,
	}
}

// GenerateCode gera código Go a partir do IR otimizado
func (cg *CodeGenerator) GenerateCode(module *ir.Module) string {
	cg.module = module

	// Pipeline de otimização
	pipeline := NewPipeline(module)

	// Compilar
	return pipeline.Compile()
}

// GenerateFromAST gera código diretamente da AST (atalho)
func (cg *CodeGenerator) GenerateFromAST(prog *parser.Program) string {
	// 1. Gerar IR
	generator := ir.NewGenerator(cg.checker)
	module := generator.Generate(prog)

	// 2. Otimizar
	optimizer := ir.NewOptimizer(module)
	optimizer.Optimize()

	// 3. Gerar código
	return cg.GenerateCode(module)
}
