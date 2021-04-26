package rapina

import "io"

type FIIStore interface {
	CNPJ(code string) (string, error)
	StoreFIIDetails(stream []byte) error
	StoreFIIDividends(stream map[string]string) error
}

type StockStore interface {
	CsvToDB(stream io.ReadCloser, code string) error
}
