package token

type Offset struct {
	Line  int
	Index int
}

type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   any
	Offset    Offset
}

var Keywords = map[string]TokenType{
	"and":      AND,
	"class":    CLASS,
	"else":     ELSE,
	"false":    FALSE,
	"fun":      FUN,
	"for":      FOR,
	"if":       IF,
	"nil":      NIL,
	"or":       OR,
	"print":    PRINT,
	"return":   RETURN,
	"super":    SUPER,
	"this":     THIS,
	"true":     TRUE,
	"var":      VAR,
	"while":    WHILE,
	"comment":  COMMENT,
	"T":        TRUE,
	"F":        FALSE,
	"break":    BREAK,
	"continue": CONTINUE,
}
