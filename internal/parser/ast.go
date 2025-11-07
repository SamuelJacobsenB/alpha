package parser

// Stmt representa uma instrução/declaração
type Stmt interface{ stmtNode() }

// Expr representa uma expressão
type Expr interface{ exprNode() }

// Type representa um tipo na linguagem
type Type interface{ typeNode() }

// PrimitiveType representa tipos básicos como int, string, etc.
type PrimitiveType struct {
	Name string // ex: "int", "string", "float"
}

func (*PrimitiveType) typeNode() {}

// ArrayType representa um tipo de array
type ArrayType struct {
	ElementType Type
	Size        Expr // pode ser nil para arrays dinâmicos
}

func (*ArrayType) typeNode() {}

type Program struct {
	Body []Stmt
}

// Statements / Declarações
type VarDecl struct {
	Name string
	Type Type // pode ser nil se inferido (ex: var x = 10)
	Init Expr // pode ser nil
}

func (v *VarDecl) stmtNode() {}

type ConstDecl struct {
	Name string
	Init Expr
}

func (c *ConstDecl) stmtNode() {}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) stmtNode() {}

type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (i *IfStmt) stmtNode() {}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
}

func (w *WhileStmt) stmtNode() {}

type ForStmt struct {
	Init Stmt // pode ser nil
	Cond Expr // pode ser nil
	Post Stmt // pode ser nil
	Body []Stmt
}

func (f *ForStmt) stmtNode() {}

type ForInStmt struct {
	Index    *Identifier // pode ser nil (apenas item)
	Item     *Identifier
	Iterable Expr
	Body     []Stmt
}

func (f *ForInStmt) stmtNode() {}

// FunctionDecl representa uma declaração de função
type FunctionDecl struct {
	Name       string
	Generics   []*GenericParam // Parâmetros genéricos [T, U]
	Params     []*Param        // Parâmetros da função
	ReturnType Type            // Tipo de retorno
	Body       []Stmt          // Corpo da função
}

func (f *FunctionDecl) stmtNode() {}

// GenericParam representa um parâmetro genérico como [T]
type GenericParam struct {
	Name string
}

func (g *GenericParam) typeNode() {}

// Param representa um parâmetro de função
type Param struct {
	Name string
	Type Type
}

// FunctionType representa um tipo de função
type FunctionType struct {
	Params     []Type
	ReturnType Type
}

func (f *FunctionType) typeNode() {}

// FunctionExpr representa uma expressão de função (função anônima)
type FunctionExpr struct {
	Generics   []*GenericParam
	Params     []*Param
	ReturnType Type
	Body       []Stmt
}

func (f *FunctionExpr) exprNode() {}

// Adicione esta struct no arquivo ast.go se ainda não existir
type GenericSpecialization struct {
	Callee   Expr
	TypeArgs []Type
}

func (g *GenericSpecialization) exprNode() {}

type ReturnStmt struct {
	Value Expr // pode ser nil
}

func (r *ReturnStmt) stmtNode() {}

type BlockStmt struct {
	Body []Stmt
}

func (b *BlockStmt) stmtNode() {}

// Expressões
type Identifier struct{ Name string }

func (i *Identifier) exprNode() {}

type IntLiteral struct{ Value int64 }

func (i *IntLiteral) exprNode() {}

type FloatLiteral struct{ Value float64 }

func (f *FloatLiteral) exprNode() {}

type StringLiteral struct{ Value string }

func (s *StringLiteral) exprNode() {}

type BoolLiteral struct{ Value bool }

func (b *BoolLiteral) exprNode() {}

type NullLiteral struct{}

func (n *NullLiteral) exprNode() {}

type UnaryExpr struct {
	Op      string
	Expr    Expr
	Postfix bool // true para i++, false para ++i
}

func (u *UnaryExpr) exprNode() {}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) exprNode() {}

type CallExpr struct {
	Callee Expr
	Args   []Expr
}

func (c *CallExpr) exprNode() {}

type AssignExpr struct {
	Left  Expr // geralmente Identifier
	Right Expr
}

func (a *AssignExpr) exprNode() {}

// ArrayLiteral representa um literal de array {1, 2, 3}
type ArrayLiteral struct {
	Elements []Expr
}

func (a *ArrayLiteral) exprNode() {}

// IndexExpr representa acesso a array: arr[index]
type IndexExpr struct {
	Array Expr
	Index Expr
}

func (i *IndexExpr) exprNode() {}
