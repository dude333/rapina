package parsers

import (
	"database/sql"

	"github.com/pkg/errors"
)

type company struct {
	id   int
	name string
}

func loadCompanies(db *sql.DB) (map[string]company, error) {
	companies := make(map[string]company)

	selectCompanies := `SELECT ID, CNPJ, NAME from companies`
	rows, err := db.Query(selectCompanies)
	if err != nil {
		return companies, errors.Wrap(err, "falha ao ler banco de dados")
	}

	var id int
	var cnpj, name string

	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id, &cnpj, &name)
		companies[cnpj] = company{id, name}
	}

	return companies, nil
}

func saveCompanies(db *sql.DB, companies map[string]company) error {
	insert := `INSERT OR IGNORE INTO companies (ID,CNPJ,NAME) VALUES (?,?,?);`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return errors.Wrap(err, "erro ao preparar insert da lista de empresas")
	}
	defer stmt.Close()

	for cnpj, value := range companies {
		_, err := stmt.Exec(value.id, cnpj, value.name)
		if err != nil {
			return errors.Wrap(err, "falha ao inserir empresa")
		}
	}

	return nil
}

//
// updateCompanies inserts a new company to the map
//
func updateCompanies(companies map[string]company, cnpj, name string) {
	if _, exists := companies[cnpj]; !exists {
		companies[cnpj] = company{
			len(companies) + 100,
			name,
		}
	}
}
