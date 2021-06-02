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
	colorYellow = "\033[33m"
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

var p *Progress

func init() {
	p = &Progress{out: os.Stderr}
}

func Cursor(show bool) {
	if p.out != os.Stdout && p.out != os.Stderr {
		return
	}
	if show {
		output([]byte("\033[?25h")) // Show cursor
	} else {
		output([]byte("\033[?25l")) // Hide cursor
	}
}

func Status(format string, a ...interface{}) {
	if len(p.running) > 0 {
		clearLine()
		output([]byte(colorCyan))
	}

	outputln("[>] " + fmt.Sprintf(format, a...))

	if len(p.running) > 0 {
		output([]byte(colorReset))
		output(p.running)
	}
}

func Error(err error) {
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorRed))
	outputln(fmt.Sprintf(evError.format, err))
	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}

}

func ErrorMsg(format string, a ...interface{}) {
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorRed))
	outputln("[✗] " + fmt.Sprintf(format, a...))
	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}

}

func Warning(format string, a ...interface{}) {
	if len(p.running) > 0 {
		clearLine()
	}

	output([]byte(colorYellow))
	outputln("[!] " + fmt.Sprintf(format, a...))
	output([]byte(colorReset))

	if len(p.running) > 0 {
		output(p.running)
	}

}

func Running(msg string) {
	p.running = []byte(fmt.Sprintf(evRunning.format, msg))
	output(p.running)
}

func Spinner() {
	output([]byte{'\r', '[', spinners[p.seq], ']'})
	p.seq = (p.seq + 1) % len(spinners)
}

func RunOK() {
	outputln(evRunOk.format)
	p.running = p.running[:0]
}

func RunFail() {
	output([]byte(colorRed))

	if len(p.running) > 0 {
		clearLine()
		output(p.running)
	}

	outputln(evRunFail.format)
	p.running = p.running[:0]

	output([]byte(colorReset))
}

func Download(a string) {
	output([]byte("[          ] " + a))
}

/* ------ static ---------
func Status(format string, a ...interface{}) {
	_progress.Status(format, a...)
}

func Warning(format string, a ...interface{}) {
	_progress.Warning(format, a...)
}

func Error(err error) {
	_progress.Error(err)
}

func ErrorMsg(format string, a ...interface{}) {
	_progress.ErrorMsg(format, a...)
}

func Download(a string) {
	_progress.Download(a)
}

/* ------- output ------- */

func clearLine() {
	if len(p.running) == 0 {
		return
	}
	buf := bytes.Repeat([]byte(" "), len(p.running)+2)
	buf[0] = byte('\r')
	buf[len(buf)-1] = byte('\r')
	_, _ = p.out.Write(buf)
}

func output(buf []byte) {
	_, _ = p.out.Write(buf)
}

func outputln(s string) {
	if p.out == nil {
		return
	}
	var buf []byte
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}
	output(buf)
}
