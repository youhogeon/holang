package bytecode

//go:generate stringer -type=OpCode
type OpCode byte

// ================================================================
// OP codes
// --------
// 1. const 정의
// 1. (필요 시) operandsCount 추가
// 1. operation.go 구현
// ================================================================

const (
	// CONSTANT
	OP_CONSTANT OpCode = iota
	OP_TRUE
	OP_FALSE
	OP_NIL
	OP_CONSTANT_M1
	OP_CONSTANT_0
	OP_CONSTANT_1
	OP_CONSTANT_2
	OP_CONSTANT_3
	OP_CONSTANT_4
	OP_CONSTANT_5

	// UNARY, TERNARY
	OP_NEGATE
	OP_NOT
	// OP_TERNARY

	// BINARY
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_EQUAL
	OP_NOT_EQUAL
	OP_GREATER
	OP_LESS
	OP_GREATER_EQUAL
	OP_LESS_EQUAL

	// SPECIAL
	OP_RETURN
	OP_PRINT
)

var operandsCount = map[OpCode]int{
	OP_CONSTANT: 1,
}

func (op OpCode) OperandsCount() int {
	if c, ok := operandsCount[op]; ok {
		return c
	}

	return 0
}
