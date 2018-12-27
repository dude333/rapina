package parsers

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
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
// SectorsToYaml parses the excel file from B3 and saves its data into a YAML file
// excelFile =(input)=>[XXXX]=(output)=> yamlFile
//
func SectorsToYaml(excelFile, yamlFile string) (err error) {
	x, err := excelize.OpenFile(excelFile)
	if err != nil {
		return
	}

	var sector, subsector, segment string
	s := S{}

	for _, row := range x.GetRows("Plan3") {
		if len(row) < 5 || row[0] == "SETOR ECONÔMICO" || row[3] == "CÓDIGO" {
			continue
		}
		if row[0] == "(DR1) BDR Nível 1" {
			break
		}
		if len(row[0]) > 0 {
			// fmt.Printf("SETOR %s\n", row[0])
			sector = row[0]
			x := Sector{Name: sector}
			s.Sectors = append(s.Sectors, x)
		}
		if len(row[1]) > 0 {
			subsector = row[1]
			x := Subsector{Name: subsector}
			s.Sectors[len(s.Sectors)-1].Subsectors = append(s.Sectors[len(s.Sectors)-1].Subsectors, x)
		}
		if len(row[2]) > 0 && len(row[3]) == 0 {
			segment = row[2]
			x := Segment{Name: segment}
			l := len(s.Sectors[len(s.Sectors)-1].Subsectors) - 1
			s.Sectors[len(s.Sectors)-1].Subsectors[l].Segments = append(s.Sectors[len(s.Sectors)-1].Subsectors[l].Segments, x)
		}

		if len(row[2]) > 0 && len(row[3]) > 0 {
			seg := trim(row[4])
			if len(seg) > 0 {
				seg = " " + seg
			}
			str := fmt.Sprintf("%s [%s%s]", trim(row[2]), trim(row[3]), seg)
			l1 := len(s.Sectors[len(s.Sectors)-1].Subsectors) - 1
			l2 := len(s.Sectors[len(s.Sectors)-1].Subsectors[l1].Segments) - 1
			s.Sectors[len(s.Sectors)-1].Subsectors[l1].Segments[l2].Companies = append(s.Sectors[len(s.Sectors)-1].Subsectors[l1].Segments[l2].Companies, str)
		}
	}

	m, err := yaml.Marshal(&s)
	err = ioutil.WriteFile(yamlFile, m, 0644)
	if err != nil {
		log.Fatalf("WriteFile: %v", err)
	}

	return
}

func trim(s string) string {
	return strings.Trim(s, " ")
}
