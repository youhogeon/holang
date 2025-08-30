package log

func LangError(msg string, where string, line int, fields ...Field) {
	f := make([]Field, 0, len(fields)+2)
	f = append(f, fields...)
	f = append(f, S("where", where), A("line", line))

	Error(msg, f...)

	StdOutf("[line %d] Error at %s: %s\n", line, where, msg)
}
