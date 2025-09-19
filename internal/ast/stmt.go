package ast

import "internal/token"

type Stmt interface {
	Accept(visitor StmtVisitor) any
	AcceptString(visitor StmtVisitor) string
}

type StmtVisitor interface {
	VisitBlockStmt(stmt *Block) any
	VisitClassStmt(stmt *Class) any
	VisitExpressionStmt(stmt *Expression) any
	VisitFunctionStmt(stmt *Function) any
	VisitIfStmt(stmt *If) any
	VisitPrintStmt(stmt *Print) any
	VisitReturnStmt(stmt *Return) any
	VisitVarStmt(stmt *Var) any
	VisitWhileStmt(stmt *While) any
	VisitBreakStmt(stmt *Break) any
	VisitContinueStmt(stmt *Continue) any
}

type Block struct {
	Statements []Stmt
	Offset     token.Offset
	NodeId     NodeIdType
}

func (s *Block) Accept(visitor StmtVisitor) any {
	return visitor.VisitBlockStmt(s)
}

func (s *Block) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Class struct {
	Name       *token.Token
	Superclass *Variable
	Methods    []*Function
	Offset     token.Offset
	NodeId     NodeIdType
}

func (c *Class) Accept(visitor StmtVisitor) any {
	return visitor.VisitClassStmt(c)
}

func (c *Class) AcceptString(visitor StmtVisitor) string {
	return c.Accept(visitor).(string)
}

type Expression struct {
	Expression Expr
	Offset     token.Offset
	NodeId     NodeIdType
}

func (s *Expression) Accept(visitor StmtVisitor) any {
	return visitor.VisitExpressionStmt(s)
}

func (s *Expression) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Function struct {
	Name   *token.Token
	Params []*token.Token
	Body   []Stmt
	Offset token.Offset
	NodeId NodeIdType
}

func (f *Function) Accept(visitor StmtVisitor) any {
	return visitor.VisitFunctionStmt(f)
}

func (f *Function) AcceptString(visitor StmtVisitor) string {
	return f.Accept(visitor).(string)
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
	Offset     token.Offset
	NodeId     NodeIdType
}

func (s *If) Accept(visitor StmtVisitor) any {
	return visitor.VisitIfStmt(s)
}

func (s *If) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Print struct {
	Expression Expr
	Offset     token.Offset
	NodeId     NodeIdType
}

func (s *Print) Accept(visitor StmtVisitor) any {
	return visitor.VisitPrintStmt(s)
}

func (s *Print) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Return struct {
	Keyword *token.Token
	Value   Expr
	Offset  token.Offset
	NodeId  NodeIdType
}

func (s *Return) Accept(visitor StmtVisitor) any {
	return visitor.VisitReturnStmt(s)
}

func (s *Return) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Var struct {
	Name        *token.Token
	Initializer Expr
	Offset      token.Offset
	NodeId      NodeIdType
}

func (s *Var) Accept(visitor StmtVisitor) any {
	return visitor.VisitVarStmt(s)
}

func (s *Var) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type While struct {
	Condition Expr
	Body      Stmt
	Post      Expr

	Offset token.Offset
	NodeId NodeIdType
}

func (s *While) Accept(visitor StmtVisitor) any {
	return visitor.VisitWhileStmt(s)
}

func (s *While) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Break struct {
	Offset token.Offset
	NodeId NodeIdType
}

func (s *Break) Accept(visitor StmtVisitor) any {
	return visitor.VisitBreakStmt(s)
}

func (s *Break) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Continue struct {
	Offset token.Offset
	NodeId NodeIdType
}

func (s *Continue) Accept(visitor StmtVisitor) any {
	return visitor.VisitContinueStmt(s)
}

func (s *Continue) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}
