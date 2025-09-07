package interpreter

import (
	"fmt"
	"internal/ast"
	"time"
)

type Callable interface {
	Arity() int
	Call(interpreter *Interpreter, arguments []any) (any, error)
}

type Function struct {
	declaration *ast.Function
	clousure    *Environment
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
			return returnSig.value, nil
		}

		return nil, err
	}

	return nil, nil
}

type BuiltInFnPrint struct{}

func (b *BuiltInFnPrint) Arity() int {
	return 1
}

func (b *BuiltInFnPrint) Call(interpreter *Interpreter, arguments []any) (any, error) {
	fmt.Println(arguments[0])

	return nil, nil
}

type BuiltInFnClock struct{}

func (b *BuiltInFnClock) Arity() int {
	return 0
}

func (b *BuiltInFnClock) Call(interpreter *Interpreter, arguments []any) (any, error) {
	return int64(time.Now().UnixNano() / 1e9), nil
}
