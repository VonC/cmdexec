// Get a shell. Can be called to execute a command.
// Will block on the execution, bug a log of stdout and stderr will be produced
// as the command execute
package shell

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"time"
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

type Status struct {
	success bool
	stdout  string
	exit    string
}

func (st *Status) IsSuccessful() bool {
	return st.success
}
func (st *Status) Stdout() string {
	return st.stdout
}
func (st *Status) Exit() string {
	return st.exit
}

type Shell struct {
	cmd    *exec.Cmd
	scmd   string
	start  *time.Time
	stdout *io.ReadCloser
	stderr *io.ReadCloser
	stdin  *io.WriteCloser
	cout   <-chan string
	status Status
}

type stateFn func(*Shell) stateFn

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
			// fmt.Println("startRead: Read")
			n, err := (*pipe).Read(buf)
			if n > 0 {
				// fmt.Println("startRead: sending " + strconv.Itoa(n) + " charaters")
				c <- string(buf[0:n])
				// fmt.Println("startRead: sent " + strconv.Itoa(n) + " charaters")
			}
			if err != nil {
				// fmt.Println("startRead: break")
				break
			}
		}
	}()
	return c
}

func waitForEnd(s *Shell) stateFn {
	r := regexp.MustCompile(
		"(?m)^.*?(?:>)?" +
			regexp.QuoteMeta("echo %ERRORLEVEL%~~~:"+s.start.String()) +
			"\\s*" +
			"([^%]*?)" + regexp.QuoteMeta("~~~:"+s.start.String()) +
			"",
	)
	// fmt.Println("waitForEnd: looking for '" + r.String() + "'")
	for readout := range s.cout {
		// fmt.Println("waitForEnd: received '" + readout + "'")
		s.status.stdout = s.status.stdout + readout
		if loc := r.FindStringSubmatchIndex(s.status.stdout); loc != nil {
			s.status.exit = s.status.stdout[loc[2]:loc[3]]
			if s.status.exit == "0" {
				s.status.success = true
			}
			// fmt.Println("waitForEnd: found status '" + s.status.exit + "'")
			s.status.stdout = s.status.stdout[:loc[0]]
			return nil
		}
	}
	// If the end command isn't found, block forever.
	return nil
}

func waitForCmd(s *Shell) stateFn {
	timeout := time.After(2 * time.Second)
	r := regexp.MustCompilePOSIX("^(.*?>)?" + regexp.QuoteMeta(s.scmd) + "[[:space:]]*?$[\r\n]*")
	// fmt.Println("waitForCmd: looking for '" + r.String() + "'")
	for {
		select {
		case readout := <-s.cout:
			// fmt.Println("waitForCmd: received '" + readout + "'")
			s.status.stdout = s.status.stdout + readout
			if loc := r.FindStringIndex(s.status.stdout); loc != nil {
				s.status.stdout = s.status.stdout[loc[1]:]
				return waitForEnd(s)
			}
		case _ = <-timeout:
			fmt.Println("waitForCmd: No cmd detected on shell?! '" + s.scmd + "'")
			return nil // should actually Panic
		}
	}
	return nil
}

func (s *Shell) run(end chan int) {
	fmt.Println("run: closing")
	close(end)
}

// Synchronous function (will block until the command sent to the shell 
// complete)
func (s *Shell) Exec(cmd string) *Status {

	now := time.Now()
	s.start = &now
	s.scmd = cmd

	end := "~~~:" + now.String()
	_, err := (*s.stdin).Write([]byte(cmd + "\n"))
	checkError(err)
	_, err = (*s.stdin).Write([]byte("echo %ERRORLEVEL%" + end + "\n"))
	checkError(err)

	s.status.stdout = ""
	for state := waitForCmd(s); state != nil; {
		state = state(s)
	}
	// fmt.Println("Exec: out of the select and the for range endch")
	return &s.status
}
