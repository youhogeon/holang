package bytecode

import (
	"encoding/binary"
	"fmt"
	"internal/util/log"
)

type Value any

type Chunk struct {
	Code      []byte
	Constants []Value
	Lines     []int
}

func NewChunk() *Chunk {
	return &Chunk{}
}

func (c *Chunk) Write(line int, op OpCode, operands ...int64) {
	c.Code = append(c.Code, byte(op))
	c.Lines = append(c.Lines, line)

	operandsCount := op.OperandsCount()
	if len(operands) != operandsCount {
		log.Error("operands count mismatch", log.I("expected", operandsCount), log.I("got", len(operands)), log.I("line", line), log.S("operator", op.String()), log.A("operands", operands))
	}

	for _, operand := range operands {
		tmp := make([]byte, binary.MaxVarintLen32)
		k := binary.PutVarint(tmp, operand)
		c.Code = append(c.Code, tmp[:k]...)
	}
}

func (c *Chunk) AddConstant(value Value) int64 {
	c.Constants = append(c.Constants, value)

	return int64(len(c.Constants) - 1)
}

func (c *Chunk) GetConstant(index int64) Value {
	return c.Constants[index]
}

func (c *Chunk) GetOperator(index int) OpCode {
	return OpCode(c.Code[index])
}

func (c *Chunk) GetOperand(index int) (int64, int) {
	return binary.Varint(c.Code[index:])
}

func (c *Chunk) Clear() {
	c.Code = c.Code[:0]
	c.Constants = c.Constants[:0]
}

func (c *Chunk) Size() int {
	return len(c.Code)
}

func (c *Chunk) Disassemble() []string {
	var dis []string

	opIdx := 0
	for pos := 0; pos < len(c.Code); pos++ {
		operator := OpCode(c.Code[pos])

		operandsCount := operator.OperandsCount()
		operands := make([]any, operandsCount)

		for j := range operandsCount {
			x, n := c.GetOperand(pos + 1 + j)
			pos += n

			if operator == OP_COSNTANT {
				operands[j] = fmt.Sprintf("%d, value %v", x, c.GetConstant(x))
			} else {
				operands[j] = x
			}
		}

		dis = append(dis, operator.String())
		log.Debug("Disassemble", log.I("pos", pos), log.I("line", c.Lines[opIdx]), log.A("operator", operator), log.A("operands", operands))

		opIdx += 1
	}

	return dis
}
