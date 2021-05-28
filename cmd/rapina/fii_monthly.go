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

type fiiMonthlyFlags struct {
	format string // output format of the report
}

// fiiMonthlyCmd represents the rendimentos command
var fiiMonthlyCmd = &cobra.Command{
	Use:     "mensal",
	Aliases: []string{"monthly"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Lista os informes mensais de um FII",
	Long:    `Lista os informes mensais de um Fundos de Investiment Imobiliários (FII).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Number of reports
		n := flags.fii.num

		parms := make(map[string]string)
		// Verbose
		if flags.verbose {
			parms[Fverbose] = "true"
		}
		// Report format
		parms[Fformat] = flags.fii.monthly.format

		if err := FIIMonthly(parms, args, n); err != nil {
			log.Println(err)
		}

	},
}

func init() {
	fiiCmd.AddCommand(fiiMonthlyCmd)
	fiiMonthlyCmd.Flags().StringVarP(&flags.fii.monthly.format, Fformat,
		"f", "tabela", "formato do relatório: tabela|csv|csvrend")
}

//
// FIIMonthly prints the monthly reports from 'code' for 'n' months,
// starting from latest.
//
func FIIMonthly(parms map[string]string, codes []string, n int) error {
	for i := 0; i < len(codes); i++ {
		codes[i] = strings.ToUpper(codes[i])
	}

	db, err := openDatabase()
	if err != nil {
		return err
	}

	opts := reports.FIITerminalOptions{
		APIKey:  viper.GetString("apikey"),
		DataDir: dataDir,
	}

	r, err := reports.NewFIITerminal(db, opts)
	if err != nil {
		return err
	}

	r.SetParms(parms)

	err = r.Monthly(codes, n)
	if err != nil {
		return err
	}

	return nil
}
