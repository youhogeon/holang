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

	// CONSTANT
	OP_CONSTANT
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

	// UNARY
	OP_NEGATE
	OP_NOT

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
)

var strings = map[OpCode]string{
	OP_RETURN:        "OP_RETURN",
	OP_CONSTANT:      "OP_CONSTANT",
	OP_TRUE:          "OP_TRUE",
	OP_FALSE:         "OP_FALSE",
	OP_NIL:           "OP_NIL",
	OP_CONSTANT_M1:   "OP_CONSTANT_M1",
	OP_CONSTANT_0:    "OP_CONSTANT_0",
	OP_CONSTANT_1:    "OP_CONSTANT_1",
	OP_CONSTANT_2:    "OP_CONSTANT_2",
	OP_CONSTANT_3:    "OP_CONSTANT_3",
	OP_CONSTANT_4:    "OP_CONSTANT_4",
	OP_CONSTANT_5:    "OP_CONSTANT_5",
	OP_NEGATE:        "OP_NEGATE",
	OP_NOT:           "OP_NOT",
	OP_ADD:           "OP_ADD",
	OP_SUBTRACT:      "OP_SUBTRACT",
	OP_MULTIPLY:      "OP_MULTIPLY",
	OP_DIVIDE:        "OP_DIVIDE",
	OP_EQUAL:         "OP_EQUAL",
	OP_NOT_EQUAL:     "OP_NOT_EQUAL",
	OP_GREATER:       "OP_GREATER",
	OP_LESS:          "OP_LESS",
	OP_GREATER_EQUAL: "OP_GREATER_EQUAL",
	OP_LESS_EQUAL:    "OP_LESS_EQUAL",
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
