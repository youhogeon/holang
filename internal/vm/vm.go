package vm

import (
	"internal/bytecode"
	"internal/util/log"
)

type VM struct {
	chunk   *bytecode.Chunk
	ip      int
	stack   []bytecode.Value
	objects *ObjectList
}

func NewVM() *VM {
	return &VM{}
}

func (vm *VM) Free() {
	if vm.chunk != nil {
		vm.chunk.Clear()
		vm.chunk = nil
	}
}

func (vm *VM) Interpret(chunk *bytecode.Chunk) InterpretResult {
	vm.chunk = chunk
	vm.ip = 0
	vm.stack = vm.stack[:0]
	vm.objects = NewObjectList()

	return vm.run()
}

func (vm *VM) peekOp() bytecode.OpCode {
	return vm.chunk.GetOperator(vm.ip)
}

func (vm *VM) getOp() bytecode.OpCode {
	op := vm.peekOp()
	vm.ip++

	return op
}

func (vm *VM) peekOperand() int64 {
	v, _ := vm.chunk.GetOperand(vm.ip)

	return v
}

func (vm *VM) getOperand() int64 {
	v, n := vm.chunk.GetOperand(vm.ip)

	vm.ip += n

	return v
}

func (vm *VM) peekConstant() bytecode.Value {
	constIndex := vm.peekOperand()

	return vm.chunk.GetConstant(constIndex)
}

func (vm *VM) getConstant() bytecode.Value {
	constIndex := vm.getOperand()

	return vm.chunk.GetConstant(constIndex)
}

func (vm *VM) push(value bytecode.Value) {
	vm.stack = append(vm.stack, value)
}

func (vm *VM) pop() bytecode.Value {
	stackTop := len(vm.stack) - 1
	if stackTop < 0 {
		return nil
	}

	value := vm.stack[stackTop]
	vm.stack = vm.stack[:stackTop]

	return value
}

func (vm *VM) run() InterpretResult {
	for vm.ip < vm.chunk.Size() {
		instruction := vm.getOp()

		log.DebugIfEnabled("VM run", func() []log.Field {
			return []log.Field{
				log.I("ip", vm.ip-1),
				log.A("instruction", instruction),
				log.A("stack", vm.stack),
			}
		})

		fn := OP_FUNCS[instruction]
		if fn == nil {
			log.Error("Unknown opcode", log.A("opcode", instruction))

			return InterpretResultRuntimeError
		}

		result := fn(vm)
		if result != InterpretResultOK {
			return result
		}
	}

	return InterpretResultOK
}
