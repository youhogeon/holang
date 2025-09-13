package bytecode

import (
	"encoding/binary"
	"fmt"
	"internal/util/log"
)

type Value any

type Offset struct {
	Line  int
	Index int
}

type Chunk struct {
	code      []byte
	constants []Value
	offsets   []Offset
}

func NewChunk() *Chunk {
	return &Chunk{}
}

func (c *Chunk) AddOperator(offset Offset, op OpCode, operands ...int64) {
	c.AddCode(op)
	c.offsets = append(c.offsets, offset)

	operandsCount := op.OperandsCount()
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

func (c *Chunk) Disassemble() []string {
	var dis []string

	opIdx := 0
	for pos := 0; pos < len(c.code); pos++ {
		operator := OpCode(c.code[pos])
		_pos := pos

		operandsCount := operator.OperandsCount()
		operands := make([]any, operandsCount)

		for j := range operandsCount {
			x, n := c.GetOperand(pos + 1 + j)
			pos += n

			if operator == OP_CONSTANT {
				operands[j] = fmt.Sprintf("%d, value %v", x, c.GetConstant(x))
			} else {
				operands[j] = x
			}
		}

		dis = append(dis, operator.String())
		log.Debug("Disassemble", log.I("pos", _pos), log.A("offset", c.offsets[opIdx]), log.A("operator", operator), log.A("operands", operands))

		opIdx += 1
	}

	return dis
}
