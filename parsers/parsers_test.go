package parsers

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const (
	fileDB  = "/tmp/rapina_test.db"
	fileBPA = "/tmp/bpa.csv"
)

func openDB() (*sql.DB, error) {
	os.Remove(fileDB)
	return sql.Open("sqlite3", fileDB)
}

func samples() error {
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
	err := ioutil.WriteFile(fileBPA, bpa, 0600)

	return err
}

func TestExec(t *testing.T) {
	var db *sql.DB
	var err error

	if db, err = openDB(); err != nil {
		t.Errorf("Fail to open db: %v", err)
	}
	if err = samples(); err != nil {
		t.Errorf("Fail to create samples: %v", err)
	}
	if err = Exec(db, "BPA", fileBPA); err != nil {
		t.Errorf("Fail to parse: %v", err)
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
