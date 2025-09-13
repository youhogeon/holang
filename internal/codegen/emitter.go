package codegen

import "internal/bytecode"

type Emitter interface {
	Emit(offset bytecode.Offset, op bytecode.OpCode, operands ...int64)
	MakeConstant(value bytecode.Value) int64
	EmitJump(offset bytecode.Offset, op bytecode.OpCode) int
	PatchJump(at int)
	EmitLoop(offset bytecode.Offset, loopStart int)
}

type ChunkEmitter struct {
	chunk *bytecode.Chunk
}

func NewChunkEmitter(chunk *bytecode.Chunk) *ChunkEmitter {
	return &ChunkEmitter{
		chunk: chunk,
	}
}

func (e *ChunkEmitter) Emit(offset bytecode.Offset, op bytecode.OpCode, operands ...int64) {
	e.chunk.AddOperator(offset, op, operands...)
}

func (e *ChunkEmitter) MakeConstant(value bytecode.Value) int64 {
	return e.chunk.AddConstant(value)
}

func (e *ChunkEmitter) EmitJump(offset bytecode.Offset, op bytecode.OpCode) int {
	return 1
}

func (e *ChunkEmitter) PatchJump(at int) {

}

func (e *ChunkEmitter) EmitLoop(offset bytecode.Offset, loopStart int) {

}
