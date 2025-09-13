package bytecode

type OpCode byte

const (
	OP_RETURN OpCode = iota
	OP_COSNTANT
	OP_ADD
)

var strings = map[OpCode]string{
	OP_RETURN:   "OP_RETURN",
	OP_COSNTANT: "OP_CONSTANT",
	OP_ADD:      "OP_ADD",
}

var operandsCount = map[OpCode]int{
	OP_COSNTANT: 1,
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
