/*
Copyright © 2021 Adriano P <dev@dude333.com>
Distributed under the MIT License.
*/
package main

import (
	"log"
	"strings"

	"github.com/dude333/rapina/reports"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type fiiDividendsFlags struct {
	format string // output format of the report
}

// fiiDividendsCmd represents the rendimentos command
var fiiDividendsCmd = &cobra.Command{
	Use:     "rendimentos",
	Aliases: []string{"rend", "dividendos", "dividends", "div"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Lista os rendimentos de um FII",
	Long:    `Lista os rendimentos de um Fundos de Investiment Imobiliários (FII).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Number of reports
		n := flags.fii.num
		if n <= 0 {
			n = 1
		}

		parms := make(map[string]string)
		// Verbose
		if flags.verbose {
			parms[Fverbose] = "true"
		}
		// Report format
		parms[Fformat] = flags.fii.dividends.format

		if err := FIIDividends(parms, args, n); err != nil {
			log.Println(err)
		}

	},
}

func init() {
	fiiCmd.AddCommand(fiiDividendsCmd)
	fiiDividendsCmd.Flags().StringVarP(&flags.fii.dividends.format, Fformat,
		"f", "tabela", "formato do relatório: tabela|csv|csvrend")
}

//
// FIIDividends prints the dividends from 'code' for 'n' months,
// starting from latest.
//
func FIIDividends(parms map[string]string, codes []string, n int) error {
	for i := 0; i < len(codes); i++ {
		codes[i] = strings.ToUpper(codes[i])
	}

	db, err := openDatabase()
	if err != nil {
		return err
	}

	r, err := reports.NewFIITerminal(db, viper.GetString("apikey"), dataDir)
	if err != nil {
		return err
	}

	r.SetParms(parms)

	err = r.Dividends(codes, n)
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
