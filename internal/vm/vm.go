package vm

import (
	"internal/bytecode"
	"internal/runtime"
	"internal/util/log"
)

type callFrame struct {
	closure *runtime.ObjClosure
	ip      int
	sp      int
}

func (cf *callFrame) getChunk() *bytecode.Chunk {
	return cf.closure.Function.Chunk
}

type VM struct {
	callFrames []*callFrame
	stack      []bytecode.Value

	globals map[string]bytecode.Value
	builtin map[string]bytecode.Value

	openUpvalues []*runtime.ObjUpvalue
}

func NewVM() *VM {
	return &VM{}
}

func (vm *VM) Free() {
	vm.callFrames = vm.callFrames[:0]
	vm.stack = vm.stack[:0]

	vm.globals = make(map[string]bytecode.Value)
	vm.builtin = make(map[string]bytecode.Value)
	vm.openUpvalues = vm.openUpvalues[:0]

	vm.initNativeFunctions()
}

func (vm *VM) initNativeFunctions() {
	for _, fn := range runtime.NativeFunctions {
		vm.builtin[fn.Name] = fn
	}
}

func (vm *VM) Run(fn *runtime.ObjFunction) InterpretResult {
	vm.Free()

	frame := &callFrame{
		closure: runtime.NewObjClosure(fn),
		ip:      0,
		sp:      0,
	}
	vm.callFrames = append(vm.callFrames, frame)

	return vm.run()
}

// ================================================================
// callFrame
// ================================================================

func (vm *VM) currentFrame() *callFrame {
	return vm.callFrames[len(vm.callFrames)-1]
}

// ================================================================
// CODE
// ================================================================

func (vm *VM) peekOp() bytecode.OpCode {
	frame := vm.currentFrame()
	chunk := frame.getChunk()

	return chunk.GetOperator(frame.ip)
}

func (vm *VM) getOp() bytecode.OpCode {
	frame := vm.currentFrame()

	op := vm.peekOp()
	frame.ip++

	return op
}

func (vm *VM) peekOperand() int64 {
	frame := vm.currentFrame()
	chunk := frame.getChunk()

	v, _ := chunk.GetOperand(frame.ip)

	return v
}

func (vm *VM) getOperand() int64 {
	frame := vm.currentFrame()
	chunk := frame.getChunk()

	v, n := chunk.GetOperand(frame.ip)

	frame.ip += n

	return v
}

func (vm *VM) getConstant() bytecode.Value {
	frame := vm.currentFrame()
	chunk := frame.getChunk()

	constIndex := vm.getOperand()

	return chunk.GetConstant(constIndex)
}

// ================================================================
// STACK
// ================================================================

func (vm *VM) push(value bytecode.Value) {
	vm.stack = append(vm.stack, value)
}

func (vm *VM) getStack(idx int) bytecode.Value {
	frame := vm.currentFrame()

	return vm.stack[frame.sp+idx]
}

func (vm *VM) setStack(idx int, value bytecode.Value) {
	frame := vm.currentFrame()

	vm.stack[frame.sp+idx] = value
}

func (vm *VM) peek(idx int) bytecode.Value {
	stackTop := len(vm.stack) - 1

	return vm.stack[stackTop-idx]
}

func (vm *VM) pop() bytecode.Value {
	return vm.popN(1)
}

func (vm *VM) popN(n int) bytecode.Value {
	stackTop := len(vm.stack) - n

	if stackTop < 0 {
		vm.stack = vm.stack[:0]

		return nil
	}

	value := vm.stack[stackTop]
	vm.stack = vm.stack[:stackTop]

	return value
}

func (vm *VM) run() InterpretResult {
	defer func() {
		if len(vm.stack) > 0 {
			log.Warn("Stack not empty", log.A("stack", vm.stack))
		}
	}()

	for {
		frame := vm.currentFrame()
		chunk := frame.getChunk()

		if frame.ip >= chunk.Size() {
			break
		}

		instruction := vm.getOp()
		ip := frame.ip - 1

		fn := OP_FUNCS[instruction]
		if fn == nil {
			log.Error("Unknown opcode", log.A("opcode", instruction))

			return InterpretResultRuntimeError
		}

		result := fn(vm)

		log.DebugIfEnabled("VM run completed", func() []log.Field {
			return []log.Field{
				log.S("function", frame.closure.Function.Name),
				log.I("ip", ip),
				log.A("instruction", instruction),
				log.A("stack", vm.stack),
				log.A("result", result),
				log.A("globals", vm.globals),
			}
		})

		if result != InterpretResultOK {
			return result
		}
	}

	return InterpretResultOK

}
