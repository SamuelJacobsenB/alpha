package parser

// ============================
// INTERFACES DA AST
// ============================

// Node é a interface comum a todos os nós da AST
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

// ============================
// NÓ RAIZ (PROGRAMA)
// ============================

// Program representa um programa completo
type Program struct {
	Body []Stmt
}

// ============================
// DECLARAÇÕES DE MÓDULO
// ============================

// PackageDecl representa uma declaração de pacote
type PackageDecl struct {
	Name string
}

func (p *PackageDecl) stmtNode() {}
func (p *PackageDecl) nodePos()  {}

// ImportDecl representa uma declaração de importação
type ImportDecl struct {
	Path    string
	Imports []*ImportSpec // nil para importar tudo
}

func (i *ImportDecl) stmtNode() {}
func (i *ImportDecl) nodePos()  {}

// ImportSpec representa um item de importação
type ImportSpec struct {
	Name  string
	Alias string // vazio se não houver alias
}

func (i *ImportSpec) nodePos() {}

// ExportDecl representa uma declaração de exportação
type ExportDecl struct {
	Exports []*ExportSpec
}

func (e *ExportDecl) stmtNode() {}
func (e *ExportDecl) nodePos()  {}

// ExportSpec representa um item de exportação
type ExportSpec struct {
	Name  string
	Alias string // vazio se não houver alias
}

func (e *ExportSpec) nodePos() {}

// ============================
// STATEMENTS DE DECLARAÇÃO
// ============================

// VarDecl representa uma declaração de variável
type VarDecl struct {
	Name string
	Type Type
	Init Expr
}

func (v *VarDecl) stmtNode() {}
func (v *VarDecl) nodePos()  {}

// ConstDecl representa uma declaração de constante
type ConstDecl struct {
	Name string
	Init Expr
}

func (c *ConstDecl) stmtNode() {}
func (c *ConstDecl) nodePos()  {}

// FunctionDecl representa uma declaração de função
type FunctionDecl struct {
	Name       string
	Generics   []*GenericParam
	Params     []*Param
	ReturnType Type
	Body       []Stmt
}

func (f *FunctionDecl) stmtNode() {}
func (f *FunctionDecl) nodePos()  {}

// StructDecl representa a definição de dados de uma estrutura
type StructDecl struct {
	Name     string
	Generics []*GenericParam
	Fields   []*FieldDecl
}

func (s *StructDecl) stmtNode() {}
func (s *StructDecl) nodePos()  {}

// ImplDecl representa um bloco de implementação (métodos e init)
type ImplDecl struct {
	TargetName string    // Nome da struct que está sendo implementada
	Init       *InitDecl // Construtor (opcional)
	Methods    []*MethodDecl
}

func (i *ImplDecl) stmtNode() {}
func (i *ImplDecl) nodePos()  {}

// TypeDecl representa uma declaração de alias de tipo
type TypeDecl struct {
	Name     string
	Generics []*GenericParam
	Type     Type
}

func (t *TypeDecl) stmtNode() {}
func (t *TypeDecl) nodePos()  {}

// ============================
// COMPONENTES ESTRUTURAIS
// ============================

// GenericParam representa um parâmetro genérico
type GenericParam struct {
	Name string
}

func (g *GenericParam) typeNode() {}
func (g *GenericParam) nodePos()  {}

// Param representa um parâmetro de função/método
type Param struct {
	Name string
	Type Type
}

func (p *Param) nodePos() {}

// FieldDecl representa uma declaração de campo
type FieldDecl struct {
	Name      string
	Type      Type
	IsPrivate bool // Flag para campos privados
}

func (f *FieldDecl) nodePos() {}

// InitDecl representa um construtor (init)
type InitDecl struct {
	Params []*Param
	Body   []Stmt
}

func (i *InitDecl) nodePos() {}

// MethodDecl representa uma declaração de método
type MethodDecl struct {
	Name       string
	Generics   []*GenericParam
	Params     []*Param
	ReturnType Type
	Body       []Stmt
}

func (m *MethodDecl) nodePos() {}

// ============================
// STATEMENTS DE CONTROLE DE FLUXO
// ============================

// ExprStmt representa um statement de expressão
type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) stmtNode() {}
func (e *ExprStmt) nodePos()  {}

// IfStmt representa um statement if-else
type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (i *IfStmt) stmtNode() {}
func (i *IfStmt) nodePos()  {}

// WhileStmt representa um loop while
type WhileStmt struct {
	Cond Expr
	Body []Stmt
}

func (w *WhileStmt) stmtNode() {}
func (w *WhileStmt) nodePos()  {}

// DoWhileStmt representa um loop do-while
type DoWhileStmt struct {
	Body []Stmt
	Cond Expr
}

func (d *DoWhileStmt) stmtNode() {}
func (d *DoWhileStmt) nodePos()  {}

// ForStmt representa um for loop tradicional
type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Body []Stmt
}

func (f *ForStmt) stmtNode() {}
func (f *ForStmt) nodePos()  {}

// ForInStmt representa um for-in loop
type ForInStmt struct {
	Index    *Identifier
	Item     *Identifier
	Iterable Expr
	Body     []Stmt
}

func (f *ForInStmt) stmtNode() {}
func (f *ForInStmt) nodePos()  {}

// SwitchStmt representa um statement switch
type SwitchStmt struct {
	Expr  Expr
	Cases []*CaseClause
}

func (s *SwitchStmt) stmtNode() {}
func (s *SwitchStmt) nodePos()  {}

// CaseClause representa um caso em um switch
type CaseClause struct {
	Value Expr
	Body  []Stmt
}

func (c *CaseClause) nodePos() {}

// ============================
// STATEMENTS DE RETORNO E CONTROLE
// ============================

// ReturnStmt representa um statement de retorno
type ReturnStmt struct {
	Value Expr
}

func (r *ReturnStmt) stmtNode() {}
func (r *ReturnStmt) nodePos()  {}

// BreakStmt representa um statement break
type BreakStmt struct{}

func (b *BreakStmt) stmtNode() {}
func (b *BreakStmt) nodePos()  {}

// ContinueStmt representa um statement continue
type ContinueStmt struct{}

func (c *ContinueStmt) stmtNode() {}
func (c *ContinueStmt) nodePos()  {}

// ============================
// STATEMENTS DE BLOCO
// ============================

// BlockStmt representa um bloco de statements
type BlockStmt struct {
	Body []Stmt
}

func (b *BlockStmt) stmtNode() {}
func (b *BlockStmt) nodePos()  {}

// ============================
// EXPRESSÕES LITERAIS
// ============================

// Identifier representa um identificador (nome de variável/função)
type Identifier struct {
	Name string
}

func (i *Identifier) exprNode() {}
func (i *Identifier) nodePos()  {}

// IntLiteral representa um literal inteiro
type IntLiteral struct {
	Value int64
}

func (i *IntLiteral) exprNode() {}
func (i *IntLiteral) nodePos()  {}

// FloatLiteral representa um literal de ponto flutuante
type FloatLiteral struct {
	Value float64
}

func (f *FloatLiteral) exprNode() {}
func (f *FloatLiteral) nodePos()  {}

// StringLiteral representa um literal de string
type StringLiteral struct {
	Value string
}

func (s *StringLiteral) exprNode() {}
func (s *StringLiteral) nodePos()  {}

// BoolLiteral representa um literal booleano
type BoolLiteral struct {
	Value bool
}

func (b *BoolLiteral) exprNode() {}
func (b *BoolLiteral) nodePos()  {}

// NullLiteral representa o literal null
type NullLiteral struct{}

func (n *NullLiteral) exprNode() {}
func (n *NullLiteral) nodePos()  {}

// ============================
// EXPRESSÕES DE OPERADORES
// ============================

// UnaryExpr representa uma expressão unária
type UnaryExpr struct {
	Op      string
	Expr    Expr
	Postfix bool
}

func (u *UnaryExpr) exprNode() {}
func (u *UnaryExpr) nodePos()  {}

// BinaryExpr representa uma expressão binária
type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) exprNode() {}
func (b *BinaryExpr) nodePos()  {}

// TernaryExpr representa uma expressão ternária (cond ? true : false)
type TernaryExpr struct {
	Cond      Expr
	TrueExpr  Expr
	FalseExpr Expr
}

func (t *TernaryExpr) exprNode() {}
func (t *TernaryExpr) nodePos()  {}

// AssignExpr representa uma expressão de atribuição
type AssignExpr struct {
	Left  Expr
	Right Expr
}

func (a *AssignExpr) exprNode() {}
func (a *AssignExpr) nodePos()  {}

// ============================
// EXPRESSÕES DE CHAMADA E ACESSO
// ============================

// CallExpr representa uma chamada de função
type CallExpr struct {
	Callee Expr
	Args   []Expr
}

func (c *CallExpr) exprNode() {}
func (c *CallExpr) nodePos()  {}

// IndexExpr representa um acesso por índice (array/map)
type IndexExpr struct {
	Array Expr
	Index Expr
}

func (i *IndexExpr) exprNode() {}
func (i *IndexExpr) nodePos()  {}

// MemberExpr representa um acesso a membro (objeto.membro)
type MemberExpr struct {
	Object Expr
	Member string
}

func (m *MemberExpr) exprNode() {}
func (m *MemberExpr) nodePos()  {}

// SelfExpr representa a referência à própria instância
type SelfExpr struct{}

func (s *SelfExpr) exprNode() {}
func (s *SelfExpr) nodePos()  {}

// ============================
// EXPRESSÕES DE COLEÇÕES
// ============================

// ArrayLiteral representa um literal de array
type ArrayLiteral struct {
	Elements []Expr
}

func (a *ArrayLiteral) exprNode() {}
func (a *ArrayLiteral) nodePos()  {}

// SetLiteral representa um literal de conjunto
type SetLiteral struct {
	Elements []Expr
}

func (s *SetLiteral) exprNode() {}
func (s *SetLiteral) nodePos()  {}

// MapLiteral representa um literal de mapa
type MapLiteral struct {
	Entries []*MapEntry
}

func (m *MapLiteral) exprNode() {}
func (m *MapLiteral) nodePos()  {}

// MapEntry representa uma entrada de mapa (chave: valor)
type MapEntry struct {
	Key   Expr
	Value Expr
}

func (m *MapEntry) nodePos() {}

// StructLiteral representa um literal de estrutura
type StructLiteral struct {
	Fields []*StructField
}

func (s *StructLiteral) exprNode() {}
func (s *StructLiteral) nodePos()  {}

// StructField representa um campo em um literal de estrutura
type StructField struct {
	Name  string
	Value Expr
}

func (s *StructField) nodePos() {}

// ============================
// EXPRESSÕES ESPECIAIS
// ============================

// FunctionExpr representa uma expressão de função (função anônima)
type FunctionExpr struct {
	Generics   []*GenericParam
	Params     []*Param
	ReturnType Type
	Body       []Stmt
}

func (f *FunctionExpr) exprNode() {}
func (f *FunctionExpr) nodePos()  {}

// ReferenceExpr representa uma expressão de referência (&var)
type ReferenceExpr struct {
	Expr Expr
}

func (r *ReferenceExpr) exprNode() {}
func (r *ReferenceExpr) nodePos()  {}

// GenericCallExpr representa uma chamada de função genérica
type GenericCallExpr struct {
	Callee   Expr
	TypeArgs []Type
	Args     []Expr
}

func (g *GenericCallExpr) exprNode() {}
func (g *GenericCallExpr) nodePos()  {}

// GenericSpecialization representa uma especialização genérica
type GenericSpecialization struct {
	Callee   Expr
	TypeArgs []Type
}

func (g *GenericSpecialization) exprNode() {}
func (g *GenericSpecialization) nodePos()  {}

// ============================
// TIPOS PRIMITIVOS E BÁSICOS
// ============================

// PrimitiveType representa um tipo primitivo (int, float, etc.)
type PrimitiveType struct {
	Name string
}

func (p *PrimitiveType) typeNode() {}
func (p *PrimitiveType) nodePos()  {}

// IdentifierType representa um tipo identificador
type IdentifierType struct {
	Name string
}

func (i *IdentifierType) typeNode() {}
func (i *IdentifierType) nodePos()  {}

// GenericType representa um tipo genérico
type GenericType struct {
	Name     string
	TypeArgs []Type
}

func (g *GenericType) typeNode() {}
func (g *GenericType) nodePos()  {}

// ============================
// TIPOS MODIFICADOS
// ============================

// ArrayType representa um tipo de array
type ArrayType struct {
	ElementType Type
	Size        Expr
}

func (a *ArrayType) typeNode() {}
func (a *ArrayType) nodePos()  {}

// NullableType representa um tipo anulável (T?)
type NullableType struct {
	BaseType Type
}

func (n *NullableType) typeNode() {}
func (n *NullableType) nodePos()  {}

// PointerType representa um tipo ponteiro (T*)
type PointerType struct {
	BaseType Type
}

func (p *PointerType) typeNode() {}
func (p *PointerType) nodePos()  {}

// SetType representa um tipo conjunto (Set<T>)
type SetType struct {
	ElementType Type
}

func (s *SetType) typeNode() {}
func (s *SetType) nodePos()  {}

// MapType representa um tipo mapa (Map<K,V>)
type MapType struct {
	KeyType   Type
	ValueType Type
}

func (m *MapType) typeNode() {}
func (m *MapType) nodePos()  {}

// UnionType representa um tipo união (T1 | T2 | T3)
type UnionType struct {
	Types []Type
}

func (u *UnionType) typeNode() {}
func (u *UnionType) nodePos()  {}

// StructType representa um tipo estrutura
type StructType struct {
	Fields []*FieldDecl
}

func (s *StructType) typeNode() {}
func (s *StructType) nodePos()  {}

// FunctionType representa um tipo função
type FunctionType struct {
	Params     []Type
	ReturnType Type
}

func (f *FunctionType) typeNode() {}
func (f *FunctionType) nodePos()  {}
