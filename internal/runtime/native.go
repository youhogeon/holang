package runtime

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
	"unicode/utf8"
)

var NativeFunctions = []*ObjNativeFunction{
	NewObjNativeFunction("print", 1, nativePrint),
	NewObjNativeFunction("input", 1, nativeInput),
	NewObjNativeFunction("clock", 0, nativeClock),

	NewObjNativeFunction("str", 1, nativeToString),
	NewObjNativeFunction("int", 1, nativeToInt),
	NewObjNativeFunction("float", 1, nativeToFloat),

	NewObjNativeFunction("randInt", 1, nativeRandInt),
	NewObjNativeFunction("sleep", 1, nativeSleep),
	NewObjNativeFunction("clear", 0, nativeClear),

	NewObjNativeFunction("strLen", 1, nativeStrLen),
	NewObjNativeFunction("subString", 3, nativeSubString),

	NewObjNativeFunction("getCh", 0, nativeGetCh),
}

type NativeFunctionType func(args ...Value) (Value, error)

type ObjNativeFunction struct {
	Name     string
	Arity    int
	Function NativeFunctionType
}

func NewObjNativeFunction(name string, arity int, function NativeFunctionType) *ObjNativeFunction {
	return &ObjNativeFunction{
		Name:     name,
		Arity:    arity,
		Function: function,
	}
}

func (f *ObjNativeFunction) ObjectType() ObjectType {
	return OBJ_NATIVE_FUNCTION
}

func (f *ObjNativeFunction) String() string {
	return "<native fun " + f.Name + ">"
}

func (f *ObjNativeFunction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + f.String() + `"`), nil
}

func (f *ObjNativeFunction) Serialize(w io.Writer) error {
	if _, err := w.Write([]byte{byte(f.ObjectType())}); err != nil {
		return err
	}

	if err := writeString(w, f.Name); err != nil {
		return err
	}

	if err := writeInt(w, int64(f.Arity)); err != nil {
		return err
	}

	return nil
}

func (f *ObjNativeFunction) Deserialize(data []byte) (any, []byte, error) {
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

	var nativeFunc NativeFunctionType
	for _, nf := range NativeFunctions {
		if nf.Name == name && nf.Arity == int(arity) {
			nativeFunc = nf.Function
			break
		}
	}
	if nativeFunc == nil {
		return nil, nil, errors.New("unknown native function: " + name)
	}

	return &ObjNativeFunction{
		Name:     name,
		Arity:    int(arity),
		Function: nativeFunc,
	}, data, nil
}

// ================================================================
// Implementation of native functions
// ================================================================

func nativePrint(args ...Value) (Value, error) {
	fmt.Println(args[0])

	return nil, nil
}

func nativeInput(args ...Value) (Value, error) {
	var input string
	fmt.Print(args[0])
	_, err := fmt.Scanln(&input)

	if err != nil {
		return nil, errors.New("failed to read input")
	}

	return input, nil
}

func nativeClock(args ...Value) (Value, error) {
	return int64(time.Now().UnixNano() / 1e6), nil
}

func nativeToString(args ...Value) (Value, error) {
	return fmt.Sprint(args[0]), nil
}

func nativeToInt(args ...Value) (Value, error) {
	return strconv.ParseInt(fmt.Sprint(args[0]), 10, 64)
}

func nativeToFloat(args ...Value) (Value, error) {
	return strconv.ParseFloat(fmt.Sprint(args[0]), 64)
}

func nativeRandInt(args ...Value) (Value, error) {
	var n int64
	switch v := args[0].(type) {
	case int64:
		n = v
	case float64:
		n = int64(v)
	default:
		// try parsing string rep
		parsed, err := strconv.ParseInt(fmt.Sprint(args[0]), 10, 64)
		if err != nil {
			return nil, errors.New("randInt argument must be a number")
		}
		n = parsed
	}
	if n <= 0 {
		return nil, errors.New("randInt argument must be > 0")
	}
	return int64(rand.Int63n(n)), nil
}

func nativeSleep(args ...Value) (Value, error) {
	var ms int64

	switch v := args[0].(type) {
	case int64:
		ms = v
	case float64:
		ms = int64(v)
	default:
		parsed, err := strconv.ParseInt(fmt.Sprint(args[0]), 10, 64)
		if err != nil {
			return nil, errors.New("sleep argument must be a number (milliseconds)")
		}
		ms = parsed
	}
	if ms < 0 {
		return nil, errors.New("sleep argument must be >= 0")
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return nil, nil
}

func nativeClear(args ...Value) (Value, error) {
	fmt.Print("\033[2J\033[H")
	return nil, nil
}

func nativeStrLen(args ...Value) (Value, error) {
	s := fmt.Sprint(args[0])
	return int64(utf8.RuneCountInString(s)), nil
}

func nativeSubString(args ...Value) (Value, error) {
	s := fmt.Sprint(args[0])
	start, ok1 := toInt(args[1])
	end, ok2 := toInt(args[2])
	if !ok1 || !ok2 {
		return nil, errors.New("substring indices must be numbers")
	}
	runes := []rune(s)
	if start < 0 || end < 0 || start > end || int(end) > len(runes) {
		return nil, errors.New("substring index out of range")
	}
	return string(runes[start:end]), nil
}

func nativeGetCh(args ...Value) (Value, error) {
	reader := bufio.NewReader(os.Stdin)
	r, _, err := reader.ReadRune()
	if err != nil {
		return nil, errors.New("failed to read char")
	}
	// If user pressed Enter first, try next rune
	if r == '\n' || r == '\r' {
		r, _, err = reader.ReadRune()
		if err != nil {
			return nil, errors.New("failed to read char")
		}
	}
	return string(r), nil
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int64:
		if n < 0 || n > int64(int(n)) { // overflow check
			return 0, false
		}
		return int(n), true
	case float64:
		if n < 0 || n > float64(int(n)) { // not whole or overflow
			return 0, false
		}
		return int(n), true
	default:
		parsed, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		if err != nil || parsed < 0 || parsed > int64(int(parsed)) {
			return 0, false
		}
		return int(parsed), true
	}
}
