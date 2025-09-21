package runtime

import (
	"io"
	"math"
)

type serializable interface {
	Serialize(w io.Writer) error
	Deserialize(data []byte) (any, []byte, error)
}

func writeString(w io.Writer, s string) error {
	err := writeInt(w, int64(len(s)))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}

func readString(data []byte) (string, []byte, error) {
	if len(data) < 4 {
		return "", data, io.EOF
	}

	strLen, data, err := readInt(data)
	if err != nil {
		return "", data, err
	}

	if len(data) < int(strLen) {
		return "", data, io.EOF
	}

	s := string(data[:strLen])
	data = data[strLen:]

	return s, data, nil
}

func writeInt(w io.Writer, n int64) error {
	b := []byte{
		byte(n >> 56),
		byte(n >> 48),
		byte(n >> 40),
		byte(n >> 32),
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	}
	_, err := w.Write(b)
	return err
}

func readInt(data []byte) (int64, []byte, error) {
	if len(data) < 8 {
		return 0, data, io.EOF
	}

	n := int64(data[0])<<56 |
		int64(data[1])<<48 |
		int64(data[2])<<40 |
		int64(data[3])<<32 |
		int64(data[4])<<24 |
		int64(data[5])<<16 |
		int64(data[6])<<8 |
		int64(data[7])

	data = data[8:]

	return n, data, nil
}

func writeFloat(w io.Writer, f float64) error {
	bits := math.Float64bits(f)
	return writeInt(w, int64(bits))
}

func readFloat(data []byte) (float64, []byte, error) {
	if len(data) < 8 {
		return 0, data, io.EOF
	}

	bits, data, err := readInt(data)
	if err != nil {
		return 0, data, err
	}

	f := math.Float64frombits(uint64(bits))
	return f, data, nil
}
