package reports

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/dude333/rapina/parsers"
	"github.com/pkg/errors"
)

type accItems struct {
	code    uint32
	cdConta string
	dsConta string
}

//
// accountsItems returns all accounts codes and descriptions, e.g.:
// [1 Ativo Total, 1.01 Ativo Circulante, ...]
//
func (r report) accountsItems(cid int) (items []accItems, err error) {
	selectItems := fmt.Sprintf(`
	SELECT DISTINCT
		CODE, CD_CONTA, DS_CONTA
	FROM
		dfp
	WHERE
		ID_CIA = "%d"
		AND ORDEM_EXERC LIKE "_LTIMO"

	ORDER BY
		CD_CONTA, DS_CONTA
	;`, cid)

	rows, err := r.db.Query(selectItems)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var item accItems
	for rows.Next() {
		err = rows.Scan(&item.code, &item.cdConta, &item.dsConta)
		if err != nil {
			return
		}
		items = append(items, item)
	}

	return
}

type account struct {
	code     uint32
	year     string
	denomCia string
	escala   string
	vlConta  float32
}

//
// accountsValues stores the values for each account into a map using a hash
// of the account code and description as its key
//
func (r report) accountsValues(cid, year int, penult bool) (values map[uint32]float32, err error) {

	period := "_LTIMO"
	if penult {
		period = "PEN_LTIMO"
		year++
	}

	selectReport := fmt.Sprintf(`
	SELECT
		CODE,
		NAME,
		ORDEM_EXERC,
		YEAR,
		VL_CONTA
	FROM
		dfp
	JOIN
		companies ON dfp.ID_CIA=companies.ID
	WHERE
		ID_CIA = "%d"
		AND ORDEM_EXERC LIKE "%s"
		AND YEAR = "%d"
	;`, cid, period, year)

	values = make(map[uint32]float32)
	st := account{}

	rows, err := r.db.Query(selectReport)
	if err != nil {
		return
	}
	defer rows.Close()

	var denomCia, orderExec string
	var dtRefer int
	for rows.Next() {
		rows.Scan(
			&st.code,
			&denomCia,
			&orderExec,
			&dtRefer,
			&st.vlConta,
		)

		values[st.code] = st.vlConta
	}

	return
}

//
// accountsAverage stores the average of all companies of the same sector
// for each account into a map using a hash of the account code and
// description as its key
//
func (r report) accountsAverage(company string, year int, penult bool) (values map[uint32]float32, err error) {

	companies, _, err := r.fromSector(company)
	if len(companies) <= 1 || err != nil {
		err = errors.Wrap(err, "erro ao ler arquivo de setores "+r.yamlFile)
		return
	}

	if len(companies) == 0 {
		err = errors.Errorf("erro ao procurar empresas")
		return
	}

	// PERIOD (last or before last year)
	period := "_LTIMO"
	if penult {
		period = "PEN_LTIMO"
		year++
	}

	cids := make([]string, len(companies))
	for i, co := range companies {
		if id, err := cid(r.db, co); err == nil {
			cids[i] = strconv.Itoa(id)
		}
	}

	selectReport := fmt.Sprintf(`
	SELECT
		CODE,
		ORDEM_EXERC,
		AVG(VL_CONTA) AS MD_CONTA
	FROM
		dfp
	WHERE
		ID_CIA IN ("%s")
		AND ORDEM_EXERC LIKE "%s"
		AND YEAR = "%d"
	GROUP BY
		CODE, ORDEM_EXERC;
	`, strings.Join(cids, "\", \""), period, year)

	values = make(map[uint32]float32)
	st := account{}

	rows, err := r.db.Query(selectReport)
	if err != nil {
		return
	}
	defer rows.Close()

	var orderExec string
	for rows.Next() {
		rows.Scan(
			&st.code,
			&orderExec,
			&st.vlConta,
		)

		values[st.code] = st.vlConta
	}

	return
}

func (r report) fromSector(company string) (companies []string, sectorName string, err error) {
	// Companies from the same sector
	secCo, secName, err := parsers.FromSector(company, r.yamlFile)
	if len(secCo) <= 1 || err != nil {
		err = errors.Wrap(err, "erro ao ler arquivo dos setores "+r.yamlFile)
		return
	}

	// All companies stored on db
	list, err := ListCompanies(r.db)
	if err != nil {
		err = errors.Wrap(err, "erro ao listar empresas")
		return
	}

	// Translate company names to match the name stored on db
	for _, s := range secCo {
		z := parsers.FuzzyFind(s, list, 3)
		if len(z) > 0 {
			companies = append(companies, z)
		}
	}

	return removeDuplicates(companies), secName, nil
}

// CompanyInfo contains the company name and CNPJ
type CompanyInfo struct {
	id   int
	name string
}

//
// companies returns available companies in the DB
//
func companies(db *sql.DB) ([]CompanyInfo, error) {

	selectCompanies := `
		SELECT ID, NAME
		FROM companies
		ORDER BY NAME;`

	rows, err := db.Query(selectCompanies)
	if err != nil {
		err = errors.Wrap(err, "falha ao ler banco de dados")
		return nil, err
	}
	defer rows.Close()

	var info CompanyInfo
	var list []CompanyInfo
	for rows.Next() {
		rows.Scan(&info.id, &info.name)
		list = append(list, info)
	}

	return list, nil
}

//
// cid returns the company ID
//
func cid(db *sql.DB, company string) (int, error) {
	selectID := fmt.Sprintf(`SELECT DISTINCT ID FROM companies WHERE NAME LIKE "%s%%"`, company)
	var cid int
	err := db.QueryRow(selectID).Scan(&cid)
	if err != nil {
		return 0, err
	}
	return cid, nil
}

//
// isCompany returns true if company exists on DB
//
func (r report) isCompany(company string) bool {
	selectCompany := fmt.Sprintf(`
	SELECT
		NAME
	FROM
		companies
	WHERE
		NAME LIKE "%s%%";`, company)

	var c string
	err := r.db.QueryRow(selectCompany).Scan(&c)
	if err != nil {
		return false
	}

	return true
}

//
// timeRange returns the begin=min(year) and end=max(year)
//
func timeRange(db *sql.DB) (begin, end int, err error) {

	selectYears := `
	SELECT
		MIN(CAST(YEAR AS INTEGER)),
		MAX(CAST(YEAR AS INTEGER))
	FROM dfp;`

	rows, err := db.Query(selectYears)
	if err != nil {
		err = errors.Wrap(err, "falha ao ler banco de dados")
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&begin, &end)
	}

	// Check year
	if begin < 1900 || begin > 2100 || end < 1900 || end > 2100 {
		err = errors.Wrap(err, "ano inválido")
		return
	}
	if begin > end {
		aux := end
		end = begin
		begin = aux
	}

	return
}

func removeDuplicates(elements []string) []string { // change string to int here if required
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{} // change string to int here if required
	result := []string{}             // change string to int here if required

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

type profit struct {
	year   int
	profit float32
}

func companyProfits(db *sql.DB, companyID int) (profits []profit, err error) {

	selectProfits := fmt.Sprintf(`
	SELECT
		ORDEM_EXERC,
		YEAR,
		VL_CONTA
	FROM
		dfp
	WHERE
		ID_CIA = "%d"
		AND CODE = "%d"
		AND (ORDEM_EXERC LIKE "_LTIMO"
			OR (
				ORDEM_EXERC LIKE "PEN_LTIMO"
				AND YEAR = (SELECT MIN(YEAR) FROM dfp WHERE ID_CIA = %d)
			)
		)
	ORDER BY
		YEAR;`, companyID, parsers.LucLiq, companyID)

	rows, err := db.Query(selectProfits)
	if err != nil {
		err = errors.Wrap(err, "falha ao ler banco de dados")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var period string
		var year int
		var val float32
		rows.Scan(&period, &year, &val)
		if period == "PENÚLTIMO" {
			year--
		}
		profits = append(profits, profit{year, val})
	}

	return
}
