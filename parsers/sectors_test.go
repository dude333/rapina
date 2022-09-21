package parsers

import (
	"os"
	"testing"
)

func TestFromSector(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "rapina-test")
	filename := tempDir + "/test_sectors.yml"

	createYaml(filename)
	s, _, _ := FromSector("GRENDENE S.A.", filename)
	expected := [...]string{"ALPARGATAS S.A.", "CAMBUCI S.A.", "GRENDENE S.A.", "VULCABRAS/AZALEIA S.A."}
	if len(s) != 4 {
		t.Errorf("\n- Expected:  %v\n- Got:       %v", expected, s)
	}

	var arr [4]string
	copy(arr[:], s)
	if arr != expected {
		t.Errorf("\n- Expected:  %v\n- Got:       %v", expected, s)
	}

	os.Remove(filename)
}

func createYaml(filename string) {
	yaml := []byte(
		`Setores:
- Setor: Bens Industriais
  Subsetores:
    - Subsetor: Comércio
      Segmentos:
        - Segmento: Material de Transporte
          Empresas:
            - MINASMAQUINAS S.A.
            - WLM PART. E COMÉRCIO DE MÁQUINAS E VEÍCULOS S.A.
- Setor: Consumo Cíclico
  Subsetores:
    - Subsetor: Tecidos. Vestuário e Calçados
      Segmentos:
        - Segmento: Acessórios
          Empresas:
          - MUNDIAL S.A. - PRODUTOS DE CONSUMO
          - TECHNOS S.A.
        - Segmento: Calçados
          Empresas:
            - ALPARGATAS S.A.
            - CAMBUCI S.A.
            - GRENDENE S.A.
            - VULCABRAS/AZALEIA S.A.`)

	_ = os.WriteFile(filename, yaml, 0644)
}
