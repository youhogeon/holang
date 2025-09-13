package codegen

import (
	"errors"
	"internal/ast"
	"internal/bytecode"
	"internal/scanner"
)

type CodeGenerator struct {
	em Emitter
}

func NewCodeGenerator(em Emitter) *CodeGenerator {
	return &CodeGenerator{
		em: em,
	}
}

func (g *CodeGenerator) Generate(statements []ast.Stmt) error {
	for _, stmt := range statements {
		err := g.genStmt(stmt)

		if err != nil {
			return err
		}
	}

	g.emit(ast.Offset{}, bytecode.OP_RETURN)

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

func (g *CodeGenerator) emit(offset ast.Offset, op bytecode.OpCode, operands ...int64) {
	g.em.Emit(bytecode.Offset(offset), op, operands...)
}

func (g *CodeGenerator) emitConstant(offset ast.Offset, value bytecode.Value) {
	g.em.EmitConstant(bytecode.Offset(offset), value)
}

// ================================================================
// Expr
// ================================================================

func (g *CodeGenerator) VisitAssignExpr(expr *ast.Assign) any {
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
	case scanner.PLUS:
		g.emit(expr.Offset, bytecode.OP_ADD)

	case scanner.MINUS:
		g.emit(expr.Offset, bytecode.OP_SUBTRACT)

	case scanner.STAR:
		g.emit(expr.Offset, bytecode.OP_MULTIPLY)

	case scanner.SLASH:
		g.emit(expr.Offset, bytecode.OP_DIVIDE)

	case scanner.GREATER:
		return nil
	case scanner.GREATER_EQUAL:
		return nil
	case scanner.LESS:
		return nil
	case scanner.LESS_EQUAL:
		return nil
	case scanner.EQUAL_EQUAL:
		return nil
	case scanner.BANG_EQUAL:
		return nil
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
	return nil
}

func (g *CodeGenerator) VisitLiteralExpr(expr *ast.Literal) any {
	g.emitConstant(expr.Offset, expr.Value)

	return nil
}

func (g *CodeGenerator) VisitLogicalExpr(expr *ast.Logical) any {
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
	case scanner.MINUS:
		g.emit(expr.Offset, bytecode.OP_NEGATE)

	case scanner.BANG:
		g.emit(expr.Offset, bytecode.OP_NOT)

	default:
		return errors.New("unknown unary operator: " + expr.Operator.Lexeme)
	}

	return nil
}

func (g *CodeGenerator) VisitVariableExpr(expr *ast.Variable) any {
	return nil
}

// ================================================================
// Stmt
// ================================================================

func (g *CodeGenerator) VisitBlockStmt(stmt *ast.Block) any {
	return nil
}

func (g *CodeGenerator) VisitClassStmt(stmt *ast.Class) any {
	return nil
}

func (g *CodeGenerator) VisitExpressionStmt(stmt *ast.Expression) any {
	return stmt.Expression.Accept(g)
}

func (g *CodeGenerator) VisitFunctionStmt(stmt *ast.Function) any {
	return nil
}

func (g *CodeGenerator) VisitIfStmt(stmt *ast.If) any {
	return nil
}

func (g *CodeGenerator) VisitPrintStmt(stmt *ast.Print) any {
	return nil
}

func (g *CodeGenerator) VisitReturnStmt(stmt *ast.Return) any {
	g.emit(stmt.Offset, bytecode.OP_RETURN)

	return nil
}

func (g *CodeGenerator) VisitVarStmt(stmt *ast.Var) any {
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
