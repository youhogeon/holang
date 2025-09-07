package resolver

type FunctionType int

const (
	NOT_FUNCTION_TYPE FunctionType = iota
	FUNCTION
	METHOD
	INITIALIZER
)

type ClassType int

const (
	NOT_CLASS_TYPE ClassType = iota
	CLASS
	SUBCLASS
)
