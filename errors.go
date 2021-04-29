package rapina

import "errors"

var (
	ErrInvalidDate    = errors.New("invalid date format")
	ErrFileNotUpdated = errors.New("file not updated")
)
