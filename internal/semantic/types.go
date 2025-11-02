package semantic

import "fmt"

// Type representa um tipo na linguagem
type Type interface {
	String() string
	Equals(Type) bool
}

// Tipos primitivos concretos
type IntType struct{}
type FloatType struct{}
type StringType struct{}
type BoolType struct{}
type NullType struct{}
type AnyType struct{} // usado como fallback / dinâmico

func (IntType) String() string    { return "int" }
func (FloatType) String() string  { return "float" }
func (StringType) String() string { return "string" }
func (BoolType) String() string   { return "bool" }
func (NullType) String() string   { return "null" }
func (AnyType) String() string    { return "any" }

func (IntType) Equals(t Type) bool    { _, ok := t.(IntType); return ok }
func (FloatType) Equals(t Type) bool  { _, ok := t.(FloatType); return ok }
func (StringType) Equals(t Type) bool { _, ok := t.(StringType); return ok }
func (BoolType) Equals(t Type) bool   { _, ok := t.(BoolType); return ok }
func (NullType) Equals(t Type) bool   { _, ok := t.(NullType); return ok }
func (AnyType) Equals(t Type) bool    { _, ok := t.(AnyType); return ok }

// FuncType representa tipo de função (simples)
type FuncType struct {
	Params []Type
	Ret    Type
}

func (f FuncType) String() string {
	ps := ""
	for i, p := range f.Params {
		if i > 0 {
			ps += ", "
		}
		ps += p.String()
	}
	return fmt.Sprintf("fn(%s) -> %s", ps, f.Ret.String())
}

func (f FuncType) Equals(t Type) bool {
	o, ok := t.(FuncType)
	if !ok {
		return false
	}
	if len(o.Params) != len(f.Params) {
		return false
	}
	for i := range f.Params {
		if !f.Params[i].Equals(o.Params[i]) {
			return false
		}
	}
	return f.Ret.Equals(o.Ret)
}
