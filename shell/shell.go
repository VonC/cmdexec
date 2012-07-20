// Get a shell and execute a command
package shell

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Shell struct {
	cmd    *exec.Cmd
	start  *time.Time
	stdout *AsyncReader
	stderr *AsyncReader
	stdin  io.WriteCloser
}

type Status struct {
	Success bool
	stdout  string
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

type AsyncReader struct {
	pipe  *io.ReadCloser
	outch chan string
	n     int
	out   string
	buf   []byte
}

func (ar *AsyncReader) assyncRead() {
	for {
		fmt.Println("pipe.read")
		ar.n, _ = (*ar.pipe).Read(ar.buf)
		fmt.Println("read! " + strconv.Itoa(ar.n) + ", '" + string(ar.buf[0:ar.n]) + "'")
		if ar.n > 0 {
			fmt.Println("Sending")
			ar.outch <- string(ar.buf[0:ar.n])
			fmt.Println("Sending SENT")
		} else {
			fmt.Println("Sending empty")
			ar.outch <- ""
			fmt.Println("Sending SENT empty")
		}
		fmt.Println("sent")
	}
}
func newAsyncReader(apipe *io.ReadCloser) *AsyncReader {
	anAsyncReader := &AsyncReader{
		pipe:  apipe,
		outch: make(chan string),
		buf:   make([]byte, 32*1024),
	}
	fmt.Println("start read")
	return anAsyncReader
}

func (ar *AsyncReader) read() int {
	ar.out = ""
	fmt.Println("wait on outch")
	select {
	case n, ok := <-ar.outch:
		if ok {
			fmt.Println("done wait on outch: '" + n + "'")
			ar.out = <-ar.outch
		} else {
			fmt.Println("outch is closed!")
			break
		}
	default:
		fmt.Println("default read")
	}
	return len(ar.out)
}
func NewShell() *Shell {

	acmd := exec.Command("cmd", "/K")

	// Create stdout, stderr streams of type io.Reader
	astdout, err := acmd.StdoutPipe()
	checkError(err)
	anAsyncStdout := newAsyncReader(&astdout)
	astderr, err := acmd.StderrPipe()
	checkError(err)
	anAsyncStderr := newAsyncReader(&astderr)
	fmt.Println("stds")
	astdin, err := acmd.StdinPipe()
	checkError(err)

	// Start command
	err = acmd.Start()
	checkError(err)

	go anAsyncStdout.assyncRead()

	// Don't let main() exit before our command has finished running
	// defer acmd.Wait() // Doesn't block

	// Non-blockingly echo command output to terminal
	//go io.Copy(os.Stdout, astdout)
	//go io.Copy(os.Stderr, stderr)
	fmt.Println("Wait")
	//go anAsyncStdout.assyncRead()
	go acmd.Wait()
	fmt.Println("Waiting")
	return &Shell{
		cmd:    acmd,
		stdout: anAsyncStdout,
		stdin:  astdin,
		stderr: anAsyncStderr,
	}
}

func (s *Shell) Exec(cmd string) *Status {

	now := time.Now()
	s.start = &now
	_, err := s.stdin.Write([]byte(cmd + "\n"))
	checkError(err)
	_, err = s.stdin.Write([]byte("echo %ERRORLEVEL%~~~:" + now.String() + "\n"))
	checkError(err)

	fmt.Println("Exec " + cmd)

	time.Sleep(1 * 1e9)
	b := make([]byte, 32*1024)
	//n, _ := (*s.stdout.pipe).Read(b)
	n := 0
	if n > 0 {
		fmt.Println("read ====> " + strconv.Itoa(n) + ", '" + string(b[0:n]) + "'")
	}

	sout := ""
	mustBreakAfterNextStdout := false
	mustBreak := false
	lim := 10
	for {
		switch nr := s.stdout.read(); true {
		case nr < 0:
			fmt.Println("nr <!< '", nr, "'")
		case nr == 0: // EOF
			//fmt.Println("nr !>! done reading")
			if mustBreakAfterNextStdout {
				mustBreak = true
			}
			lim = lim - 1
			if lim <= 0 {
				mustBreak = true
			}
		case nr > 0:
			m := s.stdout.out
			fmt.Println("m >>> '", m, "'")
			sout = sout + m
			fmt.Println("m ??? '", now.String(), " contains? "+strconv.FormatBool(strings.Contains(sout, "~~~:"+now.String())))
			if strings.Contains(m, "~~~:"+now.String()) {
				mustBreakAfterNextStdout = true
			}
		default:
			fmt.Println("default")
		}
		if mustBreak {
			break
		}
	}
	return &Status{
		Success: true,
	}
}
