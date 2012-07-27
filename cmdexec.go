// add comment for demo
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
	exec(s, "del /P .mm1")
	exec(s, "del /P .mm", "n")
	exec(s, "echo Hello World!")
}

func exec(s *shell.Shell, cmd string, ins ...string) {
	fmt.Println("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	fmt.Println("Exec " + cmd)
	// fmt.Println("vvvvvvvvvvv")
	status := s.Exec(cmd, ins...)
	// fmt.Println("^^^^^^^^^^^")
	fmt.Println("done: " + strconv.FormatBool(status.IsSuccessful()) + ": '" + status.Exit() + "'")
	fmt.Println("out: '" + status.Stdout() + "'")
	fmt.Println("outputs:")
	for _, anOutput := range status.Outputs() {
		fmt.Print("OUT: '" + anOutput.Stdout + "' - ")
		fmt.Println("ERR: '" + anOutput.Stderr + "'")
	}
	if !status.IsSuccessful() && status.Exit() == "0" {
		fmt.Println("WARNING: exit status 0 with messages on stderr!?")
	}
}
