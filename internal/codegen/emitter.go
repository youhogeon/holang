package codegen

import (
	"encoding/binary"
	"internal/bytecode"
	"internal/token"
	"internal/util/log"
)

type Emitter interface {
	Emit(offset token.Offset, op bytecode.OpCode, operands ...int64) int
	MakeConstant(value bytecode.Value) int64
	EmitJump(offset token.Offset, op bytecode.OpCode) int
	PatchJump(at int)
	EmitLoop(offset token.Offset, loopStart int)
}

type ChunkEmitter struct {
	chunk *bytecode.Chunk
}

func NewChunkEmitter(chunk *bytecode.Chunk) *ChunkEmitter {
	return &ChunkEmitter{
		chunk: chunk,
	}
}

func (e *ChunkEmitter) Emit(offset token.Offset, op bytecode.OpCode, operands ...int64) int {
	at := e.chunk.Size()
	e.chunk.AddOperator(offset, op, operands...)

	return at
}

func (e *ChunkEmitter) MakeConstant(value bytecode.Value) int64 {
	return e.chunk.AddConstant(value)
}

func (e *ChunkEmitter) EmitJump(offset token.Offset, op bytecode.OpCode) int {
	at := e.chunk.Size()

	e.Emit(offset, op, 0xfffff) // reserve 3 bytes for jump offset (will be patched later)

	return at
}

func (e *ChunkEmitter) PatchJump(at int) {
	jump := e.chunk.Size() - at - 4

	tmp := make([]byte, binary.MaxVarintLen32)
	k := binary.PutVarint(tmp, int64(jump))

	if k > 3 {
		log.Fatal("Too much to jump", log.I("jump", jump))
	}

	// jump offset는 3 bytes로 고정
	tmp[0] = tmp[0] | 128
	tmp[1] = tmp[1] | 128

	e.chunk.OverWrite(at+1, tmp[0])
	e.chunk.OverWrite(at+2, tmp[1])
	e.chunk.OverWrite(at+3, tmp[2])
}

func (e *ChunkEmitter) EmitLoop(offset token.Offset, loopStart int) {

}
