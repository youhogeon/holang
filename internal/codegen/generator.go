package codegen

import (
	"errors"
	"internal/ast"
	"internal/bytecode"
	"internal/resolver"
	"internal/runtime"
	"internal/token"
)

const _POPN_THRESHOLD = 2

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
	bindings resolver.ResolverResult

	loopCtx []*loopCtx
	fnCtx   []*functionCtx
}

func NewCodeGenerator(bindings resolver.ResolverResult) *CodeGenerator {
	return &CodeGenerator{
		bindings: bindings,
	}
}

func (g *CodeGenerator) Generate(program *ast.Program) (*runtime.ObjFunction, error) {
	rootFn := runtime.NewObjFunction("<script>", 0, runtime.FUNCTION_TYPE_SCRIPT)

	g.beginFunction(rootFn)

	err := g.genStmts(program.Statements)

	if err != nil {
		return nil, err.(error)
	}

	g.endFunction()

	return rootFn, nil
}

func (g *CodeGenerator) genStmts(statements []ast.Stmt) any {
	for _, stmt := range statements {
		if result := stmt.Accept(g); result != nil {
			return result
		}
	}

	return nil
}

func (g *CodeGenerator) getCurrentFnCtx() *functionCtx {
	return g.fnCtx[len(g.fnCtx)-1]
}

func (g *CodeGenerator) beginFunction(fn *runtime.ObjFunction) {
	ch := fn.Chunk
	em := NewChunkEmitter(ch)

	g.fnCtx = append(g.fnCtx, &functionCtx{
		fn: fn,
		em: em,
	})
}

func (g *CodeGenerator) endFunction() *runtime.ObjFunction {
	fc := g.getCurrentFnCtx()

	if fc.fn.Type != runtime.FUNCTION_TYPE_SCRIPT {
		g.emitReturn(token.Offset{Line: -1, Index: -1}, nil)
	}

	g.fnCtx = g.fnCtx[:len(g.fnCtx)-1]

	return fc.fn
}

// ================================================================
// Emitter Helper
// ================================================================

func (g *CodeGenerator) getEmitter() Emitter {
	return g.getCurrentFnCtx().em
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
	if popCnt <= 0 {
		return
	}

	if popCnt >= _POPN_THRESHOLD {
		g.emit(offset, bytecode.OP_POPN, int64(popCnt))

		return
	}

	for range popCnt {
		g.emit(offset, bytecode.OP_POP)
	}
}

func (g *CodeGenerator) emitCloseUpvalues(offset token.Offset, n int) {
	if n <= 0 {
		return
	}

	g.emit(offset, bytecode.OP_CLOSE_UPVALUE, int64(n))
}

func (g *CodeGenerator) emitPopOrCloseUpvalues(offset token.Offset, isCapturedList []bool) {
	if len(isCapturedList) == 0 {
		return
	}

	flush := func(close bool, n int) {
		if n <= 0 {
			return
		}

		if close {
			g.emitCloseUpvalues(offset, n)
		} else {
			g.emitPop(offset, n)
		}
	}

	runKind := isCapturedList[0]
	runLen := 1

	for i := 1; i < len(isCapturedList); i++ {
		if isCapturedList[i] == runKind {
			runLen++
			continue
		}

		flush(runKind, runLen)
		runKind = isCapturedList[i]
		runLen = 1
	}

	flush(runKind, runLen)
}

func (g *CodeGenerator) emitConstant(offset token.Offset, value runtime.Value) {
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

func (g *CodeGenerator) makeConstant(value runtime.Value) int64 {
	em := g.getEmitter()

	return em.MakeConstant(value)
}

func (g *CodeGenerator) emitReturn(offset token.Offset, value ast.Expr) any {
	if value != nil {
		if err := value.Accept(g); err != nil {
			return err
		}
	} else {
		fc := g.getCurrentFnCtx()

		if fc.fn.Type == runtime.FUNCTION_TYPE_INITIALIZER {
			g.emit(offset, bytecode.OP_GET_LOCAL, 0)
		} else {
			g.emit(offset, bytecode.OP_NIL)
		}
	}

	g.emit(offset, bytecode.OP_RETURN)

	return nil
}

// ================================================================
// Expr
// ================================================================

func (g *CodeGenerator) VisitAssignExpr(expr *ast.Assign) any {
	binding := g.bindings[expr.NodeId]

	if err := expr.Value.Accept(g); err != nil {
		return err
	}

	op := bytecode.OP_SET_LOCAL
	v := int64(binding.BindingIndex)

	if binding.BindingKind == resolver.BindUpvalue {
		op = bytecode.OP_SET_UPVALUE
	} else if binding.BindingKind == resolver.BindGlobal {
		op = bytecode.OP_SET_GLOBAL
		v = g.makeConstant(expr.Name.Lexeme)
	}

	g.emit(expr.Offset, op, v)

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
	if err := expr.Callee.Accept(g); err != nil {
		return err
	}

	for _, arg := range expr.Arguments {
		if err := arg.Accept(g); err != nil {
			return err
		}
	}

	g.emit(expr.Offset, bytecode.OP_CALL, int64(len(expr.Arguments)))

	return nil
}

func (g *CodeGenerator) VisitGetExpr(expr *ast.Get) any {
	if err := expr.Object.Accept(g); err != nil {
		return err
	}

	fieldName := g.makeConstant(expr.Name.Lexeme)
	g.emit(expr.Offset, bytecode.OP_GET_PROPERTY, fieldName)

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
	if err := expr.Object.Accept(g); err != nil {
		return err
	}

	if err := expr.Value.Accept(g); err != nil {
		return err
	}

	fieldName := g.makeConstant(expr.Name.Lexeme)
	g.emit(expr.Offset, bytecode.OP_SET_PROPERTY, fieldName)

	return nil
}

func (g *CodeGenerator) VisitSuperExpr(expr *ast.Super) any {
	return nil
}

func (g *CodeGenerator) VisitThisExpr(expr *ast.This) any {
	binding := g.bindings[expr.NodeId]

	op := bytecode.OP_GET_LOCAL
	v := int64(binding.BindingIndex)

	if binding.BindingKind == resolver.BindUpvalue {
		op = bytecode.OP_GET_UPVALUE
	}

	g.emit(expr.Offset, op, v)

	return nil
}

func (g *CodeGenerator) VisitTernaryExpr(expr *ast.Ternary) any {
	if err := expr.Left.Accept(g); err != nil {
		return err
	}

	jump := g.emitJump(expr.Offset, bytecode.OP_JUMP_IF_FALSE)

	g.emit(expr.Offset, bytecode.OP_POP)
	if err := expr.Mid.Accept(g); err != nil {
		return err
	}

	endJump := g.emitJump(expr.Offset, bytecode.OP_JUMP)

	g.patchJump(jump)

	g.emit(expr.Offset, bytecode.OP_POP)
	if err := expr.Right.Accept(g); err != nil {
		return err
	}

	g.patchJump(endJump)

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

	op := bytecode.OP_GET_LOCAL
	v := int64(binding.BindingIndex)

	if binding.BindingKind == resolver.BindUpvalue {
		op = bytecode.OP_GET_UPVALUE
	} else if binding.BindingKind == resolver.BindGlobal {
		op = bytecode.OP_GET_GLOBAL
		v = g.makeConstant(expr.Name.Lexeme)
	}

	g.emit(expr.Offset, op, v)

	return nil
}

// ================================================================
// Stmt
// ================================================================

func (g *CodeGenerator) VisitBlockStmt(stmt *ast.Block) any {
	b := g.bindings[stmt.NodeId]

	if err := g.genStmts(stmt.Statements); err != nil {
		return err
	}

	g.emitPopOrCloseUpvalues(stmt.Offset, b.Pops)

	return nil
}

func (g *CodeGenerator) VisitClassStmt(stmt *ast.Class) any {
	binding := g.bindings[stmt.NodeId]

	// emit clss
	nameConst := g.makeConstant(stmt.Name.Lexeme)
	g.emit(stmt.Offset, bytecode.OP_CLASS, nameConst)

	if binding.BindingKind == resolver.BindGlobal {
		g.emit(stmt.Offset, bytecode.OP_DEFINE_GLOBAL, nameConst)
		g.emit(stmt.Offset, bytecode.OP_GET_GLOBAL, nameConst)
	} else {
		g.emit(stmt.Offset, bytecode.OP_GET_LOCAL, int64(binding.BindingIndex))
	}

	// emit methods
	for _, method := range stmt.Methods {
		if err := method.Accept(g); err != nil {
			return err
		}

		g.emit(stmt.Offset, bytecode.OP_METHOD)
	}

	g.emit(stmt.Offset, bytecode.OP_POP)

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
	binding := g.bindings[stmt.NodeId]

	fnType := runtime.FUNCTION_TYPE_FUN
	if stmt.IsMethod && stmt.Name.Lexeme == "init" {
		fnType = runtime.FUNCTION_TYPE_INITIALIZER
	}

	fnObj := runtime.NewObjFunction(
		stmt.Name.Lexeme,
		len(stmt.Params),
		fnType,
	)

	// gen code
	g.beginFunction(fnObj)

	err := g.genStmts(stmt.Body)

	g.endFunction()

	if err != nil {
		return err
	}

	// closure
	uvCount := len(binding.Upvalues)
	argCount := int64(1 + 2*uvCount)
	fnObj.UpvalueCount = uvCount
	args := make([]int64, 0, argCount+1)
	constant := g.makeConstant(fnObj)

	args = append(args, argCount)
	args = append(args, constant)

	for _, up := range binding.Upvalues {
		if up.IsLocal {
			args = append(args, 1, int64(up.Index))
		} else {
			args = append(args, 0, int64(up.Index))
		}
	}

	// emit
	g.emit(stmt.Offset, bytecode.OP_CLOSURE, args...)

	if binding.BindingKind == resolver.BindGlobal {
		nameConst := g.makeConstant(stmt.Name.Lexeme)
		g.emit(stmt.Offset, bytecode.OP_DEFINE_GLOBAL, nameConst)
	}

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
	return g.emitReturn(stmt.Offset, stmt.Value)
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

	if binding.BindingKind == resolver.BindGlobal {
		constant := g.makeConstant(stmt.Name.Lexeme)
		g.emit(stmt.Offset, bytecode.OP_DEFINE_GLOBAL, constant)
	}

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

	b := g.bindings[stmt.NodeId]

	g.emitPopOrCloseUpvalues(stmt.Offset, b.Pops)

	jump := g.emitJump(stmt.Offset, bytecode.OP_JUMP)
	loopCtx.breaks = append(loopCtx.breaks, jump)

	return nil
}

func (g *CodeGenerator) VisitContinueStmt(stmt *ast.Continue) any {
	if len(g.loopCtx) == 0 {
		return errors.New("continue statement not within a loop")
	}

	loopCtx := g.loopCtx[len(g.loopCtx)-1]

	b := g.bindings[stmt.NodeId]

	g.emitPopOrCloseUpvalues(stmt.Offset, b.Pops)

	jump := g.emitJump(stmt.Offset, bytecode.OP_JUMP)
	loopCtx.continues = append(loopCtx.continues, jump)

	return nil
}
