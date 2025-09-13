package vm

type InterpretResult byte

const (
	InterpretResultOK InterpretResult = iota
	InterpretResultCompileError
	InterpretResultRuntimeError
)
