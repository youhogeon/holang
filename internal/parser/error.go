package parser

import (
	"internal/token"
	"internal/util/log"
)

type ParseError struct {
	Message string
}

func NewParseErrorWithLog(message string, token *token.Token) *ParseError {
	err := &ParseError{
		Message: message,
	}

	log.Error("Parse error", log.E(err), log.A("token", token))

	return err
}

func (e *ParseError) Error() string {
	return e.Message
}
