package rapina

import "io"

type FIIParser interface {
	CNPJ(code string) (string, error)
	StoreFIIDetails(stream []byte) error
	Dividends(code, monthYear string) (*[]Dividend, error)
	SaveDividend(stream map[string]string) (*Dividend, error)
}

type StockStore interface {
	Save(stream io.Reader, code string) (int, error)
	SaveB3Quotes(filename string) error
	Quote(code, date string) (float64, error)
}

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
}
