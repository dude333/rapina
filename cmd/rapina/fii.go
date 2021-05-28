/*
Copyright © 2021 Adriano P <dev@dude333.com>
Distributed under the MIT License.
*/
package main

import (
	"github.com/spf13/cobra"
)

type fiiFlags struct {
	num       int // number of months since current
	dividends fiiDividendsFlags
	monthly   fiiMonthlyFlags
}

// fiiCmd represents the fii command
var fiiCmd = &cobra.Command{
	Use:   "fii",
	Short: "Comando relacionados aos FIIs",
	Long:  `Comando relacionado aos Fundos de Investiment Imobiliários (FII).`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(fiiCmd)
	fiiCmd.PersistentFlags().IntVarP(&flags.fii.num,
		Fnum, "n", 1, "número de meses desde o último disponível")
}
