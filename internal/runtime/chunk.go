package runtime

import (
	"encoding/binary"
	"fmt"
	"internal/bytecode"
	"internal/token"
	"internal/util/log"
	"io"
)

type Chunk struct {
	code      []byte
	constants []Value
	offsets   []token.Offset

	constantCache map[Value]int64
}

func NewChunk() *Chunk {
	return &Chunk{
		constantCache: make(map[Value]int64),
	}
}

func (c *Chunk) AddOperator(offset token.Offset, op bytecode.OpCode, operands ...int64) {
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
		case bytecode.OpCode:
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
	if idx, ok := c.constantCache[value]; ok {
		return idx
	}

	c.constantCache[value] = int64(len(c.constants))

	c.constants = append(c.constants, value)

	return int64(len(c.constants) - 1)
}

func (c *Chunk) GetConstant(index int64) Value {
	return c.constants[index]
}

func (c *Chunk) CountConstants() int {
	return len(c.constants)
}

func (c *Chunk) GetOperator(index int) bytecode.OpCode {
	return bytecode.OpCode(c.code[index])
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
		operator := bytecode.OpCode(c.code[pos])
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

			if operator == bytecode.OP_CONSTANT {
				operands[j] = fmt.Sprintf("%d (value: %v)", x, c.GetConstant(x))
			} else if operator == bytecode.OP_DEFINE_GLOBAL || operator == bytecode.OP_GET_GLOBAL || operator == bytecode.OP_SET_GLOBAL {
				operands[j] = fmt.Sprintf("%d (name: %v)", x, c.GetConstant(x))
			} else if operator == bytecode.OP_JUMP || operator == bytecode.OP_JUMP_IF_FALSE {
				operands[j] = fmt.Sprintf("%d (target: %d)", x, int64(pos)+x+1)
			} else {
				operands[j] = x
			}
		}

		log.Debug("Bytecode", log.I("pos", _pos), log.A("offset", c.offsets[opIdx]), log.A("operator", operator), log.A("operands", operands))

		opIdx += 1
	}
}

func (c *Chunk) Serialize(w io.Writer) error {
	{
		len := int64(len(c.code))
		if err := writeInt(w, len); err != nil {
			return err
		}

		if _, err := w.Write(c.code); err != nil {
			return err
		}
	}

	{
		len := int64(len(c.constants))
		if err := writeInt(w, len); err != nil {
			return err
		}

		for _, constant := range c.constants {
			if err := SerializeValue(w, constant); err != nil {
				return err
			}
		}
	}

	{
		len := int64(len(c.offsets))
		if err := writeInt(w, len); err != nil {
			return err
		}

		for _, offset := range c.offsets {
			if err := writeInt(w, int64(offset.Line)); err != nil {
				return err
			}

			if err := writeInt(w, int64(offset.Index)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Chunk) Deserialize(data []byte) (any, []byte, error) {
	codeLen, rest, err := readInt(data)
	if err != nil {
		return nil, nil, err
	}

	if int64(len(rest)) < codeLen {
		return nil, nil, io.EOF
	}

	code := rest[:codeLen]
	rest = rest[codeLen:]

	constCount, rest2, err := readInt(rest)
	if err != nil {
		return nil, nil, err
	}

	constants := make([]Value, 0, constCount)
	for i := int64(0); i < constCount; i++ {
		constant, _rest, err := DeserializeValue(rest2)
		if err != nil {
			return nil, nil, err
		}
		constants = append(constants, constant)
		rest2 = _rest
	}

	offsetCount, rest3, err := readInt(rest2)
	if err != nil {
		return nil, nil, err
	}

	offsets := make([]token.Offset, 0, offsetCount)
	for i := int64(0); i < offsetCount; i++ {
		line, _r, err := readInt(rest3)
		if err != nil {
			return nil, nil, err
		}

		index, _r2, err := readInt(_r)
		if err != nil {
			return nil, nil, err
		}

		offsets = append(offsets, token.Offset{Line: int(line), Index: int(index)})
		rest3 = _r2
	}

	chunk := &Chunk{code: code, constants: constants, offsets: offsets}
	return chunk, rest3, nil
}
