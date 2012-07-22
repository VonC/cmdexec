package main

import (
	"cmdexec/shell"
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("hello cmdexec! ")
	s := shell.NewShell()
	exec(s, "dir")
	exec(s, "echo a & dummyCommand & echo b")
	exec(s, "echo Hello World!")
}

func exec(s *shell.Shell, cmd string) {
	fmt.Println("Exec " + cmd)
	// fmt.Println("vvvvvvvvvvv")
	status := s.Exec(cmd)
	// fmt.Println("^^^^^^^^^^^")
	fmt.Println("done: " + strconv.FormatBool(status.IsSuccessful()) + ": '" + status.Exit() + "'")
	fmt.Println("out: '" + status.Stdout() + "'")
}
