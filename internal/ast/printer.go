package ast

import (
	"fmt"
	"internal/scanner"
	"strings"
)

type AstPrinter struct{}

func NewAstPrinter() *AstPrinter {
	return &AstPrinter{}
}

func (p *AstPrinter) PrintExpr(expr_arg Expr) string {
	result := expr_arg.AcceptString(p)
	return result
}

func (p *AstPrinter) PrintStmt(stmt_arg Stmt) string {
	result := stmt_arg.AcceptString(p)
	return result
}

// -----------------------------------------------------------------------------
// ExprVisitor implementations
// -----------------------------------------------------------------------------

func (p *AstPrinter) VisitAssignExpr(e *Assign) any {
	return p.parenthesize2("=", e.Name.Lexeme, e.Value)
}

func (p *AstPrinter) VisitBinaryExpr(e *Binary) any {
	return p.parenthesize(e.Operator.Lexeme, e.Left, e.Right)
}

func (p *AstPrinter) VisitCallExpr(e *Call) any {
	return p.parenthesize2("call", e.Callee, e.Arguments)
}

func (p *AstPrinter) VisitGetExpr(e *Get) any {
	return p.parenthesize2(".", e.Object, e.Name.Lexeme)
}

func (p *AstPrinter) VisitGroupingExpr(e *Grouping) any {
	return p.parenthesize("group", e.Expression)
}

func (p *AstPrinter) VisitLiteralExpr(e *Literal) any {
	if e.Value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", e.Value)
}

func (p *AstPrinter) VisitLogicalExpr(e *Logical) any {
	return p.parenthesize(e.Operator.Lexeme, e.Left, e.Right)
}

func (p *AstPrinter) VisitSetExpr(e *Set) any {
	return p.parenthesize2("=", e.Object, e.Name.Lexeme, e.Value)
}

func (p *AstPrinter) VisitSuperExpr(e *Super) any {
	return p.parenthesize2("super", e.Method)
}

func (p *AstPrinter) VisitThisExpr(e *This) any {
	return "this"
}

func (p *AstPrinter) VisitUnaryExpr(e *Unary) any {
	return p.parenthesize(e.Operator.Lexeme, e.Right)
}

func (p *AstPrinter) VisitVariableExpr(e *Variable) any {
	return e.Name.Lexeme
}

// -----------------------------------------------------------------------------
// StmtVisitor implementations
// -----------------------------------------------------------------------------

func (p *AstPrinter) VisitBlockStmt(s *Block) any {
	var builder strings.Builder

	builder.WriteString("(block")
	for _, statement := range s.Statements {
		builder.WriteString(" ")
		builder.WriteString(statement.AcceptString(p))
	}

	builder.WriteString(")")
	return builder.String()
}

func (p *AstPrinter) VisitClassStmt(s *Class) any {
	var builder strings.Builder
	builder.WriteString("(class " + s.Name.Lexeme)

	if s.Superclass != nil {
		builder.WriteString(" < " + s.Superclass.AcceptString(p))
	}

	for _, method := range s.Methods {
		builder.WriteString(" ")
		builder.WriteString(method.AcceptString(p))
	}

	builder.WriteString(")")
	return builder.String()
}

func (p *AstPrinter) VisitExpressionStmt(s *Expression) any {
	return p.parenthesize(";", s.Expression)
}

func (p *AstPrinter) VisitFunctionStmt(s *Function) any {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(fun %s (", s.Name.Lexeme))

	for i, param := range s.Params {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(param.Lexeme)
	}

	builder.WriteString(") ")

	for _, bodyStmt := range s.Body {
		builder.WriteString(bodyStmt.AcceptString(p))
	}

	builder.WriteString(")")
	return builder.String()
}

func (p *AstPrinter) VisitIfStmt(s *If) any {
	if s.ElseBranch == nil {
		return p.parenthesize2("if", s.Condition, s.ThenBranch)
	}
	return p.parenthesize2("if-else", s.Condition, s.ThenBranch, s.ElseBranch)
}

func (p *AstPrinter) VisitPrintStmt(s *Print) any {
	return p.parenthesize("print", s.Expression)
}

func (p *AstPrinter) VisitReturnStmt(s *Return) any {
	if s.Value == nil {
		return "(return)"
	}
	return p.parenthesize("return", s.Value)
}

func (p *AstPrinter) VisitVarStmt(s *Var) any {
	if s.Initializer == nil {
		return p.parenthesize2("var", s.Name.Lexeme)
	}
	return p.parenthesize2("var", s.Name.Lexeme, "=", s.Initializer)
}

func (p *AstPrinter) VisitWhileStmt(s *While) any {
	return p.parenthesize2("while", s.Condition, s.Body)
}

// -----------------------------------------------------------------------------
// print-utilities

func (p *AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(%s", name))
	for _, exprArg := range exprs {
		builder.WriteString(" ")
		builder.WriteString(exprArg.AcceptString(p))
	}
	builder.WriteString(")")
	return builder.String()
}

func (p *AstPrinter) parenthesize2(name string, parts ...interface{}) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(%s", name))
	p.transform(&builder, parts...)
	builder.WriteString(")")
	return builder.String()
}

func (p *AstPrinter) transform(builder *strings.Builder, parts ...interface{}) {
	for _, part := range parts {
		builder.WriteString(" ")
		switch v := part.(type) {
		case Expr:
			builder.WriteString(v.AcceptString(p))
		case Stmt:
			builder.WriteString(v.AcceptString(p))
		case scanner.Token:
			builder.WriteString(v.Lexeme)
		case []Expr:
			for _, exprItem := range v {
				p.transform(builder, exprItem)
			}
		case []Stmt:
			for _, stmtItem := range v {
				p.transform(builder, stmtItem)
			}
		case []scanner.Token:
			for _, tokenItem := range v {
				p.transform(builder, tokenItem)
			}
		default:
			builder.WriteString(fmt.Sprintf("%v", v))
		}
	}
}
