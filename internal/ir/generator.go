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
	// Pré-passo: Registrar structs
	for _, stmt := range prog.Body {
		if s, ok := stmt.(*parser.StructDecl); ok {
			g.builder.Module.Structs = append(g.builder.Module.Structs, s)
		}
	}

	// Geração de código
	for _, stmt := range prog.Body {
		g.genGlobalStmt(stmt)
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
		// Variáveis globais normalmente vão para uma função 'init' ou segmento de dados
		// Simplificação: Adicionando a um 'init' implícito ou estrutura global
		// TODO: Implementar inicialização global
	}
}

func (g *Generator) genFunction(fn *parser.FunctionDecl) {
	irFunc := &Function{
		Name:       fn.Name,
		TempCount:  0,
		LabelCount: 0,
		IsExported: isExported(fn.Name), // Lógica baseada em capitalização ou keyword export
	}

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
	}
}

func (g *Generator) genVarDecl(decl *parser.VarDecl) {
	// ALLOCA type
	typ := semantic.ToType(decl.Type) // Assumindo conversão disponível
	varOp := Var(decl.Name, typ)

	// IR: %var = ALLOCA type
	g.builder.Emit(ALLOCA, &Operand{Kind: OpType, Type: typ}, nil, varOp)

	if decl.Init != nil {
		val := g.genExpr(decl.Init)
		// IR: STORE %var, %val
		// Nota: Dependendo da semântica, pode ser MOV se for SSA/Register-based
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
		// Multi-return: Pode ser implementado retornando uma struct ou ponteiro
		// Simplificação: Emitimos múltiplos RETs ou um RET com ponteiro agregado
		// Aqui assumiremos que o backend lida com múltiplos valores
		for _, expr := range ret.Values {
			val := g.genExpr(expr)
			// Pseudocódigo para push no stack de retorno
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
	g.builder.EmitJump(endLabel)

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

	g.builder.EmitLabel(startLabel)

	cond := g.genExpr(stmt.Cond)
	g.builder.Emit(JMP_FALSE, cond, endLabel, nil)

	for _, s := range stmt.Body {
		// TODO: Adicionar suporte a break/continue empilhando labels
		g.genStmt(s)
	}

	g.builder.EmitJump(startLabel)
	g.builder.EmitLabel(endLabel)
}

func (g *Generator) genFor(stmt *parser.ForStmt) {
	// Escopo do for
	if stmt.Init != nil {
		g.genStmt(stmt.Init)
	}

	startLabel := g.builder.NewLabel("for_start")
	endLabel := g.builder.NewLabel("for_end")

	g.builder.EmitLabel(startLabel)

	if stmt.Cond != nil {
		cond := g.genExpr(stmt.Cond)
		g.builder.Emit(JMP_FALSE, cond, endLabel, nil)
	}

	for _, s := range stmt.Body {
		g.genStmt(s)
	}

	if stmt.Post != nil {
		g.genStmt(stmt.Post)
	}

	g.builder.EmitJump(startLabel)
	g.builder.EmitLabel(endLabel)
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
	default:
		// Fallback para outros tipos não implementados aqui
		return g.builder.NewTemp(nil)
	}
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
	default:
		panic("Unknown operator " + e.Op)
	}

	result := g.builder.NewTemp(nil) // Tipo deve vir do Semantic
	g.builder.Emit(op, left, right, result)
	return result
}

func (g *Generator) genLogicalShortCircuit(e *parser.BinaryExpr) *Operand {
	// Implementação de &&:
	// t1 = left
	// JMP_FALSE t1, end
	// t1 = right
	// end:

	result := g.builder.NewTemp(nil)
	left := g.genExpr(e.Left)

	// Copia left para result
	g.builder.Emit(MOV, left, nil, result)

	endLabel := g.builder.NewLabel("logic_end")

	if e.Op == "&&" {
		g.builder.Emit(JMP_FALSE, result, endLabel, nil)
	} else { // ||
		g.builder.Emit(JMP_TRUE, result, endLabel, nil)
	}

	right := g.genExpr(e.Right)
	g.builder.Emit(MOV, right, nil, result)

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

	// Emitir PARAM instructions ou passar lista no CALL
	// Simulação de PARAM instructions (estilo assembly)
	for i := len(args) - 1; i >= 0; i-- {
		// PUSH param (opcional, dependendo do backend)
	}

	result := g.builder.NewTemp(nil)

	// Codificar argumentos na instrução CALL para simplificar TAC
	// Hack: Usamos Value do Arg1 para armazenar lista de argumentos ou criamos instrução customizada
	// Aqui emitiremos uma sequência

	instr := g.builder.Emit(CALL, callee, nil, result)
	// Para um IR mais robusto, Instruction teria lista de Args, mas aqui vamos simular
	// anexando metadados ou assumindo que o backend olha para as instruções anteriores
	_ = instr // evitar unused

	return result
}

func (g *Generator) genBuiltin(name string, args []*Operand) *Operand {
	res := g.builder.NewTemp(nil)

	// Nota: Muitos built-ins na sua sintaxe recebem ponteiros (&arr)
	// O IR assume que args[0] já é o endereço ou a referência correta
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
		g.builder.Emit(SUB, args[0], args[1], res) // No IR tratamos como uma operação de remoção
	case "removeIndex":
		// removeIndex(&arr, index)
		g.builder.Emit(GET_INDEX, args[0], args[1], nil) // Marca para remoção
		g.builder.Emit(NOP, nil, nil, res)
	case "delete":
		// delete(&map, key)
		g.builder.Emit(SUB, args[0], args[1], res)
	case "add":
		// add(&set, value)
		g.builder.Emit(ADD, args[0], args[1], res)
	case "clear":
		// clear(&map)
		g.builder.Emit(MOV, Literal("0", nil), nil, args[0])
	case "has":
		// has(&set, value) -> retorna bool
		g.builder.Emit(EQ, args[0], args[1], res)
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
	// Simplificação: genExpr retorna valor, precisaríamos de genAddr para o lado esquerdo
	// Assumindo genAddr para AssignExpr:
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
		return Var(e.Name, nil) // Em TAC, Var pode ser tratado como endr
	case *parser.IndexExpr:
		// Calcula endereço do elemento
		arr := g.genExpr(e.Array)
		idx := g.genExpr(e.Index)
		res := g.builder.NewTemp(nil)
		g.builder.Emit(GET_ADDR, arr, idx, res) // Opcode específico para ptr math
		return res
	default:
		panic("Cannot assign to this expression")
	}
}

func isBuiltin(name string) bool {
	builtins := map[string]bool{
		"append": true, "length": true, "remove": true, "delete": true,
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
