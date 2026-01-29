package ir

import (
	"github.com/alpha/internal/parser"
	"github.com/alpha/internal/semantic"
)

type Generator struct {
	builder *IRBuilder
	checker *semantic.Checker // Para consultar tipos resolvidos
}

func NewGenerator(checker *semantic.Checker) *Generator {
	mod := &Module{
		Name:      "main", // Padrão, pode vir do PackageDecl
		Functions: make([]*Function, 0),
		Globals:   make([]*Instruction, 0),
	}
	return &Generator{
		builder: NewBuilder(mod),
		checker: checker,
	}
}

func (g *Generator) Generate(prog *parser.Program) *Module {
	// 1. Pré-passo: Registrar structs
	for _, stmt := range prog.Body {
		if s, ok := stmt.(*parser.StructDecl); ok {
			g.builder.Module.Structs = append(g.builder.Module.Structs, s)
		}
	}

	// 2. Geração de código (isso preencherá as funções e o corpo do init)
	for _, stmt := range prog.Body {
		g.genGlobalStmt(stmt)
	}

	// 3. FINALIZAÇÃO: Selar a função init
	// Procuramos a função 'init' que foi criada pelo ensureInitFunction
	for _, fn := range g.builder.Module.Functions {
		if fn.Name == "init" {
			// Verificamos se a última instrução já não é um RET
			// para evitar retornos duplicados
			lastIdx := len(fn.Instructions) - 1
			if lastIdx < 0 || fn.Instructions[lastIdx].Op != RET {
				// Forçamos o builder a apontar para o init e emitimos o retorno final
				g.builder.CurrentFunc = fn
				g.builder.Emit(RET, nil, nil, nil)
			}
		}
	}

	return g.builder.Module
}

// ============================
// Declarações Globais
// ============================

func (g *Generator) genGlobalStmt(stmt parser.Stmt) {
	switch s := stmt.(type) {
	case *parser.FunctionDecl:
		g.genFunction(s)
	case *parser.PackageDecl:
		g.builder.Module.Name = s.Name
	case *parser.VarDecl:
		g.genGlobalVarDecl(s)
	case *parser.ConstDecl:
		g.genGlobalConstDecl(s)
	}
}

func (g *Generator) genGlobalVarDecl(decl *parser.VarDecl) {
	// Inicialização de variáveis globais
	typ := g.resolveType(decl.Type)
	globalOp := &Operand{
		Kind:  OpVar,
		Value: decl.Name,
		Type:  typ,
	}

	// Armazenar no módulo como uma instrução de alocação
	g.builder.Module.Globals = append(g.builder.Module.Globals, &Instruction{
		Op:     ALLOCA,
		Result: globalOp,
		Arg1:   &Operand{Kind: OpType, Type: typ},
	})

	if decl.Init != nil {
		initFunc := g.ensureInitFunction()
		currentFunc := g.builder.CurrentFunc
		g.builder.CurrentFunc = initFunc

		val := g.genExpr(decl.Init)
		// Emitir MOV ao invés de STORE para globais
		globalOp := &Operand{
			Kind:  OpVar,
			Value: decl.Name,
			Type:  val.Type,
		}
		g.builder.Emit(MOV, val, nil, globalOp)

		g.builder.CurrentFunc = currentFunc
	}
}

func (g *Generator) genGlobalConstDecl(decl *parser.ConstDecl) {
	// Constantes globais são resolvidas em tempo de compilação
	// Usamos genExpr para obter o operando e seu tipo
	valOperand := g.genExpr(decl.Init)

	globalOp := &Operand{
		Kind:  OpVar,
		Value: decl.Name,
		Type:  valOperand.Type,
	}

	g.builder.Module.Globals = append(g.builder.Module.Globals, &Instruction{
		Op:     MOV,
		Result: globalOp,
		Arg1:   valOperand,
	})
}

func (g *Generator) ensureInitFunction() *Function {
	// Verifica se já existe uma função init
	for _, fn := range g.builder.Module.Functions {
		if fn.Name == "init" {
			return fn
		}
	}

	// Cria função init se não existir
	initFunc := &Function{
		Name:       "init",
		TempCount:  0,
		LabelCount: 0,
		IsExported: false,
	}

	// Adiciona ao módulo
	g.builder.Module.Functions = append(g.builder.Module.Functions, initFunc)

	// Se estamos no meio da geração de outra função,
	// precisamos garantir que a função init tenha um RET
	savedFunc := g.builder.CurrentFunc
	g.builder.CurrentFunc = initFunc

	// Restaura a função atual
	if savedFunc != nil {
		g.builder.CurrentFunc = savedFunc
	}

	return initFunc
}

func (g *Generator) resolveType(parserType parser.Type) semantic.Type {
	if parserType == nil {
		return nil
	}

	// Usamos o checker para resolver o tipo
	switch t := parserType.(type) {
	case *parser.PrimitiveType:
		return &semantic.ParserTypeWrapper{Type: t}
	case *parser.IdentifierType:
		// Para tipos definidos pelo usuário, por enquanto, retornamos um wrapper básico
		return &semantic.ParserTypeWrapper{Type: t}
	case *parser.ArrayType:
		// Por enquanto, não resolvemos o tipo do elemento recursivamente
		return &semantic.ParserTypeWrapper{
			Type: &parser.ArrayType{
				ElementType: &parser.PrimitiveType{Name: "any"}, // Simplificação
				Size:        t.Size,
			},
		}
	case *parser.PointerType:
		// Por enquanto, não resolvemos o tipo base recursivamente
		return &semantic.ParserTypeWrapper{
			Type: &parser.PointerType{
				BaseType: &parser.PrimitiveType{Name: "any"}, // Simplificação
			},
		}
	default:
		// Para outros tipos, retorna um tipo "any"
		return &semantic.ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}
	}
}

func (g *Generator) genFunction(fn *parser.FunctionDecl) {
	irFunc := &Function{
		Name:       fn.Name,
		TempCount:  0,
		LabelCount: 0,
		IsExported: isExported(fn.Name),
		ReturnType: semantic.ToType(fn.ReturnTypes[0]),
	}

	generics := make([]string, 0)
	if fn.Generics != nil {
		for _, gen := range fn.Generics {
			generics = append(generics, gen.Name)
		}
	}
	irFunc.Generics = generics

	// Configurar builder para a nova função
	g.builder.CurrentFunc = irFunc

	// Processar parâmetros
	for _, param := range fn.Params {
		operand := Var(param.Name, semantic.ToType(param.Type))
		irFunc.Params = append(irFunc.Params, operand)
		// Em algumas arquiteturas, precisamos fazer STORE do param registro -> stack
	}

	// Gerar corpo
	for _, stmt := range fn.Body {
		g.genStmt(stmt)
	}

	g.builder.Module.Functions = append(g.builder.Module.Functions, irFunc)
}

// ============================
// Statements
// ============================

func (g *Generator) genStmt(stmt parser.Stmt) {
	switch s := stmt.(type) {
	case *parser.VarDecl:
		g.genVarDecl(s)
	case *parser.ExprStmt:
		g.genExpr(s.Expr) // Avalia expressão (efeitos colaterais)
	case *parser.ReturnStmt:
		g.genReturn(s)
	case *parser.IfStmt:
		g.genIf(s)
	case *parser.WhileStmt:
		g.genWhile(s)
	case *parser.ForStmt:
		g.genFor(s)
	case *parser.BlockStmt:
		for _, sub := range s.Body {
			g.genStmt(sub)
		}
	case *parser.SwitchStmt:
		g.genSwitch(s)
	case *parser.BreakStmt:
		g.genBreak()
	case *parser.ContinueStmt:
		g.genContinue()
	}
}

func (g *Generator) genVarDecl(decl *parser.VarDecl) {
	// ALLOCA type
	typ := semantic.ToType(decl.Type)
	varOp := Var(decl.Name, typ)

	// IR: %var = ALLOCA type
	g.builder.Emit(ALLOCA, &Operand{Kind: OpType, Type: typ}, nil, varOp)

	if decl.Init != nil {
		val := g.genExpr(decl.Init)
		// IR: STORE %var, %val
		g.builder.Emit(STORE, varOp, val, nil)
	}
}

func (g *Generator) genReturn(ret *parser.ReturnStmt) {
	if len(ret.Values) == 0 {
		g.builder.Emit(RET, nil, nil, nil)
		return
	}

	if len(ret.Values) == 1 {
		val := g.genExpr(ret.Values[0])
		g.builder.Emit(RET, val, nil, nil)
	} else {
		// Multi-return: criar uma struct temporária
		// Para simplificação, emitimos múltiplos valores em registros
		// (backend precisará lidar com isso)
		for _, expr := range ret.Values {
			val := g.genExpr(expr)
			g.builder.Emit(RET, val, nil, nil)
		}
	}
}

func (g *Generator) genIf(stmt *parser.IfStmt) {
	cond := g.genExpr(stmt.Cond)

	trueLabel := g.builder.NewLabel("if_then")
	elseLabel := g.builder.NewLabel("if_else")
	endLabel := g.builder.NewLabel("if_end")

	targetElse := elseLabel
	if stmt.Else == nil {
		targetElse = endLabel
	}

	// Se for falso, pula para o else/fim
	g.builder.Emit(JMP_FALSE, cond, targetElse, nil)

	// Bloco Then
	g.builder.EmitLabel(trueLabel)
	for _, s := range stmt.Then {
		g.genStmt(s)
	}
	g.builder.Emit(JMP, endLabel, nil, nil)

	// Bloco Else (se existir)
	if stmt.Else != nil {
		g.builder.EmitLabel(elseLabel)
		for _, s := range stmt.Else {
			g.genStmt(s)
		}
	}

	g.builder.EmitLabel(endLabel)
}

func (g *Generator) genWhile(stmt *parser.WhileStmt) {
	startLabel := g.builder.NewLabel("while_start")
	endLabel := g.builder.NewLabel("while_end")

	// Empilhar labels para break/continue
	prevBreak := g.builder.CurrentFunc.LabelCount
	g.builder.CurrentFunc.LabelCount += 2

	g.builder.EmitLabel(startLabel)

	cond := g.genExpr(stmt.Cond)
	g.builder.Emit(JMP_FALSE, cond, endLabel, nil)

	for _, s := range stmt.Body {
		g.genStmt(s)
	}

	g.builder.Emit(JMP, startLabel, nil, nil)
	g.builder.EmitLabel(endLabel)

	// Restaurar labels
	g.builder.CurrentFunc.LabelCount = prevBreak
}

func (g *Generator) genFor(stmt *parser.ForStmt) {
	// Escopo do for
	if stmt.Init != nil {
		g.genStmt(stmt.Init)
	}

	startLabel := g.builder.NewLabel("for_start")
	condLabel := g.builder.NewLabel("for_cond")
	endLabel := g.builder.NewLabel("for_end")

	g.builder.Emit(JMP, condLabel, nil, nil)
	g.builder.EmitLabel(startLabel)

	for _, s := range stmt.Body {
		g.genStmt(s)
	}

	if stmt.Post != nil {
		g.genStmt(stmt.Post)
	}

	g.builder.EmitLabel(condLabel)
	if stmt.Cond != nil {
		cond := g.genExpr(stmt.Cond)
		g.builder.Emit(JMP_TRUE, cond, startLabel, nil)
	} else {
		g.builder.Emit(JMP, startLabel, nil, nil)
	}

	g.builder.EmitLabel(endLabel)
}

func (g *Generator) genSwitch(stmt *parser.SwitchStmt) {
	// Implementação simplificada de switch
	expr := g.genExpr(stmt.Expr)
	endLabel := g.builder.NewLabel("switch_end")

	for _, clause := range stmt.Cases {
		caseLabel := g.builder.NewLabel("case")

		if clause.Value != nil {
			caseVal := g.genExpr(clause.Value)
			cond := g.builder.NewTemp(nil)
			g.builder.Emit(EQ, expr, caseVal, cond)
			g.builder.Emit(JMP_TRUE, cond, caseLabel, nil)
		} else {
			// default case
			g.builder.EmitLabel(caseLabel)
		}

		for _, bodyStmt := range clause.Body {
			g.genStmt(bodyStmt)
		}
		g.builder.Emit(JMP, endLabel, nil, nil)
	}

	g.builder.EmitLabel(endLabel)
}

func (g *Generator) genBreak() {
	// Implementação simplificada: pular para o fim do loop mais interno
	endLabel := g.builder.NewLabel("loop_end")
	g.builder.Emit(JMP, endLabel, nil, nil)
}

func (g *Generator) genContinue() {
	// Implementação simplificada: pular para o início do loop mais interno
	startLabel := g.builder.NewLabel("loop_start")
	g.builder.Emit(JMP, startLabel, nil, nil)
}

// ============================
// Expressões
// ============================

func (g *Generator) genExpr(expr parser.Expr) *Operand {
	switch e := expr.(type) {
	case *parser.IntLiteral:
		return IntLiteral(e.Value)
	case *parser.BoolLiteral:
		return BoolLiteral(e.Value)
	case *parser.StringLiteral:
		return Literal(e.Value, nil) // Tipo string
	case *parser.Identifier:
		// Assumimos que semantic check já resolveu se existe
		return Var(e.Name, nil)
	case *parser.BinaryExpr:
		return g.genBinaryExpr(e)
	case *parser.CallExpr:
		return g.genCallExpr(e)
	case *parser.AssignExpr:
		return g.genAssign(e)
	case *parser.IndexExpr:
		return g.genIndexExpr(e)
	case *parser.MemberExpr:
		return g.genMemberExpr(e)
	case *parser.UnaryExpr:
		return g.genUnaryExpr(e)
	case *parser.TernaryExpr:
		return g.genTernaryExpr(e)
	case *parser.TypeCastExpr:
		return g.genTypeCast(e)
	default:
		// Fallback para outros tipos não implementados aqui
		return g.builder.NewTemp(nil)
	}
}

func (g *Generator) genUnaryExpr(e *parser.UnaryExpr) *Operand {
	expr := g.genExpr(e.Expr)
	res := g.builder.NewTemp(nil)

	if e.Postfix {
		// Para pós-fixo (i++), retornar valor original, depois incrementar
		temp := g.builder.NewTemp(nil)
		g.builder.Emit(MOV, expr, nil, temp)
		g.builder.Emit(ADD, expr, Literal("1", nil), expr)
		return temp
	}

	switch e.Op {
	case "-":
		g.builder.Emit(SUB, Literal("0", nil), expr, res)
	case "!":
		g.builder.Emit(XOR, expr, Literal("1", nil), res)
	case "&":
		// Operador de endereço
		return g.genAddr(e.Expr)
	case "*":
		// Dereferenciação de ponteiro
		g.builder.Emit(LOAD, expr, nil, res)
	default:
		g.builder.Emit(MOV, expr, nil, res)
	}
	return res
}

func (g *Generator) genTernaryExpr(e *parser.TernaryExpr) *Operand {
	cond := g.genExpr(e.Cond)
	trueLabel := g.builder.NewLabel("ternary_true")
	falseLabel := g.builder.NewLabel("ternary_false")
	endLabel := g.builder.NewLabel("ternary_end")

	result := g.builder.NewTemp(nil)

	g.builder.Emit(JMP_FALSE, cond, falseLabel, nil)
	g.builder.EmitLabel(trueLabel)
	trueVal := g.genExpr(e.TrueExpr)
	g.builder.Emit(MOV, trueVal, nil, result)
	g.builder.Emit(JMP, endLabel, nil, nil)

	g.builder.EmitLabel(falseLabel)
	falseVal := g.genExpr(e.FalseExpr)
	g.builder.Emit(MOV, falseVal, nil, result)

	g.builder.EmitLabel(endLabel)
	return result
}

func (g *Generator) genTypeCast(e *parser.TypeCastExpr) *Operand {
	expr := g.genExpr(e.Expr)
	res := g.builder.NewTemp(semantic.ToType(e.Type))

	// Emitir instrução CAST
	g.builder.Emit(CAST, expr, nil, res)
	return res
}

func (g *Generator) genBinaryExpr(e *parser.BinaryExpr) *Operand {
	// Curto circuito para && e ||
	if e.Op == "&&" || e.Op == "||" {
		return g.genLogicalShortCircuit(e)
	}

	left := g.genExpr(e.Left)
	right := g.genExpr(e.Right)

	// Determina OpCode
	var op OpCode
	switch e.Op {
	case "+":
		op = ADD
	case "-":
		op = SUB
	case "*":
		op = MUL
	case "/":
		op = DIV
	case "%":
		op = MOD
	case "==":
		op = EQ
	case "!=":
		op = NEQ
	case "<":
		op = LT
	case ">":
		op = GT
	case "<=":
		op = LE
	case ">=":
		op = GE
	case "&":
		op = AND
	case "|":
		op = OR
	case "^":
		op = XOR
	case "<<":
		op = SHL
	case ">>":
		op = SHR
	default:
		panic("Unknown operator " + e.Op)
	}

	result := g.builder.NewTemp(nil)
	g.builder.Emit(op, left, right, result)
	return result
}

func (g *Generator) genLogicalShortCircuit(e *parser.BinaryExpr) *Operand {
	result := g.builder.NewTemp(nil)
	left := g.genExpr(e.Left)

	endLabel := g.builder.NewLabel("logic_end")

	if e.Op == "&&" {
		// left && right
		g.builder.Emit(JMP_FALSE, left, endLabel, nil)
		right := g.genExpr(e.Right)
		g.builder.Emit(MOV, right, nil, result)
	} else { // ||
		// left || right
		g.builder.Emit(JMP_TRUE, left, endLabel, nil)
		right := g.genExpr(e.Right)
		g.builder.Emit(MOV, right, nil, result)
	}

	g.builder.EmitLabel(endLabel)
	return result
}

func (g *Generator) genCallExpr(e *parser.CallExpr) *Operand {
	var args []*Operand
	for _, arg := range e.Args {
		args = append(args, g.genExpr(arg))
	}

	// Resolve callee
	var callee *Operand
	if ident, ok := e.Callee.(*parser.Identifier); ok {
		// Trata built-ins
		if isBuiltin(ident.Name) {
			return g.genBuiltin(ident.Name, args)
		}
		callee = &Operand{Kind: OpFunction, Value: ident.Name}
	} else {
		callee = g.genExpr(e.Callee) // Ponteiro de função
	}

	result := g.builder.NewTemp(nil)

	instr := g.builder.Emit(CALL, callee, nil, result)
	instr.Args = args

	return result
}

func (g *Generator) genBuiltin(name string, args []*Operand) *Operand {
	res := g.builder.NewTemp(nil)

	switch name {
	case "length":
		g.builder.Emit(LEN, args[0], nil, res)
	case "append":
		if len(args) < 2 {
			panic("append requires 2 arguments")
		}
		g.builder.Emit(APPEND, args[0], args[1], res)
	case "remove":
		// remove(&arr, element)
		g.builder.Emit(REMOVE, args[0], args[1], res)
	case "removeIndex":
		// removeIndex(&arr, index)
		g.builder.Emit(REMOVE_INDEX, args[0], args[1], res)
	case "delete":
		// delete(&map, key)
		g.builder.Emit(DELETE, args[0], args[1], res)
	case "add":
		// add(&set, value)
		g.builder.Emit(ADD, args[0], args[1], res)
	case "clear":
		// clear(&map)
		g.builder.Emit(CLEAR, args[0], nil, res)
	case "has":
		// has(&set, value) -> retorna bool
		g.builder.Emit(HAS, args[0], args[1], res)
	default:
		// Para funções não mapeadas, tratamos como uma chamada genérica de sistema
		g.builder.Emit(CALL, &Operand{Kind: OpFunction, Value: "builtin_" + name}, nil, res)
	}
	return res
}

func (g *Generator) genAssign(e *parser.AssignExpr) *Operand {
	val := g.genExpr(e.Right)

	// Se Left for identificador simples
	if ident, ok := e.Left.(*parser.Identifier); ok {
		dest := Var(ident.Name, nil)
		g.builder.Emit(STORE, dest, val, nil)
		return val
	}

	// Se for acesso complexo (array/struct), precisamos do endereço
	addr := g.genAddr(e.Left)
	g.builder.Emit(STORE, addr, val, nil)

	return val
}

func (g *Generator) genIndexExpr(e *parser.IndexExpr) *Operand {
	arr := g.genExpr(e.Array)
	idx := g.genExpr(e.Index)
	res := g.builder.NewTemp(nil)
	g.builder.Emit(GET_INDEX, arr, idx, res)
	return res
}

func (g *Generator) genMemberExpr(e *parser.MemberExpr) *Operand {
	obj := g.genExpr(e.Object)
	field := &Operand{Kind: OpField, Value: e.Member}
	res := g.builder.NewTemp(nil)
	g.builder.Emit(GET_FIELD, obj, field, res)
	return res
}

// Helpers
func (g *Generator) genAddr(expr parser.Expr) *Operand {
	// Lógica para obter endereço de memória ao invés do valor
	switch e := expr.(type) {
	case *parser.Identifier:
		return Var(e.Name, nil)
	case *parser.IndexExpr:
		// Calcula endereço do elemento
		arr := g.genExpr(e.Array)
		idx := g.genExpr(e.Index)
		res := g.builder.NewTemp(nil)
		g.builder.Emit(GET_ADDR, arr, idx, res)
		return res
	case *parser.MemberExpr:
		obj := g.genExpr(e.Object)
		field := &Operand{Kind: OpField, Value: e.Member}
		res := g.builder.NewTemp(nil)
		g.builder.Emit(GET_ADDR, obj, field, res)
		return res
	default:
		panic("Cannot take address of this expression")
	}
}

func isBuiltin(name string) bool {
	builtins := map[string]bool{
		"append": true, "length": true, "remove": true, "delete": true,
		"add": true, "clear": true, "has": true, "removeIndex": true,
	}
	return builtins[name]
}

func isExported(name string) bool {
	// Exemplo: letra maiúscula exporta
	if len(name) > 0 {
		return name[0] >= 'A' && name[0] <= 'Z'
	}
	return false
}
