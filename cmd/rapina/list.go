// Copyright © 2018 Adriano P
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"fmt"

	"github.com/dude333/rapina/reports"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista informações armazenadas no banco de dados",
}

func init() {
	var (
		listCompanies bool
		sector        string
		netProfitRate float32
	)

	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listCompanies, "empresas", "e", false, "Lista todas as empresas disponíveis")
	listCmd.Flags().StringVarP(&sector, "setor", "s", "", "Lista todas as empresas do mesmo setor")
	listCmd.Flags().Float32VarP(&netProfitRate, "lucroLiquido", "l", -0.8, "Lista empresas com lucros lucros positivos e com a taxa de crescimento definida")

	listCmd.Run = func(cmd *cobra.Command, args []string) {
		var err error

		if listCmd.Flags().NFlag() == 0 {
			_ = listCmd.Help()
			return
		}

		if listCompanies {
			err = ListCompanies()
		} else if sector != "" {
			err = ListSector(sector, yamlFile)
		} else if listCmd.Flags().Changed("lucroLiquido") {
			err = ListCompaniesProfits(netProfitRate)
		}
		if err != nil {
			fmt.Println("[x]", err)
		}
	}

}

//
// ListCompanies a company from DB to Excel
//
func ListCompanies() (err error) {
	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	com, err := reports.ListCompanies(db)
	if err != nil {
		return errors.Wrap(err, "erro ao listar empresas")
	}
	for _, c := range com {
		fmt.Println(c)
	}

	return
}

//
// ListSector shows all companies from the same sector as 'company'
//
func ListSector(company, yamlFile string) (err error) {
	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	err = reports.ListSector(db, company, yamlFile)
	if err != nil {
		return errors.Wrap(err, "erro ao listar empresas")
	}

	return
}

//
// ListCompaniesProfits lists companies profits
//
func ListCompaniesProfits(rate float32) (err error) {
	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	err = reports.ListCompaniesProfits(db, rate)
	if err != nil {
		return errors.Wrap(err, "erro ao listar lucros")
	}

	return
}
