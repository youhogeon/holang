package parser

import (
	"errors"
	"internal/ast"
	"internal/scanner"
)

type Parser struct {
	tokens  []scanner.Token
	current int
}

func NewParser(tokens []scanner.Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() (ast.Expr, error) {
	return p.expression()
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.equality()
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
			Operator: *operator,
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
			Operator: *operator,
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
			Operator: *operator,
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
			Operator: *operator,
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
			Operator: *operator,
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

		_, err = p.consume(scanner.RIGHT_PAREN, "Expect ')' after expression.")

		if err != nil {
			return nil, err
		}

		return &ast.Grouping{Expression: expr}, nil
	}

	return nil, errors.New("expect expression")
}

func (p *Parser) match(types ...scanner.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
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

func (p *Parser) consume(t scanner.TokenType, message string) (*scanner.Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return nil, errors.New(message)
}
