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
