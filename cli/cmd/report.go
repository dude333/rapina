/// Copyright © 2018 Adriano P
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

// Flags
var scriptMode bool
var outputDir string

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report [-s] nome_empresa",
	Short: "Cria planilha com dados da companhia escolhida",
	Long:  "Cria planilha com dados da companhia escolhida",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		report(args[0])
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	reportCmd.Flags().BoolVarP(&scriptMode, "scriptMode", "s", false, "Para modo script (escolhe a empresa com nome mais próximo)")
	reportCmd.Flags().StringVarP(&outputDir, "outputDir", "d", "", "Diretório onde o relatório será salvo")
}

func report(company string) {
	company = rapina.SelectCompany(company, scriptMode)
	if company == "" {
		fmt.Println("[x] Empresa não encontrada")
		return
	}
	fmt.Println()
	fmt.Printf("[✓] Criando relatório para %s ========\n", company)
	err := rapina.Report(company, outputDir, "./setores.yaml")
	if err != nil {
		fmt.Println("[x]", err)
	}
}
