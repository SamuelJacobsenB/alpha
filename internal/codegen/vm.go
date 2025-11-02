package codegen

import (
	"errors"
	"fmt"

	"github.com/alpha/internal/ir"
)

// Value é o tipo runtime usado pelo VM: int64, float64, string, bool, nil
type Value = interface{}

// VM representa o estado da máquina para um módulo
type VM struct {
	mod     *ir.Module
	globals map[string]Value        // variáveis globais (por name)
	funcs   map[string]*ir.Function // função por nome

	// builtins
	builtins map[string]BuiltinFunc
}

// BuiltinFunc é a função builtin que a VM pode chamar
type BuiltinFunc func(args []Value) (Value, error)

// NewVM cria VM com módulo e builtin padrão
func NewVM(m *ir.Module) *VM {
	vm := &VM{
		mod:      m,
		globals:  map[string]Value{},
		funcs:    map[string]*ir.Function{},
		builtins: map[string]BuiltinFunc{},
	}
	for _, f := range m.Functions {
		vm.funcs[f.Name] = f
	}
	// registrar builtins básicos
	vm.builtins["print"] = builtinPrint
	return vm
}

// RunMain executa a função "main" do módulo e retorna valor ou erro
func (vm *VM) RunMain() (Value, error) {
	mainf, ok := vm.funcs["main"]
	if !ok {
		return nil, errors.New("no main function")
	}
	return vm.runFunction(mainf)
}

// runFunction executa uma função e retorna o valor de retorno (ou nil)
func (vm *VM) runFunction(fn *ir.Function) (Value, error) {
	// construir mapa de blocos por nome
	blocks := map[string]*ir.Block{}
	for _, b := range fn.Blocks {
		blocks[b.Name] = b
	}

	// temporários: map Temp -> Value
	temps := map[ir.Temp]Value{}

	// ip: current block and instruction index
	curBlock := fn.Entry
	if curBlock == nil {
		if len(fn.Blocks) == 0 {
			return nil, nil
		}
		curBlock = fn.Blocks[0]
	}
	ip := 0

	// loop de execução por bloco
	for {
		if ip >= len(curBlock.Instr) {
			// fim do bloco sem jump -> return nil
			return nil, nil
		}
		ins := curBlock.Instr[ip]

		switch ins.Op {
		case ir.OpConst:
			if ins.Dst == nil || len(ins.Args) == 0 {
				return nil, fmt.Errorf("invalid Const instr")
			}
			carg := ins.Args[0]
			switch v := carg.(type) {
			case ir.ValConst:
				temps[ins.Dst.T] = v.Literal
			default:
				temps[ins.Dst.T] = nil
			}
			ip++

		case ir.OpLoad:
			// Args: ValSym{name} or other -> load global var
			if ins.Dst == nil || len(ins.Args) == 0 {
				return nil, fmt.Errorf("invalid Load instr")
			}
			switch s := ins.Args[0].(type) {
			case ir.ValSym:
				val, ok := vm.globals[s.Name]
				if !ok {
					// undefined globals default to nil
					val = nil
				}
				temps[ins.Dst.T] = val
			case ir.ValTemp:
				// load from temp (copy)
				if v, ok := temps[s.T]; ok {
					temps[ins.Dst.T] = v
				} else {
					temps[ins.Dst.T] = nil
				}
			default:
				temps[ins.Dst.T] = nil
			}
			ip++

		case ir.OpStore:
			// Args: ValSym{name}, src
			if len(ins.Args) < 2 {
				return nil, fmt.Errorf("invalid Store instr")
			}
			sym, ok := ins.Args[0].(ir.ValSym)
			if !ok {
				return nil, fmt.Errorf("Store target must be symbol")
			}
			vv, err := vm.evalValueRef(ins.Args[1], temps)
			if err != nil {
				return nil, err
			}
			vm.globals[sym.Name] = vv
			ip++

		case ir.OpAdd, ir.OpSub, ir.OpMul, ir.OpDiv:
			if ins.Dst == nil || len(ins.Args) != 2 {
				return nil, fmt.Errorf("invalid arithmetic instr %s", ins.Op)
			}
			a, err := vm.evalValueRef(ins.Args[0], temps)
			if err != nil {
				return nil, err
			}
			b, err := vm.evalValueRef(ins.Args[1], temps)
			if err != nil {
				return nil, err
			}
			res, err := arithmeticOp(ins.Op, a, b)
			if err != nil {
				return nil, err
			}
			temps[ins.Dst.T] = res
			ip++

		case ir.OpCmpEq, ir.OpCmpLt:
			// comparisons -> bool
			if ins.Dst == nil || len(ins.Args) != 2 {
				return nil, fmt.Errorf("invalid cmp instr %s", ins.Op)
			}
			a, err := vm.evalValueRef(ins.Args[0], temps)
			if err != nil {
				return nil, err
			}
			b, err := vm.evalValueRef(ins.Args[1], temps)
			if err != nil {
				return nil, err
			}
			var out bool
			switch ins.Op {
			case ir.OpCmpEq:
				out = equals(a, b)
			case ir.OpCmpLt:
				out, err = lessThan(a, b)
				if err != nil {
					return nil, err
				}
			}
			temps[ins.Dst.T] = out
			ip++

		case ir.OpNop:
			ip++

		case ir.OpJump:
			// Arg0: ValSym{label}
			if len(ins.Args) < 1 {
				return nil, fmt.Errorf("invalid Jump instr")
			}
			lbl, ok := ins.Args[0].(ir.ValSym)
			if !ok {
				return nil, fmt.Errorf("jump target must be label symbol")
			}
			nb, ok := blocks[lbl.Name]
			if !ok {
				return nil, fmt.Errorf("unknown jump target %s", lbl.Name)
			}
			curBlock = nb
			ip = 0

		case ir.OpBranch:
			// Args: cond, thenLabel, elseLabel
			if len(ins.Args) < 3 {
				return nil, fmt.Errorf("invalid Branch instr")
			}
			condv, err := vm.evalValueRef(ins.Args[0], temps)
			if err != nil {
				return nil, err
			}
			condb, ok := condv.(bool)
			if !ok {
				return nil, fmt.Errorf("branch condition is not boolean: %T", condv)
			}
			var lbl ir.ValSym
			if condb {
				lbl, ok = ins.Args[1].(ir.ValSym)
				if !ok {
					return nil, fmt.Errorf("branch then target must be label")
				}
			} else {
				lbl, ok = ins.Args[2].(ir.ValSym)
				if !ok {
					return nil, fmt.Errorf("branch else target must be label")
				}
			}
			nb, ok := blocks[lbl.Name]
			if !ok {
				return nil, fmt.Errorf("unknown branch target %s", lbl.Name)
			}
			curBlock = nb
			ip = 0

		case ir.OpCall:
			// Args: callee (ValSym or ValTemp/ValConst), rest are args
			if ins.Dst == nil {
				// allow call without dst (ignore result)
			}
			if len(ins.Args) < 1 {
				return nil, fmt.Errorf("call without callee")
			}
			calleeRef := ins.Args[0]
			// collect args values
			args := []Value{}
			for i := 1; i < len(ins.Args); i++ {
				v, err := vm.evalValueRef(ins.Args[i], temps)
				if err != nil {
					return nil, err
				}
				args = append(args, v)
			}
			// if callee is ValSym and builtin exists, call builtin
			switch c := calleeRef.(type) {
			case ir.ValSym:
				if bf, ok := vm.builtins[c.Name]; ok {
					ret, err := bf(args)
					if err != nil {
						return nil, err
					}
					if ins.Dst != nil {
						temps[ins.Dst.T] = ret
					}
					ip++
					continue
				}
				// else if it's a function in module, run it (no args handling of params here)
				if ffn, ok := vm.funcs[c.Name]; ok {
					// naive: ignore passing args to user functions (params not wired)
					ret, err := vm.runFunction(ffn)
					if err != nil {
						return nil, err
					}
					if ins.Dst != nil {
						temps[ins.Dst.T] = ret
					}
					ip++
					continue
				}
				return nil, fmt.Errorf("unknown callee %s", c.Name)
			default:
				// callee is temp: evaluate and treat as function pointer (not supported)
				return nil, fmt.Errorf("call of non-symbol callee not supported")
			}

		case ir.OpReturn:
			// if args present, evaluate first and return
			if len(ins.Args) >= 1 {
				v, err := vm.evalValueRef(ins.Args[0], temps)
				if err != nil {
					return nil, err
				}
				return v, nil
			}
			return nil, nil

		default:
			return nil, fmt.Errorf("unsupported opcode %s", ins.Op)
		}
	}
}
