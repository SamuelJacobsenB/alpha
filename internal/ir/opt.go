package ir

// implementações simples de otimização (esqueleto)

// ConstFold percorre instruções e substitui op com args constantes por Const
func ConstFold(m *Module) {
	for _, f := range m.Functions {
		for _, b := range f.Blocks {
			newInstr := []*Instr{}
			for _, ins := range b.Instr {
				// exemplo: Add com dois ValConst -> substituir por Const
				if (ins.Op == OpAdd || ins.Op == OpSub || ins.Op == OpMul || ins.Op == OpDiv) && len(ins.Args) == 2 {
					a0, ok0 := ins.Args[0].(ValConst)
					a1, ok1 := ins.Args[1].(ValConst)
					if ok0 && ok1 {
						// tenta operar nos literais; só trata int64/float64 básicos aqui
						switch ins.Op {
						case OpAdd:
							switch v0 := a0.Literal.(type) {
							case int64:
								if v1, ok := a1.Literal.(int64); ok {
									newInstr = append(newInstr, &Instr{Op: OpConst, Dst: ins.Dst, Args: []ValueRef{ValConst{Literal: v0 + v1}}})
									continue
								}
							case float64:
								if v1, ok := a1.Literal.(float64); ok {
									newInstr = append(newInstr, &Instr{Op: OpConst, Dst: ins.Dst, Args: []ValueRef{ValConst{Literal: v0 + v1}}})
									continue
								}
							}
						case OpSub:
							switch v0 := a0.Literal.(type) {
							case int64:
								if v1, ok := a1.Literal.(int64); ok {
									newInstr = append(newInstr, &Instr{Op: OpConst, Dst: ins.Dst, Args: []ValueRef{ValConst{Literal: v0 - v1}}})
									continue
								}
							case float64:
								if v1, ok := a1.Literal.(float64); ok {
									newInstr = append(newInstr, &Instr{Op: OpConst, Dst: ins.Dst, Args: []ValueRef{ValConst{Literal: v0 - v1}}})
									continue
								}
							}
						}
					}
				}
				newInstr = append(newInstr, ins)
			}
			b.Instr = newInstr
		}
	}
}
