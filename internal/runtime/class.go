package runtime

type ObjClass struct {
	Name string
}

func NewObjClass(name string) *ObjClass {
	return &ObjClass{
		Name: name,
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
