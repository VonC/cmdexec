// Get a shell and execute a command
package shell

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

type Shell struct {
	cmd    *exec.Cmd
	start  *time.Time
	stdout *io.ReadCloser
	stderr *io.ReadCloser
	stdin  *io.WriteCloser
	cout   <-chan string
}

type stateFn func(*Shell) stateFn

type Status struct {
	Success bool
	Stdout  string
}

func NewShell() *Shell {

	acmd := exec.Command("cmd", "/K")

	// Create stdout, stderr streams of type io.Reader
	astdout, err := acmd.StdoutPipe()
	checkError(err)
	astderr, err := acmd.StderrPipe()
	checkError(err)
	astdin, err := acmd.StdinPipe()
	checkError(err)

	// Start command
	err = acmd.Start()
	checkError(err)

	go acmd.Wait()
	return &Shell{
		cmd:    acmd,
		stdout: &astdout,
		stdin:  &astdin,
		stderr: &astderr,
		cout:   startRead(&astdout),
	}
}

func startRead(pipe *io.ReadCloser) <-chan string {
	c := make(chan string)
	buf := make([]byte, 32*1024)
	go func() {
		for {
			fmt.Println("startRead: Read")
			n, err := (*pipe).Read(buf)
			if n > 0 {
				fmt.Println("startRead: sending " + strconv.Itoa(n) + " charaters")
				c <- string(buf[0:n])
				fmt.Println("startRead: sent " + strconv.Itoa(n) + " charaters")
			}
			if err != nil {
				fmt.Println("startRead: break")
				break
			}
		}
	}()
	return c
}

// Synchronous function (will block until the command sent to the shell 
// complete)
func (s *Shell) Exec(cmd string) *Status {

	now := time.Now()
	s.start = &now
	end := "~~~:" + now.String()
	_, err := (*s.stdin).Write([]byte(cmd + "\n"))
	checkError(err)
	_, err = (*s.stdin).Write([]byte("echo %ERRORLEVEL%" + end + "\n"))
	checkError(err)

	re, err := regexp.CompilePOSIX("^[^%]*?" + regexp.QuoteMeta(end) + "[\r\n]*?$")
	checkError(err)
	fmt.Println(re)
	lend := len(end) + 7

	out := ""
	ok := true
	readout := ""
	for {
		select {
		case readout, ok = <-s.cout:
			if ok {
				fmt.Println("Exec: received '" + readout + "'")
				out = out + readout
				if re.FindAllStringIndex(out, len(out)-lend) != nil {
					ok = false
				}
			} else {
				fmt.Println("Exec: breaking")
				break
			}
		}
		if !ok {
			break
		}
	}
	fmt.Println("out of the select")
	return &Status{
		Success: true,
		Stdout:  out,
	}
}
