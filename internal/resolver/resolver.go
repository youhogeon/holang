package resolver

import (
	"errors"
	"internal/ast"
	"internal/token"
	"internal/util/log"
)

// ================================================================
// Bindings (Result)
// ================================================================

type BindingKind uint8

const (
	BindGlobal BindingKind = iota
	BindLocal
)

type Binding struct {
	Kind BindingKind
	Slot int // Local 슬롯 인덱스. Global이면 -1
}

// ================================================================
// Resolver
// ================================================================

type Local struct {
	Name  *token.Token
	Depth int // -1: 초기화 안됨
	Slot  int
}

type Resolver struct {
	locals     []Local
	nextSlot   int
	scopeDepth int

	bindings map[int]Binding
	errors   []error
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) clear() {
	r.locals = r.locals[:0]
	r.nextSlot = 0
	r.scopeDepth = 0

	r.bindings = make(map[int]Binding)
	r.errors = r.errors[:0]
}

func (r *Resolver) Resolve(statements []ast.Stmt) (map[int]Binding, []error) {
	r.clear()

	for _, stmt := range statements {
		stmt.Accept(r)
	}

	return r.bindings, r.errors
}

func (r *Resolver) beginScope() {
	r.scopeDepth++
}

func (r *Resolver) endScope() {
	if r.scopeDepth == 0 {
		return
	}

	for len(r.locals) > 0 && r.locals[len(r.locals)-1].Depth >= r.scopeDepth {
		r.locals = r.locals[:len(r.locals)-1]
	}

	r.scopeDepth--
}

func (r *Resolver) declare(name *token.Token) (slot int, isLocal bool) {
	if r.scopeDepth == 0 {
		return -1, false
	}

	for i := len(r.locals) - 1; i >= 0; i-- {
		local := r.locals[i]

		if local.Depth < r.scopeDepth && local.Depth >= 0 {
			break
		}

		if local.Name.Lexeme == name.Lexeme {
			r.addError("Variable with this name already declared in this scope: " + name.Lexeme)

			return
		}
	}

	slot = r.addLocal(name)

	return slot, true
}

func (r *Resolver) define(name *token.Token) {
	if r.scopeDepth == 0 {
		return
	}

	for i := len(r.locals) - 1; i >= 0; i-- {
		local := r.locals[i]

		if local.Depth < r.scopeDepth && local.Depth >= 0 {
			break
		}

		if local.Name.Lexeme == name.Lexeme {
			local.Depth = r.scopeDepth
			r.locals[i] = local
			return
		}
	}
}

func (r *Resolver) findLocal(name *token.Token) (int, bool) {
	for i := len(r.locals) - 1; i >= 0; i-- {
		local := r.locals[i]

		if local.Name.Lexeme != name.Lexeme {
			continue
		}

		if local.Depth == -1 {
			r.addError("Can't read local variable in its own initializer: " + name.Lexeme)

			return -1, false
		}

		return local.Slot, true
	}

	return -1, false
}

func (r *Resolver) addLocal(name *token.Token) (slot int) {
	slot = r.nextSlot

	r.locals = append(r.locals, Local{
		Name:  name,
		Depth: -1,
		Slot:  slot,
	})

	r.nextSlot++

	return
}

func (r *Resolver) addError(message string) {
	err := errors.New(message)

	r.errors = append(r.errors, err)
	log.Error("Resolver error", log.E(err))
}

// ================================================================
// Expr
// ================================================================

func (r *Resolver) VisitAssignExpr(expr *ast.Assign) any {
	expr.Value.Accept(r)

	if slot, ok := r.findLocal(expr.Name); ok {
		r.bindings[expr.NodeId] = Binding{
			Kind: BindLocal,
			Slot: slot,
		}
	} else {
		r.bindings[expr.NodeId] = Binding{
			Kind: BindGlobal,
			Slot: -1,
		}
	}

	return nil
}

func (r *Resolver) VisitBinaryExpr(expr *ast.Binary) any {
	expr.Left.Accept(r)
	expr.Right.Accept(r)

	return nil
}

func (r *Resolver) VisitCallExpr(expr *ast.Call) any {
	expr.Callee.Accept(r)
	for _, arg := range expr.Arguments {
		arg.Accept(r)
	}

	return nil
}

func (r *Resolver) VisitGetExpr(expr *ast.Get) any {
	expr.Object.Accept(r)

	return nil
}

func (r *Resolver) VisitGroupingExpr(expr *ast.Grouping) any {
	expr.Expression.Accept(r)

	return nil
}

func (r *Resolver) VisitLiteralExpr(expr *ast.Literal) any {
	return nil
}

func (r *Resolver) VisitLogicalExpr(expr *ast.Logical) any {
	expr.Left.Accept(r)
	expr.Right.Accept(r)

	return nil
}

func (r *Resolver) VisitSetExpr(expr *ast.Set) any {
	expr.Object.Accept(r)
	expr.Value.Accept(r)

	return nil
}

func (r *Resolver) VisitSuperExpr(expr *ast.Super) any {
	return nil
}

func (r *Resolver) VisitThisExpr(expr *ast.This) any {
	return nil
}

func (r *Resolver) VisitTernaryExpr(expr *ast.Ternary) any {
	expr.Left.Accept(r)
	expr.Mid.Accept(r)
	expr.Right.Accept(r)

	return nil
}

func (r *Resolver) VisitUnaryExpr(expr *ast.Unary) any {
	expr.Right.Accept(r)

	return nil
}

func (r *Resolver) VisitVariableExpr(expr *ast.Variable) any {
	if slot, ok := r.findLocal(expr.Name); ok {
		r.bindings[expr.NodeId] = Binding{
			Kind: BindLocal,
			Slot: slot,
		}
	} else {
		r.bindings[expr.NodeId] = Binding{
			Kind: BindGlobal,
			Slot: -1,
		}
	}

	return nil
}

// ================================================================
// Stmt
// ================================================================

func (r *Resolver) VisitBlockStmt(stmt *ast.Block) any {
	r.beginScope()

	for _, s := range stmt.Statements {
		s.Accept(r)
	}

	r.endScope()

	return nil
}

func (r *Resolver) VisitClassStmt(stmt *ast.Class) any {
	if stmt.Superclass != nil {
		stmt.Superclass.Accept(r)
	}

	for _, method := range stmt.Methods {
		method.Accept(r)
	}

	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt *ast.Expression) any {
	stmt.Expression.Accept(r)

	return nil
}

func (r *Resolver) VisitFunctionStmt(stmt *ast.Function) any {
	for _, bodyStmt := range stmt.Body {
		bodyStmt.Accept(r)
	}

	return nil
}

func (r *Resolver) VisitIfStmt(stmt *ast.If) any {
	stmt.Condition.Accept(r)
	stmt.ThenBranch.Accept(r)
	if stmt.ElseBranch != nil {
		stmt.ElseBranch.Accept(r)
	}

	return nil
}

func (r *Resolver) VisitPrintStmt(stmt *ast.Print) any {
	stmt.Expression.Accept(r)

	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *ast.Return) any {
	if stmt.Value != nil {
		stmt.Value.Accept(r)
	}

	return nil
}

func (r *Resolver) VisitVarStmt(stmt *ast.Var) any {
	slot, isLocal := r.declare(stmt.Name)

	if isLocal {
		r.bindings[stmt.NodeId] = Binding{Kind: BindLocal, Slot: slot}
	} else {
		r.bindings[stmt.NodeId] = Binding{Kind: BindGlobal, Slot: -1}
	}

	if stmt.Initializer != nil {
		stmt.Initializer.Accept(r)
	}

	r.define(stmt.Name)

	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *ast.While) any {
	stmt.Condition.Accept(r)
	stmt.Body.Accept(r)

	return nil
}

func (r *Resolver) VisitBreakStmt(stmt *ast.Break) any {
	return nil
}

func (r *Resolver) VisitContinueStmt(stmt *ast.Continue) any {
	return nil
}
