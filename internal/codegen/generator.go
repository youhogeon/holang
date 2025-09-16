package codegen

import (
	"errors"
	"internal/ast"
	"internal/bytecode"
	"internal/resolver"
	"internal/token"
)

const _POPN_THRESHOLD = 3

type CodeGenerator struct {
	em       Emitter
	bindings resolver.BindingResult
}

func NewCodeGenerator(em Emitter, bindings resolver.BindingResult) *CodeGenerator {
	return &CodeGenerator{
		em:       em,
		bindings: bindings,
	}
}

func (g *CodeGenerator) Generate(statements []ast.Stmt) error {
	for _, stmt := range statements {
		err := g.genStmt(stmt)

		if err != nil {
			return err
		}
	}

	g._emit(bytecode.OP_RETURN)

	return nil
}

func (g *CodeGenerator) genExpr(e ast.Expr) error {
	if result := e.Accept(g); result != nil {
		return result.(error)
	}

	return nil
}

func (g *CodeGenerator) genStmt(s ast.Stmt) error {
	if result := s.Accept(g); result != nil {
		return result.(error)
	}

	return nil
}

func (g *CodeGenerator) _emit(op bytecode.OpCode, operands ...int64) int {
	return g.em.Emit(token.Offset{Line: -1, Index: -1}, op, operands...)
}

func (g *CodeGenerator) emit(offset token.Offset, op bytecode.OpCode, operands ...int64) int {
	return g.em.Emit(offset, op, operands...)
}

func (g *CodeGenerator) emitJump(offset token.Offset, op bytecode.OpCode) int {
	return g.em.EmitJump(offset, op)
}

func (g *CodeGenerator) patchJump(at int) {
	g.em.PatchJump(at)
}

func (g *CodeGenerator) emitConstant(offset token.Offset, value bytecode.Value) {
	switch v := value.(type) {
	case nil:
		g.em.Emit(offset, bytecode.OP_NIL)

	case bool:
		if v {
			g.em.Emit(offset, bytecode.OP_TRUE)
		} else {
			g.em.Emit(offset, bytecode.OP_FALSE)
		}

	case int64:
		switch v {
		case -1:
			g.em.Emit(offset, bytecode.OP_CONSTANT_M1)
		case 0:
			g.em.Emit(offset, bytecode.OP_CONSTANT_0)
		case 1:
			g.em.Emit(offset, bytecode.OP_CONSTANT_1)
		case 2:
			g.em.Emit(offset, bytecode.OP_CONSTANT_2)
		case 3:
			g.em.Emit(offset, bytecode.OP_CONSTANT_3)
		case 4:
			g.em.Emit(offset, bytecode.OP_CONSTANT_4)
		case 5:
			g.em.Emit(offset, bytecode.OP_CONSTANT_5)
		default:
			g.em.Emit(offset, bytecode.OP_CONSTANT, g.makeConstant(value))
		}

	default:
		g.em.Emit(offset, bytecode.OP_CONSTANT, g.makeConstant(value))
	}
}

func (g *CodeGenerator) makeConstant(value bytecode.Value) int64 {
	return g.em.MakeConstant(value)
}

// ================================================================
// Expr
// ================================================================

func (g *CodeGenerator) VisitAssignExpr(expr *ast.Assign) any {
	binding := g.bindings[expr.NodeId]

	if err := expr.Value.Accept(g); err != nil {
		return err
	}

	if binding.Kind == resolver.BindLocal {
		g.emit(expr.Offset, bytecode.OP_SET_LOCAL, int64(binding.Slot))

		return nil
	}

	constant := g.makeConstant(expr.Name.Lexeme)
	g.emit(expr.Offset, bytecode.OP_SET_GLOBAL, constant)

	return nil
}

func (g *CodeGenerator) VisitBinaryExpr(expr *ast.Binary) any {
	if err := expr.Left.Accept(g); err != nil {
		return err
	}

	if err := expr.Right.Accept(g); err != nil {
		return err
	}

	switch expr.Operator.TokenType {
	case token.PLUS:
		g.emit(expr.Offset, bytecode.OP_ADD)
	case token.MINUS:
		g.emit(expr.Offset, bytecode.OP_SUBTRACT)
	case token.STAR:
		g.emit(expr.Offset, bytecode.OP_MULTIPLY)
	case token.SLASH:
		g.emit(expr.Offset, bytecode.OP_DIVIDE)
	case token.GREATER:
		g.emit(expr.Offset, bytecode.OP_GREATER)
	case token.GREATER_EQUAL:
		g.emit(expr.Offset, bytecode.OP_GREATER_EQUAL)
	case token.LESS:
		g.emit(expr.Offset, bytecode.OP_LESS)
	case token.LESS_EQUAL:
		g.emit(expr.Offset, bytecode.OP_LESS_EQUAL)
	case token.EQUAL_EQUAL:
		g.emit(expr.Offset, bytecode.OP_EQUAL)
	case token.BANG_EQUAL:
		g.emit(expr.Offset, bytecode.OP_NOT_EQUAL)
	default:
		return errors.New("unknown binary operator: " + expr.Operator.Lexeme)
	}

	return nil
}

func (g *CodeGenerator) VisitCallExpr(expr *ast.Call) any {
	return nil
}

func (g *CodeGenerator) VisitGetExpr(expr *ast.Get) any {
	return nil
}

func (g *CodeGenerator) VisitGroupingExpr(expr *ast.Grouping) any {
	return expr.Expression.Accept(g)
}

func (g *CodeGenerator) VisitLiteralExpr(expr *ast.Literal) any {
	g.emitConstant(expr.Offset, expr.Value)

	return nil
}

func (g *CodeGenerator) VisitLogicalExpr(expr *ast.Logical) any {
	if err := expr.Left.Accept(g); err != nil {
		return err
	}

	if expr.Operator.TokenType == token.OR {
		elseJump := g.emitJump(expr.Offset, bytecode.OP_JUMP_IF_FALSE)
		endJump := g.emitJump(expr.Offset, bytecode.OP_JUMP)

		g.patchJump(elseJump)
		g.emit(expr.Offset, bytecode.OP_POP)

		if err := expr.Right.Accept(g); err != nil {
			return err
		}

		g.patchJump(endJump)
	} else {
		endJump := g.emitJump(expr.Offset, bytecode.OP_JUMP_IF_FALSE)

		g.emit(expr.Offset, bytecode.OP_POP)
		if err := expr.Right.Accept(g); err != nil {
			return err
		}

		g.patchJump(endJump)
	}

	return nil
}

func (g *CodeGenerator) VisitSetExpr(expr *ast.Set) any {
	return nil
}

func (g *CodeGenerator) VisitSuperExpr(expr *ast.Super) any {
	return nil
}

func (g *CodeGenerator) VisitThisExpr(expr *ast.This) any {
	return nil
}

func (g *CodeGenerator) VisitTernaryExpr(expr *ast.Ternary) any {
	return nil
}

func (g *CodeGenerator) VisitUnaryExpr(expr *ast.Unary) any {
	if err := expr.Right.Accept(g); err != nil {
		return err
	}

	switch expr.Operator.TokenType {
	case token.MINUS:
		g.emit(expr.Offset, bytecode.OP_NEGATE)

	case token.BANG:
		g.emit(expr.Offset, bytecode.OP_NOT)

	default:
		return errors.New("unknown unary operator: " + expr.Operator.Lexeme)
	}

	return nil
}

func (g *CodeGenerator) VisitVariableExpr(expr *ast.Variable) any {
	binding := g.bindings[expr.NodeId]

	if binding.Kind == resolver.BindLocal {
		g.emit(expr.Offset, bytecode.OP_GET_LOCAL, int64(binding.Slot))

		return nil
	}

	constant := g.makeConstant(expr.Name.Lexeme)
	g.emit(expr.Offset, bytecode.OP_GET_GLOBAL, constant)

	return nil
}

// ================================================================
// Stmt
// ================================================================

func (g *CodeGenerator) VisitBlockStmt(stmt *ast.Block) any {
	for _, s := range stmt.Statements {
		if err := g.genStmt(s); err != nil {
			return err
		}
	}

	popCnt := g.bindings[stmt.NodeId].Slot

	if popCnt >= _POPN_THRESHOLD {
		g.emit(stmt.Offset, bytecode.OP_POPN, int64(popCnt))

		return nil
	}

	for range popCnt {
		g.emit(stmt.Offset, bytecode.OP_POP)
	}

	return nil
}

func (g *CodeGenerator) VisitClassStmt(stmt *ast.Class) any {
	return nil
}

func (g *CodeGenerator) VisitExpressionStmt(stmt *ast.Expression) any {
	if err := stmt.Expression.Accept(g); err != nil {
		return err
	}

	g.emit(stmt.Offset, bytecode.OP_POP)

	return nil
}

func (g *CodeGenerator) VisitFunctionStmt(stmt *ast.Function) any {
	return nil
}

func (g *CodeGenerator) VisitIfStmt(stmt *ast.If) any {
	if err := stmt.Condition.Accept(g); err != nil {
		return err
	}

	elseJump := g.emitJump(stmt.Offset, bytecode.OP_JUMP_IF_FALSE)
	g.emit(stmt.Offset, bytecode.OP_POP)

	if err := stmt.ThenBranch.Accept(g); err != nil {
		return err
	}

	endJump := g.emitJump(stmt.Offset, bytecode.OP_JUMP)
	g.patchJump(elseJump)
	g.emit(stmt.Offset, bytecode.OP_POP)

	if stmt.ElseBranch != nil {
		if err := stmt.ElseBranch.Accept(g); err != nil {
			return err
		}
	}

	g.patchJump(endJump)

	return nil
}

func (g *CodeGenerator) VisitPrintStmt(stmt *ast.Print) any {
	if err := stmt.Expression.Accept(g); err != nil {
		return err
	}

	g.emit(stmt.Offset, bytecode.OP_PRINT)

	return nil
}

func (g *CodeGenerator) VisitReturnStmt(stmt *ast.Return) any {
	g.emit(stmt.Offset, bytecode.OP_RETURN)

	return nil
}

func (g *CodeGenerator) VisitVarStmt(stmt *ast.Var) any {
	binding := g.bindings[stmt.NodeId]

	if stmt.Initializer == nil {
		g.emit(stmt.Offset, bytecode.OP_NIL)
	} else {
		// initializer에 의해 stack에 값이 올라감
		if err := stmt.Initializer.Accept(g); err != nil {
			return err
		}
	}

	if binding.Kind == resolver.BindLocal {
		// stack에 값이 있으므로, OP_SET_LOCAL 불필요
		return nil
	}

	constant := g.makeConstant(stmt.Name.Lexeme)
	g.emit(stmt.Offset, bytecode.OP_DEFINE_GLOBAL, constant)

	return nil
}

func (g *CodeGenerator) VisitWhileStmt(stmt *ast.While) any {
	return nil
}

func (g *CodeGenerator) VisitBreakStmt(stmt *ast.Break) any {
	return nil
}

func (g *CodeGenerator) VisitContinueStmt(stmt *ast.Continue) any {
	return nil
}
