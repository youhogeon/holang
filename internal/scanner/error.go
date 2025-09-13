package scanner

import (
	"fmt"
	"internal/util/log"
)

type ScanError struct {
	Message string
	Line    int
	Where   string
}

func NewScanErrorWithLog(message string, line int, where string) *ScanError {
	err := &ScanError{
		Message: message,
		Line:    line,
		Where:   where,
	}

	log.Error("Scan error", log.E(err))

	return err
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("[line %d] Error at %s: %s", e.Line, e.Where, e.Message)
}
