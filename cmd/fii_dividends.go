/*
Copyright © 2021 Adriano P <dev@dude333.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/parsers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fiiDividendsCmd represents the rendimentos command
var fiiDividendsCmd = &cobra.Command{
	Use:     "rendimentos",
	Aliases: []string{"rend", "dividendos", "dividends", "div"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Lista os rendimentos de um FII",
	Long:    `Lista os rendimentos de um Fundos de Investiment Imobiliários (FII).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Number of reports
		n, err := cmd.Flags().GetInt("num")
		log.Println("n =", n)
		if err != nil || n <= 0 {
			n = 1
		}

		if err := FIIDividends(args[0], n); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	fiiCmd.AddCommand(fiiDividendsCmd)
}

//
// FIIDividends prints the dividends from 'code' fund for 'n' months,
// starting from latest.
//
func FIIDividends(code string, n int) error {
	db, err := openDatabase()
	if err != nil {
		return err
	}
	code = strings.ToUpper(code)

	srv, err := fetch.NewStockServer(db, viper.GetString("apikey"))
	if err != nil {
		return err
	}
	err = srv.FetchStockQuote(fix(code))
	if err != nil {
		return err
	}

	fii, _ := parsers.NewFII(db, srv.QuoteFromDB)

	cnpj, err := cnpj(fii, code)
	if err != nil {
		return err
	}
	err = fii.FetchFIIDividends(cnpj, n)
	if err != nil {
		return err
	}

	return nil
}

//
// cnpj returns the CNPJ from FII code. It first checks the DB and, if not
// found, fetches from B3.
//
func cnpj(fii *parsers.FII, code string) (string, error) {
	fiiDetails, err := fii.SelectFIIDetails(code)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("[x] error", err)
	}
	if err == nil && fiiDetails.DetailFund.CNPJ != "" {
		fmt.Println("DB", code, fiiDetails.DetailFund.CNPJ)
		return fiiDetails.DetailFund.CNPJ, nil
	}
	//
	// Fetch online if DB fails
	fiiDetails, err = fii.FetchFIIDetails(code)
	if err != nil {
		return "", err
	}

	fmt.Println("online", code, fiiDetails.DetailFund.CNPJ)

	return fiiDetails.DetailFund.CNPJ, nil
}

func fix(code string) string {
	if len(code) == 4 {
		return code + "11"
	}
	return code
}
