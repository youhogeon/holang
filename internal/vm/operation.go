package vm

import (
	"fmt"
	"internal/util/log"
)

var OP_FUNCS []func(vm *VM) InterpretResult = []func(vm *VM) InterpretResult{
	(*VM).OP_RETURN,
	(*VM).OP_CONSTANT,
	(*VM).OP_NEGATE,
	(*VM).OP_ADD,
	(*VM).OP_SUBTRACT,
	(*VM).OP_MULTIPLY,
	(*VM).OP_DIVIDE,
}

func (vm *VM) OP_RETURN() InterpretResult {
	log.Info("OP_RETURN", log.A("stack pop!", vm.pop()))

	return InterpretResultOK
}

func (vm *VM) OP_CONSTANT() InterpretResult {
	constant := vm.getConstant()

	vm.push(constant)

	return InterpretResultOK
}

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
	b := vm.pop()
	a := vm.pop()

	switch a.(type) {
	case int64:
		switch v := b.(type) {
		case int64:
			vm.push(a.(int64) - v)
		case float64:
			vm.push(float64(a.(int64)) - v)
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	case float64:
		switch v := b.(type) {
		case int64:
			vm.push(a.(float64) - float64(v))
		case float64:
			vm.push(a.(float64) - v)
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	default:
		log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

		return InterpretResultRuntimeError
	}

	return InterpretResultOK
}

func (vm *VM) OP_MULTIPLY() InterpretResult {
	b := vm.pop()
	a := vm.pop()

	switch a.(type) {
	case int64:
		switch v := b.(type) {
		case int64:
			vm.push(a.(int64) * v)
		case float64:
			vm.push(float64(a.(int64)) * v)
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	case float64:
		switch v := b.(type) {
		case int64:
			vm.push(a.(float64) * float64(v))
		case float64:
			vm.push(a.(float64) * v)
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	default:
		log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

		return InterpretResultRuntimeError
	}

	return InterpretResultOK
}

func (vm *VM) OP_DIVIDE() InterpretResult {
	b := vm.pop()
	a := vm.pop()

	switch a.(type) {
	case int64:
		switch v := b.(type) {
		case int64:
			if v == 0 {
				log.Error("Division by zero", log.A("a", a), log.A("b", b))

				return InterpretResultRuntimeError
			}
			vm.push(a.(int64) / v)
		case float64:
			if v == 0.0 {
				log.Error("Division by zero", log.A("a", a), log.A("b", b))

				return InterpretResultRuntimeError
			}
			vm.push(float64(a.(int64)) / v)
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	case float64:
		switch v := b.(type) {
		case int64:
			if v == 0 {
				log.Error("Division by zero", log.A("a", a), log.A("b", b))

				return InterpretResultRuntimeError
			}
			vm.push(a.(float64) / float64(v))
		case float64:
			if v == 0.0 {
				log.Error("Division by zero", log.A("a", a), log.A("b", b))

				return InterpretResultRuntimeError
			}
			vm.push(a.(float64) / v)
		default:
			log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

			return InterpretResultRuntimeError
		}
	default:
		log.Error("Operand must be a number", log.A("a", a), log.A("b", b))

		return InterpretResultRuntimeError
	}

	return InterpretResultOK
}
