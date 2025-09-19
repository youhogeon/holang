package runtime

import "internal/bytecode"

type FunctionType byte

const (
	FUNCTION_TYPE_FUN FunctionType = iota
	FUNCTION_TYPE_SCRIPT
)

type ObjFunction struct {
	Name  string
	Arity int
	Type  FunctionType
	Chunk *bytecode.Chunk
}

func NewObjFunction(name string, arity int, ftype FunctionType) *ObjFunction {
	chunk := bytecode.NewChunk()

	return &ObjFunction{
		Name:  name,
		Arity: arity,
		Type:  ftype,
		Chunk: chunk,
	}
}

func (of *ObjFunction) ObjectType() ObjectType {
	return OBJ_FUNCTION
}

func (of *ObjFunction) String() string {
	return "<fun " + of.Name + ">"
}
