package runtime

import (
	"internal/util/log"
	"io"
)

type ObjFunctionType byte

const (
	FUNCTION_TYPE_FUN ObjFunctionType = iota
	FUNCTION_TYPE_SCRIPT
)

type ObjFunction struct {
	Name         string
	Arity        int
	Type         ObjFunctionType
	UpvalueCount int

	Chunk *Chunk
}

func NewObjFunction(name string, arity int, ftype ObjFunctionType) *ObjFunction {
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

func (f *ObjFunction) Serialize(w io.Writer) error {
	if _, err := w.Write([]byte{byte(f.ObjectType())}); err != nil {
		return err
	}

	if err := writeString(w, f.Name); err != nil {
		return err
	}

	if err := writeInt(w, int64(f.Arity)); err != nil {
		return err
	}

	if _, err := w.Write([]byte{byte(f.Type)}); err != nil {
		return err
	}

	if err := writeInt(w, int64(f.UpvalueCount)); err != nil {
		return err
	}

	return f.Chunk.Serialize(w)
}

func (f *ObjFunction) Deserialize(data []byte) (any, []byte, error) {
	if data[0] != byte(f.ObjectType()) {
		return nil, nil, io.ErrUnexpectedEOF
	}

	name, data, err := readString(data[1:])
	if err != nil {
		return nil, nil, err
	}

	arity, data, err := readInt(data)
	if err != nil {
		return nil, nil, err
	}

	fnType := ObjFunctionType(data[0])

	upvalueCount, data, err := readInt(data[1:])
	if err != nil {
		return nil, nil, err
	}

	chunk, data, err := (&Chunk{}).Deserialize(data)
	if err != nil {
		return nil, nil, err
	}

	return &ObjFunction{
		Name:         name,
		Arity:        int(arity),
		Type:         fnType,
		UpvalueCount: int(upvalueCount),
		Chunk:        chunk.(*Chunk),
	}, data, nil
}
