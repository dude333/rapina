package parsers

import (
	"strings"
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
	ContasARecebCirc
	ContasARecebNCirc
	AtivoCirc
	AtivoNCirc
	AtivoTotal
	PassivoCirc
	PassivoNCirc
	PassivoTotal
	DividaCirc
	DividaNCirc
	DividendosJCP
	DividendosMin

	// Income Statement
	Vendas
	CustoVendas
	DespesasOp
	EBIT
	ResulFinanc
	ResulOpDescont
	LucLiq

	// DFC
	FCO
	FCI
	FCF

	// Value Added Statement
	Deprec
	JurosCapProp
	Dividendos

	// Values stored on table 'fre'
	Shares
	FreeFloat

	// Financial ratios
	EstoqueMedio
	EquityAvg

	// Financial scale (unit, thousand)
	Escala

	// Stock quote from last day of year
	Quote
)

// account code, description and bookkeeping code
type account struct {
	cdAccount string
	dsAccount string
	code      uint32
}

var _accountsTable = []account{
	// BPA
	{"1", "Ativo Total", AtivoTotal},
	{"1.01", "Ativo Circulante", AtivoCirc},
	{"1.02", "Ativo Não Circulante", AtivoNCirc},
	{"1.01.01", "Caixa e Equivalentes de Caixa", Caixa},
	{"1.01.02", "Aplicações Financeiras", AplicFinanceiras},
	{"1.01.04", "Estoques", Estoque}, // or "Títulos e Créditos a Receber" for security companies
	{"1.01.03", "Contas a Receber", ContasARecebCirc},
	{"1.02.01.03", "Contas a Receber", ContasARecebNCirc},
	{"1.02.01.04", "Contas a Receber", ContasARecebNCirc},

	// BPP
	{"2", "Passivo Total", PassivoTotal},
	{"2.01", "Passivo Circulante", PassivoCirc},
	{"2.02", "Passivo Não Circulante", PassivoNCirc},
	{"2.*", "Patrimônio Líquido Consolidado", Equity},
	{"2.01.04", "Empréstimos e Financiamentos", DividaCirc},
	{"2.02.01", "Empréstimos e Financiamentos", DividaNCirc},
	{"2.01.05.02.01", "Dividendos e JCP a Pagar", DividendosJCP},
	{"2.01.05.02.02", "Dividendo Mínimo Obrigatório a Pagar", DividendosMin},

	// DRE
	{"3.01", "", Vendas},
	{"3.02", "", CustoVendas},
	{"3.04", "", DespesasOp},
	{"3.*", "Resultado Antes do Resultado Financeiro e dos Tributos", EBIT},
	{"3.06", "Resultado Financeiro", ResulFinanc},
	{"3.07", "Resultado Financeiro", ResulFinanc},
	{"3.08", "Resultado Financeiro", ResulFinanc},
	{"3.10", "Resultado Líquido de Operações Descontinuadas", ResulOpDescont},
	{"3.11", "Resultado Líquido de Operações Descontinuadas", ResulOpDescont},
	{"3.12", "Resultado Líquido de Operações Descontinuadas", ResulOpDescont},
	{"3.*", "Lucro/Prejuízo Consolidado do Período", LucLiq},
	{"3.*", "Lucro/Prejuízo do Período", LucLiq},

	// DFC
	{"6.01", "", FCO},
	{"6.02", "", FCI},
	{"6.03", "", FCF},

	// DVA
	{"7.*", "Depreciação, Amortização e Exaustão", Deprec},
	{"7.*", "Juros sobre o Capital Próprio", JurosCapProp},
	{"7.*", "Dividendos", Dividendos},
}

// acctCode returns the code based on the account code and
// account description; if the code is not found in the table
// returns the hash.
func acctCode(cdAccount, dsAccount string) uint32 {
	dsAccount = strings.ToLower(dsAccount)

	for _, acc := range _accountsTable {
		descr := strings.ToLower(acc.dsAccount)
		l := len(acc.cdAccount)
		code := ""
		if l > 1 && acc.cdAccount[l-1] == '*' {
			code = acc.cdAccount[:l-1] // remove the '*'
		}

		if code != "" && strings.HasPrefix(cdAccount, code) {
			if descr == "" || descr == dsAccount {
				return acc.code
			}
		} else if acc.cdAccount == "" || acc.cdAccount == cdAccount {
			if descr == "" || descr == dsAccount {
				return acc.code
			}
		}
	}

	return Hash(cdAccount + dsAccount)
}
