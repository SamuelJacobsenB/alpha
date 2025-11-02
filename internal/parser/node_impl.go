package parser

// implementações de nodePos para Stmts
func (*VarDecl) nodePos()    {}
func (*ConstDecl) nodePos()  {}
func (*ExprStmt) nodePos()   {}
func (*IfStmt) nodePos()     {}
func (*WhileStmt) nodePos()  {}
func (*ForStmt) nodePos()    {}
func (*ReturnStmt) nodePos() {}
func (*BlockStmt) nodePos()  {}

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
