package codegen

import (
	"errors"
	"internal/ast"
	"internal/bytecode"
	"internal/resolver"
	"internal/runtime"
	"internal/token"
)

const _POPN_THRESHOLD = 3

type loopCtx struct {
	loopStart int
	breaks    []int
	continues []int
}

type functionCtx struct {
	fn *runtime.ObjFunction
	em Emitter
}

type CodeGenerator struct {
	bindings resolver.BindingResult

	loopCtx []*loopCtx
	funcCtx []*functionCtx
}

func NewCodeGenerator(bindings resolver.BindingResult) *CodeGenerator {
	return &CodeGenerator{
		bindings: bindings,
	}
}

func (g *CodeGenerator) Generate(statements []ast.Stmt) (*runtime.ObjFunction, error) {
	rootFn := runtime.NewObjFunction("<script>", 0, runtime.FUNCTION_TYPE_SCRIPT)
	g.beginFunction(rootFn)

	for _, stmt := range statements {
		err := g.genStmt(stmt)

		if err != nil {
			return nil, err
		}
	}

	g.endFunction()

	return rootFn, nil
}

func (g *CodeGenerator) beginFunction(fn *runtime.ObjFunction) {
	ch := fn.Chunk
	em := NewChunkEmitter(ch)

	g.funcCtx = append(g.funcCtx, &functionCtx{
		fn: fn,
		em: em,
	})
}

func (g *CodeGenerator) endFunction() *runtime.ObjFunction {
	g.emit(token.Offset{Line: -1, Index: -1}, bytecode.OP_RETURN)

	fc := g.funcCtx[len(g.funcCtx)-1]
	g.funcCtx = g.funcCtx[:len(g.funcCtx)-1]

	return fc.fn
}

func (g *CodeGenerator) getEmitter() Emitter {
	if len(g.funcCtx) == 0 {
		return nil
	}

	return g.funcCtx[len(g.funcCtx)-1].em
}

func (g *CodeGenerator) genStmt(s ast.Stmt) error {
	if result := s.Accept(g); result != nil {
		return result.(error)
	}

	return nil
}

func (g *CodeGenerator) getChunkSize() int {
	em := g.getEmitter()

	return em.Size()
}

func (g *CodeGenerator) emit(offset token.Offset, op bytecode.OpCode, operands ...int64) int {
	em := g.getEmitter()

	return em.Emit(offset, op, operands...)
}

func (g *CodeGenerator) emitJump(offset token.Offset, op bytecode.OpCode) int {
	em := g.getEmitter()

	return em.EmitJump(offset, op)
}

func (g *CodeGenerator) patchJump(jumpOpLoc int) {
	g.patchJumpTo(jumpOpLoc, g.getChunkSize())
}

func (g *CodeGenerator) patchJumpTo(jumpOpLoc int, jumpTo int) {
	em := g.getEmitter()

	em.PatchJump(jumpOpLoc, jumpTo)
}

func (g *CodeGenerator) emitPop(offset token.Offset, popCnt int) {
	if popCnt >= _POPN_THRESHOLD {
		g.emit(offset, bytecode.OP_POPN, int64(popCnt))
	}

	for range popCnt {
		g.emit(offset, bytecode.OP_POP)
	}
}

func (g *CodeGenerator) emitConstant(offset token.Offset, value bytecode.Value) {
	switch v := value.(type) {
	case nil:
		g.emit(offset, bytecode.OP_NIL)

	case bool:
		if v {
			g.emit(offset, bytecode.OP_TRUE)
		} else {
			g.emit(offset, bytecode.OP_FALSE)
		}

	case int64:
		switch v {
		case -1:
			g.emit(offset, bytecode.OP_CONSTANT_M1)
		case 0:
			g.emit(offset, bytecode.OP_CONSTANT_0)
		case 1:
			g.emit(offset, bytecode.OP_CONSTANT_1)
		case 2:
			g.emit(offset, bytecode.OP_CONSTANT_2)
		case 3:
			g.emit(offset, bytecode.OP_CONSTANT_3)
		case 4:
			g.emit(offset, bytecode.OP_CONSTANT_4)
		case 5:
			g.emit(offset, bytecode.OP_CONSTANT_5)
		default:
			g.emit(offset, bytecode.OP_CONSTANT, g.makeConstant(value))
		}

	default:
		g.emit(offset, bytecode.OP_CONSTANT, g.makeConstant(value))
	}
}

func (g *CodeGenerator) makeConstant(value bytecode.Value) int64 {
	em := g.getEmitter()

	return em.MakeConstant(value)
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

	g.emitPop(stmt.Offset, popCnt)

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
	loopStart := g.getChunkSize()

	currentLoopCtx := &loopCtx{
		loopStart: loopStart,
		breaks:    []int{},
	}
	g.loopCtx = append(g.loopCtx, currentLoopCtx)

	if err := stmt.Condition.Accept(g); err != nil {
		return err
	}

	endJump := g.emitJump(stmt.Offset, bytecode.OP_JUMP_IF_FALSE)
	g.emit(stmt.Offset, bytecode.OP_POP)

	if err := stmt.Body.Accept(g); err != nil {
		return err
	}

	// Body에서 사용된 continue 들을 이 곳으로 patch
	for _, continuePos := range currentLoopCtx.continues {
		g.patchJump(continuePos)
	}

	// Post 처리
	if stmt.Post != nil {
		if err := stmt.Post.Accept(g); err != nil {
			return err
		}

		g.emit(stmt.Offset, bytecode.OP_POP)
	}

	// goto loopStart
	j := g.emitJump(stmt.Offset, bytecode.OP_JUMP)
	g.patchJumpTo(j, loopStart)

	// loop 종료
	g.patchJump(endJump)
	g.emit(stmt.Offset, bytecode.OP_POP)

	// Body에서 사용된 break 들을 이 곳으로 patch
	for _, breakPos := range currentLoopCtx.breaks {
		g.patchJump(breakPos)
	}

	g.loopCtx = g.loopCtx[:len(g.loopCtx)-1]

	return nil
}

func (g *CodeGenerator) VisitBreakStmt(stmt *ast.Break) any {
	if len(g.loopCtx) == 0 {
		return errors.New("break statement not within a loop")
	}

	loopCtx := g.loopCtx[len(g.loopCtx)-1]

	popCount := g.bindings[stmt.NodeId].Slot
	g.emitPop(stmt.Offset, popCount)

	jump := g.emitJump(stmt.Offset, bytecode.OP_JUMP)
	loopCtx.breaks = append(loopCtx.breaks, jump)

	return nil
}

func (g *CodeGenerator) VisitContinueStmt(stmt *ast.Continue) any {
	if len(g.loopCtx) == 0 {
		return errors.New("continue statement not within a loop")
	}

	loopCtx := g.loopCtx[len(g.loopCtx)-1]

	popCount := g.bindings[stmt.NodeId].Slot
	g.emitPop(stmt.Offset, popCount)

	jump := g.emitJump(stmt.Offset, bytecode.OP_JUMP)
	loopCtx.continues = append(loopCtx.continues, jump)

	return nil
}
