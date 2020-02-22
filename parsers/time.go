package parsers

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
)

func yearRange(year string) (int64, int64, error) {
	y, err := strconv.Atoi(year)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "ano incorreto: %s.", year)
	}
	firstday := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
	lastday := firstday.AddDate(1, 0, 0).Add(time.Nanosecond * -1)

	return firstday.Unix(), lastday.Unix(), nil
}
