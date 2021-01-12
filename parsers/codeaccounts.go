package parsers

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

// acctCode returns the code based on the account code and
// account description; if the code is not found in the table
// returns the hash.
func acctCode(cdAccount, dsAccount string) uint32 {

	//  "BPA", "DRE", "DFC_MD", "DFC_MI", "DVA":

	accounts := []account{
		// BPA
		{"", "Caixa e Equivalentes de Caixa", Caixa},
		{"1.01.02", "Aplicações Financeiras", AplicFinanceiras},
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

	for _, acc := range accounts {
		if acc.cdAccount == "" || acc.cdAccount == cdAccount {
			if acc.dsAccount == "" || acc.dsAccount == dsAccount {
				return acc.code
			}
		}
	}

	return Hash(cdAccount + dsAccount)
}
