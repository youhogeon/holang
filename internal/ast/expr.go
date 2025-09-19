package ast

import "internal/token"

type NodeIdType int

type Expr interface {
	Accept(visitor ExprVisitor) any
	AcceptString(visitor ExprVisitor) string
}

type ExprVisitor interface {
	VisitAssignExpr(expr *Assign) any
	VisitBinaryExpr(expr *Binary) any
	VisitCallExpr(expr *Call) any
	VisitGetExpr(expr *Get) any
	VisitGroupingExpr(expr *Grouping) any
	VisitLiteralExpr(expr *Literal) any
	VisitLogicalExpr(expr *Logical) any
	VisitSetExpr(expr *Set) any
	VisitSuperExpr(expr *Super) any
	VisitThisExpr(expr *This) any
	VisitTernaryExpr(expr *Ternary) any
	VisitUnaryExpr(expr *Unary) any
	VisitVariableExpr(expr *Variable) any
}

type Assign struct {
	Name   *token.Token
	Value  Expr
	Offset token.Offset
	NodeId NodeIdType
}

func (a *Assign) Accept(visitor ExprVisitor) any {
	return visitor.VisitAssignExpr(a)
}

func (a *Assign) AcceptString(visitor ExprVisitor) string {
	return a.Accept(visitor).(string)
}

type Binary struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
	Offset   token.Offset
	NodeId   NodeIdType
}

func (b *Binary) Accept(visitor ExprVisitor) any {
	return visitor.VisitBinaryExpr(b)
}

func (b *Binary) AcceptString(visitor ExprVisitor) string {
	return b.Accept(visitor).(string)
}

type Call struct {
	Callee    Expr
	Arguments []Expr
	Offset    token.Offset
	NodeId    NodeIdType
}

func (c *Call) Accept(visitor ExprVisitor) any {
	return visitor.VisitCallExpr(c)
}

func (c *Call) AcceptString(visitor ExprVisitor) string {
	return c.Accept(visitor).(string)
}

type Get struct {
	Object Expr
	Name   *token.Token
	Offset token.Offset
	NodeId NodeIdType
}

func (g *Get) Accept(visitor ExprVisitor) any {
	return visitor.VisitGetExpr(g)
}

func (g *Get) AcceptString(visitor ExprVisitor) string {
	return g.Accept(visitor).(string)
}

type Grouping struct {
	Expression Expr
	Offset     token.Offset
	NodeId     NodeIdType
}

func (g *Grouping) Accept(visitor ExprVisitor) any {
	return visitor.VisitGroupingExpr(g)
}

func (g *Grouping) AcceptString(visitor ExprVisitor) string {
	return g.Accept(visitor).(string)
}

type Literal struct {
	Value  any
	Offset token.Offset
	NodeId NodeIdType
}

func (l *Literal) Accept(visitor ExprVisitor) any {
	return visitor.VisitLiteralExpr(l)
}

func (l *Literal) AcceptString(visitor ExprVisitor) string {
	return l.Accept(visitor).(string)
}

type Logical struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
	Offset   token.Offset
	NodeId   NodeIdType
}

func (l *Logical) Accept(visitor ExprVisitor) any {
	return visitor.VisitLogicalExpr(l)
}

func (l *Logical) AcceptString(visitor ExprVisitor) string {
	return l.Accept(visitor).(string)
}

type Set struct {
	Object Expr
	Name   *token.Token
	Value  Expr
	Offset token.Offset
	NodeId NodeIdType
}

func (s *Set) Accept(visitor ExprVisitor) any {
	return visitor.VisitSetExpr(s)
}

func (s *Set) AcceptString(visitor ExprVisitor) string {
	return s.Accept(visitor).(string)
}

type Super struct {
	Keyword *token.Token
	Method  *token.Token
	Offset  token.Offset
	NodeId  NodeIdType
}

func (s *Super) Accept(visitor ExprVisitor) any {
	return visitor.VisitSuperExpr(s)
}

func (s *Super) AcceptString(visitor ExprVisitor) string {
	return s.Accept(visitor).(string)
}

type This struct {
	Keyword *token.Token
	Offset  token.Offset
	NodeId  NodeIdType
}

func (t *This) Accept(visitor ExprVisitor) any {
	return visitor.VisitThisExpr(t)
}

func (t *This) AcceptString(visitor ExprVisitor) string {
	return t.Accept(visitor).(string)
}

type Ternary struct {
	Left           Expr
	FirstOperator  *token.Token
	Mid            Expr
	SecondOperator *token.Token
	Right          Expr
	Offset         token.Offset
	NodeId         NodeIdType
}

func (u *Ternary) Accept(visitor ExprVisitor) any {
	return visitor.VisitTernaryExpr(u)
}

func (u *Ternary) AcceptString(visitor ExprVisitor) string {
	return u.Accept(visitor).(string)
}

type Unary struct {
	Operator *token.Token
	Right    Expr
	Offset   token.Offset
	NodeId   NodeIdType
}

func (u *Unary) Accept(visitor ExprVisitor) any {
	return visitor.VisitUnaryExpr(u)
}

func (u *Unary) AcceptString(visitor ExprVisitor) string {
	return u.Accept(visitor).(string)
}

type Variable struct {
	Name   *token.Token
	Offset token.Offset
	NodeId NodeIdType
}

func (v *Variable) Accept(visitor ExprVisitor) any {
	return visitor.VisitVariableExpr(v)
}

func (v *Variable) AcceptString(visitor ExprVisitor) string {
	return v.Accept(visitor).(string)
}
