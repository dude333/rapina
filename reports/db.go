package reports

import (
	"database/sql"
	"fmt"
	"time"

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
func accountsItems(db *sql.DB, company string) (items []accItems, err error) {
	selectItems := fmt.Sprintf(`
	SELECT DISTINCT
		CODE, CD_CONTA, DS_CONTA
	FROM
		dfp
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "_LTIMO"

	ORDER BY
		CD_CONTA, DS_CONTA
	;`, company)

	rows, err := db.Query(selectItems)
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

	// genericPrint(rows)

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
func accountsValues(db *sql.DB, company string, year int, penult bool) (values map[uint32]float32, err error) {

	period := "_LTIMO"
	if penult {
		period = "PEN_LTIMO"
		year++
	}

	layout := "2006-01-02"
	var t [2]time.Time
	for i, y := range [2]int{year, year + 1} {
		t[i], err = time.Parse(layout, fmt.Sprintf("%d-01-01", y))
		if err != nil {
			err = errors.Wrapf(err, "data invalida %d", year)
			return
		}
	}

	selectReport := fmt.Sprintf(`
	SELECT
		CODE,
		DENOM_CIA,
		ORDEM_EXERC,
		DT_REFER,
		VL_CONTA
	FROM
		dfp
	WHERE
		DENOM_CIA LIKE "%s%%"
		AND ORDEM_EXERC LIKE "%s"
		AND DT_REFER >= "%v" AND DT_REFER < "%v"
	;`, company, period, t[0].Unix(), t[1].Unix())

	values = make(map[uint32]float32)
	st := account{}

	rows, err := db.Query(selectReport)
	if err != nil {
		return
	}
	defer rows.Close()

	// var y time.Time
	for rows.Next() {
		rows.Scan(
			&st.code,
			nil, // denom_cia
			nil, // ordem_exerc
			nil, // dt_refer
			&st.vlConta,
		)

		values[st.code] = st.vlConta
	}

	return
}

//
// genericPrint prints the entire row
//
func genericPrint(rows *sql.Rows) (err error) {
	limit := 0
	cols, _ := rows.Columns()
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		// m := make(map[string]interface{})
		for i := range cols {
			val := columnPointers[i].(*interface{})
			// m[colName] = *val
			// fmt.Println(colName, *val)

			switch (*val).(type) {
			default:
				fmt.Print(*val, ";")
			case []uint8:
				y := *val
				var x = y.([]uint8)
				fmt.Print(string(x[:]), ";")
			}
		}
		fmt.Println()

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		// fmt.Println(m)
		limit++
		if limit >= 4000 {
			break
		}
	}

	return
}

//
// companies returns available companies in the DB
//
func companies(db *sql.DB) (list []string, err error) {

	selectCompanies := `
		SELECT DISTINCT
			DENOM_CIA
		FROM
			dfp
		ORDER BY
			DENOM_CIA;`

	rows, err := db.Query(selectCompanies)
	if err != nil {
		err = errors.Wrap(err, "falha ao ler banco de dados")
		return
	}
	defer rows.Close()

	var companyName string
	for rows.Next() {
		rows.Scan(&companyName)
		list = append(list, companyName)
	}

	return
}

//
// isCompany returns true if company exists on DB
//
func isCompany(db *sql.DB, company string) bool {
	selectCompany := fmt.Sprintf(`
	SELECT DISTINCT
		DENOM_CIA
	FROM
		dfp
	WHERE
		DENOM_CIA LIKE "%s%%";`, company)

	var c string
	err := db.QueryRow(selectCompany).Scan(&c)
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
		MIN(CAST(strftime('%Y', DT_REFER, 'unixepoch') AS INTEGER)),
		MAX(CAST(strftime('%Y', DT_REFER, 'unixepoch') AS INTEGER))
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
