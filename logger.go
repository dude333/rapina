package rapina

import "io"

// Logger interface contains the methods needed to poperly display log messages.
type Logger interface {
	Run(format string, v ...interface{})
	Ok()
	Nok()
	Printf(format string, v ...interface{})
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	SetOut(out io.Writer)
}
