package reports

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	p "github.com/dude333/rapina/parsers"
)

// metric parameters
type metric struct {
	descr  string
	val    float32
	format int // mapped by constants NUMBER, INDEX, PERCENT
}

//
// Report company from DB to Excel
//
func Report(db *sql.DB, company string, path string) (err error) {
	f, err := filename(path, company)
	if err != nil {
		return err
	}

	e := newExcel()
	sheet, _ := e.newSheet(company)

	var lastStatementsRow, lastMetricsRow int

	// Company name
	sheet.mergeCell("A1", "B1")
	sheet.printRows("A1", &[]string{company}, LEFT, true)

	// ACCOUNT NUMBERING AND DESCRIPTION (COLS A AND B) ===============\/

	// Print accounts codes and descriptions in columns A and B
	// starting on row 2. Adjust space related to the group, e.g.:
	// 3.02 ABC <== print in bold if base item and stores the row position in baseItems[]
	//   3.02.01 ABC
	accounts, _ := accountsItems(db, company)
	row := 2
	baseItems := make([]bool, len(accounts)+row)
	for _, it := range accounts {
		var sp string
		sp, baseItems[row] = ident(it.cdConta)
		cell := "A" + strconv.Itoa(row)
		sheet.printRows(cell, &[]string{sp + it.cdConta, sp + it.dsConta}, LEFT, baseItems[row])
		row++
	}
	lastStatementsRow = row - 1
	row += 2
	// Metrics descriptions
	for _, metric := range metricsList(nil) {
		if metric.descr != "" {
			cell := "B" + strconv.Itoa(row)
			sheet.printRows(cell, &[]string{metric.descr}, RIGHT, false)
		}
		row++
	}
	lastMetricsRow = row - 1

	begin, end, err := timeRange(db)
	if err != nil {
		return
	}

	// 	VALUES (COLS C, D, E...) / PER YEAR ===========================\/

	// Print accounts values ONE YEAR PER COLUMN, starting from C, row 2
	var values map[uint32]float32
	cols := "CDEFGHIJKLMONPQRSTUVWXYZ"
	start := begin - 1
	for y := start; y <= end; y++ {
		if y-start >= len(cols) {
			break
		}
		col := string(cols[y-start])
		cell := col + "1"
		sheet.printTitle(cell, "["+strconv.Itoa(y)+"]") // Print year as title in row 1

		values, _ = accountsValues(db, company, y, y == start)
		row = 2
		for _, acct := range accounts {
			cell := col + strconv.Itoa(row)
			sheet.printValue(cell, values[acct.code], NUMBER, baseItems[row])
			row++
		}

		// FINANCIAL METRICS (COLS C, D, E...) / YEAR ===================\/

		// Print financial metrics
		row++
		cell = fmt.Sprintf("%s%d", col, row)
		sheet.printTitle(cell, "["+strconv.Itoa(y)+"]") // Print year as title
		row++

		// Print report in the sequence defined in financialMetricsList()
		for _, metric := range metricsList(values) {
			if metric.format != EMPTY {
				cell := col + strconv.Itoa(row)
				sheet.printValue(cell, metric.val, metric.format, false)
			}
			row++
		}

	} // next year

	_ = lastMetricsRow

	//
	// VERTICAL ANALYSIS
	//
	// CODES | DESCRIPTION | Y1 | Y2 | Yn | sp | v1 | v2 | v3
	//
	wide := (end - begin) + 1
	year := begin
	top := 2
	for col := 2; col <= 2+wide; col++ {
		formulaCol := col + wide + 2
		sheet.printTitle(axis(formulaCol, 1), "'"+strconv.Itoa(year)) // Print year
		year++
		var ref string
		for row := top; row <= lastStatementsRow; row++ {
			idx := row - top
			if idx < 0 || idx >= len(accounts) {
				break
			}
			if len(accounts[idx].cdConta) == 0 {
				break
			}
			n, _ := strconv.Atoi(accounts[idx].cdConta[:1])
			if n > 3 {
				break
			}
			switch accounts[idx].cdConta {
			case "1", "2", "3.01":
				ref = axis(col, row)
			}
			val := axis(col, row)
			formula := fmt.Sprintf(`=IfError(%s/%s, "-")`, val, ref)

			sheet.printFormula(axis(formulaCol, row), formula, PERCENT, baseItems[row])
		}
	}

	sheet.autoWidth()

	sheet2, _ := e.newSheet("Setor")
	row, col := 2, 0
	reportSector(db, sheet2, &row, &col, true, company)
	row = 2
	col++
	reportSector(db, sheet2, &row, &col, false, "ALPARGATAS")
	row = 2
	col++
	reportSector(db, sheet2, &row, &col, false, "VULCABRAS")

	err = e.saveAndCloseExcel(f)
	if err == nil {
		fmt.Printf("[✓] Dados salvos em %s\n", f)
	}

	return
}

//
// reportSector reports all companies from the same segment into the
// 'Setor' sheet.
//
func reportSector(db *sql.DB, sheet *Sheet, row, col *int, printDescr bool, company string) (err error) {
	begin, end, err := timeRange(db)
	if err != nil {
		return
	}

	// Company name
	if printDescr {
		*col++
	}
	sheet.mergeCell(axis(*col, *row), axis(*col+end-begin+1, *row))
	f := setFormat(DEFAULT, CENTER, true, formatBorder{})
	f.Font = &formatFont{Size: 16}
	sheet.printCell(*row, *col, company, f)
	if printDescr {
		*col--
	}
	*row++

	// Save starting row
	r := *row

	// Set width for the description col
	if printDescr {
		sheet.setColWidth(*col, 16.4)
		*col++
	}

	// Print values ONE YEAR PER COLUMN
	var values map[uint32]float32
	start := begin - 1
	for y := start; y <= end; y++ {
		values, _ = accountsValues(db, company, y, y == start)

		*row = r

		// Print year
		sheet.printTitle(axis(*col, *row), "["+strconv.Itoa(y)+"]")
		*row++

		// Print financial metrics
		for _, metric := range metricsList(values) {
			if metric.format != EMPTY {
				if printDescr {
					f := setFormat(DEFAULT, RIGHT, false, formatBorder{})
					sheet.printCell(*row, *col-1, metric.descr, f)
				}
				f := setFormat(metric.format, DEFAULT, false, formatBorder{})
				sheet.printCell(*row, *col, metric.val, f)
			}
			*row++
		}

		printDescr = false
		*col++
	} // next year

	return
}

//
// metricsList returns the sequence to be printed after the financial statements
//
func metricsList(v map[uint32]float32) (metrics []metric) {
	dividaBruta := v[p.DividaCirc] + v[p.DividaNCirc]
	caixa := v[p.Caixa] + v[p.AplicFinanceiras]
	dividaLiquida := dividaBruta - caixa
	EBITDA := v[p.EBIT] - v[p.Deprec]
	proventos := v[p.Dividendos] + v[p.JurosCapProp]

	var roe float32
	if v[p.LucLiq] > 0 && v[p.Equity] > 0 {
		roe = zeroIfNeg(safeDiv(v[p.LucLiq], v[p.Equity]))
	}

	return []metric{
		{"Patrimônio Líquido", v[p.Equity], NUMBER},
		{"", 0, EMPTY},

		{"Receita Líquida", v[p.Vendas], NUMBER},
		{"EBITDA", EBITDA, NUMBER},
		{"D&A", v[p.Deprec], NUMBER},
		{"EBIT", v[p.EBIT], NUMBER},
		{"Lucro Líquido", v[p.LucLiq], NUMBER},
		{"", 0, EMPTY},

		{"Marg. EBITDA", zeroIfNeg(safeDiv(v[p.EBIT], v[p.Vendas])), PERCENT},
		{"Marg. EBIT", zeroIfNeg(safeDiv(v[p.EBIT], v[p.Vendas])), PERCENT},
		{"Marg. Líq.", zeroIfNeg(safeDiv(v[p.LucLiq], v[p.Vendas])), PERCENT},
		{"ROE", roe, PERCENT},
		{"", 0, EMPTY},

		{"Caixa", caixa, NUMBER},
		{"Dívida Bruta", dividaBruta, NUMBER},
		{"Dívida Líq.", dividaLiquida, NUMBER},
		{"Dív. Bru./PL", zeroIfNeg(safeDiv(dividaBruta, v[p.Equity])), PERCENT},
		{"Dív.Líq./EBITDA", zeroIfNeg(safeDiv(dividaLiquida, EBITDA)), INDEX},
		{"", 0, EMPTY},

		{"FCO", v[p.FCO], NUMBER},
		{"FCI", v[p.FCI], NUMBER},
		{"FCF", v[p.FCF], NUMBER},
		{"Fluxo de Caixa Total", v[p.FCO] + v[p.FCI] + v[p.FCF], NUMBER},
		{"", 0, EMPTY},

		{"Proventos", proventos, NUMBER},
		{"Payout", zeroIfNeg(safeDiv(proventos, v[p.LucLiq])), PERCENT},
	}
}

func zeroIfNeg(n float32) float32 {
	if n < 0 {
		return 0
	}
	return n
}

func safeDiv(n, d float32) float32 {
	if d == 0 {
		return 0
	}
	return n / d
}

//
// ident returns the number of spaces according to the code level, e.g.:
// "1.1 ABC"   => "  " (2 spaces)
// "1.1.1 ABC" => "    " (4 spaces)
// For items equal or above 3, only returns spaces after 2nd level:
// "3.01 ABC"    => ""
// "3.01.01 ABC" => "  "
//
func ident(str string) (spaces string, baseItem bool) {
	num := strings.SplitN(str, ".", 2)[0]
	c := strings.Count(str, ".")
	if num != "1" && num != "2" && c > 0 {
		c--
	}
	if c > 0 {
		spaces = strings.Repeat("  ", c)
	}

	if num == "1" || num == "2" {
		baseItem = c <= 1
	} else {
		baseItem = c == 0
	}

	return
}
