package parser

// implementações de nodePos para Stmts
func (*PrimitiveType) nodePos()         {}
func (*ArrayType) nodePos()             {}
func (*VarDecl) nodePos()               {}
func (*ConstDecl) nodePos()             {}
func (*ExprStmt) nodePos()              {}
func (*IfStmt) nodePos()                {}
func (*WhileStmt) nodePos()             {}
func (*DoWhileStmt) nodePos()           {}
func (*ForStmt) nodePos()               {}
func (*ForInStmt) nodePos()             {}
func (*FunctionDecl) nodePos()          {}
func (*IdentifierType) nodePos()        {}
func (*GenericParam) nodePos()          {}
func (*Param) nodePos()                 {}
func (*FunctionType) nodePos()          {}
func (*FunctionExpr) nodePos()          {}
func (*GenericSpecialization) nodePos() {}
func (*ReturnStmt) nodePos()            {}
func (*BlockStmt) nodePos()             {}

// implementações de nodePos para Exprs
func (*Identifier) nodePos()    {}
func (*IntLiteral) nodePos()    {}
func (*FloatLiteral) nodePos()  {}
func (*StringLiteral) nodePos() {}
func (*BoolLiteral) nodePos()   {}
func (*NullLiteral) nodePos()   {}
func (*UnaryExpr) nodePos()     {}
func (*BinaryExpr) nodePos()    {}
func (*CallExpr) nodePos()      {}
func (*AssignExpr) nodePos()    {}
func (*ArrayLiteral) nodePos()  {}
func (*IndexExpr) nodePos()     {}

func (*NullableType) nodePos()  {}
func (*PointerType) nodePos()   {}
func (*SetType) nodePos()       {}
func (*MapType) nodePos()       {}
func (*UnionType) nodePos()     {}
func (*SetLiteral) nodePos()    {}
func (*MapLiteral) nodePos()    {}
func (*MapEntry) nodePos()      {}
func (*ReferenceExpr) nodePos() {}
