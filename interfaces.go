package rapina

import "io"

// FIIParser is the interface that contains the methods needed to parse
// fetched FII data from any source to a storage.
//
// CNPJ returns the CNPJ for a given company with a given stock 'code'.
//
// StoreFIIDetails saves the fetched FII details to the storage.
//
// Dividends returns the dividends data from a given company stock 'code' at
// a given 'date'.
//
// SaveDividend parses and stores a fetched stream with dividends data and
// returns a structured dividend object.
type FIIParser interface {
	Details(code string) (*FIIDetails, error)
	StoreFIIDetails(stream []byte) error
	Dividends(code, monthYear string) (*[]Dividend, error)
	SaveDividend(stream map[string]string) (*Dividend, error)
}

// StockParser is the interface that contains the methods needed to parse
// fetched stock quotes from any source to a storage.
//
// Save parses a stream to extract the stock quotes and saves them on the
// storage, returning the number o quotes saved.
//
// Quote returns the quote of a company stock 'code' at a given 'date'.
type StockParser interface {
	Save(stream io.Reader, code string) (int, error)
	SaveB3Quotes(filename string) error
	Quote(code, date string) (float64, error)
	Code(companyName, stockType string) (string, error)
}

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
}
