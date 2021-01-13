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
		dfp a
	WHERE
		ID_CIA = "%d"
		AND VERSAO = (SELECT MAX(VERSAO) FROM dfp WHERE ID_CIA = a.ID_CIA AND YEAR = a.YEAR)
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
func (r report) accountsValues(cid, year int) (map[uint32]float32, error) {
	lastYear, isITR, err := r.lastYear()

	if err == nil && year == lastYear && isITR {
		return ttm(r.db, cid)
	}

	return dfp(r.db, cid, year)
}

//
// lastYear considers the current year as the latest year recorded on the DB.
// returns the last date, if it's to use the ITR table (instead of the DFP)
// and the error, if any.
//
func (r report) lastYear() (int, bool, error) {
	if r.cid == 0 {
		return 0, false, fmt.Errorf("customer ID not set")
	}

	selectDfpLastYear := `SELECT MAX(CAST(YEAR AS INTEGER)) YEAR FROM dfp WHERE ID_CIA = ?;`
	dfp := 0
	err := r.db.QueryRow(selectDfpLastYear, r.cid).Scan(&dfp)
	if err != nil {
		return 0, false, err
	}

	selectItrLastYear := `SELECT MAX(CAST(YEAR AS INTEGER)) YEAR FROM itr WHERE ID_CIA = ?;`
	itr := 0
	err = r.db.QueryRow(selectItrLastYear, r.cid).Scan(&itr)
	if err != nil {
		return 0, false, err
	}

	if itr > dfp {
		return itr, true, nil // Use ITR
	}

	return dfp, false, nil // Use DFP
}

func dfp(db *sql.DB, cid, year int) (map[uint32]float32, error) {
	selectReport := `
	SELECT
		CODE, VL_CONTA
	FROM
		dfp a
	WHERE
		ID_CIA = $1 
		AND YEAR = $2
		AND VERSAO = (SELECT MAX(VERSAO) FROM dfp WHERE ID_CIA = a.ID_CIA AND YEAR = a.YEAR)
	;`

	rows, err := db.Query(selectReport, cid, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[uint32]float32)
	for rows.Next() {
		var code uint32
		var vlConta float32
		rows.Scan(&code, &vlConta)
		values[code] = vlConta
	}

	return values, nil
}

//
// itrNumQuarters returns the number os quarters in the last year.
//
func itrNumQuarters(db *sql.DB, cid int) (int, error) {
	validate := `
	SELECT COUNT(*) FROM (
		SELECT 
			DISTINCT date(DT_FIM_EXERC, 'unixepoch') 
		FROM dfp d 
		WHERE 
			ID_CIA = $1 
			AND YEAR = strftime('%Y', 'now', '-1 year')
			AND VERSAO = (SELECT MAX(VERSAO) FROM dfp WHERE ID_CIA = d.ID_CIA AND YEAR = d.YEAR)

		UNION

		SELECT
			DISTINCT date(DT_FIM_EXERC, 'unixepoch')
		FROM
			itr i
		WHERE
			ID_CIA = $1
			AND YEAR >= strftime('%Y', 'now', '-1 year')
			AND VERSAO = (SELECT MAX(VERSAO) FROM itr WHERE ID_CIA = i.ID_CIA AND DT_FIM_EXERC = i.DT_FIM_EXERC)
	);`

	row := db.QueryRow(validate, cid)
	count := 0
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

//
// lastBalance returns a hash with the '[code] = value' from the balance sheet
// with the newest date available on the dfp or itr tables.
//
func lastBalance(db *sql.DB, cid int) (map[uint32]float32, error) {
	selectBalance := `
		SELECT MAX(DT), CODE, TOTAL FROM (

			SELECT 
				date(DT_FIM_EXERC, 'unixepoch') DT, CODE, SUM(VL_CONTA) TOTAL
			FROM dfp d
			WHERE 
				ID_CIA = $1
				AND VERSAO = (SELECT MAX(VERSAO) FROM dfp WHERE ID_CIA = d.ID_CIA AND YEAR = d.YEAR)
				AND CAST(substr(CD_CONTA, 1, 1) as decimal) <= 2
			GROUP BY
				DT_FIM_EXERC, CODE, CD_CONTA
				
			UNION

			SELECT 
				date(DT_FIM_EXERC, 'unixepoch') DT, CODE, SUM(VL_CONTA) TOTAL
			FROM itr i
			WHERE 
				ID_CIA = $1
				AND VERSAO = (SELECT MAX(VERSAO) FROM itr WHERE ID_CIA = i.ID_CIA AND DT_FIM_EXERC = i.DT_FIM_EXERC)
				AND CAST(substr(CD_CONTA, 1, 1) as decimal) <= 2
			GROUP BY
				DT_FIM_EXERC, CODE, CD_CONTA
			
		) 
		GROUP BY CODE
		ORDER BY CODE;
	`

	rows, err := db.Query(selectBalance, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balance := make(map[uint32]float32)
	var maxDt string
	var code uint32
	var vlConta float32

	for rows.Next() {
		rows.Scan(&maxDt, &code, &vlConta)
		balance[code] = vlConta
	}

	return balance, nil
}

//
// ttm (twelve trailling months) returns a hash with the '[code] = value'
// with the last year dfp value subtracted of the sum of last quarters from
// last year, but > 1 year ago, and then sums it with the current year's
// quarters.
//
func ttm(db *sql.DB, cid int) (map[uint32]float32, error) {
	selectQuarters := `
		SELECT CODE, SUM(TOTAL) TOTAL FROM (
			SELECT CODE, SUM(TOTAL) TOTAL, COUNT(*) N FROM (	
				SELECT 
					CODE, SUM(VL_CONTA) TOTAL
				FROM dfp d 
				WHERE 
					ID_CIA = $1
					AND YEAR = strftime('%Y', 'now', '-1 year')
					AND VERSAO = (SELECT MAX(VERSAO) FROM dfp WHERE ID_CIA = d.ID_CIA AND YEAR = d.YEAR)	
					AND CAST(substr(CD_CONTA, 1, 1) as decimal) > 2 -- IGNORE BALANCE SHEETS	
				GROUP BY
					CODE, CD_CONTA

				UNION

				SELECT 
					CODE, -1 * SUM(VL_CONTA) TOTAL 
				FROM itr i
				WHERE 
					ID_CIA = $1
					AND YEAR = strftime('%Y', 'now', '-1 year')
					AND VERSAO = (SELECT MAX(VERSAO) FROM itr WHERE ID_CIA = i.ID_CIA AND DT_FIM_EXERC = i.DT_FIM_EXERC)
					AND CAST(substr(CD_CONTA, 1, 1) as decimal) > 2 -- IGNORE BALANCE SHEETS
				GROUP BY
					CODE, CD_CONTA
			)
			GROUP BY CODE

			UNION

			SELECT
				CODE, SUM(VL_CONTA), COUNT(*) N
			FROM
				itr i
			WHERE
				ID_CIA = $1
				AND date(DT_FIM_EXERC, 'unixepoch') > date('now', '-1 year')
				AND VERSAO = (SELECT MAX(VERSAO) FROM itr WHERE ID_CIA = i.ID_CIA AND DT_FIM_EXERC = i.DT_FIM_EXERC)
				AND CAST(substr(CD_CONTA, 1, 1) as decimal) > 2 -- IGNORE BALANCE SHEETS
			GROUP BY
				CODE
		)		
		GROUP BY
			CODE
		HAVING 
			SUM(N) >= 5;
	`

	rows, err := db.Query(selectQuarters, cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ttm := make(map[uint32]float32)
	var code uint32
	var vlConta float32
	for rows.Next() {
		rows.Scan(&code, &vlConta)
		ttm[code] = vlConta
	}

	bal, err := lastBalance(db, cid)
	if err != nil {
		return nil, err
	}
	for k, v := range bal {
		ttm[k] = v
	}

	return ttm, nil
}

//
// accountsAverage stores the average of all companies of the same sector
// for each account into a map using a hash of the account code and
// description as its key
//
func (r report) accountsAverage(company string, year int) (map[uint32]float32, error) {

	companies, _, err := r.fromSector(company)
	if len(companies) <= 1 || err != nil {
		err = errors.Wrap(err, "erro ao ler arquivo de setores "+r.yamlFile)
		return nil, err
	}

	if len(companies) == 0 {
		err = errors.Errorf("erro ao procurar empresas")
		return nil, err
	}

	cids := make([]string, len(companies))
	for i, co := range companies {
		if id, err := cid(r.db, co); err == nil {
			cids[i] = strconv.Itoa(id)
		}
	}

	selectReport := fmt.Sprintf(`
	SELECT
		CODE, AVG(VL_CONTA)
	FROM
		dfp a
	WHERE
		ID_CIA IN (%s)
		AND YEAR = "%d"
		AND VERSAO = (SELECT MAX(VERSAO) FROM dfp WHERE ID_CIA = a.ID_CIA AND YEAR = a.YEAR)
	GROUP BY
		CODE;
	`, strings.Join(cids, ","), year)

	rows, err := r.db.Query(selectReport)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[uint32]float32)
	for rows.Next() {
		var code uint32
		var vlConta float32
		rows.Scan(
			&code,
			&vlConta,
		)
		values[code] = vlConta
	}

	return values, nil
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
func timeRange(db *sql.DB) (int, int, error) {

	selectYears := `
	SELECT
		MIN(CAST(YEAR AS INTEGER)),
		MAX(CAST(YEAR AS INTEGER))
	FROM dfp;`
	begin := 0
	end := 0
	err := db.QueryRow(selectYears).Scan(&begin, &end)
	if err != nil {
		return 0, 0, err
	}

	selectItrYears := `
	SELECT
		MAX(CAST(YEAR AS INTEGER))
	FROM itr;`
	end2 := 0
	err = db.QueryRow(selectItrYears).Scan(&end2)
	if err == nil && end2 > end {
		end = end2
	}

	// Check year
	if begin < 1900 || begin > 2100 || end < 1900 || end > 2100 {
		err = errors.Wrap(err, "ano inválido")
		return 0, 0, err
	}
	if begin > end {
		aux := end
		end = begin
		begin = aux
	}

	return begin, end, nil
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

func companyProfits(db *sql.DB, companyID int) ([]profit, error) {

	selectProfits := fmt.Sprintf(`
	SELECT
		YEAR,
		VL_CONTA
	FROM
		dfp a
	WHERE
		ID_CIA = "%d"
		AND CODE = "%d"
		AND VERSAO = (SELECT MAX(VERSAO) FROM dfp WHERE ID_CIA = a.ID_CIA AND YEAR = a.YEAR)
	ORDER BY
		YEAR;`, companyID, parsers.LucLiq)

	rows, err := db.Query(selectProfits)
	if err != nil {
		err = errors.Wrap(err, "falha ao ler banco de dados")
		return nil, err
	}
	defer rows.Close()

	var profits []profit
	for rows.Next() {
		var year int
		var val float32
		rows.Scan(&year, &val)
		profits = append(profits, profit{year, val})
	}

	return profits, nil
}
