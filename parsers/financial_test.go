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

func Test_prepareFields(t *testing.T) {
	companies := make(map[string]company)
	companies["54321"] = company{1, "A"}

	type args struct {
		hash      uint32
		header    map[string]int
		fields    []string
		companies map[string]company
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"dt_refer not found",
			args{
				5454,
				map[string]int{"a": 0, "b": 1},
				[]string{"a", "b"},
				companies,
			},
			true,
		}, {
			"should work",
			args{
				393723,
				map[string]int{"x": 0, "y": 1, "DT_REFER": 2, "CNPJ_CIA": 3},
				[]string{"X", "Y", "2020-02-25", "54321"},
				companies,
			},
			false,
		}, {
			"cnpj not found",
			args{
				393724,
				map[string]int{"x": 0, "y": 2, "DT_REFER": 1},
				[]string{"X", "2020-02-25", "Y"},
				companies,
			},
			true,
		}, {
			"dt_refer not found",
			args{
				393724,
				map[string]int{"x": 0, "y": 2, "DT_REFER": 1},
				[]string{"X", "202", "Y"},
				companies,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := prepareFields(tt.args.hash, tt.args.header, tt.args.fields, tt.args.companies)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func BenchmarkPrepareFields(b *testing.B) {
	companies := make(map[string]company)
	companies["54321"] = company{1, "A"}

	h := make(map[string]int)
	h = map[string]int{"x": 0, "y": 1, "DT_REFER": 2, "CNPJ_CIA": 3}

	f := []string{"X", "Y", "2020-02-25", "54321"}

	// run the prepareFields function b.N times
	for n := 0; n < b.N; n++ {
		_, err := prepareFields(4433555, h, f, companies)
		if err != nil {
			b.Errorf("error: %v", err)
			return
		}
	}
}
