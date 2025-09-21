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
	BindUpvalue
	Block
	Break
	Continue
	Return
)

type Binding struct {
	Kind BindingKind

	// BindGlobal:		-1
	// BindLocal:		Index 슬롯 인덱스
	// BindUpvalue:		Upvalues배열의 인덱스
	Index int

	// Block, Break, Continue, Return: 제거할 로컬 수
	// (capture여부)의 배열
	Pops []bool

	Upvalues []*UpValue
}

type BindingResult map[ast.NodeIdType]Binding

// ================================================================
// Resolver
// ================================================================

type Local struct {
	Name       *token.Token
	Depth      int // -1: 초기화 안됨
	Slot       int
	IsCaptured bool
}

type UpValue struct {
	Index   int
	IsLocal bool
}

type functionCtx struct {
	locals     []*Local
	nextSlot   int
	scopeDepth int

	upvalues      []*UpValue
	upIndexByName map[string]int

	loopDepthStack []int
}

type Resolver struct {
	fnCtx []*functionCtx

	bindings BindingResult
	errors   []error
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) clear() {
	r.fnCtx = r.fnCtx[:0]
	r.addFunctionCtx()

	r.fnCtx[0].nextSlot = 0

	r.bindings = make(BindingResult)
	r.errors = r.errors[:0]
}

func (r *Resolver) addFunctionCtx() *functionCtx {
	ctx := &functionCtx{
		upIndexByName: make(map[string]int),
		nextSlot:      1, // 0은 callee(this)
	}

	r.fnCtx = append(r.fnCtx, ctx)

	return ctx
}

func (r *Resolver) currentCtx() *functionCtx {
	return r.fnCtx[len(r.fnCtx)-1]
}

func (r *Resolver) Resolve(program *ast.Program) (BindingResult, []error) {
	r.clear()

	for _, stmt := range program.Statements {
		stmt.Accept(r)
	}

	return r.bindings, r.errors
}

func (r *Resolver) beginScope() {
	ctx := r.currentCtx()
	ctx.scopeDepth++
}

func (r *Resolver) endScope() {
	ctx := r.currentCtx()

	if ctx.scopeDepth == 0 {
		return
	}

	for len(ctx.locals) > 0 && ctx.locals[len(ctx.locals)-1].Depth >= ctx.scopeDepth {
		ctx.locals = ctx.locals[:len(ctx.locals)-1]
		ctx.nextSlot--
	}

	ctx.scopeDepth--
}

func (r *Resolver) declare(name *token.Token) (slot int, isLocal bool) {
	ctx := r.currentCtx()

	if ctx.scopeDepth == 0 {
		return -1, false
	}

	for i := len(ctx.locals) - 1; i >= 0; i-- {
		local := ctx.locals[i]

		if local.Depth < ctx.scopeDepth && local.Depth >= 0 {
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
	ctx := r.currentCtx()

	if ctx.scopeDepth == 0 {
		return
	}

	for i := len(ctx.locals) - 1; i >= 0; i-- {
		local := ctx.locals[i]

		if local.Depth < ctx.scopeDepth && local.Depth >= 0 {
			break
		}

		if local.Name.Lexeme == name.Lexeme {
			local.Depth = ctx.scopeDepth
			ctx.locals[i] = local
			return
		}
	}
}
func (r *Resolver) findLocal(name *token.Token) (int, bool) {
	ctx := r.currentCtx()

	return r.findLocalIn(name, ctx)
}

func (r *Resolver) findLocalIn(name *token.Token, ctx *functionCtx) (int, bool) {
	for i := len(ctx.locals) - 1; i >= 0; i-- {
		local := ctx.locals[i]

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
	ctx := r.currentCtx()

	slot = ctx.nextSlot

	ctx.locals = append(ctx.locals, &Local{
		Name:  name,
		Depth: -1,
		Slot:  slot,
	})

	ctx.nextSlot++

	return
}

func (r *Resolver) addUpvalue(ctx *functionCtx, name string, index int, isLocal bool) int {
	if upIndex, ok := ctx.upIndexByName[name]; ok {
		return upIndex
	}

	upIndex := len(ctx.upvalues)

	ctx.upvalues = append(ctx.upvalues, &UpValue{
		Index:   index,
		IsLocal: isLocal,
	})

	ctx.upIndexByName[name] = upIndex

	return upIndex
}

func (r *Resolver) resolveUpvalue(name *token.Token) (int, bool) {
	ctx := r.currentCtx()

	for i := len(r.fnCtx) - 2; i >= 0; i-- {
		outerCtx := r.fnCtx[i]

		if slot, ok := r.findLocalIn(name, outerCtx); ok {
			outerCtx.locals[slot].IsCaptured = true

			isLocal := true
			for j := i + 1; j < len(r.fnCtx); j++ {
				innerCtx := r.fnCtx[j]
				slot = r.addUpvalue(innerCtx, name.Lexeme, slot, isLocal)
				isLocal = false
			}

			return slot, true
		}

		if upIndex, ok := outerCtx.upIndexByName[name.Lexeme]; ok {
			return r.addUpvalue(ctx, name.Lexeme, upIndex, false), true
		}
	}

	return -1, false
}

func (r *Resolver) makePopInfo(fromSlot int) []bool {
	ctx := r.currentCtx()

	localInfo := []bool{}

	for i := len(ctx.locals) - 1; i >= 0; i-- {
		l := ctx.locals[i]

		if l.Slot < fromSlot {
			break
		}

		localInfo = append(localInfo, l.IsCaptured)
	}

	return localInfo
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
			Kind:  BindLocal,
			Index: slot,
		}

		return nil
	}

	if uv, ok := r.resolveUpvalue(expr.Name); ok {
		r.bindings[expr.NodeId] = Binding{
			Kind:  BindUpvalue,
			Index: uv,
		}

		return nil
	}

	r.bindings[expr.NodeId] = Binding{
		Kind:  BindGlobal,
		Index: -1,
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
			Kind:  BindLocal,
			Index: slot,
		}

		return nil
	}

	if uv, ok := r.resolveUpvalue(expr.Name); ok {
		r.bindings[expr.NodeId] = Binding{
			Kind:  BindUpvalue,
			Index: uv,
		}

		return nil
	}

	r.bindings[expr.NodeId] = Binding{
		Kind:  BindGlobal,
		Index: -1,
	}

	return nil
}

// ================================================================
// Stmt
// ================================================================

func (r *Resolver) VisitBlockStmt(stmt *ast.Block) any {
	ctx := r.currentCtx()

	r.beginScope()

	prevSlot := ctx.nextSlot

	for _, s := range stmt.Statements {
		s.Accept(r)
	}

	locals := r.makePopInfo(prevSlot)

	r.endScope()

	r.bindings[stmt.NodeId] = Binding{
		Kind: Block,
		Pops: locals,
	}

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
	slot, isLocal := r.declare(stmt.Name)

	r.define(stmt.Name)

	// 새로운 frame
	ctx := r.addFunctionCtx()
	// 새로운 scope
	r.beginScope()

	for _, p := range stmt.Params {
		r.declare(p)
		r.define(p)
	}

	for _, bodyStmt := range stmt.Body {
		bodyStmt.Accept(r)
	}

	r.endScope()
	r.fnCtx = r.fnCtx[:len(r.fnCtx)-1]

	if isLocal {
		r.bindings[stmt.NodeId] = Binding{Kind: BindLocal, Index: slot, Upvalues: ctx.upvalues}
	} else {
		r.bindings[stmt.NodeId] = Binding{Kind: BindGlobal, Index: -1, Upvalues: ctx.upvalues}
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

	locals := r.makePopInfo(0)

	r.bindings[stmt.NodeId] = Binding{
		Kind: Return,
		Pops: locals,
	}

	return nil
}

func (r *Resolver) VisitVarStmt(stmt *ast.Var) any {
	slot, isLocal := r.declare(stmt.Name)

	if isLocal {
		r.bindings[stmt.NodeId] = Binding{Kind: BindLocal, Index: slot}
	} else {
		r.bindings[stmt.NodeId] = Binding{Kind: BindGlobal, Index: -1}
	}

	if stmt.Initializer != nil {
		stmt.Initializer.Accept(r)
	}

	r.define(stmt.Name)

	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *ast.While) any {
	ctx := r.currentCtx()

	stmt.Condition.Accept(r)

	if stmt.Post != nil {
		stmt.Post.Accept(r)
	}

	ctx.loopDepthStack = append(ctx.loopDepthStack, ctx.nextSlot)
	stmt.Body.Accept(r)
	ctx.loopDepthStack = ctx.loopDepthStack[:len(ctx.loopDepthStack)-1]

	return nil
}

func (r *Resolver) VisitBreakStmt(stmt *ast.Break) any {
	ctx := r.currentCtx()

	if len(ctx.loopDepthStack) == 0 {
		r.addError("break statement not within a loop")

		return nil
	}

	topBaseSlot := ctx.loopDepthStack[len(ctx.loopDepthStack)-1]
	locals := r.makePopInfo(topBaseSlot)

	r.bindings[stmt.NodeId] = Binding{
		Kind: Break,
		Pops: locals,
	}

	return nil
}

func (r *Resolver) VisitContinueStmt(stmt *ast.Continue) any {
	ctx := r.currentCtx()

	if len(ctx.loopDepthStack) == 0 {
		r.addError("continue statement not within a loop")

		return nil
	}

	topBaseSlot := ctx.loopDepthStack[len(ctx.loopDepthStack)-1]
	locals := r.makePopInfo(topBaseSlot)

	r.bindings[stmt.NodeId] = Binding{
		Kind: Continue,
		Pops: locals,
	}

	return nil
}
