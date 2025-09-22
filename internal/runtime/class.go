package runtime

type ObjClass struct {
	Name    string
	Methods map[string]*ObjClosure
}

func NewObjClass(name string) *ObjClass {
	return &ObjClass{
		Name:    name,
		Methods: make(map[string]*ObjClosure),
	}
}

func (c *ObjClass) ObjectType() ObjectType {
	return OBJ_NATIVE_FUNCTION
}

func (c *ObjClass) String() string {
	return "<class " + c.Name + ">"
}

func (c *ObjClass) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.String() + `"`), nil
}
