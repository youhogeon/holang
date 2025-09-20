package runtime

type ObjClosure struct {
	Function *ObjFunction
}

func NewObjClosure(function *ObjFunction) *ObjClosure {
	return &ObjClosure{
		Function: function,
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
