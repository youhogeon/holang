package codegen

import (
	"encoding/binary"
	"internal/bytecode"
	"internal/runtime"
	"internal/token"
	"internal/util/log"
)

type Emitter interface {
	Emit(offset token.Offset, op bytecode.OpCode, operands ...int64) int
	MakeConstant(value runtime.Value) int64
	EmitJump(offset token.Offset, op bytecode.OpCode) int
	PatchJump(jumpOpLoc int, jumpTo int)
	Size() int
}

type ChunkEmitter struct {
	chunk *runtime.Chunk
}

func NewChunkEmitter(chunk *runtime.Chunk) *ChunkEmitter {
	return &ChunkEmitter{
		chunk: chunk,
	}
}

func (e *ChunkEmitter) Emit(offset token.Offset, op bytecode.OpCode, operands ...int64) int {
	at := e.chunk.Size()
	e.chunk.AddOperator(offset, op, operands...)

	return at
}

func (e *ChunkEmitter) MakeConstant(value runtime.Value) int64 {
	return e.chunk.AddConstant(value)
}

func (e *ChunkEmitter) EmitJump(offset token.Offset, op bytecode.OpCode) int {
	at := e.chunk.Size()

	e.Emit(offset, op, 0xfffff) // reserve 3 bytes for jump offset (will be patched later)

	return at
}

func (e *ChunkEmitter) PatchJump(jumpOpLoc int, jumpTo int) {
	jump := jumpTo - jumpOpLoc - 4

	tmp := make([]byte, binary.MaxVarintLen32)
	k := binary.PutVarint(tmp, int64(jump))

	if k > 3 {
		log.Fatal("Too much to jump", log.I("jump", jump))
	}

	// jump offset는 3 bytes로 고정
	tmp[0] = tmp[0] | 128
	tmp[1] = tmp[1] | 128

	e.chunk.OverWrite(jumpOpLoc+1, tmp[0])
	e.chunk.OverWrite(jumpOpLoc+2, tmp[1])
	e.chunk.OverWrite(jumpOpLoc+3, tmp[2])
}

func (e *ChunkEmitter) Size() int {
	return e.chunk.Size()
}
