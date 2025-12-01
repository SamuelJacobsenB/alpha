package parser

// implementações de nodePos para Types
func (*PrimitiveType) nodePos()  {}
func (*ArrayType) nodePos()      {}
func (*IdentifierType) nodePos() {}
func (*GenericParam) nodePos()   {}
func (*FunctionType) nodePos()   {}
func (*NullableType) nodePos()   {}
func (*PointerType) nodePos()    {}
func (*SetType) nodePos()        {}
func (*MapType) nodePos()        {}
func (*UnionType) nodePos()      {}
func (*StructType) nodePos()     {}

// implementações de nodePos para Stmts
func (*VarDecl) nodePos()               {}
func (*ConstDecl) nodePos()             {}
func (*ExprStmt) nodePos()              {}
func (*IfStmt) nodePos()                {}
func (*WhileStmt) nodePos()             {}
func (*DoWhileStmt) nodePos()           {}
func (*ForStmt) nodePos()               {}
func (*ForInStmt) nodePos()             {}
func (*SwitchStmt) nodePos()            {}
func (*CaseClause) nodePos()            {}
func (*FunctionDecl) nodePos()          {}
func (*FunctionExpr) nodePos()          {}
func (*GenericSpecialization) nodePos() {}
func (*ReturnStmt) nodePos()            {}
func (*BlockStmt) nodePos()             {}
func (*ClassDecl) nodePos()             {}
func (*FieldDecl) nodePos()             {}
func (*ConstructorDecl) nodePos()       {}
func (*MethodDecl) nodePos()            {}
func (*TypeDecl) nodePos()              {}

// implementações de nodePos para Params
func (*Param) nodePos() {}

// implementações de nodePos para Exprs
func (*Identifier) nodePos()    {}
func (*IntLiteral) nodePos()    {}
func (*FloatLiteral) nodePos()  {}
func (*StringLiteral) nodePos() {}
func (*BoolLiteral) nodePos()   {}
func (*NullLiteral) nodePos()   {}
func (*UnaryExpr) nodePos()     {}
func (*BinaryExpr) nodePos()    {}
func (*TernaryExpr) nodePos()   {}
func (*CallExpr) nodePos()      {}
func (*AssignExpr) nodePos()    {}
func (*ArrayLiteral) nodePos()  {}
func (*IndexExpr) nodePos()     {}
func (*SetLiteral) nodePos()    {}
func (*MapLiteral) nodePos()    {}
func (*MapEntry) nodePos()      {}
func (*ReferenceExpr) nodePos() {}
func (*NewExpr) nodePos()       {}
func (*MemberExpr) nodePos()    {}
func (*ThisExpr) nodePos()      {}
func (*StructLiteral) nodePos() {}
func (*StructField) nodePos()   {}
