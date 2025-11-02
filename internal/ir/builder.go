package ir

import (
	"fmt"

	"github.com/alpha/internal/parser"
)

// Builder produz IR a partir da AST (versão simples, funções planas)
type Builder struct {
	mod      *Module
	curFunc  *Function
	curBlock *Block
	tempCnt  int
}

// NewBuilder cria builder com um módulo vazio
func NewBuilder() *Builder { return &Builder{mod: NewModule(), tempCnt: 0} }

func (b *Builder) newTemp() ValTemp {
	b.tempCnt++
	return ValTemp{T: Temp(b.tempCnt)}
}

func (b *Builder) StartFunction(name string, params []string) {
	f := &Function{Name: name, Params: params}
	entry := NewBlock(name + ".entry")
	f.Entry = entry
	f.Blocks = []*Block{entry}
	b.mod.AddFunction(f)
	b.curFunc = f
	b.curBlock = entry
}

func (b *Builder) NewBlock(name string) *Block {
	nb := NewBlock(name)
	b.curFunc.Blocks = append(b.curFunc.Blocks, nb)
	return nb
}

func (b *Builder) emit(op Opcode, dst *ValTemp, args ...ValueRef) *Instr {
	ins := &Instr{Op: op, Dst: dst, Args: args}
	b.curBlock.Emit(ins)
	return ins
}

// BuildModule percorre Program e gera um módulo simples (procedural top-level)
func (b *Builder) BuildModule(p *parser.Program) (*Module, error) {
	// criar uma "main" que contém o body top-level
	b.StartFunction("main", nil)
	for _, s := range p.Body {
		if _, err := b.lowerStmt(s); err != nil {
			return nil, err
		}
	}
	// garantir return (implicit)
	b.emit(OpReturn, nil)
	return b.mod, nil
}

// lowerStmt gera IR para uma statement (muito simplificado)
func (b *Builder) lowerStmt(s parser.Stmt) (ValueRef, error) {
	switch st := s.(type) {
	case *parser.VarDecl:
		if st.Init != nil {
			val, err := b.lowerExpr(st.Init)
			if err != nil {
				return nil, err
			}
			b.emit(OpStore, nil, ValSym{Name: st.Name}, val)
		}
		return nil, nil
	case *parser.ConstDecl:
		if st.Init != nil {
			val, err := b.lowerExpr(st.Init)
			if err != nil {
				return nil, err
			}
			b.emit(OpStore, nil, ValSym{Name: st.Name}, val)
		}
		return nil, nil
	case *parser.ExprStmt:
		_, err := b.lowerExpr(st.Expr)
		return nil, err
	case *parser.ReturnStmt:
		if st.Value != nil {
			v, err := b.lowerExpr(st.Value)
			if err != nil {
				return nil, err
			}
			b.emit(OpReturn, nil, v)
			return nil, nil
		}
		b.emit(OpReturn, nil)
		return nil, nil
	case *parser.IfStmt:
		condV, err := b.lowerExpr(st.Cond)
		if err != nil {
			return nil, err
		}
		thenB := b.NewBlock("if.then")
		elseB := b.NewBlock("if.else")
		mergeB := b.NewBlock("if.merge")
		// branch cond -> then/else
		b.emit(OpBranch, nil, condV, ValSym{Name: thenB.Name}, ValSym{Name: elseB.Name})

		// then
		b.curBlock = thenB
		for _, ss := range st.Then {
			if _, err := b.lowerStmt(ss); err != nil {
				return nil, err
			}
		}
		b.emit(OpJump, nil, ValSym{Name: mergeB.Name})

		// else
		b.curBlock = elseB
		for _, ss := range st.Else {
			if _, err := b.lowerStmt(ss); err != nil {
				return nil, err
			}
		}
		b.emit(OpJump, nil, ValSym{Name: mergeB.Name})

		// continue in merge
		b.curBlock = mergeB
		return nil, nil
	case *parser.WhileStmt:
		condB := b.NewBlock("while.cond")
		bodyB := b.NewBlock("while.body")
		mergeB := b.NewBlock("while.merge")
		b.emit(OpJump, nil, ValSym{Name: condB.Name})
		// cond
		b.curBlock = condB
		condV, err := b.lowerExpr(st.Cond)
		if err != nil {
			return nil, err
		}
		b.emit(OpBranch, nil, condV, ValSym{Name: bodyB.Name}, ValSym{Name: mergeB.Name})
		// body
		b.curBlock = bodyB
		for _, ss := range st.Body {
			if _, err := b.lowerStmt(ss); err != nil {
				return nil, err
			}
		}
		b.emit(OpJump, nil, ValSym{Name: condB.Name})
		// merge
		b.curBlock = mergeB
		return nil, nil
	case *parser.ForStmt:
		// desugar: init; jump cond; cond->body/merge; body->post->jump cond
		loopScopeEntry := b.NewBlock("for.init")
		condB := b.NewBlock("for.cond")
		bodyB := b.NewBlock("for.body")
		postB := b.NewBlock("for.post")
		mergeB := b.NewBlock("for.merge")

		// jump to init block
		b.emit(OpJump, nil, ValSym{Name: loopScopeEntry.Name})
		// init
		b.curBlock = loopScopeEntry
		if st.Init != nil {
			if _, err := b.lowerStmt(st.Init); err != nil {
				return nil, err
			}
		}
		b.emit(OpJump, nil, ValSym{Name: condB.Name})
		// cond
		b.curBlock = condB
		if st.Cond != nil {
			condV, err := b.lowerExpr(st.Cond)
			if err != nil {
				return nil, err
			}
			b.emit(OpBranch, nil, condV, ValSym{Name: bodyB.Name}, ValSym{Name: mergeB.Name})
		} else {
			// no cond = true -> jump to body
			b.emit(OpJump, nil, ValSym{Name: bodyB.Name})
		}
		// body
		b.curBlock = bodyB
		for _, ss := range st.Body {
			if _, err := b.lowerStmt(ss); err != nil {
				return nil, err
			}
		}
		b.emit(OpJump, nil, ValSym{Name: postB.Name})
		// post
		b.curBlock = postB
		if st.Post != nil {
			if _, err := b.lowerStmt(st.Post); err != nil {
				return nil, err
			}
		}
		b.emit(OpJump, nil, ValSym{Name: condB.Name})
		// merge
		b.curBlock = mergeB
		return nil, nil
	case *parser.BlockStmt:
		for _, ss := range st.Body {
			if _, err := b.lowerStmt(ss); err != nil {
				return nil, err
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unhandled stmt type %T", s)
	}
}

func (b *Builder) lowerExpr(e parser.Expr) (ValueRef, error) {
	switch ex := e.(type) {
	case *parser.IntLiteral:
		tmp := b.newTemp()
		b.emit(OpConst, &tmp, ValConst{Literal: ex.Value})
		return tmp, nil
	case *parser.FloatLiteral:
		tmp := b.newTemp()
		b.emit(OpConst, &tmp, ValConst{Literal: ex.Value})
		return tmp, nil
	case *parser.StringLiteral:
		tmp := b.newTemp()
		b.emit(OpConst, &tmp, ValConst{Literal: ex.Value})
		return tmp, nil
	case *parser.BoolLiteral:
		tmp := b.newTemp()
		b.emit(OpConst, &tmp, ValConst{Literal: ex.Value})
		return tmp, nil
	case *parser.Identifier:
		tmp := b.newTemp()
		b.emit(OpLoad, &tmp, ValSym{Name: ex.Name})
		return tmp, nil
	case *parser.UnaryExpr:
		operand, err := b.lowerExpr(ex.Expr)
		if err != nil {
			return nil, err
		}
		dst := b.newTemp()
		switch ex.Op {
		case "-":
			// dst = 0 - operand
			b.emit(OpConst, nil, ValConst{Literal: int64(0)})
			b.emit(OpSub, &dst, ValTemp{T: dst.T}, operand)
			return dst, nil
		case "!":
			// Not implemented generically; leave as call to runtime or nop placeholder
			b.emit(OpNop, &dst, operand)
			return dst, nil
		default:
			return nil, fmt.Errorf("unsupported unary op %q", ex.Op)
		}
	case *parser.BinaryExpr:
		la, err := b.lowerExpr(ex.Left)
		if err != nil {
			return nil, err
		}
		rb, err := b.lowerExpr(ex.Right)
		if err != nil {
			return nil, err
		}
		dst := b.newTemp()
		var op Opcode
		switch ex.Op {
		case "+":
			op = OpAdd
		case "-":
			op = OpSub
		case "*":
			op = OpMul
		case "/":
			op = OpDiv
		case "==":
			op = OpCmpEq
		case "<":
			op = OpCmpLt
		default:
			return nil, fmt.Errorf("unsupported binary op %q", ex.Op)
		}
		b.emit(op, &dst, la, rb)
		return dst, nil
	case *parser.AssignExpr:
		// left deve ser identifier
		if id, ok := ex.Left.(*parser.Identifier); ok {
			rv, err := b.lowerExpr(ex.Right)
			if err != nil {
				return nil, err
			}
			b.emit(OpStore, nil, ValSym{Name: id.Name}, rv)
			return rv, nil
		}
		return nil, fmt.Errorf("left-hand side of assignment must be identifier")
	case *parser.CallExpr:
		// lower args
		args := []ValueRef{}
		for _, a := range ex.Args {
			v, err := b.lowerExpr(a)
			if err != nil {
				return nil, err
			}
			args = append(args, v)
		}
		dst := b.newTemp()
		if id, ok := ex.Callee.(*parser.Identifier); ok {
			ins := b.emit(OpCall, &dst, ValSym{Name: id.Name})
			ins.Args = append(ins.Args, args...)
			return dst, nil
		}
		// callee is expr: lower it
		calleeTmp, err := b.lowerExpr(ex.Callee)
		if err != nil {
			return nil, err
		}
		ins := b.emit(OpCall, &dst, calleeTmp)
		ins.Args = append(ins.Args, args...)
		return dst, nil
	default:
		return nil, fmt.Errorf("unhandled expr type %T", e)
	}
}
