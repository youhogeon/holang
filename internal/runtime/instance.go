package runtime

type ObjInstance struct {
	Class  *ObjClass
	Fields map[string]Value
}

func NewObjInstance(class *ObjClass) *ObjInstance {
	return &ObjInstance{
		Class:  class,
		Fields: make(map[string]Value),
	}
}

func (i *ObjInstance) ObjectType() ObjectType {
	return OBJ_NATIVE_FUNCTION
}

func (i *ObjInstance) String() string {
	return "<instance " + i.Class.Name + ">"
}

func (i *ObjInstance) MarshalJSON() ([]byte, error) {
	return []byte(`"` + i.String() + `"`), nil
}
