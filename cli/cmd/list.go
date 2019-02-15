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

package cmd

import (
	"fmt"

	"github.com/dude333/rapina"
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
		if listCmd.Flags().NFlag() == 0 {
			listCmd.Help()
			return
		}

		if listCompanies {
			rapina.ListCompanies()
		}
		if sector != "" {
			rapina.ListSector(sector, yamlFile)
		}
		if listCmd.Flags().Changed("lucroLiquido") {
			err := rapina.ListCompaniesProfits(netProfitRate)
			if err != nil {
				fmt.Println("[x]", err)
			}
		}
	}

}
