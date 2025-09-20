package bytecode

import (
	"encoding/binary"
	"fmt"
	"internal/token"
	"internal/util/log"
)

type Value any

type Chunk struct {
	code      []byte
	constants []Value
	offsets   []token.Offset
}

func NewChunk() *Chunk {
	return &Chunk{}
}

func (c *Chunk) AddOperator(offset token.Offset, op OpCode, operands ...int64) {
	c.AddCode(op)
	c.offsets = append(c.offsets, offset)

	operandsCount := op.OperandsCount()
	if operandsCount == -1 { // 가변
		operandsCount = int(operands[0]) + 1
	}

	if len(operands) != operandsCount {
		log.Error("operands count mismatch", log.I("expected", operandsCount), log.I("got", len(operands)), log.A("offset", offset), log.S("operator", op.String()), log.A("operands", operands))
	}

	for _, operand := range operands {
		c.AddCode(operand)
	}
}

func (c *Chunk) AddCode(code ...any) {
	for _, v := range code {
		switch v := v.(type) {
		case byte:
			c.code = append(c.code, v)
		case OpCode:
			c.code = append(c.code, byte(v))
		case int64:
			tmp := make([]byte, binary.MaxVarintLen32)
			k := binary.PutVarint(tmp, v)
			c.code = append(c.code, tmp[:k]...)
		default:
			log.Fatal("unsupported code type", log.A("value", v))
		}
	}
}

func (c *Chunk) AddConstant(value Value) int64 {
	c.constants = append(c.constants, value)

	return int64(len(c.constants) - 1)
}

func (c *Chunk) GetConstant(index int64) Value {
	return c.constants[index]
}

func (c *Chunk) CountConstants() int {
	return len(c.constants)
}

func (c *Chunk) GetOperator(index int) OpCode {
	return OpCode(c.code[index])
}

func (c *Chunk) GetOperand(index int) (int64, int) {
	return binary.Varint(c.code[index:])
}

func (c *Chunk) Clear() {
	c.code = c.code[:0]
	c.constants = c.constants[:0]
}

func (c *Chunk) Size() int {
	return len(c.code)
}

func (c *Chunk) OverWrite(at int, value byte) {
	c.code[at] = value
}

func (c *Chunk) Disassemble() {
	opIdx := 0
	for pos := 0; pos < len(c.code); pos++ {
		operator := OpCode(c.code[pos])
		_pos := pos

		operandsCount := operator.OperandsCount()
		if operandsCount == -1 { // 가변
			x, _ := c.GetOperand(pos + 1)
			operandsCount = int(x) + 1
		}

		operands := make([]any, operandsCount)

		for j := range operandsCount {
			x, n := c.GetOperand(pos + 1)
			pos += n

			if operator == OP_CONSTANT {
				operands[j] = fmt.Sprintf("%d (value: %v)", x, c.GetConstant(x))
			} else if operator == OP_DEFINE_GLOBAL || operator == OP_GET_GLOBAL || operator == OP_SET_GLOBAL {
				operands[j] = fmt.Sprintf("%d (name: %v)", x, c.GetConstant(x))
			} else if operator == OP_JUMP || operator == OP_JUMP_IF_FALSE {
				operands[j] = fmt.Sprintf("%d (target: %d)", x, int64(pos)+x+1)
			} else {
				operands[j] = x
			}
		}

		log.Debug("Bytecode", log.I("pos", _pos), log.A("offset", c.offsets[opIdx]), log.A("operator", operator), log.A("operands", operands))

		opIdx += 1
	}
}
