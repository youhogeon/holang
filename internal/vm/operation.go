package vm

import (
	"fmt"
	"internal/runtime"
	"internal/util"
	"internal/util/log"
)

var OP_FUNCS []func(vm *VM) InterpretResult = []func(vm *VM) InterpretResult{
	// CONSTANT
	(*VM).OP_CONSTANT,
	(*VM).OP_TRUE,
	(*VM).OP_FALSE,
	(*VM).OP_NIL,
	(*VM).OP_CONSTANT_M1,
	(*VM).OP_CONSTANT_0,
	(*VM).OP_CONSTANT_1,
	(*VM).OP_CONSTANT_2,
	(*VM).OP_CONSTANT_3,
	(*VM).OP_CONSTANT_4,
	(*VM).OP_CONSTANT_5,

	// UNARY
	(*VM).OP_NEGATE,
	(*VM).OP_NOT,

	// BINARY
	(*VM).OP_ADD,
	(*VM).OP_SUBTRACT,
	(*VM).OP_MULTIPLY,
	(*VM).OP_DIVIDE,
	(*VM).OP_EQUAL,
	(*VM).OP_NOT_EQUAL,
	(*VM).OP_GREATER,
	(*VM).OP_LESS,
	(*VM).OP_GREATER_EQUAL,
	(*VM).OP_LESS_EQUAL,

	// VARIABLE
	(*VM).OP_DEFINE_GLOBAL,
	(*VM).OP_GET_GLOBAL,
	(*VM).OP_SET_GLOBAL,
	(*VM).OP_GET_LOCAL,
	(*VM).OP_SET_LOCAL,

	// STACK
	(*VM).OP_POP,
	(*VM).OP_POPN,

	// JUMP
	(*VM).OP_JUMP,
	(*VM).OP_JUMP_IF_FALSE,

	// FUN
	(*VM).OP_CALL,
	(*VM).OP_RETURN,
	(*VM).OP_CLOSURE,
	(*VM).OP_GET_UPVALUE,
	(*VM).OP_SET_UPVALUE,
	(*VM).OP_CLOSE_UPVALUE,

	// CLASS
	(*VM).OP_CLASS,
	(*VM).OP_GET_PROPERTY,
	(*VM).OP_SET_PROPERTY,
	(*VM).OP_METHOD,
	(*VM).OP_INVOKE,

	// SPECIAL
	(*VM).OP_PRINT,
}

// ================================================================
// CONSTANT
// ================================================================

func (vm *VM) OP_CONSTANT() InterpretResult {
	constant := vm.getConstant()

	vm.push(constant)

	return InterpretResultOK
}

func (vm *VM) OP_TRUE() InterpretResult {
	vm.push(true)

	return InterpretResultOK
}

func (vm *VM) OP_FALSE() InterpretResult {
	vm.push(false)

	return InterpretResultOK
}

func (vm *VM) OP_NIL() InterpretResult {
	vm.push(nil)

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT_M1() InterpretResult {
	vm.push(int64(-1))

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT_0() InterpretResult {
	vm.push(int64(0))

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT_1() InterpretResult {
	vm.push(int64(1))

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT_2() InterpretResult {
	vm.push(int64(2))

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT_3() InterpretResult {
	vm.push(int64(3))

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT_4() InterpretResult {
	vm.push(int64(4))

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT_5() InterpretResult {
	vm.push(int64(5))

	return InterpretResultOK
}

// ================================================================
// UNARY, TRINARY
// ================================================================

func (vm *VM) OP_NEGATE() InterpretResult {
	value := vm.pop()

	switch v := value.(type) {
	case int64:
		vm.push(-v)
	case float64:
		vm.push(-v)
	default:
		log.Error("Operand must be a number", log.A("value", value))

		return InterpretResultRuntimeError
	}

	return InterpretResultOK
}

func (vm *VM) OP_NOT() InterpretResult {
	value := vm.pop()

	vm.push(!util.IsTruthy(value))

	return InterpretResultOK
}

// ================================================================
// BINARY
// ================================================================

func (vm *VM) OP_ADD() InterpretResult {
	b := vm.pop()
	a := vm.pop()

	switch aVal := a.(type) {
	case int64:
		switch bVal := b.(type) {
		case int64:
			vm.push(aVal + bVal)
		case float64:
			vm.push(float64(aVal) + bVal)
		case string:
			vm.push(fmt.Sprintf("%d%s", aVal, bVal))
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	case float64:
		switch bVal := b.(type) {
		case int64:
			vm.push(aVal + float64(bVal))
		case float64:
			vm.push(aVal + bVal)
		case string:
			vm.push(fmt.Sprintf("%f%s", aVal, bVal))
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	case string:
		switch bVal := b.(type) {
		case int64:
			vm.push(fmt.Sprintf("%s%d", aVal, bVal))
		case float64:
			vm.push(fmt.Sprintf("%s%f", aVal, bVal))
		case string:
			vm.push(aVal + bVal)
		default:
			log.Error("Operand must be a number or string", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	default:
		log.Error("Operand must be a number or string", log.A("a", a), log.A("b", b))

		return InterpretResultRuntimeError
	}

	return InterpretResultOK
}

func (vm *VM) OP_SUBTRACT() InterpretResult {
	return vm._binary(
		func(a int64, b int64) any {
			return a - b
		}, func(a float64, b float64) any {
			return a - b
		},
	)
}

func (vm *VM) OP_MULTIPLY() InterpretResult {
	return vm._binary(
		func(a int64, b int64) any {
			return a * b
		}, func(a float64, b float64) any {
			return a * b
		},
	)
}

func (vm *VM) OP_DIVIDE() InterpretResult {
	return vm._binary(
		func(a int64, b int64) any {
			if b == 0 {
				return float64(a) / float64(b)
			} else {
				return a / b
			}
		}, func(a float64, b float64) any {
			return a / b
		},
	)
}

func (vm *VM) OP_EQUAL() InterpretResult {
	b := vm.pop()
	a := vm.pop()

	vm.push(util.IsEqual(a, b))

	return InterpretResultOK
}

func (vm *VM) OP_NOT_EQUAL() InterpretResult {
	b := vm.pop()
	a := vm.pop()

	vm.push(util.IsNotEqual(a, b))

	return InterpretResultOK
}

func (vm *VM) OP_GREATER() InterpretResult {
	return vm._binary(
		func(a int64, b int64) any {
			return a > b
		}, func(a float64, b float64) any {
			return a > b
		},
	)
}

func (vm *VM) OP_LESS() InterpretResult {
	return vm._binary(
		func(a int64, b int64) any {
			return a < b
		}, func(a float64, b float64) any {
			return a < b
		},
	)
}

func (vm *VM) OP_GREATER_EQUAL() InterpretResult {
	return vm._binary(
		func(a int64, b int64) any {
			return a >= b
		}, func(a float64, b float64) any {
			return a >= b
		},
	)
}

func (vm *VM) OP_LESS_EQUAL() InterpretResult {
	return vm._binary(
		func(a int64, b int64) any {
			return a <= b
		}, func(a float64, b float64) any {
			return a <= b
		},
	)
}

func (vm *VM) _binary(
	intFunc func(a int64, b int64) any,
	floatFunc func(a float64, b float64) any,
) InterpretResult {
	b := vm.pop()
	a := vm.pop()

	if aNum, ok := a.(int64); ok {
		if bNum, ok := b.(int64); ok {
			vm.push(intFunc(aNum, bNum))

			return InterpretResultOK
		} else if bNum, ok := b.(float64); ok {
			vm.push(floatFunc(float64(aNum), bNum))

			return InterpretResultOK
		}
	} else if aNum, ok := a.(float64); ok {
		if bNum, ok := b.(int64); ok {
			vm.push(floatFunc(aNum, float64(bNum)))

			return InterpretResultOK
		} else if bNum, ok := b.(float64); ok {
			vm.push(floatFunc(aNum, bNum))

			return InterpretResultOK
		}
	}

	log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

	return InterpretResultRuntimeError
}

// ================================================================
// VARIABLE
// ================================================================

func (vm *VM) OP_DEFINE_GLOBAL() InterpretResult {
	name := vm.getConstant()

	value := vm.pop()
	vm.globals[name.(string)] = value

	return InterpretResultOK
}

func (vm *VM) OP_GET_GLOBAL() InterpretResult {
	name := vm.getConstant()

	if value, ok := vm.globals[name.(string)]; ok {
		vm.push(value)

		return InterpretResultOK
	}

	if value, ok := vm.builtin[name.(string)]; ok {
		vm.push(value)

		return InterpretResultOK
	}

	log.Error("Undefined variable", log.A("name", name))

	return InterpretResultRuntimeError
}

func (vm *VM) OP_SET_GLOBAL() InterpretResult {
	name := vm.getConstant()
	if _, ok := vm.globals[name.(string)]; ok {
		value := vm.peek(0)
		vm.globals[name.(string)] = value

		return InterpretResultOK
	}

	log.Error("Undefined variable", log.A("name", name))

	return InterpretResultRuntimeError
}

func (vm *VM) OP_GET_LOCAL() InterpretResult {
	slot := vm.getOperand()
	vm.push(vm.getStack(int(slot)))

	return InterpretResultOK
}

func (vm *VM) OP_SET_LOCAL() InterpretResult {
	slot := vm.getOperand()
	vm.setStack(int(slot), vm.peek(0))

	return InterpretResultOK
}

// ================================================================
// STACK
// ================================================================

func (vm *VM) OP_POP() InterpretResult {
	vm.pop()

	return InterpretResultOK
}

func (vm *VM) OP_POPN() InterpretResult {
	n := vm.getOperand()
	vm.popN(int(n))

	return InterpretResultOK
}

// ================================================================
// JUMP
// ================================================================

func (vm *VM) OP_JUMP() InterpretResult {
	n := vm.getOperand()
	frame := vm.currentFrame()

	frame.ip += int(n)

	return InterpretResultOK
}

func (vm *VM) OP_JUMP_IF_FALSE() InterpretResult {
	condition := vm.peek(0)
	n := vm.getOperand()
	frame := vm.currentFrame()

	if !util.IsTruthy(condition) {
		frame.ip += int(n)
	}

	return InterpretResultOK
}

// ================================================================
// FUN
// ================================================================

func (vm *VM) OP_CALL() InterpretResult {
	argCount := int(vm.getOperand())
	callee := vm.peek(int(argCount))

	return vm.call(callee, argCount)
}

func (vm *VM) call(callee runtime.Value, argCount int) InterpretResult {
	switch callee := callee.(type) {
	case *runtime.ObjFunction:
		return vm.callFunction(callee, argCount)
	case *runtime.ObjClosure:
		return vm.callClosure(callee, argCount)
	case *runtime.ObjNativeFunction:
		return vm.callNativeFunction(callee, argCount)
	case *runtime.ObjClass:
		return vm.callClass(callee, argCount)
	case *runtime.ObjBoundMethod:
		return vm.callBoundMethod(callee, argCount)
	default:
		log.Error("Can only call functions", log.A("value", callee))

		return InterpretResultRuntimeError
	}
}

func (vm *VM) callFunction(fn *runtime.ObjFunction, argCount int) InterpretResult {
	closure := runtime.NewObjClosure(fn)

	return vm.callClosure(closure, argCount)
}

func (vm *VM) callClosure(closure *runtime.ObjClosure, argCount int) InterpretResult {
	fn := closure.Function

	if argCount != fn.Arity {
		log.Error("Wrong arguments count", log.I("expected", fn.Arity), log.I("got", argCount), log.A("function", fn.String()))

		return InterpretResultRuntimeError
	}

	if len(vm.callFrames) == FRAMES_MAX {
		log.Error("Stack overflow")

		return InterpretResultRuntimeError
	}

	frame := &callFrame{
		closure: closure,
		ip:      0,
		sp:      len(vm.stack) - argCount - 1,
	}

	vm.callFrames = append(vm.callFrames, frame)

	return InterpretResultOK
}

func (vm *VM) callNativeFunction(fn *runtime.ObjNativeFunction, argCount int) InterpretResult {
	if argCount != fn.Arity {
		log.Error("Wrong arguments count", log.I("expected", fn.Arity), log.I("got", argCount), log.A("function", fn.String()))

		return InterpretResultRuntimeError
	}

	args := make([]runtime.Value, argCount)
	for i := range argCount {
		args[argCount-i-1] = vm.pop()
	}

	result, err := fn.Function(args...)
	if err != nil {
		log.Error("Runtime error", log.E(err))

		return InterpretResultRuntimeError
	}

	vm.popN(argCount + 1)
	vm.push(result)

	return InterpretResultOK
}

func (vm *VM) callClass(class *runtime.ObjClass, argCount int) InterpretResult {
	instance := runtime.NewObjInstance(class)
	vm.setStack(len(vm.stack)-argCount-1, instance)

	if initializer, ok := class.Methods["init"]; ok {
		return vm.callClosure(initializer, argCount)
	} else if argCount != 0 {
		log.Error("Wrong arguments count", log.I("expected", 0), log.I("got", argCount), log.A("class", class.String()))

		return InterpretResultRuntimeError
	}

	return InterpretResultOK
}

func (vm *VM) callBoundMethod(boundMethod *runtime.ObjBoundMethod, argCount int) InterpretResult {
	vm.setStack(len(vm.stack)-argCount-1, boundMethod.Receiver)

	return vm.callClosure(boundMethod.Method, argCount)
}

func (vm *VM) OP_RETURN() InterpretResult {
	result := vm.pop()
	frame := vm.currentFrame()

	vm.closeUpvaluesFrom(frame.sp)

	vm.callFrames = vm.callFrames[:len(vm.callFrames)-1]
	vm.stack = vm.stack[:frame.sp]
	vm.push(result)

	return InterpretResultOK
}

func (vm *VM) OP_CLOSURE() InterpretResult {
	frame := vm.currentFrame()

	_ = int(vm.getOperand())
	constant := vm.getConstant()

	fn := constant.(*runtime.ObjFunction)
	closure := runtime.NewObjClosure(fn)

	vm.push(closure)

	uvCount := fn.UpvalueCount

	for i := range uvCount {
		isLocal := vm.getOperand()
		index := vm.getOperand()

		if isLocal == 1 {
			closure.Upvalues[i] = vm.captureUpvalue(frame.sp + int(index))
		} else {
			closure.Upvalues[i] = frame.closure.Upvalues[index]
		}
	}

	return InterpretResultOK
}

func (vm *VM) OP_GET_UPVALUE() InterpretResult {
	frame := vm.currentFrame()
	slot := vm.getOperand()

	uv := frame.closure.Upvalues[slot]
	vm.push(uv.Get())

	return InterpretResultOK
}

func (vm *VM) OP_SET_UPVALUE() InterpretResult {
	frame := vm.currentFrame()
	slot := vm.getOperand()

	uv := frame.closure.Upvalues[slot]
	uv.Set(vm.peek(0))

	return InterpretResultOK
}

func (vm *VM) OP_CLOSE_UPVALUE() InterpretResult {
	n := vm.getOperand()

	for range n {
		top := len(vm.stack) - 1

		if top >= 0 {
			vm.closeUpvalueAt(top)
			vm.pop()
		}
	}

	return InterpretResultOK
}

func (vm *VM) captureUpvalue(absIdx int) *runtime.ObjUpvalue {
	stackRef := &vm.stack
	for _, uv := range vm.openUpvalues {
		if !uv.IsClosed && uv.Stack == stackRef && uv.Index == absIdx {
			return uv
		}
	}

	uv := &runtime.ObjUpvalue{
		Stack: stackRef,
		Index: absIdx,
	}

	vm.openUpvalues = append(vm.openUpvalues, uv)

	return uv
}

func (vm *VM) closeUpvalueAt(absIdx int) {
	write := 0
	for _, uv := range vm.openUpvalues {
		if !uv.IsClosed && uv.Stack == &vm.stack && uv.Index == absIdx {
			uv.Close()
			continue
		}

		vm.openUpvalues[write] = uv
		write++
	}

	vm.openUpvalues = vm.openUpvalues[:write]
}

func (vm *VM) closeUpvaluesFrom(minIdx int) {
	for _, uv := range vm.openUpvalues {
		if !uv.IsClosed && uv.Stack == &vm.stack && uv.Index >= minIdx {
			uv.Close()
		}
	}

	write := 0
	for _, uv := range vm.openUpvalues {
		if uv.IsClosed {
			continue
		}
		vm.openUpvalues[write] = uv
		write++
	}
	vm.openUpvalues = vm.openUpvalues[:write]
}

// ================================================================
// CLASS
// ================================================================
func (vm *VM) OP_CLASS() InterpretResult {
	name := vm.getConstant()

	class := runtime.NewObjClass(name.(string))
	vm.push(class)

	return InterpretResultOK
}

func (vm *VM) OP_GET_PROPERTY() InterpretResult {
	_instance := vm.pop()

	instance, ok := _instance.(*runtime.ObjInstance)
	if !ok {
		log.Error("Only instances have properties", log.A("value", _instance))

		return InterpretResultRuntimeError
	}

	fieldName := vm.getConstant()
	if value, ok := instance.Fields[fieldName.(string)]; ok {
		vm.push(value)

		return InterpretResultOK
	}

	if method, ok := instance.Class.Methods[fieldName.(string)]; ok {
		bound := runtime.NewObjBoundMethod(method, instance)
		vm.push(bound)

		return InterpretResultOK
	}

	log.Error("Undefined property", log.A("name", fieldName), log.A("instance", instance))

	return InterpretResultRuntimeError
}

func (vm *VM) OP_SET_PROPERTY() InterpretResult {
	value := vm.pop()
	_instance := vm.pop()
	vm.push(value)

	instance, ok := _instance.(*runtime.ObjInstance)
	if !ok {
		log.Error("Only instances have properties", log.A("value", _instance))

		return InterpretResultRuntimeError
	}

	fieldName := vm.getConstant()
	instance.Fields[fieldName.(string)] = value

	return InterpretResultOK
}

func (vm *VM) OP_METHOD() InterpretResult {
	closure := vm.pop().(*runtime.ObjClosure)

	class := vm.peek(0).(*runtime.ObjClass)
	methodName := closure.Function.Name
	class.Methods[methodName] = closure

	return InterpretResultOK
}

func (vm *VM) OP_INVOKE() InterpretResult {
	methodName := vm.getConstant().(string)
	argCount := int(vm.getOperand())

	receiver := vm.peek(argCount)
	instance, ok := receiver.(*runtime.ObjInstance)
	if !ok {
		log.Error("Only instances have methods", log.A("value", receiver), log.S("method", methodName))
		return InterpretResultRuntimeError
	}

	if method, ok := instance.Class.Methods[methodName]; ok {
		bound := runtime.NewObjBoundMethod(method, instance)
		vm.setStack(len(vm.stack)-argCount-1, bound)

		return vm.callBoundMethod(bound, argCount)
	}

	if value, ok := instance.Fields[methodName]; ok {
		vm.setStack(len(vm.stack)-argCount-1, value)

		return vm.call(value, argCount)
	}

	log.Error("Undefined property", log.A("name", methodName), log.A("instance", instance))

	return InterpretResultRuntimeError
}

// ================================================================
// SPECIAL
// ================================================================

func (vm *VM) OP_PRINT() InterpretResult {
	value := vm.pop()
	fmt.Println(value)

	return InterpretResultOK
}
