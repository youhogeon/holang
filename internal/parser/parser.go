package parser

import (
	"internal/ast"
	"internal/scanner"
	"slices"
)

type Parser struct {
	tokens    []scanner.Token
	current   int
	loopDepth int
}

func NewParser(tokens []scanner.Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() ([]ast.Stmt, []error) {
	statements := make([]ast.Stmt, 0)
	errors := make([]error, 0)

	for !p.isAtEnd() {
		stmt, err := p.declaration()

		if err != nil {
			p.synchronize()

			errors = append(errors, err)

			continue
		}

		statements = append(statements, stmt)
	}

	return statements, errors
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(scanner.VAR) {
		return p.varDecl()
	}

	return p.statement()
}

func (p *Parser) varDecl() (*ast.Var, error) {
	name, err := p.consumeOrError(scanner.IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var initializer ast.Expr

	if p.match(scanner.EQUAL) {
		initializer, err = p.expression()

		if err != nil {
			return nil, err
		}
	}

	_, err = p.consumeOrError(scanner.SEMICOLON, "Expect ';' after value.")

	if err != nil {
		return nil, err
	}

	return &ast.Var{
		Name:        name,
		Initializer: initializer,
	}, nil
}

func (p *Parser) statement() (ast.Stmt, error) {
	if p.match(scanner.LEFT_BRACE) {
		return p.block()
	}

	if p.match(scanner.PRINT) {
		return p.printStatement()
	}

	if p.match(scanner.IF) {
		return p.ifStatement()
	}

	if p.match(scanner.WHILE) {
		return p.whileStatement()
	}

	if p.match(scanner.FOR) {
		return p.forStatement()
	}

	if p.match(scanner.BREAK) {
		return p.breakStatement()
	}

	if p.match(scanner.CONTINUE) {
		return p.continueStatement()
	}

	return p.exprStatement()
}

func (p *Parser) block() (*ast.Block, error) {
	statements := make([]ast.Stmt, 0)

	for !p.check(scanner.RIGHT_BRACE) && !p.isAtEnd() {
		stmt, err := p.declaration()

		if err != nil {
			return &ast.Block{}, err
		}

		statements = append(statements, stmt)
	}

	_, err := p.consumeOrError(scanner.RIGHT_BRACE, "Expect '}' after block.")
	if err != nil {
		return &ast.Block{}, err
	}

	return &ast.Block{Statements: statements}, nil
}

func (p *Parser) exprStatement() (*ast.Expression, error) {
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

func (p *Parser) printStatement() (*ast.Print, error) {
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

func (p *Parser) ifStatement() (*ast.If, error) {
	_, err := p.consumeOrError(scanner.LEFT_PAREN, "Expect '(' after if.")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(scanner.RIGHT_PAREN, "Expect ')' after if condition.")
	if err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch ast.Stmt

	if p.match(scanner.ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return &ast.If{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) whileStatement() (*ast.While, error) {
	p.loopDepth++
	defer func() { p.loopDepth-- }()

	_, err := p.consumeOrError(scanner.LEFT_PAREN, "Expect '(' after while.")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(scanner.RIGHT_PAREN, "Expect ')' after while condition.")
	if err != nil {
		return nil, err
	}

	branch, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &ast.While{
		Condition: condition,
		Body:      branch,
	}, nil
}

func (p *Parser) forStatement() (ast.Stmt, error) {
	p.loopDepth++
	defer func() { p.loopDepth-- }()

	_, err := p.consumeOrError(scanner.LEFT_PAREN, "Expect '(' after for.")
	if err != nil {
		return nil, err
	}

	var initializer ast.Stmt

	if p.match(scanner.VAR) {
		initializer, err = p.varDecl()
		if err != nil {
			return nil, err
		}
	} else if !p.check(scanner.SEMICOLON) { // for any expr stmt
		initializer, err = p.exprStatement()
		if err != nil {
			return nil, err
		}
	}

	var condition ast.Expr

	if !p.check(scanner.SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consumeOrError(scanner.SEMICOLON, "Expect ';' after for condition.")
	if err != nil {
		return nil, err
	}

	var increment ast.Expr

	if !p.check(scanner.RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consumeOrError(scanner.RIGHT_PAREN, "Expect ')' after for clauses.")
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	// desugar for loop into while loop
	if increment != nil {
		body = &ast.Block{
			Statements: []ast.Stmt{
				body,
				&ast.Expression{Expression: increment},
			},
		}
	}

	if condition == nil {
		condition = &ast.Literal{Value: true}
	}

	body = &ast.While{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = &ast.Block{
			Statements: []ast.Stmt{
				initializer,
				body,
			},
		}
	}

	return body, nil
}

func (p *Parser) breakStatement() (*ast.Break, error) {
	if p.loopDepth == 0 {
		return nil, NewParseErrorWithLog("break statement not within a loop", p.previous())
	}

	_, err := p.consumeOrError(scanner.SEMICOLON, "Expect ';' after break.")
	if err != nil {
		return nil, err
	}

	return &ast.Break{}, nil
}

func (p *Parser) continueStatement() (*ast.Continue, error) {
	if p.loopDepth == 0 {
		return nil, NewParseErrorWithLog("continue statement not within a loop", p.previous())
	}

	_, err := p.consumeOrError(scanner.SEMICOLON, "Expect ';' after continue.")
	if err != nil {
		return nil, err
	}

	return &ast.Continue{}, nil
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (ast.Expr, error) {
	expr, err := p.ternary()

	if err != nil {
		return nil, err
	}

	if p.match(scanner.EQUAL) {
		equals := p.previous()
		value, err := p.assignment()

		if err != nil {
			return nil, err
		}

		if variable, ok := expr.(*ast.Variable); ok {
			name := variable.Name

			return &ast.Assign{
				Name:  name,
				Value: value,
			}, nil
		}

		return nil, NewParseErrorWithLog("invalid assignment target", equals)
	}

	return expr, nil
}

func (p *Parser) ternary() (ast.Expr, error) {
	expr, err := p.logicOr()

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

func (p *Parser) logicOr() (ast.Expr, error) {
	expr, err := p.logicAnd()

	if err != nil {
		return nil, err
	}

	for p.match(scanner.AND) {
		operator := p.previous()
		right, err := p.logicAnd()

		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) logicAnd() (ast.Expr, error) {
	expr, err := p.equality()

	if err != nil {
		return nil, err
	}

	for p.match(scanner.AND) {
		operator := p.previous()
		right, err := p.equality()

		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
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

	if p.match(scanner.IDENTIFIER) {
		return &ast.Variable{Name: p.previous()}, nil
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
