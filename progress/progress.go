// progress prints the program progress on screen. It's similar to a logger, but with
// better formatting.
package progress

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// event stores logs IDs and messages.
type event struct {
	id     int
	format string
}

// Log messages as new Events
var (
	// evStatus  = event{1, "[>] %v"}
	evError   = event{1, "[✗] %v"}
	evRunning = event{2, "[ ] %v"}
	evRunOk   = event{3, "\r[✓]"}
	evRunFail = event{4, "\r[✗]"}
)

const spinners = `/-\|`

const (
	colorReset = "\033[0m"

	colorRed = "\033[31m"
	// colorGreen  = "\033[32m"
	// colorYellow = "\033[33m"
	// colorBlue   = "\033[34m"
	// colorPurple = "\033[35m"
	colorCyan = "\033[36m"
	// colorWhite  = "\033[37m"
)

type Progress struct {
	out     io.Writer // destination for output, usually os.Stderr
	running []byte
	seq     int // sequence of spinners
}

func Open(out io.Writer) *Progress {
	p := Progress{out: out}
	if out == os.Stdout || out == os.Stderr {
		p.output([]byte("\033[?25l")) // Hide cursor
	}
	return &p
}

func (p *Progress) Close() {
	if p.out == os.Stdout || p.out == os.Stderr {
		p.output([]byte("\033[?25h")) // Show cursor
	}
}

func (p *Progress) Status(format string, a ...interface{}) {
	if len(p.running) > 0 {
		p.clearLine()
		p.output([]byte(colorCyan))
	}

	p.outputln("[>] " + fmt.Sprintf(format, a...))

	if len(p.running) > 0 {
		p.output([]byte(colorReset))
		p.output(p.running)
	}
}

func (p *Progress) Error(err error) {
	if len(p.running) > 0 {
		p.clearLine()
	}

	p.output([]byte(colorRed))
	p.outputln(fmt.Sprintf(evError.format, err))
	p.output([]byte(colorReset))

	if len(p.running) > 0 {
		p.output(p.running)
	}

}

func (p *Progress) Running(msg string) {
	p.running = []byte(fmt.Sprintf(evRunning.format, msg))
	p.output(p.running)
}

func (p *Progress) Spinner() {
	p.output([]byte{'\r', '[', spinners[p.seq], ']'})
	p.seq = (p.seq + 1) % len(spinners)
}

func (p *Progress) RunOK() {
	p.outputln(evRunOk.format)
	p.running = p.running[:0]
}

func (p *Progress) RunFail() {
	p.output([]byte(colorRed))

	if len(p.running) > 0 {
		p.clearLine()
		p.output(p.running)
	}

	p.outputln(evRunFail.format)
	p.running = p.running[:0]

	p.output([]byte(colorReset))
}

/* ------- output ------- */

func (p *Progress) clearLine() {
	if len(p.running) == 0 {
		return
	}
	buf := bytes.Repeat([]byte(" "), len(p.running)+2)
	buf[0] = byte('\r')
	buf[len(buf)-1] = byte('\r')
	_, _ = p.out.Write(buf)
}

func (p *Progress) output(buf []byte) {
	_, _ = p.out.Write(buf)
}

func (p *Progress) outputln(s string) {
	if p.out == nil {
		return
	}
	var buf []byte
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	p.output(buf)
}
