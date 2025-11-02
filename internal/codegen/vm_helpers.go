package codegen

import (
	"fmt"

	"github.com/alpha/internal/ir"
)

// evalValueRef avalia um ValueRef (ValTemp, ValConst, ValSym) usando map de temps e globals
func (vm *VM) evalValueRef(vref ir.ValueRef, temps map[ir.Temp]Value) (Value, error) {
	switch vv := vref.(type) {
	case ir.ValTemp:
		val, ok := temps[vv.T]
		if !ok {
			return nil, nil
		}
		return val, nil
	case ir.ValConst:
		return vv.Literal, nil
	case ir.ValSym:
		// sym pode ser variable name ou label; try globals first
		if val, ok := vm.globals[vv.Name]; ok {
			return val, nil
		}
		// if not found, return symbol name as string? better return nil
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown ValueRef type %T", vref)
	}
}

func arithmeticOp(op ir.Opcode, a, b Value) (Value, error) {
	// handle int64 and float64
	switch av := a.(type) {
	case int64:
		switch bv := b.(type) {
		case int64:
			switch op {
			case ir.OpAdd:
				return av + bv, nil
			case ir.OpSub:
				return av - bv, nil
			case ir.OpMul:
				return av * bv, nil
			case ir.OpDiv:
				if bv == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return av / bv, nil
			}
		case float64:
			af := float64(av)
			switch op {
			case ir.OpAdd:
				return af + bv, nil
			case ir.OpSub:
				return af - bv, nil
			case ir.OpMul:
				return af * bv, nil
			case ir.OpDiv:
				if bv == 0.0 {
					return nil, fmt.Errorf("division by zero")
				}
				return af / bv, nil
			}
		}
	case float64:
		switch bv := b.(type) {
		case int64:
			bf := float64(bv)
			switch op {
			case ir.OpAdd:
				return av + bf, nil
			case ir.OpSub:
				return av - bf, nil
			case ir.OpMul:
				return av * bf, nil
			case ir.OpDiv:
				if bf == 0.0 {
					return nil, fmt.Errorf("division by zero")
				}
				return av / bf, nil
			}
		case float64:
			switch op {
			case ir.OpAdd:
				return av + bv, nil
			case ir.OpSub:
				return av - bv, nil
			case ir.OpMul:
				return av * bv, nil
			case ir.OpDiv:
				if bv == 0.0 {
					return nil, fmt.Errorf("division by zero")
				}
				return av / bv, nil
			}
		}
	}
	return nil, fmt.Errorf("unsupported arithmetic types: %T %T for %s", a, b, op)
}

func equals(a, b Value) bool {
	// nil equals nil
	if a == nil && b == nil {
		return true
	}
	switch av := a.(type) {
	case int64:
		if bv, ok := b.(int64); ok {
			return av == bv
		}
		if bv, ok := b.(float64); ok {
			return float64(av) == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
		if bv, ok := b.(int64); ok {
			return av == float64(bv)
		}
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	}
	return false
}

func lessThan(a, b Value) (bool, error) {
	switch av := a.(type) {
	case int64:
		switch bv := b.(type) {
		case int64:
			return av < bv, nil
		case float64:
			return float64(av) < bv, nil
		}
	case float64:
		switch bv := b.(type) {
		case float64:
			return av < bv, nil
		case int64:
			return av < float64(bv), nil
		}
	case string:
		if bv, ok := b.(string); ok {
			return av < bv, nil
		}
	}
	return false, fmt.Errorf("unsupported types for <: %T %T", a, b)
}
