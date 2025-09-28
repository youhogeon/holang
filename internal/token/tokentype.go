package token

import "fmt"

type TokenType int

const (
	// Single-character tokens.
	LEFT_PAREN TokenType = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR
	QUESTION
	COLON
	PERCENT

	// One or two character tokens.
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals.
	IDENTIFIER
	STRING
	NUMBER_INT
	NUMBER_REAL

	// Keywords.
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE
	BREAK
	CONTINUE

	// ETC
	COMMENT
	MULTI_COMMENT
	EOF
)

var tokenNames = map[TokenType]string{
	//0
	LEFT_PAREN:  "LEFT_PAREN",
	RIGHT_PAREN: "RIGHT_PAREN",
	LEFT_BRACE:  "LEFT_BRACE",
	RIGHT_BRACE: "RIGHT_BRACE",
	COMMA:       "COMMA",
	DOT:         "DOT",
	MINUS:       "MINUS",
	PLUS:        "PLUS",
	SEMICOLON:   "SEMICOLON",
	SLASH:       "SLASH",
	//10
	QUESTION:      "QUESTION",
	COLON:         "COLON",
	PERCENT:       "PERCENT",
	STAR:          "STAR",
	BANG:          "BANG",
	BANG_EQUAL:    "BANG_EQUAL",
	EQUAL:         "EQUAL",
	EQUAL_EQUAL:   "EQUAL_EQUAL",
	GREATER:       "GREATER",
	GREATER_EQUAL: "GREATER_EQUAL",
	//20
	LESS:        "LESS",
	LESS_EQUAL:  "LESS_EQUAL",
	IDENTIFIER:  "IDENTIFIER",
	STRING:      "STRING",
	NUMBER_INT:  "NUMBER_INT",
	NUMBER_REAL: "NUMBER_REAL",
	AND:         "AND",
	CLASS:       "CLASS",
	ELSE:        "ELSE",
	FALSE:       "FALSE",
	//30
	FUN:    "FUN",
	FOR:    "FOR",
	IF:     "IF",
	NIL:    "NIL",
	OR:     "OR",
	PRINT:  "PRINT",
	RETURN: "RETURN",
	SUPER:  "SUPER",
	THIS:   "THIS",
	TRUE:   "TRUE",
	//40
	VAR:           "VAR",
	WHILE:         "WHILE",
	BREAK:         "BREAK",
	CONTINUE:      "CONTINUE",
	COMMENT:       "COMMENT",
	MULTI_COMMENT: "MULTI_COMMENT",
	EOF:           "EOF",
}

func (t *TokenType) String() string {
	if name, ok := tokenNames[*t]; ok {
		return name
	}

	return fmt.Sprintf("UNKNOWN(%d)", *t)
}
