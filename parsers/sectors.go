package parsers

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
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
// xlsSectorsToYaml parses the excel file from B3 and saves its data into a YAML file
// excelFile =(input)=>[XXXX]=(output)=> yamlFile
//
func xlsSectorsToYaml(excelFile, yamlFile string) (err error) {
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
		errors.Wrapf(err, "WriteFile: %v", err)
	}

	return
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
				if FuzzyMatch(company, segment.Companies, 4) {
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
	src = RemoveDiacritics(src)
	for _, l := range list {
		l = RemoveDiacritics(l)
		r := fuzzy.RankMatchFold(src, l)
		if r >= 0 && r <= distance {
			return true
		}
		r = fuzzy.RankMatchFold(l, src)
		if r >= 0 && r <= distance {
			return true
		}
	}
	return false
}

//
// FuzzyFind returns the most approximate string inside 'list' that
// matches the 'src' string within a maximum 'distance'.
//
func FuzzyFind(src string, list []string, distance int) string {
	clean := make([]string, len(list))
	txt := RemoveDiacritics(src)
	for i, l := range list {
		clean[i] = RemoveDiacritics(l)
	}

	rank := fuzzy.RankFindFold(txt, clean)
	if len(rank) > 0 {
		sort.Sort(rank)
		if rank[0].Distance <= distance {
			i := rank[0].OriginalIndex
			return list[i]
		}
	}

	return ""
}

func trim(s string) string {
	return strings.Trim(s, " ")
}

//
// removeExtras removes the extra info (stock name and segment) from
// the list of companies and filters out companies not listed in
// special segments of the B3 listing: Bovespa Mais,
// Bovespa Mais Nível 2, Novo Mercado, Nível 2 and Nível 1
//
// pattern: COMPANY NAME [STCK SGM]
//
// (NM) Novo Mercado
// (N1) Nível 1 of Corporate Governance
// (N2) Nível 2 of Corporate Governance
// (MA) Bovespa Mais
// (M2) Bovespa Mais Nível 2
//
// Ignored:
// (MB) Traditional Org. OTC
// (DR1) Level 1 BDR
// (DR2) Level 2 BDR
// (DR3) Level 3 BDR
// (DRN) Unsponsored BDRs
//
func removeExtras(companies []string) (list []string) {
	var ok int
	valid := []string{"NM", "N1", "N2", "MA", "M2"}

	for _, co := range companies {

		// Check if is a valid segment (the SGM in [STCK SGM])
		p := strings.Index(co, "[")
		if p > 1 {
			v := strings.Split(co[p:], " ")
			if len(v) == 2 {
				sgmt := v[1][:len(v[1])-1]
				ok = 0
				for _, v := range valid {
					if sgmt == v {
						ok++
					}
				}
			}
		}

		if ok > 0 {
			// Leave the company name only
			for _, p := range []int{
				strings.Index(co, "S/A") - 1,
				strings.Index(co, "S.A") - 1,
				strings.Index(co, "[") - 1,
			} {
				if p > 0 {
					list = append(list, co[:p])
					break
				}
			}
		}
	}

	return
}
