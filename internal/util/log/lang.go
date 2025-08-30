package log

import "fmt"

func LangError(msg string, where string, line int, fields ...Field) error {
	f := make([]Field, 0, len(fields)+2)
	f = append(f, fields...)
	f = append(f, S("where", where), A("line", line))

	Error(msg, f...)

	err := fmt.Errorf("[line %d] Error at %s: %s", line, where, msg)

	return err
}
