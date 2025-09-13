package interpreter

import (
	"errors"
	"internal/ast"
	"internal/scanner"
	"internal/util/log"
)

type Resolver struct {
	interpreter  *Interpreter
	scopes       []map[string]bool
	currentFunc  FunctionType
	currentClass ClassType
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		interpreter: interpreter,
		scopes:      make([]map[string]bool, 0),
		currentFunc: NOT_FUNCTION_TYPE,
	}
}

func (r *Resolver) Resolve(statements []ast.Stmt) error {
	err := r.resolveStmts(statements)
	if err != nil {
		log.Error("Resolve error", log.E(err))
	}

	return err
}

func (r *Resolver) resolveStmts(statements []ast.Stmt) error {
	for _, stmt := range statements {
		err := stmt.Accept(r)

		if err, ok := err.(error); ok {
			return err
		}
	}

	return nil
}

func (r *Resolver) VisitAssignExpr(expr *ast.Assign) any {
	err := expr.Value.Accept(r)
	if err != nil {
		return err
	}

	r.resolveLocal(expr, expr.Name)

	return nil
}

func (r *Resolver) VisitBinaryExpr(expr *ast.Binary) any {
	err := expr.Left.Accept(r)
	if err != nil {
		return err
	}

	err = expr.Right.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitCallExpr(expr *ast.Call) any {
	err := expr.Callee.Accept(r)
	if err != nil {
		return err
	}

	for _, arg := range expr.Arguments {
		err := arg.Accept(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) VisitGetExpr(expr *ast.Get) any {
	err := expr.Object.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitGroupingExpr(expr *ast.Grouping) any {
	err := expr.Expression.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitLiteralExpr(expr *ast.Literal) any {
	return nil
}

func (r *Resolver) VisitLogicalExpr(expr *ast.Logical) any {
	err := expr.Left.Accept(r)
	if err != nil {
		return err
	}

	err = expr.Right.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitSetExpr(expr *ast.Set) any {
	err := expr.Value.Accept(r)
	if err != nil {
		return err
	}

	err = expr.Object.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitSuperExpr(expr *ast.Super) any {
	if r.currentClass == NOT_CLASS_TYPE {
		return errors.New("cannot use 'super' outside of a class")
	} else if r.currentClass != SUBCLASS {
		return errors.New("cannot use 'super' in a class with no superclass")
	}

	r.resolveLocal(expr, expr.Keyword)

	return nil
}

func (r *Resolver) VisitThisExpr(expr *ast.This) any {
	if r.currentClass == NOT_CLASS_TYPE {
		return errors.New("cannot use 'this' outside of a class")
	}

	r.resolveLocal(expr, expr.Keyword)

	return nil
}

func (r *Resolver) VisitTernaryExpr(expr *ast.Ternary) any {
	err := expr.Left.Accept(r)
	if err != nil {
		return err
	}

	err = expr.Mid.Accept(r)
	if err != nil {
		return err
	}

	err = expr.Right.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitUnaryExpr(expr *ast.Unary) any {
	err := expr.Right.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitVariableExpr(expr *ast.Variable) any {
	if len(r.scopes) != 0 {
		if defined, ok := r.scopes[len(r.scopes)-1][expr.Name.Lexeme]; ok && !defined {
			return errors.New("Cannot read local variable in its own initializer: " + expr.Name.Lexeme)
		}
	}

	r.resolveLocal(expr, expr.Name)

	return nil
}

func (r *Resolver) VisitBlockStmt(stmt *ast.Block) any {
	r.beginScope()

	err := r.resolveStmts(stmt.Statements)
	if err != nil {
		return err
	}

	r.endScope()

	return nil
}

func (r *Resolver) VisitClassStmt(stmt *ast.Class) any {
	enclosingClass := r.currentClass
	r.currentClass = CLASS

	r.declare(stmt.Name)
	r.define(stmt.Name)

	if stmt.Superclass != nil {
		if stmt.Name.Lexeme == stmt.Superclass.Name.Lexeme {
			return errors.New("a class cannot inherit from itself")
		}

		r.currentClass = SUBCLASS
		err := stmt.Superclass.Accept(r)
		if err != nil {
			return err
		}

		r.beginScope()
		r.scopes[len(r.scopes)-1]["super"] = true
	}

	r.beginScope()
	r.scopes[len(r.scopes)-1]["this"] = true

	for _, method := range stmt.Methods {
		fnType := METHOD

		if method.Name.Lexeme == "init" {
			fnType = INITIALIZER
		}

		err := r.resolveFunction(method, fnType)
		if err != nil {
			return err
		}
	}

	r.endScope()
	if stmt.Superclass != nil {
		r.endScope()
	}

	r.currentClass = enclosingClass

	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt *ast.Expression) any {
	return stmt.Expression.Accept(r)
}

func (r *Resolver) VisitFunctionStmt(stmt *ast.Function) any {
	err := r.declare(stmt.Name)
	if err != nil {
		return err
	}

	r.define(stmt.Name)

	err = r.resolveFunction(stmt, FUNCTION)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) resolveFunction(function *ast.Function, fnType FunctionType) error {
	prevFunc := r.currentFunc
	r.currentFunc = fnType
	r.beginScope()

	for _, param := range function.Params {
		err := r.declare(param)
		if err != nil {
			return err
		}

		r.define(param)
	}

	err := r.resolveStmts(function.Body)
	if err != nil {
		return err
	}

	r.endScope()
	r.currentFunc = prevFunc

	return nil
}

func (r *Resolver) VisitIfStmt(stmt *ast.If) any {
	err := stmt.Condition.Accept(r)
	if err != nil {
		return err
	}

	err = stmt.ThenBranch.Accept(r)
	if err != nil {
		return err
	}

	if stmt.ElseBranch != nil {
		err = stmt.ElseBranch.Accept(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) VisitPrintStmt(stmt *ast.Print) any {
	return stmt.Expression.Accept(r)
}

func (r *Resolver) VisitReturnStmt(stmt *ast.Return) any {
	if r.currentFunc == NOT_FUNCTION_TYPE {
		return errors.New("cannot return from top-level code")
	}

	if stmt.Value != nil {
		if r.currentFunc == INITIALIZER {
			return errors.New("cannot return a value from an initializer")
		}

		err := stmt.Value.Accept(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) VisitVarStmt(stmt *ast.Var) any {
	err := r.declare(stmt.Name)
	if err != nil {
		return err
	}

	if stmt.Initializer != nil {
		err := stmt.Initializer.Accept(r)
		if err != nil {
			return err
		}
	}

	r.define(stmt.Name)

	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *ast.While) any {
	err := stmt.Condition.Accept(r)
	if err != nil {
		return err
	}

	err = stmt.Body.Accept(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitBreakStmt(stmt *ast.Break) any {
	return nil
}

func (r *Resolver) VisitContinueStmt(stmt *ast.Continue) any {
	return nil
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(name *scanner.Token) error {
	if len(r.scopes) == 0 {
		return nil
	}

	if _, ok := r.scopes[len(r.scopes)-1][name.Lexeme]; ok {
		return errors.New("Variable with this name already declared in this scope: " + name.Lexeme)
	}

	scope := r.scopes[len(r.scopes)-1]
	scope[name.Lexeme] = false

	return nil
}

func (r *Resolver) define(name *scanner.Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes[len(r.scopes)-1]
	scope[name.Lexeme] = true
}

func (r *Resolver) resolveLocal(expr ast.Expr, name *scanner.Token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name.Lexeme]; ok {
			r.interpreter.Resolve(expr, len(r.scopes)-1-i)
			return
		}
	}
}
