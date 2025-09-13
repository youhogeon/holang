package bytecode

type OpCode byte

// ================================================================
// OP codes
// --------
// 1. const 정의
// 1. (필요 시) operandsCount 추가
// 1. operation.go 구현
// ================================================================

const (
	OP_RETURN OpCode = iota
	OP_CONSTANT
	OP_NEGATE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NOT
)

var strings = map[OpCode]string{
	OP_RETURN:   "OP_RETURN",
	OP_CONSTANT: "OP_CONSTANT",
	OP_NEGATE:   "OP_NEGATE",
	OP_ADD:      "OP_ADD",
	OP_SUBTRACT: "OP_SUBTRACT",
	OP_MULTIPLY: "OP_MULTIPLY",
	OP_DIVIDE:   "OP_DIVIDE",
	OP_NOT:      "OP_NOT",
}

var operandsCount = map[OpCode]int{
	OP_CONSTANT: 1,
}

func (op OpCode) String() string {
	if s, ok := strings[op]; ok {
		return s
	}

	return "OP_UNKNOWN"
}

func (op OpCode) OperandsCount() int {
	if c, ok := operandsCount[op]; ok {
		return c
	}

	return 0
}
