package serialize

import (
	"bytes"
	"errors"
	"internal/runtime"
)

type Program struct {
	Signature [8]byte
	Version   [8]byte

	RootFunction *runtime.ObjFunction
}

func (p *Program) Serialize(buffer *bytes.Buffer) error {
	_, err := buffer.Write(p.Signature[:])
	if err != nil {
		return err
	}

	_, err = buffer.Write(p.Version[:])
	if err != nil {
		return err
	}

	return p.RootFunction.Serialize(buffer)
}

func Serialize(fn *runtime.ObjFunction, version string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	version = version + "        "

	program := &Program{
		Signature:    [8]byte{'H', 'O', 'L', 'A', 'N', 'G', 0, 0},
		Version:      [8]byte([]byte(version[:8])),
		RootFunction: fn,
	}

	err := program.Serialize(buffer)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func Deserialize(data []byte) (*Program, error) {
	fn, _, err := (&runtime.ObjFunction{}).Deserialize(data[16:])

	if err != nil {
		return nil, err
	}

	program := &Program{
		Signature:    [8]byte(data[:8]),
		Version:      [8]byte(data[8:16]),
		RootFunction: fn.(*runtime.ObjFunction),
	}

	if program.Signature != [8]byte{'H', 'O', 'L', 'A', 'N', 'G', 0, 0} {
		return nil, errors.New("invalid signature")
	}

	return program, nil
}
