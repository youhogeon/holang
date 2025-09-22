package runtime

type ObjBoundMethod struct {
	Method   *ObjClosure
	Receiver Value
}

func NewObjBoundMethod(method *ObjClosure, receiver Value) *ObjBoundMethod {
	return &ObjBoundMethod{
		Method:   method,
		Receiver: receiver,
	}
}

func (m *ObjBoundMethod) ObjectType() ObjectType {
	return OBJ_BOUND_METHOD
}

func (m *ObjBoundMethod) String() string {
	return "<method " + m.Method.Function.Name + ">"
}

func (m *ObjBoundMethod) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}
