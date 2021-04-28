package reports

import (
	"fmt"
	"io"
)

type Logger struct {
	out io.Writer // destination for output
	buf []byte    // for accumulating text to write
}

// New creates a new Logger
func New(out io.Writer) *Logger {
	return &Logger{out: out}
}

// Run prints a message before running a process.
func (l *Logger) Run(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	l.output("[ ] " + s)
}

// Ok prints a checkmark after a successful Run()
func (l *Logger) Ok() {
	l.outputln("\r[✓]")
}

// Nok prints a x mark after a unsuccessful Run()
func (l *Logger) Nok() {
	l.outputln("\r[✗]")
}

// Printf prints the plain text.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.output(fmt.Sprintf(format, v...))
}

// Trace for very low level logs.
func (l *Logger) Trace(format string, v ...interface{}) {
	l.outputln("[TRACE] " + fmt.Sprintf(format, v...))
}

// Debug for debugging information.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.outputln("[DEBUG] " + fmt.Sprintf(format, v...))
}

// Info for something noteworthy.
func (l *Logger) Info(format string, v ...interface{}) {
	l.outputln("[INFO] " + fmt.Sprintf(format, v...))
}

// Warn for a warning message.
func (l *Logger) Warn(format string, v ...interface{}) {
	l.outputln("[WARN] " + fmt.Sprintf(format, v...))
}

// Error message.
func (l *Logger) Error(format string, v ...interface{}) {
	l.outputln("[ERROR] " + fmt.Sprintf(format, v...))
}

func (l *Logger) output(s string) {
	l.buf = l.buf[:0]
	l.buf = append(l.buf, s...)
	_, _ = l.out.Write(l.buf)
}

func (l *Logger) outputln(s string) {
	l.buf = l.buf[:0]
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, _ = l.out.Write(l.buf)
}
