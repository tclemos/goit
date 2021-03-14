package log

import "fmt"

func Log(args ...interface{}) {
	prefix()
	fmt.Print(args...)
	fmt.Println()
}

func Logf(format string, args ...interface{}) {
	prefix()
	fmt.Printf(format, args...)
	fmt.Println()
}

func Error(err error, args ...interface{}) {
	prefix()
	fmt.Print(args...)
	fmt.Print(" ")
	fmt.Print("err: ")
	fmt.Print(err)
	fmt.Println()
}

func Errorf(err error, format string, args ...interface{}) {
	prefix()
	fmt.Printf(format, args...)
	fmt.Print(" ")
	fmt.Print("err: ")
	fmt.Print(err)
	fmt.Println()
}

func prefix() {
	fmt.Print("[goit]: ")
}
