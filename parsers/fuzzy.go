package parsers

import (
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

//
// FuzzyMatch measures the Levenshtein distance between
// the source and the list, returning true if the distance
// is less or equal the 'distance'.
// Diacritics are removed from 'src' and 'list'.
//
func FuzzyMatch(src string, list []string, distance int) bool {
	return FuzzyFind(src, list, distance) != ""
}

//
// FuzzyFind returns the most approximate string inside 'list' that
// matches the 'src' string within a maximum 'distance'.
//
func FuzzyFind(source string, targets []string, maxDistance int) (found string) {
	for _, target := range targets {
		src := fix(source)
		trg := fix(target)
		if strings.HasPrefix(src, trg) || strings.HasPrefix(trg, src) {
			return target
		}
		distance := fuzzy.LevenshteinDistance(src, trg)
		if distance <= maxDistance {
			maxDistance = distance
			found = target
		}
	}

	if found == "" {
		for _, target := range targets {
			src := strings.Split(fix(source), " ")
			trg := strings.Split(fix(target), " ")

			if len(src) > 2 && len(trg) > 2 {
				if src[0] == trg[0] && src[1] == trg[1] {
					return target
				}
			}
		}
	}

	return
}

func fix(txt string) string {
	txt = strings.ToUpper(txt)
	txt = strings.Replace(txt, "BCO ", "BANCO ", 1)
	return RemoveDiacritics(txt)
}
