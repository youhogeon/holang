package serialize

import (
	"bytes"
	"errors"
	"internal/runtime"
)

type Program struct {
	Signature [8]byte
	Version   [8]byte
	Flag      [8]byte

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

	_, err = buffer.Write(p.Flag[:])
	if err != nil {
		return err
	}

	return p.RootFunction.Serialize(buffer)
}

func (p *Program) validate(availableVersion []string) error {
	if p.Signature != [8]byte{'H', 'O', 'L', 'A', 'N', 'G', 0, 0} {
		return errors.New("invalid signature")
	}

	if !p.validateVersion(availableVersion) {
		return errors.New("incompatible version")
	}

	return nil
}

func (p *Program) validateVersion(availableVersion []string) bool {
	version := string(bytes.Trim(p.Version[:], "\x00 "))
	for _, v := range availableVersion {
		if v == version {
			return true
		}
	}

	return false
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

func Deserialize(data []byte, availableVersion []string) (*Program, error) {
	program := &Program{
		Signature: [8]byte(data[:8]),
		Version:   [8]byte(data[8:16]),
		Flag:      [8]byte(data[16:24]),
	}

	if err := program.validate(availableVersion); err != nil {
		return nil, err
	}

	fn, _, err := (&runtime.ObjFunction{}).Deserialize(data[24:])

	if err != nil {
		return nil, err
	}

	program.RootFunction = fn.(*runtime.ObjFunction)

	return program, nil
}
