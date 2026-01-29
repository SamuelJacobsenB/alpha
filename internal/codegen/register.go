package codegen

import "github.com/alpha/internal/ir"

type RegisterAllocator struct {
	registers map[string]int // variável -> registro
	nextReg   int
}

func NewRegisterAllocator() *RegisterAllocator {
	return &RegisterAllocator{
		registers: make(map[string]int),
		nextReg:   0,
	}
}

func (ra *RegisterAllocator) Allocate(fn *ir.Function) {
	// Para cada variável, atribui um "registro" (na prática, otimiza o uso de variáveis)

	// 1. Coleta variáveis vivas
	liveRanges := ra.computeLiveRanges(fn)

	// 2. Ordena por uso
	sortedVars := ra.sortByUse(fn)

	// 3. Atribui "registros"
	for _, varName := range sortedVars {
		if range_, exists := liveRanges[varName]; exists {
			// Tenta reusar registro
			reg := ra.findAvailableRegister(range_)
			ra.registers[varName] = reg
		}
	}
}

func (ra *RegisterAllocator) computeLiveRanges(fn *ir.Function) map[string][]int {
	ranges := make(map[string][]int)

	for i, instr := range fn.Instructions {
		// Definições
		if instr.Result != nil {
			varName := instr.Result.Value
			ranges[varName] = append(ranges[varName], i)
		}

		// Usos
		if instr.Arg1 != nil {
			varName := instr.Arg1.Value
			ranges[varName] = append(ranges[varName], i)
		}
		if instr.Arg2 != nil {
			varName := instr.Arg2.Value
			ranges[varName] = append(ranges[varName], i)
		}
	}

	return ranges
}

func (ra *RegisterAllocator) sortByUse(fn *ir.Function) []string {
	useCount := make(map[string]int)

	for _, instr := range fn.Instructions {
		if instr.Arg1 != nil {
			useCount[instr.Arg1.Value]++
		}
		if instr.Arg2 != nil {
			useCount[instr.Arg2.Value]++
		}
	}

	// Ordena por frequência de uso (decrescente)
	var sorted []string
	for varName := range useCount {
		sorted = append(sorted, varName)
	}

	// TODO: Implementar ordenação por uso
	return sorted
}

func (ra *RegisterAllocator) findAvailableRegister(range_ []int) int {
	// Algoritmo simplificado: retorna próximo registro disponível
	reg := ra.nextReg
	ra.nextReg++
	return reg
}

func (ra *RegisterAllocator) GetRegister(varName string) int {
	if reg, ok := ra.registers[varName]; ok {
		return reg
	}
	return -1
}
