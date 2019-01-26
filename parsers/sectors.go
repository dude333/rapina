package parsers

import (
	"io/ioutil"

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
				if FuzzyMatch(company, segment.Companies, 3) {
					return segment.Companies, nil
				}
			}
		}
	}

	return
}
