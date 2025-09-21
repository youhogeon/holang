package runtime

import "io"

type Value any

type valueType byte

const (
	VALUE_TYPE_NIL valueType = iota
	VALUE_TYPE_BOOL
	VALUE_TYPE_INT
	VALUE_TYPE_FLOAT
	VALUE_TYPE_STRING
	VALUE_TYPE_OBJ
)

func SerializeValue(w io.Writer, v Value) error {
	switch val := v.(type) {
	case nil:
		if _, err := w.Write([]byte{byte(VALUE_TYPE_NIL)}); err != nil {
			return err
		}
	case bool:
		if _, err := w.Write([]byte{byte(VALUE_TYPE_BOOL)}); err != nil {
			return err
		}

		var b byte = 0
		if val {
			b = 1
		}

		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}
	case int64:
		if _, err := w.Write([]byte{byte(VALUE_TYPE_INT)}); err != nil {
			return err
		}

		if err := writeInt(w, val); err != nil {
			return err
		}
	case float64:
		if _, err := w.Write([]byte{byte(VALUE_TYPE_FLOAT)}); err != nil {
			return err
		}

		if err := writeFloat(w, val); err != nil {
			return err
		}
	case string:
		if _, err := w.Write([]byte{byte(VALUE_TYPE_STRING)}); err != nil {
			return err
		}

		if err := writeString(w, val); err != nil {
			return err
		}
	case SerializableObject:
		if _, err := w.Write([]byte{byte(VALUE_TYPE_OBJ)}); err != nil {
			return err
		}

		if err := val.Serialize(w); err != nil {
			return err
		}
	default:
		return io.ErrUnexpectedEOF
	}

	return nil
}

func DeserializeValue(data []byte) (any, []byte, error) {
	if len(data) < 1 {
		return nil, nil, io.EOF
	}

	vType := valueType(data[0])
	data = data[1:]

	switch vType {
	case VALUE_TYPE_NIL:
		return nil, data, nil
	case VALUE_TYPE_BOOL:
		if len(data) < 1 {
			return nil, nil, io.EOF
		}

		b := data[0] != 0
		data = data[1:]

		return b, data, nil
	case VALUE_TYPE_INT:
		n, data, err := readInt(data)
		if err != nil {
			return nil, nil, err
		}

		return n, data, nil
	case VALUE_TYPE_FLOAT:
		f, data, err := readFloat(data)
		if err != nil {
			return nil, nil, err
		}

		return f, data, nil
	case VALUE_TYPE_STRING:
		s, data, err := readString(data)
		if err != nil {
			return nil, nil, err
		}

		return s, data, nil
	case VALUE_TYPE_OBJ:
		if len(data) < 1 {
			return nil, nil, io.EOF
		}

		var obj SerializableObject
		switch ObjectType(data[0]) {
		case OBJ_FUNCTION:
			obj = &ObjFunction{}
		case OBJ_NATIVE_FUNCTION:
			obj = &ObjNativeFunction{}
		default:
			return nil, nil, io.ErrUnexpectedEOF
		}

		o, data, err := obj.Deserialize(data)
		if err != nil {
			return nil, nil, err
		}

		return o, data, nil
	default:
		return nil, nil, io.ErrUnexpectedEOF
	}

}
