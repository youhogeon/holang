package parser

import (
	"internal/ast"
	"internal/token"
	"slices"
)

type Parser struct {
	tokens    []token.Token
	current   int
	loopDepth int

	nextId int
}

func NewParser(tokens []token.Token) *Parser {
	tokensWithoutComments := make([]token.Token, 0, len(tokens))
	for _, t := range tokens {
		if t.TokenType != token.COMMENT && t.TokenType != token.MULTI_COMMENT {
			tokensWithoutComments = append(tokensWithoutComments, t)
		}
	}

	return &Parser{
		tokens: tokensWithoutComments,
	}
}

func (p *Parser) Parse() (*ast.Program, []error) {
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

	program := &ast.Program{
		Statements: statements,
	}

	return program, errors
}

func (p *Parser) getNodeId() int {
	id := p.nextId
	p.nextId++

	return id
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(token.VAR) {
		return p.varDecl()
	}

	if p.match(token.CLASS) {
		return p.classDecl()
	}

	if p.match(token.FUN) {
		return p.funDecl()
	}

	return p.statement()
}

func (p *Parser) varDecl() (*ast.Var, error) {
	name, err := p.consumeOrError(token.IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var initializer ast.Expr

	if p.match(token.EQUAL) {
		initializer, err = p.expression()

		if err != nil {
			return nil, err
		}
	}

	_, err = p.consumeOrError(token.SEMICOLON, "Expect ';' after value.")

	if err != nil {
		return nil, err
	}

	return &ast.Var{
		Name:        name,
		Initializer: initializer,
		Offset:      name.Offset,
		NodeId:      p.getNodeId(),
	}, nil
}

func (p *Parser) classDecl() (*ast.Class, error) {
	name, err := p.consumeOrError(token.IDENTIFIER, "Expect class name.")
	if err != nil {
		return nil, err
	}

	var superclass *ast.Variable

	if p.match(token.LESS) {
		superToken, err := p.consumeOrError(token.IDENTIFIER, "Expect superclass name.")
		if err != nil {
			return nil, err
		}

		superclass = &ast.Variable{
			Name:   superToken,
			Offset: superToken.Offset,
			NodeId: p.getNodeId(),
		}

		if name.Lexeme == superToken.Lexeme {
			return nil, NewParseErrorWithLog("a class cannot inherit from itself", superToken)
		}
	}

	_, err = p.consumeOrError(token.LEFT_BRACE, "Expect '{' before class body.")
	if err != nil {
		return nil, err
	}

	methods := make([]*ast.Function, 0)

	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		method, err := p.funDecl()
		if err != nil {
			return nil, err
		}

		methods = append(methods, method)
	}

	_, err = p.consumeOrError(token.RIGHT_BRACE, "Expect '}' after class body.")
	if err != nil {
		return nil, err
	}

	return &ast.Class{
		Name:       name,
		Methods:    methods,
		Superclass: superclass,
		Offset:     name.Offset,
		NodeId:     p.getNodeId(),
	}, nil
}

func (p *Parser) funDecl() (*ast.Function, error) {
	name, err := p.consumeOrError(token.IDENTIFIER, "Expect function name.")
	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(token.LEFT_PAREN, "Expect '(' after function name.")
	if err != nil {
		return nil, err
	}

	parameters := make([]*token.Token, 0)

	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				return nil, NewParseErrorWithLog("can't have more than 255 parameters", p.peek())
			}

			param, err := p.consumeOrError(token.IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, err
			}

			parameters = append(parameters, param)

			if !p.match(token.COMMA) {
				break
			}
		}
	}

	_, err = p.consumeOrError(token.RIGHT_PAREN, "Expect ')' after parameters.")
	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(token.LEFT_BRACE, "Expect '{' before function body.")
	if err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return &ast.Function{
		Name:   name,
		Params: parameters,
		Body:   body.Statements,
		Offset: name.Offset,
		NodeId: p.getNodeId(),
	}, nil
}

func (p *Parser) statement() (ast.Stmt, error) {
	if p.match(token.LEFT_BRACE) {
		return p.block()
	}

	if p.match(token.PRINT) {
		return p.printStatement()
	}

	if p.match(token.IF) {
		return p.ifStatement()
	}

	if p.match(token.WHILE) {
		return p.whileStatement()
	}

	if p.match(token.FOR) {
		return p.forStatement()
	}

	if p.match(token.BREAK) {
		return p.breakStatement()
	}

	if p.match(token.CONTINUE) {
		return p.continueStatement()
	}

	if p.match(token.RETURN) {
		return p.returnStatement()
	}

	return p.exprStatement()
}

func (p *Parser) block() (*ast.Block, error) {
	offset := p.previous().Offset
	statements := make([]ast.Stmt, 0)

	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {

		stmt, err := p.declaration()

		if err != nil {
			return &ast.Block{
				Offset: offset,
				NodeId: p.getNodeId(),
			}, err
		}

		statements = append(statements, stmt)
	}

	_, err := p.consumeOrError(token.RIGHT_BRACE, "Expect '}' after block.")
	if err != nil {
		return &ast.Block{
			Offset: offset,
			NodeId: p.getNodeId(),
		}, err
	}

	return &ast.Block{
		Statements: statements,
		Offset:     offset,
		NodeId:     p.getNodeId(),
	}, nil
}

func (p *Parser) exprStatement() (*ast.Expression, error) {
	offset := p.peek().Offset
	expr, err := p.expression()

	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(token.SEMICOLON, "Expect ';' after value.")

	if err != nil {
		return nil, err
	}

	return &ast.Expression{
		Expression: expr,
		Offset:     offset,
		NodeId:     p.getNodeId(),
	}, nil
}

func (p *Parser) printStatement() (*ast.Print, error) {
	offset := p.previous().Offset
	expr, err := p.expression()

	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(token.SEMICOLON, "Expect ';' after value.")

	if err != nil {
		return nil, err
	}

	return &ast.Print{
		Expression: expr,
		Offset:     offset,
		NodeId:     p.getNodeId(),
	}, nil
}

func (p *Parser) ifStatement() (*ast.If, error) {
	offset := p.previous().Offset
	_, err := p.consumeOrError(token.LEFT_PAREN, "Expect '(' after if.")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(token.RIGHT_PAREN, "Expect ')' after if condition.")
	if err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch ast.Stmt

	if p.match(token.ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return &ast.If{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
		Offset:     offset,
		NodeId:     p.getNodeId(),
	}, nil
}

func (p *Parser) whileStatement() (*ast.While, error) {
	offset := p.previous().Offset
	p.loopDepth++
	defer func() { p.loopDepth-- }()

	_, err := p.consumeOrError(token.LEFT_PAREN, "Expect '(' after while.")
	if err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consumeOrError(token.RIGHT_PAREN, "Expect ')' after while condition.")
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
		Offset:    offset,
		NodeId:    p.getNodeId(),
	}, nil
}

func (p *Parser) forStatement() (ast.Stmt, error) {
	offset := p.previous().Offset

	p.loopDepth++
	defer func() { p.loopDepth-- }()

	_, err := p.consumeOrError(token.LEFT_PAREN, "Expect '(' after for.")
	if err != nil {
		return nil, err
	}

	var initializer ast.Stmt

	if p.match(token.VAR) {
		initializer, err = p.varDecl()
		if err != nil {
			return nil, err
		}
	} else if !p.check(token.SEMICOLON) { // for any expr stmt
		initializer, err = p.exprStatement()
		if err != nil {
			return nil, err
		}
	}

	var condition ast.Expr

	if !p.check(token.SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consumeOrError(token.SEMICOLON, "Expect ';' after for condition.")
	if err != nil {
		return nil, err
	}

	var increment ast.Expr

	if !p.check(token.RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consumeOrError(token.RIGHT_PAREN, "Expect ')' after for clauses.")
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
				&ast.Expression{
					Expression: increment,
					Offset:     offset,
					NodeId:     p.getNodeId(),
				},
			},
			Offset: offset,
			NodeId: p.getNodeId(),
		}
	}

	if condition == nil {
		condition = &ast.Literal{Value: true}
	}

	body = &ast.While{
		Condition: condition,
		Body:      body,
		Offset:    offset,
		NodeId:    p.getNodeId(),
	}

	if initializer != nil {
		body = &ast.Block{
			Statements: []ast.Stmt{
				initializer,
				body,
			},
			Offset: offset,
			NodeId: p.getNodeId(),
		}
	}

	return body, nil
}

func (p *Parser) breakStatement() (*ast.Break, error) {
	offset := p.previous().Offset

	if p.loopDepth == 0 {
		return nil, NewParseErrorWithLog("break statement not within a loop", p.previous())
	}

	_, err := p.consumeOrError(token.SEMICOLON, "Expect ';' after break.")
	if err != nil {
		return nil, err
	}

	return &ast.Break{
		Offset: offset,
		NodeId: p.getNodeId(),
	}, nil
}

func (p *Parser) continueStatement() (*ast.Continue, error) {
	offset := p.previous().Offset

	if p.loopDepth == 0 {
		return nil, NewParseErrorWithLog("continue statement not within a loop", p.previous())
	}

	_, err := p.consumeOrError(token.SEMICOLON, "Expect ';' after continue.")
	if err != nil {
		return nil, err
	}

	return &ast.Continue{
		Offset: offset,
		NodeId: p.getNodeId(),
	}, nil
}

func (p *Parser) returnStatement() (*ast.Return, error) {
	keyword := p.previous()

	var value ast.Expr
	var err error

	if !p.check(token.SEMICOLON) {
		value, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consumeOrError(token.SEMICOLON, "Expect ';' after return value.")
	if err != nil {
		return nil, err
	}

	return &ast.Return{
		Keyword: keyword,
		Value:   value,
		Offset:  keyword.Offset,
		NodeId:  p.getNodeId(),
	}, nil
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (ast.Expr, error) {
	expr, err := p.ternary()

	if err != nil {
		return nil, err
	}

	if p.match(token.EQUAL) {
		equals := p.previous()
		value, err := p.assignment()

		if err != nil {
			return nil, err
		}

		if variable, ok := expr.(*ast.Variable); ok {
			name := variable.Name

			return &ast.Assign{
				Name:   name,
				Value:  value,
				Offset: name.Offset,
				NodeId: p.getNodeId(),
			}, nil
		}

		if get, ok := expr.(*ast.Get); ok {
			return &ast.Set{
				Object: get.Object,
				Name:   get.Name,
				Value:  value,
				Offset: get.Name.Offset,
				NodeId: p.getNodeId(),
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

	if p.match(token.QUESTION) {
		firstOp := p.previous()
		mid, err := p.expression()

		if err != nil {
			return nil, err
		}

		secondOp, err := p.consumeOrError(token.COLON, "Expect ':' after expression.")

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
			Offset:         firstOp.Offset,
			NodeId:         p.getNodeId(),
		}, nil
	}

	return expr, nil
}

func (p *Parser) logicOr() (ast.Expr, error) {
	expr, err := p.logicAnd()

	if err != nil {
		return nil, err
	}

	for p.match(token.OR) {
		operator := p.previous()
		right, err := p.logicAnd()

		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
			Offset:   operator.Offset,
			NodeId:   p.getNodeId(),
		}
	}

	return expr, nil
}

func (p *Parser) logicAnd() (ast.Expr, error) {
	expr, err := p.equality()

	if err != nil {
		return nil, err
	}

	for p.match(token.AND) {
		operator := p.previous()
		right, err := p.equality()

		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{
			Left:     expr,
			Operator: operator,
			Right:    right,
			Offset:   operator.Offset,
			NodeId:   p.getNodeId(),
		}
	}

	return expr, nil
}

func (p *Parser) equality() (ast.Expr, error) {
	expr, err := p.comparison()

	if err != nil {
		return nil, err
	}

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
			Offset:   operator.Offset,
			NodeId:   p.getNodeId(),
		}
	}

	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.term()

	if err != nil {
		return nil, err
	}

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
			Offset:   operator.Offset,
			NodeId:   p.getNodeId(),
		}
	}

	return expr, nil
}

func (p *Parser) term() (ast.Expr, error) {
	expr, err := p.factor()

	if err != nil {
		return nil, err
	}

	for p.match(token.MINUS, token.PLUS) {
		operator := p.previous()
		right, err := p.factor()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
			Offset:   operator.Offset,
			NodeId:   p.getNodeId(),
		}
	}

	return expr, nil
}

func (p *Parser) factor() (ast.Expr, error) {
	expr, err := p.unary()

	if err != nil {
		return nil, err
	}

	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,
			Offset:   operator.Offset,
			NodeId:   p.getNodeId(),
		}
	}

	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(token.BANG, token.MINUS) {
		operator := p.previous()
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		return &ast.Unary{
			Operator: operator,
			Right:    right,
			Offset:   operator.Offset,
			NodeId:   p.getNodeId(),
		}, nil
	}

	return p.call()
}

func (p *Parser) call() (ast.Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(token.LEFT_PAREN) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(token.DOT) {
			name, err := p.consumeOrError(token.IDENTIFIER, "Expect property name after '.'.")
			if err != nil {
				return nil, err
			}

			expr = &ast.Get{
				Object: expr,
				Name:   name,
				Offset: name.Offset,
				NodeId: p.getNodeId(),
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	arguments := make([]ast.Expr, 0)

	if !p.check(token.RIGHT_PAREN) {
		for {
			arg, err := p.expression()
			if err != nil {
				return nil, err
			}

			arguments = append(arguments, arg)

			if !p.match(token.COMMA) {
				break
			}
		}
	}

	if len(arguments) >= 255 {
		return nil, NewParseErrorWithLog("can't have more than 255 arguments", p.peek())
	}

	paren, err := p.consumeOrError(token.RIGHT_PAREN, "Expect ')' after arguments.")
	if err != nil {
		return nil, err
	}

	return &ast.Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: arguments,
		Offset:    paren.Offset,
		NodeId:    p.getNodeId(),
	}, nil
}

func (p *Parser) primary() (ast.Expr, error) {
	offset := p.peek().Offset

	if p.match(token.FALSE) {
		return &ast.Literal{
			Value:  false,
			Offset: offset,
			NodeId: p.getNodeId(),
		}, nil
	}

	if p.match(token.TRUE) {
		return &ast.Literal{
			Value:  true,
			Offset: offset,
			NodeId: p.getNodeId(),
		}, nil
	}

	if p.match(token.NIL) {
		return &ast.Literal{
			Value:  nil,
			Offset: offset,
			NodeId: p.getNodeId(),
		}, nil
	}

	if p.match(token.THIS) {
		return &ast.This{
			Keyword: p.previous(),
			Offset:  offset,
			NodeId:  p.getNodeId(),
		}, nil
	}

	if p.match(token.NUMBER_INT, token.NUMBER_REAL, token.STRING) {
		return &ast.Literal{
			Value:  p.previous().Literal,
			Offset: offset,
			NodeId: p.getNodeId(),
		}, nil
	}

	if p.match(token.IDENTIFIER) {
		return &ast.Variable{
			Name:   p.previous(),
			Offset: offset,
			NodeId: p.getNodeId(),
		}, nil
	}

	if p.match(token.LEFT_PAREN) {
		expr, err := p.expression()

		if err != nil {
			return nil, err
		}

		_, err = p.consumeOrError(token.RIGHT_PAREN, "Expect ')' after expression.")

		if err != nil {
			return nil, err
		}

		return &ast.Grouping{
			Expression: expr,
			Offset:     offset,
			NodeId:     p.getNodeId(),
		}, nil
	}

	if p.match(token.SUPER) {
		keyword := p.previous()

		_, err := p.consumeOrError(token.DOT, "Expect '.' after 'super'.")
		if err != nil {
			return nil, err
		}

		method, err := p.consumeOrError(token.IDENTIFIER, "Expect superclass method name.")
		if err != nil {
			return nil, err
		}

		return &ast.Super{
			Keyword: keyword,
			Method:  method,
			Offset:  offset,
			NodeId:  p.getNodeId(),
		}, nil
	}

	return nil, NewParseErrorWithLog("expect expression", p.peek())
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().TokenType == token.SEMICOLON {
			return
		}

		switch p.peek().TokenType {
		case token.CLASS, token.FUN, token.VAR, token.FOR,
			token.IF, token.WHILE, token.PRINT, token.RETURN:
			return
		}

		p.advance()
	}
}

// ----------------------------------------------------------------
// helpers
// ----------------------------------------------------------------

func (p *Parser) match(types ...token.TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

func (p *Parser) check(t token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().TokenType == t
}

func (p *Parser) advance() *token.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == token.EOF
}

func (p *Parser) peek() *token.Token {
	return &p.tokens[p.current]
}

func (p *Parser) previous() *token.Token {
	return &p.tokens[p.current-1]
}

func (p *Parser) consumeOrError(t token.TokenType, message string) (*token.Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return nil, NewParseErrorWithLog(message, p.peek())
}
