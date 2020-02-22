package parsers

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// Bookkeeping account codes
// If you add new const values, run 'go generate'
// to update the generated code
const (
	UNDEF uint32 = iota
	SPACE

	// Balance Sheet
	Caixa
	AplicFinanceiras
	Estoque
	Equity
	DividaCirc
	DividaNCirc

	// Income Statement
	Vendas
	CustoVendas
	DespesasOp
	EBIT
	LucLiq

	// DFC
	FCO
	FCI
	FCF

	// Value Added Statement
	Deprec
	JurosCapProp
	Dividendos
)

// account code, description and bookkeeping code
type account struct {
	cdAccount string
	dsAccount string
	code      uint32
}

// CodeAccounts code all the accounts with the bookkeeping accounts constants
// An account is a line in the financial statement (e.g., equity, cash, EBITDA,
// etc.)
// How to check entries:
// cat b??_cia_aberta_con_201* | iconv -f WINDOWS-1252 -t UTF-8 |  awk -F ';' '{print $11 " "$12}' | sort | uniq -c | grep -Ei "Caixa"
// cat d??_cia_aberta_con_201* | iconv -f WINDOWS-1252 -t UTF-8 |  awk -F ';' '{print $12 " "$13}' | sort | uniq -c | grep -Ei " 3\.01 "
func CodeAccounts(db *sql.DB) (err error) {

	//  "BPA", "DRE", "DFC_MD", "DFC_MI", "DVA":

	accounts := []account{
		// BPA
		{"", "Caixa e Equivalentes de Caixa", Caixa},
		{"", "Aplicações Financeiras", AplicFinanceiras},
		{"1.01.04", "", Estoque}, // or "Títulos e Créditos a Receber" for security companies

		// BPP
		{"", "Patrimônio Líquido Consolidado", Equity},
		{"2.01.04", "Empréstimos e Financiamentos", DividaCirc},
		{"2.02.01", "Empréstimos e Financiamentos", DividaNCirc},

		// DRE
		{"3.01", "", Vendas},
		{"3.02", "", CustoVendas},
		{"3.04", "", DespesasOp},
		{"", "Resultado Antes do Resultado Financeiro e dos Tributos", EBIT},
		{"", "Lucro/Prejuízo Consolidado do Período", LucLiq},

		// DFC
		{"6.01", "", FCO},
		{"6.02", "", FCI},
		{"6.03", "", FCF},

		// DVA
		{"", "Depreciação, Amortização e Exaustão", Deprec},
		{"", "Juros sobre o Capital Próprio", JurosCapProp},
		{"", "Dividendos", Dividendos},
	}

	// <== DFP TABLE ==>

	// Tweak table indexes to optmize i/o
	dropIndexes(db)
	createIndexes(db, 1)

	// Set code for every account based on 'accounts' list ======
	for _, acc := range accounts {
		// WHERE clause
		where := ""
		if acc.cdAccount != "" {
			where = "CD_CONTA = '" + acc.cdAccount + "' "
			if acc.dsAccount != "" {
				where += "AND "
			}
		}
		if acc.dsAccount != "" {
			where += "DS_CONTA = '" + acc.dsAccount + "'"
		}

		// UPDATE table 'dfp'
		update := fmt.Sprintf("UPDATE dfp SET CODE = %d WHERE %s;", acc.code, where)
		_, err = db.Exec(update)
		if err != nil {
			return errors.Wrap(err, "erro ao atualizar dados")
		}
	} // next acc

	// Tweak table indexes to optmize i/o
	dropIndexes(db)
	createIndexes(db, 2)

	return
}
