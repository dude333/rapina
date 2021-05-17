package rapina

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// IsDate checks if date is in format YYYY-MM-DD.
func IsDate(date string) bool {
	if len(date) != len("2021-04-26") || strings.Count(date, "-") != 2 {
		return false
	}

	y, errY := strconv.Atoi(date[0:4])
	m, errM := strconv.Atoi(date[5:7])
	d, errD := strconv.Atoi(date[8:10])
	if errY != nil || errM != nil || errD != nil {
		return false
	}

	// Ok, we'll still be using this in 2200 :)
	if y < 1970 || y > 2200 {
		return false
	}
	if m < 1 || m > 12 {
		return false
	}
	nDays := [13]int{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if d < 1 || d > nDays[m] {
		return false
	}
	return true
}

// IsURL returns true if 'str' is a valid URL.
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// JoinURL joins strings as URL paths
func JoinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
}

var _timeNow = time.Now

// MonthsFromToday returns a list of months including the current.
// Date formatted as YYYY-MM-DD.
func MonthsFromToday(n int) []string {
	if n < 1 {
		n = 1
	}
	if n > 100 {
		n = 100
	}

	now := _timeNow()
	now = time.Date(now.Year(), now.Month(), 15, 12, 0, 0, 0, time.UTC)

	var monthYears []string
	for ; n > 0; n-- {
		monthYears = append(monthYears, now.Format("2006-01"))
		now = now.AddDate(0, -1, 0)
	}

	return monthYears
}

// LastBusinessDayOfYear returns the last business day of the 'year' (the business
// day before Dec 30).
// Returns date as YYYY-MM-DD.
func LastBusinessDayOfYear(year int) string {
	date := time.Date(year, time.December, 29, 12, 0, 0, 0, time.UTC)

	if date.Weekday() == time.Saturday {
		date = date.AddDate(0, 0, -1)
	}
	if date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, -2)
	}

	return date.Format("2006-01-02")
}

// LastBusinessDay returns the most recent business day 'n' days before today.
// Returns date as YYYY-MM-DD.
func LastBusinessDay(n int) string {
	date := time.Now()
	if n > 0 {
		date = date.AddDate(0, 0, -n)
	}

	if date.Weekday() == time.Saturday {
		date = date.AddDate(0, 0, -1)
	}
	if date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, -2)
	}

	return date.Format("2006-01-02")
}
