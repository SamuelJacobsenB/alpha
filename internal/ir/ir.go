package ir

import "fmt"

// Temp representa um registrador virtual
type Temp int

// ValueRef é operando: Temp, ConstLiteral ou Symbol(name)
type ValueRef interface{ isVal() }
type ValTemp struct{ T Temp }
type ValConst struct{ Literal interface{} }
type ValSym struct{ Name string }

func (ValTemp) isVal()  {}
func (ValConst) isVal() {}
func (ValSym) isVal()   {}

type Opcode string

const (
	OpConst  Opcode = "Const"
	OpLoad   Opcode = "Load"
	OpStore  Opcode = "Store"
	OpAdd    Opcode = "Add"
	OpSub    Opcode = "Sub"
	OpMul    Opcode = "Mul"
	OpDiv    Opcode = "Div"
	OpCmpEq  Opcode = "CmpEq"
	OpCmpLt  Opcode = "CmpLt"
	OpJump   Opcode = "Jump"
	OpBranch Opcode = "Branch"
	OpCall   Opcode = "Call"
	OpReturn Opcode = "Return"
	OpNop    Opcode = "Nop"
)

type Instr struct {
	Op   Opcode
	Dst  *ValTemp // nil se sem destino
	Args []ValueRef
	Meta interface{} // opcional: source pos
}

type Block struct {
	Name  string
	Instr []*Instr
}

type Function struct {
	Name   string
	Params []string
	Blocks []*Block
	Entry  *Block
}

type Module struct {
	Functions []*Function
	Globals   map[string]ValueRef
}

func NewModule() *Module {
	return &Module{Functions: []*Function{}, Globals: map[string]ValueRef{}}
}

func (m *Module) AddFunction(f *Function) { m.Functions = append(m.Functions, f) }

func NewBlock(name string) *Block { return &Block{Name: name, Instr: []*Instr{}} }
func (b *Block) Emit(i *Instr)    { b.Instr = append(b.Instr, i) }

func NewTemp(n int) Temp { return Temp(n) }

func (i *Instr) String() string {
	dst := ""
	if i.Dst != nil {
		dst = fmt.Sprintf("t%d = ", i.Dst.T)
	}
	args := ""
	for j, a := range i.Args {
		if j > 0 {
			args += ", "
		}
		switch v := a.(type) {
		case ValTemp:
			args += fmt.Sprintf("t%d", v.T)
		case ValConst:
			args += fmt.Sprintf("%v", v.Literal)
		case ValSym:
			args += fmt.Sprintf("%s", v.Name)
		default:
			args += fmt.Sprintf("%v", v)
		}
	}
	return fmt.Sprintf("%s%s %s", dst, i.Op, args)
}
