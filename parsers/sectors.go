package parsers

import (
	"io/ioutil"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// S contains the sectors
type S struct {
	Sectors []Sector `yaml:"Setores"`
}

// Sector is divided into subsectors
type Sector struct {
	Name       string      `yaml:"Setor"`
	Subsectors []Subsector `yaml:"Subsetores"`
}

// Subsector is divided into segments
type Subsector struct {
	Name     string    `yaml:"Subsetor"`
	Segments []Segment `yaml:"Segmentos"`
}

// Segment contains companies from the same sector/subsector/segment
type Segment struct {
	Name      string   `yaml:"Segmento"`
	Companies []string `yaml:"Empresas"`
}

//
// FromSector returns all companies from the same sector as the 'company'
//
func FromSector(company, yamlFile string) (companies []string, err error) {

	y, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		err = errors.Wrapf(err, "ReadFile: %v", err)
		return
	}

	s := S{}
	yaml.Unmarshal(y, &s)

	for _, sector := range s.Sectors {
		for _, subsector := range sector.Subsectors {
			for _, segment := range subsector.Segments {
				if FuzzyMatch(company, segment.Companies, 6) {
					return segment.Companies, nil
				}
			}
		}
	}

	return
}

//
// FuzzyMatch measures the Levenshtein distance between
// the source and the list, returning true if the distance
// is less or equal the 'distance'.
// Diacritics are removed from 'src' and 'list'.
//
func FuzzyMatch(src string, list []string, distance int) bool {
	if FuzzyFind(src, list, distance) != "" {
		return true
	}
	return false
}

//
// FuzzyFind returns the most approximate string inside 'list' that
// matches the 'src' string within a maximum 'distance'.
//
func FuzzyFind(source string, targets []string, maxDistance int) (found string) {
	for _, target := range targets {
		distance := fuzzy.LevenshteinDistance(fix(source), target)
		if distance <= maxDistance {
			maxDistance = distance
			found = target
		}
	}
	return
}

func fix(txt string) string {
	switch txt {
	case "BCO BRASIL S.A.":
		return "BANCO DO BRASIL S.A."
	}

	return txt
}
