package parser

// Node é a interface comum a todos os nós da AST.
type Node interface {
	nodePos()
}

// Stmt representa uma instrução/declaração
type Stmt interface {
	stmtNode()
	nodePos()
}

// Expr representa uma expressão
type Expr interface {
	exprNode()
	nodePos()
}

// Type representa um tipo na linguagem
type Type interface {
	typeNode()
	nodePos()
}

// Program
type Program struct {
	Body []Stmt
}

// Statements / Declarações
type VarDecl struct {
	Name string
	Type Type
	Init Expr
}

func (v *VarDecl) stmtNode() {}
func (v *VarDecl) nodePos()  {}

type ConstDecl struct {
	Name string
	Init Expr
}

func (c *ConstDecl) stmtNode() {}
func (c *ConstDecl) nodePos()  {}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) stmtNode() {}
func (e *ExprStmt) nodePos()  {}

type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (i *IfStmt) stmtNode() {}
func (i *IfStmt) nodePos()  {}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
}

func (w *WhileStmt) stmtNode() {}
func (w *WhileStmt) nodePos()  {}

type DoWhileStmt struct {
	Body []Stmt
	Cond Expr
}

func (d *DoWhileStmt) stmtNode() {}
func (d *DoWhileStmt) nodePos()  {}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Body []Stmt
}

func (f *ForStmt) stmtNode() {}
func (f *ForStmt) nodePos()  {}

type ForInStmt struct {
	Index    *Identifier
	Item     *Identifier
	Iterable Expr
	Body     []Stmt
}

func (f *ForInStmt) stmtNode() {}
func (f *ForInStmt) nodePos()  {}

type SwitchStmt struct {
	Expr  Expr
	Cases []*CaseClause
}

func (s *SwitchStmt) stmtNode() {}
func (s *SwitchStmt) nodePos()  {}

type CaseClause struct {
	Value Expr
	Body  []Stmt
}

func (c *CaseClause) nodePos() {}

type FunctionDecl struct {
	Name       string
	Generics   []*GenericParam
	Params     []*Param
	ReturnType Type
	Body       []Stmt
}

func (f *FunctionDecl) stmtNode() {}
func (f *FunctionDecl) nodePos()  {}

type GenericParam struct {
	Name string
}

func (g *GenericParam) typeNode() {}
func (g *GenericParam) nodePos()  {}

type Param struct {
	Name string
	Type Type
}

func (p *Param) nodePos() {}

type FunctionType struct {
	Params     []Type
	ReturnType Type
}

func (f *FunctionType) typeNode() {}
func (f *FunctionType) nodePos()  {}

type FunctionExpr struct {
	Generics   []*GenericParam
	Params     []*Param
	ReturnType Type
	Body       []Stmt
}

func (f *FunctionExpr) exprNode() {}
func (f *FunctionExpr) nodePos()  {}

type GenericSpecialization struct {
	Callee   Expr
	TypeArgs []Type
}

func (g *GenericSpecialization) exprNode() {}
func (g *GenericSpecialization) nodePos()  {}

type ClassDecl struct {
	Name        string
	Generics    []*GenericParam
	Fields      []*FieldDecl
	Constructor *ConstructorDecl
	Methods     []*MethodDecl
}

func (c *ClassDecl) stmtNode() {}
func (c *ClassDecl) nodePos()  {}

type FieldDecl struct {
	Name string
	Type Type
}

func (f *FieldDecl) nodePos() {}

type ConstructorDecl struct {
	Params []*Param
	Body   []Stmt
}

func (c *ConstructorDecl) nodePos() {}

type MethodDecl struct {
	Name       string
	Generics   []*GenericParam
	Params     []*Param
	ReturnType Type
	Body       []Stmt
}

func (m *MethodDecl) nodePos() {}

type TypeDecl struct {
	Name     string
	Generics []*GenericParam
	Type     Type
}

func (t *TypeDecl) stmtNode() {}
func (t *TypeDecl) nodePos()  {}

type ReturnStmt struct {
	Value Expr
}

func (r *ReturnStmt) stmtNode() {}
func (r *ReturnStmt) nodePos()  {}

type BlockStmt struct {
	Body []Stmt
}

func (b *BlockStmt) stmtNode() {}
func (b *BlockStmt) nodePos()  {}

// Expressões
type Identifier struct{ Name string }

func (i *Identifier) exprNode() {}
func (i *Identifier) nodePos()  {}

type IntLiteral struct{ Value int64 }

func (i *IntLiteral) exprNode() {}
func (i *IntLiteral) nodePos()  {}

type FloatLiteral struct{ Value float64 }

func (f *FloatLiteral) exprNode() {}
func (f *FloatLiteral) nodePos()  {}

type StringLiteral struct{ Value string }

func (s *StringLiteral) exprNode() {}
func (s *StringLiteral) nodePos()  {}

type BoolLiteral struct{ Value bool }

func (b *BoolLiteral) exprNode() {}
func (b *BoolLiteral) nodePos()  {}

type NullLiteral struct{}

func (n *NullLiteral) exprNode() {}
func (n *NullLiteral) nodePos()  {}

type UnaryExpr struct {
	Op      string
	Expr    Expr
	Postfix bool
}

func (u *UnaryExpr) exprNode() {}
func (u *UnaryExpr) nodePos()  {}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) exprNode() {}
func (b *BinaryExpr) nodePos()  {}

type TernaryExpr struct {
	Cond      Expr
	TrueExpr  Expr
	FalseExpr Expr
}

func (t *TernaryExpr) exprNode() {}
func (t *TernaryExpr) nodePos()  {}

type CallExpr struct {
	Callee Expr
	Args   []Expr
}

func (c *CallExpr) exprNode() {}
func (c *CallExpr) nodePos()  {}

type AssignExpr struct {
	Left  Expr
	Right Expr
}

func (a *AssignExpr) exprNode() {}
func (a *AssignExpr) nodePos()  {}

type ArrayLiteral struct {
	Elements []Expr
}

func (a *ArrayLiteral) exprNode() {}
func (a *ArrayLiteral) nodePos()  {}

type IndexExpr struct {
	Array Expr
	Index Expr
}

func (i *IndexExpr) exprNode() {}
func (i *IndexExpr) nodePos()  {}

type SetLiteral struct {
	Elements []Expr
}

func (*SetLiteral) exprNode() {}
func (*SetLiteral) nodePos()  {}

type MapLiteral struct {
	Entries []*MapEntry
}

func (*MapLiteral) exprNode() {}
func (*MapLiteral) nodePos()  {}

type MapEntry struct {
	Key   Expr
	Value Expr
}

func (*MapEntry) nodePos() {}

type ReferenceExpr struct {
	Expr Expr
}

func (*ReferenceExpr) exprNode() {}
func (*ReferenceExpr) nodePos()  {}

type NewExpr struct {
	TypeName string
	TypeArgs []Type
	Args     []Expr
}

func (n *NewExpr) exprNode() {}
func (n *NewExpr) nodePos()  {}

type MemberExpr struct {
	Object Expr
	Member string
}

func (m *MemberExpr) exprNode() {}
func (m *MemberExpr) nodePos()  {}

type ThisExpr struct{}

func (t *ThisExpr) exprNode() {}
func (t *ThisExpr) nodePos()  {}

type StructLiteral struct {
	Fields []*StructField
}

func (s *StructLiteral) exprNode() {}
func (s *StructLiteral) nodePos()  {}

type StructField struct {
	Name  string
	Value Expr
}

func (s *StructField) nodePos() {}

// Tipos
type PrimitiveType struct {
	Name string
}

func (*PrimitiveType) typeNode() {}
func (*PrimitiveType) nodePos()  {}

type IdentifierType struct {
	Name string
}

func (*IdentifierType) typeNode() {}
func (*IdentifierType) nodePos()  {}

type ArrayType struct {
	ElementType Type
	Size        Expr
}

func (*ArrayType) typeNode() {}
func (*ArrayType) nodePos()  {}

type NullableType struct {
	BaseType Type
}

func (*NullableType) typeNode() {}
func (*NullableType) nodePos()  {}

type PointerType struct {
	BaseType Type
}

func (*PointerType) typeNode() {}
func (*PointerType) nodePos()  {}

type SetType struct {
	ElementType Type
}

func (*SetType) typeNode() {}
func (*SetType) nodePos()  {}

type MapType struct {
	KeyType   Type
	ValueType Type
}

func (*MapType) typeNode() {}
func (*MapType) nodePos()  {}

type UnionType struct {
	Types []Type
}

func (*UnionType) typeNode() {}
func (*UnionType) nodePos()  {}

type StructType struct {
	Fields []*FieldDecl
}

func (s *StructType) typeNode() {}
func (s *StructType) nodePos()  {}

type BreakStmt struct{}

func (b *BreakStmt) stmtNode() {}
func (b *BreakStmt) nodePos()  {}

type ContinueStmt struct{}

func (c *ContinueStmt) stmtNode() {}
func (c *ContinueStmt) nodePos()  {}
