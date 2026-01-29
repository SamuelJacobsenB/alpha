package ir

import (
	"fmt"

	"github.com/alpha/internal/parser"
	"github.com/alpha/internal/semantic"
)

type IRBuilder struct {
	CurrentFunc *Function
	Module      *Module
}

func NewBuilder(mod *Module) *IRBuilder {
	return &IRBuilder{Module: mod}
}

// ============================
// Gerenciamento de Temporários e Labels
// ============================

func (b *IRBuilder) NewTemp(t semantic.Type) *Operand {
	id := b.CurrentFunc.TempCount
	b.CurrentFunc.TempCount++
	return &Operand{
		Kind:  OpTemp,
		Value: fmt.Sprintf("t%d", id),
		Type:  t,
	}
}

func (b *IRBuilder) NewLabel(prefix string) *Operand {
	id := b.CurrentFunc.LabelCount
	b.CurrentFunc.LabelCount++
	return &Operand{
		Kind:  OpLabel,
		Value: fmt.Sprintf("%s_%d", prefix, id),
	}
}

func (b *IRBuilder) Emit(op OpCode, arg1, arg2, result *Operand) *Instruction {
	instr := &Instruction{
		Op:     op,
		Arg1:   arg1,
		Arg2:   arg2,
		Result: result,
	}
	b.CurrentFunc.Instructions = append(b.CurrentFunc.Instructions, instr)
	return instr
}

func (b *IRBuilder) EmitLabel(label *Operand) {
	b.Emit(LABEL, label, nil, nil)
}

func (b *IRBuilder) EmitJump(label *Operand) {
	b.Emit(JMP, label, nil, nil)
}

func (b *IRBuilder) EmitCondJump(cond, trueLabel, falseLabel *Operand) {
	// Implementação corrigida de salto condicional
	if falseLabel != nil {
		b.Emit(JMP_FALSE, cond, falseLabel, nil)
	}
	if trueLabel != nil {
		b.Emit(JMP_TRUE, cond, trueLabel, nil)
	}
}

// ============================
// Helpers de Operandos
// ============================

func Literal(val string, t semantic.Type) *Operand {
	return &Operand{Kind: OpLiteral, Value: val, Type: t}
}

func Var(name string, t semantic.Type) *Operand {
	return &Operand{Kind: OpVar, Value: name, Type: t}
}

func IntLiteral(val int64) *Operand {
	return &Operand{
		Kind:  OpLiteral,
		Value: fmt.Sprintf("%d", val),
		Type:  &semantic.ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "int"}},
	}
}

func BoolLiteral(val bool) *Operand {
	strVal := "false"
	if val {
		strVal = "true"
	}
	return &Operand{
		Kind:  OpLiteral,
		Value: strVal,
		Type:  &semantic.ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "bool"}},
	}
}
