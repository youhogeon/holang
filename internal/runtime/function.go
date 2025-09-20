package runtime

import (
	"internal/util/log"
)

type FunctionType byte

const (
	FUNCTION_TYPE_FUN FunctionType = iota
	FUNCTION_TYPE_SCRIPT
	FUNCTION_TYPE_NATIVE
)

// ================================================================
// ObjFunction
// ================================================================

type ObjFunction struct {
	Name         string
	Arity        int
	Type         FunctionType
	UpvalueCount int

	Chunk *Chunk
}

func NewObjFunction(name string, arity int, ftype FunctionType) *ObjFunction {
	chunk := NewChunk()

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

func (f *ObjFunction) String() string {
	return "<fun " + f.Name + ">"
}

func (f *ObjFunction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + f.String() + `"`), nil
}

func (f *ObjFunction) Disassemble() {
	log.Debug("================================================================")
	log.Debug("Disassemble ObjFunction", log.S("function", f.String()))
	f.Chunk.Disassemble()

	fns := make([]*ObjFunction, 0)

	for i := range f.Chunk.CountConstants() {
		constant := f.Chunk.GetConstant(int64(i))

		log.Debug("Constant", log.I("index", i), log.A("value", constant))

		if v, ok := constant.(*ObjFunction); ok {
			fns = append(fns, v)
		}
	}

	for _, fn := range fns {
		fn.Disassemble()
	}
}

// ================================================================
// ObjNativeFunction
// ================================================================

type NativeFunctionType func(args ...Value) (Value, error)

type ObjNativeFunction struct {
	Name     string
	Arity    int
	Function NativeFunctionType
}

func NewObjNativeFunction(name string, arity int, function NativeFunctionType) *ObjNativeFunction {
	return &ObjNativeFunction{
		Name:     name,
		Arity:    arity,
		Function: function,
	}
}

func (f *ObjNativeFunction) ObjectType() ObjectType {
	return OBJ_NATIVE_FUNCTION
}

func (f *ObjNativeFunction) String() string {
	return "<native fun " + f.Name + ">"
}

func (f *ObjNativeFunction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + f.String() + `"`), nil
}
