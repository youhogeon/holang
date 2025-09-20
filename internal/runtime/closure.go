package runtime

import "internal/bytecode"

type ObjClosure struct {
	Function *ObjFunction
	Upvalues []*ObjUpvalue
}

func NewObjClosure(function *ObjFunction) *ObjClosure {
	return &ObjClosure{
		Function: function,
		Upvalues: make([]*ObjUpvalue, function.UpvalueCount),
	}
}

func (c *ObjClosure) ObjectType() ObjectType {
	return OBJ_CLOSURE
}

func (c *ObjClosure) String() string {
	return c.Function.String()
}

func (c *ObjClosure) MarshalJSON() ([]byte, error) {
	return c.Function.MarshalJSON()
}

type ObjUpvalue struct {
	Stack *[]bytecode.Value
	Index int

	Closed   bytecode.Value
	IsClosed bool
}

func (uv *ObjUpvalue) ObjectType() ObjectType {
	return OBJ_UPVALUE
}

func (uv *ObjUpvalue) Get() bytecode.Value {
	if uv.IsClosed {
		return uv.Closed
	}

	return (*uv.Stack)[uv.Index]
}

func (uv *ObjUpvalue) Set(v bytecode.Value) {
	if uv.IsClosed {
		uv.Closed = v

		return
	}

	(*uv.Stack)[uv.Index] = v
}

func (uv *ObjUpvalue) Close() {
	if uv.IsClosed {
		return
	}

	uv.Closed = (*uv.Stack)[uv.Index]
	uv.IsClosed = true
	uv.Stack = nil
}
