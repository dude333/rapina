package parsers

import (
	"hash/fnv"
	"unicode"

	"golang.org/x/text/runes"
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
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ = transform.String(t, original)

	return
}
