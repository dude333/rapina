package parsers

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func tempFilename(t *testing.T) string {
	f, err := ioutil.TempFile("", "rapina-test-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func samples(filename string) error {
	bpa := []byte(`
CNPJ_CIA;DT_REFER;VERSAO;DENOM_CIA;CD_CVM;GRUPO_DFP;MOEDA;ESCALA_MOEDA;ORDEM_EXERC;DT_FIM_EXERC;CD_CONTA;DS_CONTA;VL_CONTA
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1;Ativo Total;1162167882.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.01;Caixa e Equivalentes de Caixa;68841638.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.02;Aplica��es Financeiras;110019404.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.02.01;Aplica��es Financeiras Avaliadas a Valor Justo;109376121.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.02.01.01;T�tulos para Negocia��o;18991047.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.02.01.02;T�tulos Dispon�veis para Venda;90385074.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.02.02;Aplica��es Financeiras Avaliadas ao Custo Amortizado;643283.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.02.02.01;T�tulos Mantidos at� o Vencimento;643283.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.03;Empr�stimos e Receb�veis;755821983.00
00.000.000/0001-91;2013-12-31;4;BANCO DO BRASIL S.A.;1023;DF Consolidado - Balan�o Patrimonial Ativo;REAL;MILHAR;�LTIMO;2013-12-31;1.04;Tributos Diferidos;21954460.00
	`)
	err := ioutil.WriteFile(filename, bpa, 0600)

	return err
}

func TestImportCsv(t *testing.T) {
	var db *sql.DB
	var err error
	fileBPA := tempFilename(t)
	defer os.Remove(fileBPA)
	fileDB := tempFilename(t)
	defer os.Remove(fileDB)

	if db, err = sql.Open("sqlite3", fileDB); err != nil {
		t.Errorf("Fail to open db: %v", err)
	}
	defer db.Close()

	if err = samples(fileBPA); err != nil {
		t.Errorf("Fail to create samples: %v", err)
	}

	if err = ImportCsv(db, "BPA", fileBPA); err != nil {
		t.Errorf("Fail to parse: %v", err)
	}

	for _, tp := range []string{"BPA", "MD5"} {
		if v, table := dbVersion(db, tp); v != currentDbVersion {
			t.Errorf("Expecting table %s on version %d, received %d", table, currentDbVersion, v)
		}
	}

	isNew, err := isNewFile(db, fileBPA)
	if isNew && err == nil {
		t.Errorf("Expecting processed file, got new file")
	}

}

func TestGetHash(t *testing.T) {
	table := []struct {
		s string
		h uint32
	}{
		{"test1", 2569220284},
		{"random data", 1626193638},
		{"excel", 1973829744},
		{"One More...12345!", 2258028052},
	}
	for _, x := range table {
		h := GetHash(x.s)
		if h != x.h {
			t.Errorf("Hash was incorrect, got: %d, want: %d.", h, x.h)
		}
	}
}

func TestRemoveDiacritics(t *testing.T) {
	list := []struct {
		str string
		exp string
	}{
		{"ITAÚ", "ITAU"},
		{"SÃO", "SAO"},
		{"São Paulo", "Sao Paulo"},
		{"ÁÉÍÓÚáéíóúÀàÃÕãõÇç", "AEIOUaeiouAaAOaoCc"},
	}

	for _, l := range list {
		if RemoveDiacritics(l.str) != l.exp {
			t.Errorf("Expecting %s, received %s", l.exp, RemoveDiacritics(l.str))
		}
	}
}

func TestPrepareFields(t *testing.T) {
	list := []struct {
		hash   uint32
		header map[string]int
		fields []string
	}{
		{
			5454,
			map[string]int{"a": 1, "b": 2},
			[]string{"x", "y", "z"},
		},
	}

	for _, l := range list {
		f, err := prepareFields(l.hash, l.header, l.fields)
		if err != nil {
			t.Error("prepareFields returned error ", err)
		}

		ok := f[0] == l.hash
		if !ok {
			t.Error("field 0 error")
		}
	}
}

func BenchmarkPrepareFields(b *testing.B) {
	p := struct {
		hash   uint32
		header map[string]int
		fields []string
	}{
		5454,
		map[string]int{"field_1": 1, "field_2": 2, "field_3": 3},
		[]string{"FIELD_1_HEADER", "FIELD_2_HEADER", "FIELD_3_HEADER"},
	}

	// run the prepareFields function b.N times
	for n := 0; n < b.N; n++ {
		prepareFields(p.hash, p.header, p.fields)
	}
}
