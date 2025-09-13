package ast

import "internal/scanner"

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
	Offset     Offset
}

func (s *Block) Accept(visitor StmtVisitor) any {
	return visitor.VisitBlockStmt(s)
}

func (s *Block) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Class struct {
	Name       *scanner.Token
	Superclass *Variable
	Methods    []*Function
	Offset     Offset
}

func (c *Class) Accept(visitor StmtVisitor) any {
	return visitor.VisitClassStmt(c)
}

func (c *Class) AcceptString(visitor StmtVisitor) string {
	return c.Accept(visitor).(string)
}

type Expression struct {
	Expression Expr
	Offset     Offset
}

func (s *Expression) Accept(visitor StmtVisitor) any {
	return visitor.VisitExpressionStmt(s)
}

func (s *Expression) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Function struct {
	Name   *scanner.Token
	Params []*scanner.Token
	Body   []Stmt
	Offset Offset
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
	Offset     Offset
}

func (s *If) Accept(visitor StmtVisitor) any {
	return visitor.VisitIfStmt(s)
}

func (s *If) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Print struct {
	Expression Expr
	Offset     Offset
}

func (s *Print) Accept(visitor StmtVisitor) any {
	return visitor.VisitPrintStmt(s)
}

func (s *Print) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Return struct {
	Keyword *scanner.Token
	Value   Expr
	Offset  Offset
}

func (s *Return) Accept(visitor StmtVisitor) any {
	return visitor.VisitReturnStmt(s)
}

func (s *Return) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Var struct {
	Name        *scanner.Token
	Initializer Expr
	Offset      Offset
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
	Offset    Offset
}

func (s *While) Accept(visitor StmtVisitor) any {
	return visitor.VisitWhileStmt(s)
}

func (s *While) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Break struct {
	Offset Offset
}

func (s *Break) Accept(visitor StmtVisitor) any {
	return visitor.VisitBreakStmt(s)
}

func (s *Break) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}

type Continue struct {
	Offset Offset
}

func (s *Continue) Accept(visitor StmtVisitor) any {
	return visitor.VisitContinueStmt(s)
}

func (s *Continue) AcceptString(visitor StmtVisitor) string {
	return s.Accept(visitor).(string)
}
