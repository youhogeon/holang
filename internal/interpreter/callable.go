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
