package rapina

import "io"

type FIIParser interface {
	CNPJ(code string) (string, error)
	StoreFIIDetails(stream []byte) error
	Dividends(code, monthYear string) (*[]Dividend, error)
	SaveDividend(stream map[string]string) (*Dividend, error)
}

type StockStore interface {
	CsvToDB(stream io.ReadCloser, code string) error
	Quote(code, date string) (float64, error)
}

type Logger interface {
	Printf(format string, v ...interface{})
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}
