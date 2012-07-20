package main

import (
	"cmdexec/shell"
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("hello cmdexec! ")
	s := shell.NewShell()
	status := s.Exec("dir")
	fmt.Println("done: " + strconv.FormatBool(status.Success))

	status = s.Exec("echo Hello World!")
	fmt.Println("done: " + strconv.FormatBool(status.Success))

}
