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
	env *Environment
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		env: NewEnvironment(nil),
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

	i.env.Assign(expr.Name.Lexeme, value)

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
			func(a, b int64) any { return a / b },
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
	return &valueAndError{nil, nil}
}

func (i *Interpreter) VisitGetExpr(expr *ast.Get) any {
	return &valueAndError{nil, nil}
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
	return &valueAndError{nil, nil}
}

func (i *Interpreter) VisitSuperExpr(expr *ast.Super) any {
	return &valueAndError{nil, nil}
}

func (i *Interpreter) VisitThisExpr(expr *ast.This) any {
	return &valueAndError{nil, nil}
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
	v, err := i.env.Get(expr.Name.Lexeme)

	return &valueAndError{v, err}
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
	return nil
}

func (i *Interpreter) VisitExpressionStmt(stmt *ast.Expression) any {
	_, err := i.evaluate(stmt.Expression)

	return err
}

func (i *Interpreter) VisitFunctionStmt(stmt *ast.Function) any {
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
	return nil
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
			if _, ok := err.(*BreakSignal); ok {
				break
			}

			if _, ok := err.(*ContinueSignal); ok {
				continue
			}

			return err
		}
	}

	return nil
}

func (i *Interpreter) VisitBreakStmt(stmt *ast.Break) any {
	return &BreakSignal{}
}

func (i *Interpreter) VisitContinueStmt(stmt *ast.Continue) any {
	return &ContinueSignal{}
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
