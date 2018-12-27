package parsers

import "testing"

func TestSectorsToYaml(t *testing.T) {
	err := SectorsToYaml("../cli/.data/B3sectors.xlsx", "../cli/.data/sectors.yaml")
	if err != nil {
		t.Errorf("Error: %v\n", err)
	}
}
