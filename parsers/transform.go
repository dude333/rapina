package parsers

import (
	"hash/fnv"
	"strconv"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// fnvHash is a global var set to speed up Hash
var fnvHash = fnv.New32a()

//
// Hash returns the FNV-1 non-cryptographic hash
//
func Hash(s string) uint32 {
	fnvHash.Write([]byte(s))
	defer fnvHash.Reset()

	return fnvHash.Sum32()
}

//
// RemoveDiacritics transforms, for example, "žůžo" into "zuzo"
//
func RemoveDiacritics(original string) (result string) {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}

	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ = transform.String(t, original)

	return
}

//
// yearRange transforms a year string (e.g., "2020") in two int64
// containing the 1st and last day of that year in Unix timestamp.
//
func yearRange(year string) (int64, int64, error) {
	y, err := strconv.Atoi(year)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "ano incorreto: %s.", year)
	}
	firstday := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
	lastday := firstday.AddDate(1, 0, 0).Add(time.Nanosecond * -1)

	return firstday.Unix(), lastday.Unix(), nil
}
