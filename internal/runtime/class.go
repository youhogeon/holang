package runtime

import "io"

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

func (c *ObjClass) Serialize(w io.Writer) error {
	if _, err := w.Write([]byte{byte(c.ObjectType())}); err != nil {
		return err
	}

	if err := writeString(w, c.Name); err != nil {
		return err
	}

	return nil
}

func (c *ObjClass) Deserialize(data []byte) (any, []byte, error) {
	if data[0] != byte(c.ObjectType()) {
		return nil, nil, io.ErrUnexpectedEOF
	}

	name, data, err := readString(data[1:])
	if err != nil {
		return nil, nil, err
	}

	return &ObjClass{
		Name: name,
	}, data, nil
}
