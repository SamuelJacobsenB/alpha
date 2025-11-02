package parser

// AST simples: Stmt (decl/stmt) e Expr (expressões)

// Stmt representa uma instrução/declaração
type Stmt interface{ stmtNode() }

// Expr representa uma expressão
type Expr interface{ exprNode() }

type Program struct {
	Body []Stmt
}

// Statements / Declarações
type VarDecl struct {
	Name string
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
	Op   string
	Expr Expr
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
