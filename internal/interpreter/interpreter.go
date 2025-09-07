package interpreter

import (
	"fmt"
	"internal/ast"
	"internal/scanner"
)

type valueAndError struct {
	value any
	err   error
}

type Interpreter struct {
	env     *Environment
	globals *Environment
	locals  map[ast.Expr]int
}

func NewInterpreter() *Interpreter {
	globals := NewEnvironment(nil)

	globals.Define("print", &BuiltInFnPrint{})
	globals.Define("input", &BuiltInFnInput{})
	globals.Define("clock", &BuiltInFnClock{})
	globals.Define("str", &BuiltInFnToString{})
	globals.Define("int", &BuiltInFnToInt{})
	globals.Define("float", &BuiltInFnToFloat{})
	globals.Define("rand", &BuiltInFnRand{})
	globals.Define("randInt", &BuiltInFnRandInt{})
	globals.Define("sleep", &BuiltInFnSleep{})
	globals.Define("clear", &BuiltInFnClear{})
	globals.Define("strlen", &BuiltInFnStrLen{})
	globals.Define("substring", &BuiltInFnSubstring{})
	globals.Define("getch", &BuiltInFnGetch{})

	return &Interpreter{
		env:     globals,
		globals: globals,
		locals:  make(map[ast.Expr]int),
	}
}

func (i *Interpreter) Interpret(program []ast.Stmt) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewRuntimeErrorWithLog(fmt.Sprint(r))
		}
	}()

	for _, stmt := range program {
		err = i.execute(stmt)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) Resolve(expr ast.Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) execute(stmt ast.Stmt) error {
	result := stmt.Accept(i)

	if result == nil {
		return nil
	}

	return result.(error)
}

func (i *Interpreter) evaluate(expr ast.Expr) (any, error) {
	if v, ok := expr.Accept(i).(*valueAndError); ok {
		return v.value, v.err
	}

	return nil, NewRuntimeErrorWithLog("interpreter error")
}

func (i *Interpreter) VisitAssignExpr(expr *ast.Assign) any {
	value, err := i.evaluate(expr.Value)

	if err != nil {
		return &valueAndError{nil, err}
	}

	if distance, ok := i.locals[expr]; ok {
		err := i.env.AssignAt(distance, expr.Name.Lexeme, value)
		return &valueAndError{value, err}
	}

	err = i.globals.Assign(expr.Name.Lexeme, value)
	if err != nil {
		return &valueAndError{nil, err}
	}
	return &valueAndError{value, nil}
}

func (i *Interpreter) VisitBinaryExpr(expr *ast.Binary) any {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return &valueAndError{nil, err}
	}

	right, err := i.evaluate(expr.Right)
	if err != nil {
		return &valueAndError{nil, err}
	}

	switch expr.Operator.TokenType {
	case scanner.PLUS:
		if ls, ok := left.(string); ok {
			if rs, ok := right.(string); ok {
				return &valueAndError{ls + rs, nil}
			}

			return &valueAndError{nil, NewRuntimeErrorWithLog("can only concatenate string to string")}
		}

		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return a + b },
			func(a, b float64) any { return a + b },
		)
	case scanner.MINUS:
		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return a - b },
			func(a, b float64) any { return a - b },
		)
	case scanner.STAR:
		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return a * b },
			func(a, b float64) any { return a * b },
		)
	case scanner.SLASH:
		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return float64(a) / float64(b) },
			func(a, b float64) any { return a / b },
		)
	case scanner.GREATER:
		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return a > b },
			func(a, b float64) any { return a > b },
		)
	case scanner.GREATER_EQUAL:
		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return a >= b },
			func(a, b float64) any { return a >= b },
		)
	case scanner.LESS:
		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return a < b },
			func(a, b float64) any { return a < b },
		)
	case scanner.LESS_EQUAL:
		return binaryNumericOp(
			left, right,
			func(a, b int64) any { return a <= b },
			func(a, b float64) any { return a <= b },
		)
	case scanner.EQUAL_EQUAL:
		return &valueAndError{left == right, nil}
	case scanner.BANG_EQUAL:
		return &valueAndError{left != right, nil}
	}

	return &valueAndError{nil, NewRuntimeErrorWithLog("unknown binary operator")}
}

func (i *Interpreter) VisitCallExpr(expr *ast.Call) any {
	callee, err := i.evaluate(expr.Callee)
	if err != nil {
		return &valueAndError{nil, err}
	}

	var arguments []any
	for _, argExpr := range expr.Arguments {
		arg, err := i.evaluate(argExpr)
		if err != nil {
			return &valueAndError{nil, err}
		}
		arguments = append(arguments, arg)
	}

	if function, ok := callee.(Callable); ok {
		if len(arguments) != function.Arity() {
			return &valueAndError{nil, NewRuntimeErrorWithLog(fmt.Sprintf("expected %d arguments but got %d", function.Arity(), len(arguments)))}
		}

		value, err := function.Call(i, arguments)

		return &valueAndError{value, err}
	}

	return &valueAndError{nil, NewRuntimeErrorWithLog("can only call functions and classes")}

}

func (i *Interpreter) VisitGetExpr(expr *ast.Get) any {
	object, err := i.evaluate(expr.Object)
	if err != nil {
		return &valueAndError{nil, err}
	}

	if instance, ok := object.(*Instance); ok {
		value, err := instance.get(expr.Name.Lexeme)

		return &valueAndError{value, err}
	}

	return &valueAndError{nil, NewRuntimeErrorWithLog("only instances have properties")}
}

func (i *Interpreter) VisitGroupingExpr(expr *ast.Grouping) any {
	v, err := i.evaluate(expr.Expression)

	return &valueAndError{v, err}
}

func (i *Interpreter) VisitLiteralExpr(expr *ast.Literal) any {
	return &valueAndError{expr.Value, nil}
}

func (i *Interpreter) VisitLogicalExpr(expr *ast.Logical) any {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return &valueAndError{nil, err}
	}

	if expr.Operator.TokenType == scanner.OR {
		if isTruthy(left) {
			return &valueAndError{left, nil}
		}
	} else {
		if !isTruthy(left) {
			return &valueAndError{left, nil}
		}
	}

	right, err := i.evaluate(expr.Right)
	if err != nil {
		return &valueAndError{nil, err}
	}

	return &valueAndError{right, nil}
}

func (i *Interpreter) VisitSetExpr(expr *ast.Set) any {
	value, err := i.evaluate(expr.Value)
	if err != nil {
		return &valueAndError{nil, err}
	}

	object, err := i.evaluate(expr.Object)
	if err != nil {
		return &valueAndError{nil, err}
	}

	if instance, ok := object.(*Instance); ok {
		instance.set(expr.Name.Lexeme, value)

		return &valueAndError{value, nil}
	}

	return &valueAndError{nil, NewRuntimeErrorWithLog("only instances have fields")}
}

func (i *Interpreter) VisitSuperExpr(expr *ast.Super) any {
	distance := i.locals[expr]
	superclass, err := i.env.GetAt(distance, "super")
	if err != nil {
		return &valueAndError{nil, err}
	}

	object, err := i.env.GetAt(distance-1, "this")
	if err != nil {
		return &valueAndError{nil, err}
	}

	cls, ok := superclass.(*Class)
	if !ok {
		return &valueAndError{nil, NewRuntimeErrorWithLog("super must be a class")}
	}

	instance, ok := object.(*Instance)
	if !ok {
		return &valueAndError{nil, NewRuntimeErrorWithLog("this must be an instance")}
	}

	method := cls.findMethod(expr.Method.Lexeme)
	if method == nil {
		return &valueAndError{nil, NewRuntimeErrorWithLog("undefined property: " + expr.Method.Lexeme)}
	}

	return &valueAndError{method.bind(instance), nil}
}

func (i *Interpreter) VisitThisExpr(expr *ast.This) any {
	v, err := i.lookupVariable(expr.Keyword, expr)
	return &valueAndError{v, err}
}

func (i *Interpreter) VisitTernaryExpr(expr *ast.Ternary) any {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return &valueAndError{nil, err}
	}

	if isTruthy(left) {
		v, err := i.evaluate(expr.Mid)

		return &valueAndError{v, err}
	}

	v, err := i.evaluate(expr.Right)
	return &valueAndError{v, err}
}

func (i *Interpreter) VisitUnaryExpr(expr *ast.Unary) any {
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return &valueAndError{nil, err}
	}

	switch expr.Operator.TokenType {
	case scanner.MINUS:
		if v, ok := right.(int64); ok {
			return &valueAndError{-v, nil}
		} else if v, ok := right.(float64); ok {
			return &valueAndError{-v, nil}
		}

		return &valueAndError{nil, NewRuntimeErrorWithLog("operand must be a number")}
	case scanner.BANG:
		return &valueAndError{!isTruthy(right), nil}
	}

	return &valueAndError{nil, NewRuntimeErrorWithLog("unknown unary operator")}
}

func (i *Interpreter) VisitVariableExpr(expr *ast.Variable) any {
	v, err := i.lookupVariable(expr.Name, expr)
	return &valueAndError{v, err}
}

func (i *Interpreter) lookupVariable(name *scanner.Token, expr ast.Expr) (any, error) {
	if distance, ok := i.locals[expr]; ok {
		return i.env.GetAt(distance, name.Lexeme)
	}

	return i.globals.Get(name.Lexeme)
}

func (i *Interpreter) VisitBlockStmt(stmt *ast.Block) any {
	return i.executeBlock(stmt.Statements, NewEnvironment(i.env))
}

func (i *Interpreter) executeBlock(stmt []ast.Stmt, environment *Environment) error {
	prevEnv := i.env
	defer func() { i.env = prevEnv }()

	i.env = environment

	for _, s := range stmt {
		err := i.execute(s)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) VisitClassStmt(stmt *ast.Class) any {
	var superclass *Class
	if stmt.Superclass != nil {
		v, err := i.evaluate(stmt.Superclass)
		if err != nil {
			return err
		}

		if sc, ok := v.(*Class); ok {
			superclass = sc
		} else {
			return NewRuntimeErrorWithLog("superclass must be a class")
		}
	}

	i.env.Define(stmt.Name.Lexeme, nil)

	if stmt.Superclass != nil {
		i.env = NewEnvironment(i.env)
		i.env.Define("super", superclass)
	}

	methods := make(map[string]*Function)
	for _, method := range stmt.Methods {
		function := &Function{
			declaration:   method,
			clousure:      i.env,
			isInitializer: method.Name.Lexeme == "init",
		}
		methods[method.Name.Lexeme] = function
	}

	cls := &Class{
		name:       stmt.Name.Lexeme,
		methods:    methods,
		superclass: superclass,
	}

	if stmt.Superclass != nil {
		i.env = i.env.enclosing
	}

	i.env.Assign(stmt.Name.Lexeme, cls)

	return nil
}

func (i *Interpreter) VisitExpressionStmt(stmt *ast.Expression) any {
	_, err := i.evaluate(stmt.Expression)

	return err
}

func (i *Interpreter) VisitFunctionStmt(stmt *ast.Function) any {
	function := &Function{declaration: stmt, clousure: i.env}
	i.env.Define(stmt.Name.Lexeme, function)

	return nil
}

func (i *Interpreter) VisitIfStmt(stmt *ast.If) any {
	value, err := i.evaluate(stmt.Condition)

	if err != nil {
		return err
	}

	if isTruthy(value) {
		err := i.execute(stmt.ThenBranch)

		if err != nil {
			return err
		}
	} else if stmt.ElseBranch != nil {
		err := i.execute(stmt.ElseBranch)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) VisitPrintStmt(stmt *ast.Print) any {
	value, err := i.evaluate(stmt.Expression)

	if err != nil {
		return err
	}

	fmt.Println(value)

	return nil
}

func (i *Interpreter) VisitReturnStmt(stmt *ast.Return) any {
	if stmt.Value == nil {
		return &returnSignal{value: nil}
	}

	value, err := i.evaluate(stmt.Value)

	if err != nil {
		return err
	}

	return &returnSignal{value: value}
}

func (i *Interpreter) VisitVarStmt(stmt *ast.Var) any {
	var value any

	if stmt.Initializer != nil {
		var err error

		value, err = i.evaluate(stmt.Initializer)

		if err != nil {
			return err
		}
	}

	i.env.Define(stmt.Name.Lexeme, value)

	return nil
}

func (i *Interpreter) VisitWhileStmt(stmt *ast.While) any {
	for {
		value, err := i.evaluate(stmt.Condition)

		if err != nil {
			return err
		}

		if !isTruthy(value) {
			break
		}

		err = i.execute(stmt.Body)
		if err != nil {
			if _, ok := err.(*breakSignal); ok {
				break
			}

			if _, ok := err.(*continueSignal); ok {
				continue
			}

			return err
		}
	}

	return nil
}

func (i *Interpreter) VisitBreakStmt(stmt *ast.Break) any {
	return &breakSignal{}
}

func (i *Interpreter) VisitContinueStmt(stmt *ast.Continue) any {
	return &continueSignal{}
}

func isTruthy(value any) bool {
	if value == nil {
		return false
	}

	if b, ok := value.(bool); ok {
		return b
	}

	return true
}

func binaryNumericOp(left, right any, opInt func(int64, int64) any, opFloat func(float64, float64) any) *valueAndError {
	lInt, lIsInt := left.(int64)
	rInt, rIsInt := right.(int64)
	lFloat, lIsFloat := left.(float64)
	rFloat, rIsFloat := right.(float64)

	// int64 + int64
	if lIsInt && rIsInt {
		return &valueAndError{opInt(lInt, rInt), nil}
	}

	// float64가 하나라도 있으면 float64로 변환
	var lf, rf float64
	if lIsInt {
		lf = float64(lInt)
	} else if lIsFloat {
		lf = lFloat
	} else {
		return &valueAndError{nil, NewRuntimeErrorWithLog("operand must be a int or float")}
	}

	if rIsInt {
		rf = float64(rInt)
	} else if rIsFloat {
		rf = rFloat
	} else {
		return &valueAndError{nil, NewRuntimeErrorWithLog("operand must be a int or float")}
	}

	return &valueAndError{opFloat(lf, rf), nil}
}
