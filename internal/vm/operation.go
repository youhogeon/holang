package vm

import (
	"fmt"
	"internal/util"
	"internal/util/log"
)

var OP_FUNCS []func(vm *VM) InterpretResult = []func(vm *VM) InterpretResult{
	(*VM).OP_RETURN,

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
}

func (vm *VM) OP_RETURN() InterpretResult {
	log.Info("OP_RETURN", log.A("stack pop!", vm.pop()))

	return InterpretResultOK
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
// UNARY
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

	switch a.(type) {
	case int64:
		switch v := b.(type) {
		case int64:
			vm.push(a.(int64) + v)
		case float64:
			vm.push(float64(a.(int64)) + v)
		case string:
			vm.push(fmt.Sprintf("%d%s", a.(int64), v))
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	case float64:
		switch v := b.(type) {
		case int64:
			vm.push(a.(float64) + float64(v))
		case float64:
			vm.push(a.(float64) + v)
		case string:
			vm.push(fmt.Sprintf("%f%s", a.(float64), v))
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	case string:
		switch v := b.(type) {
		case int64:
			vm.push(fmt.Sprintf("%s%d", a.(string), v))
		case float64:
			vm.push(fmt.Sprintf("%s%f", a.(string), v))
		case string:
			vm.push(a.(string) + v)
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
