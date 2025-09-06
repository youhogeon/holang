package interpreter

import "internal/util/log"

type RuntimeError struct {
	Message string
}

func NewRuntimeErrorWithLog(message string) *RuntimeError {
	err := &RuntimeError{
		Message: message,
	}

	log.Error("Runtime error", log.E(err))

	return err
}

func (e *RuntimeError) Error() string {
	return e.Message
}
