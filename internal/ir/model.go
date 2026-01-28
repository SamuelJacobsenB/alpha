package ir

import (
	"fmt"
	"strings"

	"github.com/alpha/internal/parser"
	"github.com/alpha/internal/semantic"
)

// OpCode representa a operação a ser executada
type OpCode int

const (
	// Aritmética e Lógica
	ADD OpCode = iota
	SUB
	MUL
	DIV
	MOD
	AND
	OR
	XOR
	SHL
	SHR

	// Comparação
	EQ
	NEQ
	LT
	GT
	LE
	GE

	// Memória e Atribuição
	MOV       // t1 = t2
	LOAD      // t1 = *t2
	STORE     // *t1 = t2
	ALLOCA    // t1 = alloc type
	GET_FIELD // t1 = t2.field (offset calculation)
	GET_INDEX // t1 = t2[t3]
	GET_ADDR  // t1 = &t2

	// Controle de Fluxo
	LABEL     // Definição de label
	JMP       // Pulo incondicional
	JMP_TRUE  // Pulo se verdadeiro
	JMP_FALSE // Pulo se falso
	CALL      // Chamada de função
	RET       // Retorno de função
	PHI       // Para SSA (opcional, incluído para extensibilidade)

	// Built-ins e Especiais
	LEN        // t1 = len(t2)
	APPEND     // t1 = append(t1, t2)
	MAKE_SLICE // t1 = make([]T, len)
	MAKE_MAP   // t1 = make(map[K]V)
	CAST       // t1 = type(t2)
	NOP        // No Operation
)

// OperandType define o tipo do operando
type OperandType int

const (
	OpVar      OperandType = iota // Variável do usuário
	OpTemp                        // Variável temporária do compilador (%t1)
	OpLiteral                     // Literal (número, string)
	OpLabel                       // Label de salto (.L1)
	OpFunction                    // Nome de função
	OpType                        // Referência a tipo (para allocs)
	OpField                       // Nome de campo de struct
)

// Operand representa um argumento de uma instrução
type Operand struct {
	Kind  OperandType
	Value string        // Representação string do valor
	Type  semantic.Type // Tipo semântico associado (para backend)
}

func (o Operand) String() string {
	switch o.Kind {
	case OpTemp:
		return "%" + o.Value
	case OpLabel:
		return "." + o.Value
	case OpLiteral:
		if o.Type != nil && strings.Contains(semantic.StringifyType(o.Type), "string") {
			return `"` + o.Value + `"`
		}
		return o.Value
	default:
		return o.Value
	}
}

// Instruction representa uma linha de código no IR (Quadruple)
// Formato: Result = Op Arg1, Arg2
type Instruction struct {
	Op     OpCode
	Arg1   *Operand
	Arg2   *Operand
	Result *Operand
	Args   [](*Operand) // Para instruções com número variável de argumentos (ex: CALL)
	// Metadados adicionais para debug ou backend específico
	Line int
}

func (i *Instruction) String() string {
	opStr := i.opToString()

	if i.Op == LABEL {
		return fmt.Sprintf("%s:", i.Arg1)
	}

	var sb strings.Builder
	if i.Result != nil {
		sb.WriteString(fmt.Sprintf("%s = ", i.Result))
	}

	sb.WriteString(opStr)

	if i.Arg1 != nil {
		sb.WriteString(" ")
		sb.WriteString(i.Arg1.String())
	}
	if i.Arg2 != nil {
		sb.WriteString(", ")
		sb.WriteString(i.Arg2.String())
	}
	return sb.String()
}

func (i *Instruction) opToString() string {
	// Mapeamento simples para debug
	names := []string{
		"ADD", "SUB", "MUL", "DIV", "MOD", "AND", "OR", "XOR", "SHL", "SHR",
		"EQ", "NEQ", "LT", "GT", "LE", "GE",
		"MOV", "LOAD", "STORE", "ALLOCA", "GET_FIELD", "GET_INDEX", "GET_ADDR",
		"LABEL", "JMP", "JMP_TRUE", "JMP_FALSE", "CALL", "RET", "PHI",
		"LEN", "APPEND", "MAKE_SLICE", "MAKE_MAP", "CAST", "NOP",
	}
	if int(i.Op) < len(names) {
		return names[i.Op]
	}
	return "UNKNOWN"
}

// BasicBlock representa uma sequência linear de instruções
type BasicBlock struct {
	Label        string
	Instructions []*Instruction
	Predecessors []*BasicBlock
	Successors   []*BasicBlock
}

// Function representa uma função compilada no IR
type Function struct {
	Name         string
	Receiver     string
	Params       []*Operand
	Instructions []*Instruction // Representação linear
	TempCount    int            // Contador para variáveis temporárias
	LabelCount   int            // Contador para labels
	ReturnType   semantic.Type
	IsExported   bool
}

// Module representa o programa inteiro (pacote)
type Module struct {
	Name      string
	Imports   []string
	Globals   []*Instruction // Inicialização de globais
	Functions []*Function
	Structs   []*parser.StructDecl // Metadados de structs para backend
}
