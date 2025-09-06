package ast

import (
	"internal/scanner"
	"testing"
)

func TestAstPrinter_LiteralAndBinary(t *testing.T) {
	p := NewAstPrinter()

	lit1 := &Literal{Value: 1}
	if got := p.PrintExpr(lit1); got != "1" {
		t.Fatalf("literal: got %q want %q", got, "1")
	}

	litNil := &Literal{Value: nil}
	if got := p.PrintExpr(litNil); got != "nil" {
		t.Fatalf("nil literal: got %q want %q", got, "nil")
	}

	bin := &Binary{
		Left:     lit1,
		Operator: &scanner.Token{Lexeme: "+"},
		Right:    &Literal{Value: 2},
	}
	if got := p.PrintExpr(bin); got != "(+ 1 2)" {
		t.Fatalf("binary: got %q want %q", got, "(+ 1 2)")
	}
}

func TestAstPrinter_GroupingUnaryVariableAndPrintStmt(t *testing.T) {
	p := NewAstPrinter()

	lit := &Literal{Value: 42}
	group := &Grouping{Expression: lit}
	if got := p.PrintExpr(group); got != "(group 42)" {
		t.Fatalf("grouping: got %q want %q", got, "(group 42)")
	}

	un := &Unary{Operator: &scanner.Token{Lexeme: "-"}, Right: lit}
	if got := p.PrintExpr(un); got != "(- 42)" {
		t.Fatalf("unary: got %q want %q", got, "(- 42)")
	}

	variable := &Variable{Name: &scanner.Token{Lexeme: "x"}}
	if got := p.PrintExpr(variable); got != "x" {
		t.Fatalf("variable: got %q want %q", got, "x")
	}

	printStmt := &Print{Expression: &Literal{Value: "hello"}}
	if got := p.PrintStmt(printStmt); got != "(print hello)" {
		t.Fatalf("print stmt: got %q want %q", got, "(print hello)")
	}
}
