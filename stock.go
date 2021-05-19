package rapina

import "io"

// StockSaver is the interface that contains the methods needed to parse and save
// a stream with stock information to a storage.
type StockSaver interface {
	Save(stream io.Reader, code string) (int, error)
}

// StockLoader is the interface that contains the methods needed to retrieve
// stock information from a storage.
type StockLoader interface {
	Quote(code, date string) (float64, error)
	Code(companyName, stockType string) (string, error)
}

// StockParser is the interface that contains the methods needed to parse, save and
// retrieve stock data to/from a storage.
type StockParser interface {
	StockSaver
	StockLoader
}
