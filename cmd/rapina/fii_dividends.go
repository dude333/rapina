/*
Copyright © 2021 Adriano P <dev@dude333.com>
Distributed under the MIT License.
*/
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dude333/rapina/reports"
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
		if err != nil || n <= 0 {
			n = 1
		}

		code := args[0]

		if len(code) != 6 {
			fmt.Println("[x] Código inválido:", code)
			fmt.Println("[i] Padrão experado: ABCD11")
			os.Exit(1)
		}

		// FII code ('ABCD11')
		if err := FIIDividends(args[0], n); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	fiiCmd.AddCommand(fiiDividendsCmd)
}

//
// FIIDividends prints the dividends from 'code' for 'n' months,
// starting from latest.
//
func FIIDividends(code string, n int) error {
	code = strings.ToUpper(code)

	db, err := openDatabase()
	if err != nil {
		return err
	}

	r, err := reports.NewFIITerminalReport(db, viper.GetString("apikey"))
	if err != nil {
		return err
	}
	err = r.Dividends(code, n)
	if err != nil {
		return err
	}
	/*
		// QUOTES
		stockStore, _ := parsers.NewStockStore(db)
		stock, err := fetch.NewStockFetch(stockStore, viper.GetString("apikey"))
		if err != nil {
			return err
		}
		q, err := stock.Quote(code, "2011-01-02")
		if err != nil {
			return err
		}
		fmt.Println(q)

		// DIVIDENDS
		fiiStore := parsers.NewFIIStore(db)
		fii := fetch.NewFII(fiiStore)
		err = fii.Dividends(code, n)
		if err != nil {
			return err
		}

		// DIVIDENDS REPORT
		r, err := reports.NewFIITerminalReport(db)
		if err != nil {
			return err
		}
		r.Dividends(code, n)
		if err != nil {
			return err
		}
	*/
	return nil
}
