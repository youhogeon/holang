package interpreter

import (
	"bufio"
	"fmt"
	"internal/ast"
	"math/rand"
	"os"
	"strconv"
	"time"
	"unicode/utf8"
)

type Callable interface {
	Arity() int
	Call(interpreter *Interpreter, arguments []any) (any, error)
}

type Function struct {
	declaration   *ast.Function
	clousure      *Environment
	isInitializer bool
}

func (f *Function) Arity() int {
	return len(f.declaration.Params)
}

func (f *Function) Call(interpreter *Interpreter, arguments []any) (any, error) {
	env := NewEnvironment(f.clousure)

	for i, param := range f.declaration.Params {
		env.Define(param.Lexeme, arguments[i])
	}

	err := interpreter.executeBlock(f.declaration.Body, env)

	if err != nil {
		if returnSig, ok := err.(*returnSignal); ok {
			if f.isInitializer {
				thisValue, _ := f.clousure.Get("this")
				return thisValue, nil
			}

			return returnSig.value, nil
		}

		return nil, err
	}

	if f.isInitializer {
		thisValue, _ := f.clousure.Get("this")
		return thisValue, nil
	}

	return nil, nil
}

func (f *Function) bind(instance *Instance) *Function {
	env := NewEnvironment(f.clousure)
	env.Define("this", instance)

	return &Function{
		declaration:   f.declaration,
		clousure:      env,
		isInitializer: f.isInitializer,
	}
}

type Class struct {
	name       string
	methods    map[string]*Function
	superclass *Class
}

func (f *Class) Arity() int {
	if initializer := f.findMethod("init"); initializer != nil {
		return initializer.Arity()
	}

	return 0
}

func (f *Class) Call(interpreter *Interpreter, arguments []any) (any, error) {
	instance := &Instance{
		class:  f,
		fields: make(map[string]any),
	}

	if initializer := f.findMethod("init"); initializer != nil {
		_, err := initializer.bind(instance).Call(interpreter, arguments)

		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}

func (f *Class) findMethod(name string) *Function {
	if method, ok := f.methods[name]; ok {
		return method
	}

	if f.superclass != nil {
		return f.superclass.findMethod(name)
	}

	return nil
}

type Instance struct {
	class  *Class
	fields map[string]any
}

func (i *Instance) get(name string) (any, error) {
	if value, ok := i.fields[name]; ok {
		return value, nil
	}

	if method := i.class.findMethod(name); method != nil {
		return method.bind(i), nil
	}

	return nil, NewRuntimeErrorWithLog("undefined property: " + name)
}

func (i *Instance) set(name string, value any) {
	i.fields[name] = value
}

// ----------------------------------------------------------------
// Built-in functions
// ----------------------------------------------------------------

type BuiltInFnPrint struct{}

func (b *BuiltInFnPrint) Arity() int {
	return 1
}

func (b *BuiltInFnPrint) Call(interpreter *Interpreter, arguments []any) (any, error) {
	fmt.Println(arguments[0])

	return nil, nil
}

type BuiltInFnInput struct{}

func (b *BuiltInFnInput) Arity() int {
	return 1
}

func (b *BuiltInFnInput) Call(interpreter *Interpreter, arguments []any) (any, error) {
	var input string
	fmt.Print(arguments[0])
	_, err := fmt.Scanln(&input)

	if err != nil {
		return nil, NewRuntimeErrorWithLog("failed to read input")
	}

	return input, nil
}

type BuiltInFnClock struct{}

func (b *BuiltInFnClock) Arity() int {
	return 0
}

func (b *BuiltInFnClock) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return int64(time.Now().UnixNano() / 1e9), nil
}

type BuiltInFnToString struct{}

func (b *BuiltInFnToString) Arity() int {
	return 1
}

func (b *BuiltInFnToString) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return fmt.Sprint(arguments[0]), nil
}

type BuiltInFnToInt struct{}

func (b *BuiltInFnToInt) Arity() int {
	return 1
}

func (b *BuiltInFnToInt) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return strconv.ParseInt(fmt.Sprint(arguments[0]), 10, 64)
}

type BuiltInFnToFloat struct{}

func (b *BuiltInFnToFloat) Arity() int {
	return 1
}

func (b *BuiltInFnToFloat) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return strconv.ParseFloat(fmt.Sprint(arguments[0]), 64)
}

type BuiltInFnRand struct{}

func (b *BuiltInFnRand) Arity() int { return 0 }
func (b *BuiltInFnRand) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return rand.Float64(), nil
}

type BuiltInFnRandInt struct{}

func (b *BuiltInFnRandInt) Arity() int { return 1 }
func (b *BuiltInFnRandInt) Call(interpreter *Interpreter, arguments []any) (any, error) {
	// Accept int64 or float64; convert via fmt then ParseInt fallback
	var n int64
	switch v := arguments[0].(type) {
	case int64:
		n = v
	case float64:
		n = int64(v)
	default:
		// try parsing string rep
		parsed, err := strconv.ParseInt(fmt.Sprint(arguments[0]), 10, 64)
		if err != nil {
			return nil, NewRuntimeErrorWithLog("randInt argument must be a number")
		}
		n = parsed
	}
	if n <= 0 {
		return nil, NewRuntimeErrorWithLog("randInt argument must be > 0")
	}
	return int64(rand.Int63n(n)), nil
}

type BuiltInFnSleep struct{}

func (b *BuiltInFnSleep) Arity() int { return 1 }
func (b *BuiltInFnSleep) Call(interpreter *Interpreter, arguments []any) (any, error) {
	var ms int64
	switch v := arguments[0].(type) {
	case int64:
		ms = v
	case float64:
		ms = int64(v)
	default:
		parsed, err := strconv.ParseInt(fmt.Sprint(arguments[0]), 10, 64)
		if err != nil {
			return nil, NewRuntimeErrorWithLog("sleep argument must be a number (milliseconds)")
		}
		ms = parsed
	}
	if ms < 0 {
		return nil, NewRuntimeErrorWithLog("sleep argument must be >= 0")
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return nil, nil
}

type BuiltInFnClear struct{}

func (b *BuiltInFnClear) Arity() int { return 0 }
func (b *BuiltInFnClear) Call(interpreter *Interpreter, arguments []any) (any, error) {
	// ANSI escape: clear screen & move cursor home
	fmt.Print("\033[2J\033[H")
	return nil, nil
}

type BuiltInFnStrLen struct{}

func (b *BuiltInFnStrLen) Arity() int { return 1 }
func (b *BuiltInFnStrLen) Call(interpreter *Interpreter, arguments []any) (any, error) {
	s := fmt.Sprint(arguments[0])
	return int64(utf8.RuneCountInString(s)), nil
}

type BuiltInFnSubstring struct{}

func (b *BuiltInFnSubstring) Arity() int { return 3 }
func (b *BuiltInFnSubstring) Call(interpreter *Interpreter, arguments []any) (any, error) {
	s := fmt.Sprint(arguments[0])
	start, ok1 := toInt(arguments[1])
	end, ok2 := toInt(arguments[2])
	if !ok1 || !ok2 {
		return nil, NewRuntimeErrorWithLog("substring indices must be numbers")
	}
	runes := []rune(s)
	if start < 0 || end < 0 || start > end || int(end) > len(runes) {
		return nil, NewRuntimeErrorWithLog("substring index out of range")
	}
	return string(runes[start:end]), nil
}

type BuiltInFnGetch struct{}

func (b *BuiltInFnGetch) Arity() int { return 0 }
func (b *BuiltInFnGetch) Call(interpreter *Interpreter, arguments []any) (any, error) {
	reader := bufio.NewReader(os.Stdin)
	r, _, err := reader.ReadRune()
	if err != nil {
		return nil, NewRuntimeErrorWithLog("failed to read char")
	}
	// If user pressed Enter first, try next rune
	if r == '\n' || r == '\r' {
		r, _, err = reader.ReadRune()
		if err != nil {
			return nil, NewRuntimeErrorWithLog("failed to read char")
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
