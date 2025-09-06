package parser

import (
	"internal/ast"
	"internal/scanner"
	"slices"
)

type Parser struct {
	tokens  []scanner.Token
	current int
}

// program        → statement* EOF ;
//
// statement      → exprStmt
//                | printStmt ;
//
// exprStmt       → expression ";" ;
// printStmt      → "print" expression ";" ;
//
// expression     → ternary ;
// ternary        → equality ("?" expression ":" expression)? ;
// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           → factor ( ( "-" | "+" ) factor )* ;
// factor         → unary ( ( "/" | "*" ) unary )* ;
// unary          → ( "!" | "-" ) unary
//                | primary ;
// primary        → NUMBER | STRING | "true" | "false" | "nil"
//                | "(" expression ")" ;

func NewParser(tokens []scanner.Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() ([]ast.Stmt, []error) {
	statements := make([]ast.Stmt, 0)
	errors := make([]error, 0)

	for !p.isAtEnd() {
		stmt, err := p.parseStatements()

		if err != nil {
			p.synchronize()

			errors = append(errors, err)

			continue
		}

		statements = append(statements, stmt)
	}

	return statements, errors
}

func (p *Parser) parseStatements() (ast.Stmt, error) {
	if p.match(scanner.PRINT) {
		return p.printStatement()
	}

	return p.exprStatement()
}

func (p *Parser) exprStatement() (ast.Stmt, error) {
	expr, err := p.expression()

	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(scanner.SEMICOLON, "Expect ';' after value.")

	if err != nil {
		return nil, err
	}

	return &ast.Expression{
		Expression: expr,
	}, nil
}

func (p *Parser) printStatement() (ast.Stmt, error) {
	expr, err := p.expression()

	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(scanner.SEMICOLON, "Expect ';' after value.")

	if err != nil {
		return nil, err
	}

	return &ast.Print{
		Expression: expr,
	}, nil
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.ternary()
}

func (p *Parser) ternary() (ast.Expr, error) {
	expr, err := p.equality()

	if err != nil {
		return nil, err
	}

	if p.match(scanner.QUESTION) {
		firstOp := p.previous()
		mid, err := p.expression()

		if err != nil {
			return nil, err
		}

		secondOp, err := p.consumeOrError(scanner.COLON, "Expect ':' after expression.")

		if err != nil {
			return nil, err
		}

		right, err := p.expression()

		if err != nil {
			return nil, err
		}

		return &ast.Ternary{
			Left:           expr,
			FirstOperator:  firstOp,
			Mid:            mid,
			SecondOperator: secondOp,
			Right:          right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) equality() (ast.Expr, error) {
	expr, err := p.comparison()

	if err != nil {
		return nil, err
	}

	for p.match(scanner.BANG_EQUAL, scanner.EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.term()

	if err != nil {
		return nil, err
	}

	for p.match(scanner.GREATER, scanner.GREATER_EQUAL, scanner.LESS, scanner.LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) term() (ast.Expr, error) {
	expr, err := p.factor()

	if err != nil {
		return nil, err
	}

	for p.match(scanner.MINUS, scanner.PLUS) {
		operator := p.previous()
		right, err := p.factor()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) factor() (ast.Expr, error) {
	expr, err := p.unary()

	if err != nil {
		return nil, err
	}

	for p.match(scanner.SLASH, scanner.STAR) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(scanner.BANG, scanner.MINUS) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		return &ast.Unary{
			Operator: operator,
			Right:    right,
		}, nil
	}

	return p.primary()
}

func (p *Parser) primary() (ast.Expr, error) {
	if p.match(scanner.FALSE) {
		return &ast.Literal{Value: false}, nil
	}

	if p.match(scanner.TRUE) {
		return &ast.Literal{Value: true}, nil
	}

	if p.match(scanner.NIL) {
		return &ast.Literal{Value: nil}, nil
	}

	if p.match(scanner.NUMBER_INT, scanner.NUMBER_REAL, scanner.STRING) {
		return &ast.Literal{Value: p.previous().Literal}, nil
	}

	if p.match(scanner.LEFT_PAREN) {
		expr, err := p.expression()

		if err != nil {
			return nil, err
		}

		_, err = p.consumeOrError(scanner.RIGHT_PAREN, "Expect ')' after expression.")

		if err != nil {
			return nil, err
		}

		return &ast.Grouping{Expression: expr}, nil
	}

	return nil, NewParseErrorWithLog("expect expression", p.peek())
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().TokenType == scanner.SEMICOLON {
			return
		}

		switch p.peek().TokenType {
		case scanner.CLASS, scanner.FUN, scanner.VAR, scanner.FOR,
			scanner.IF, scanner.WHILE, scanner.PRINT, scanner.RETURN:
			return
		}

		p.advance()
	}
}

// ----------------------------------------------------------------
// helpers
// ----------------------------------------------------------------

func (p *Parser) match(types ...scanner.TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

func (p *Parser) check(t scanner.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().TokenType == t
}

func (p *Parser) advance() *scanner.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == scanner.EOF
}

func (p *Parser) peek() *scanner.Token {
	return &p.tokens[p.current]
}

func (p *Parser) previous() *scanner.Token {
	return &p.tokens[p.current-1]
}

func (p *Parser) consumeOrError(t scanner.TokenType, message string) (*scanner.Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return nil, NewParseErrorWithLog(message, p.peek())
}
