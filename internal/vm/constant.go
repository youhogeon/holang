package vm

type InterpretResult byte

const (
	InterpretResultOK InterpretResult = iota
	InterpretResultCompileError
	InterpretResultRuntimeError
)

const FRAMES_MAX = 1024
