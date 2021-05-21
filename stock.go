package rapina

import "io"

// StockStorage is the interface that contains the methods needed to parse, save and
// retrieve stock data to/from a storage.
type StockStorage interface {
	Quote(code, date string) (float64, error)
	Code(companyName, stockType string) (string, error)
	Save(stream io.Reader, code string) (int, error)
}
