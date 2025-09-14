package codegen

import (
	"internal/bytecode"
	"internal/token"
)

type Emitter interface {
	Emit(offset token.Offset, op bytecode.OpCode, operands ...int64)
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

func (e *ChunkEmitter) Emit(offset token.Offset, op bytecode.OpCode, operands ...int64) {
	e.chunk.AddOperator(offset, op, operands...)
}

func (e *ChunkEmitter) MakeConstant(value bytecode.Value) int64 {
	return e.chunk.AddConstant(value)
}

func (e *ChunkEmitter) EmitJump(offset token.Offset, op bytecode.OpCode) int {
	return 1
}

func (e *ChunkEmitter) PatchJump(at int) {

}

func (e *ChunkEmitter) EmitLoop(offset token.Offset, loopStart int) {

}
