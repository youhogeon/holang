package log

import "fmt"

func StdOut(v ...any) (n int, err error) {
	return fmt.Print(v...)
}

func StdOutf(format string, a ...any) (n int, err error) {
	return fmt.Printf(format, a...)
}

func Print(v ...any) (n int, err error) {
	Debug("Print", A("values", v))

	return StdOut(v...)
}

func Printf(format string, a ...any) (n int, err error) {
	formatted := fmt.Sprintf(format, a...)

	Debug("Print", S("formatted", formatted), S("format", format))

	return StdOut(formatted)
}

func Println(a ...any) (n int, err error) {
	n, err = Print(a...)

	StdOut("\n")

	return
}
