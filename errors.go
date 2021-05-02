package rapina

import "errors"

var (
	ErrRecordExists   = errors.New("insert ignored, register already exists")
	ErrFileNotUpdated = errors.New("file not updated")
	ErrInvalidAPIKey  = errors.New("apiKey inválida, configure uma chave em" +
		" https://www.alphavantage.co/support/#api-key e adicione no arquivo" +
		" config.yml")
	ErrInvalidDate = errors.New("invalid date format")
)
