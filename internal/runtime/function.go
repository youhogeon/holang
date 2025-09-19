package runtime

import (
	"internal/bytecode"
	"internal/util/log"
)

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

func (f *ObjFunction) ObjectType() ObjectType {
	return OBJ_FUNCTION
}

func (f ObjFunction) String() string {
	return "<fun " + f.Name + ">"
}

func (f *ObjFunction) Disassemble() {
	log.Debug("Disassemble ObjFunction", log.S("function", f.String()))
	f.Chunk.Disassemble()

	fns := make([]*ObjFunction, 0)

	for i := range f.Chunk.CountConstants() {
		constant := f.Chunk.GetConstant(int64(i))

		log.Debug("Constant", log.I("index", i), log.A("value", constant))

		if v, ok := constant.(ObjFunction); ok {
			fns = append(fns, &v)
		}
	}

	for _, fn := range fns {
		fn.Disassemble()
	}
}
