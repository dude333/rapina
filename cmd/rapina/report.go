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

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dude333/rapina/reports"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Flags
var scriptMode bool
var all bool
var showShares bool
var extraRatios bool
var fleuriet bool
var omitSector bool
var outputDir = "reports"
var format string // output format of the report

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

	reportCmd.Flags().BoolVarP(&scriptMode, "scriptMode", "s", false, "Para modo script (escolhe a empresa com nome mais próximo)")
	reportCmd.Flags().BoolVarP(&all, "all", "a", false, "Mostra todos os indicadores")
	reportCmd.Flags().BoolVarP(&showShares, "showShares", "f", false, "Mostra o número de ações e free float")
	reportCmd.Flags().BoolVarP(&extraRatios, "extraRatios", "x", false, "Reporte de índices extras")
	reportCmd.Flags().BoolVarP(&fleuriet, "fleuriet", "F", false, "Capital de giro no modelo Fleuriet")
	reportCmd.Flags().BoolVarP(&omitSector, "omitSector", "o", false, "Omite o relatório das empresas do mesmo setor")
	reportCmd.Flags().StringVarP(&outputDir, "outputDir", "d", "reports", "Diretório onde o relatório será salvo")
	reportCmd.Flags().StringVarP(&format, "format", "r", "xlsx", "Formato do relatório: xlsx|stdout")
}

func report(company string) {
	var spcfctnCd string = "ON"
	company = SelectCompany(company, scriptMode)
	if company == "" {
		fmt.Println("[x] Empresa não encontrada")
		return
	}
	if strings.Contains(company, "@#") {
		companyWithTicker := strings.Split(company, "@#")
		company = companyWithTicker[0]
		spcfctnCd = companyWithTicker[1]
	}
	fmt.Println()
	fmt.Printf("[√] Criando relatório para %s ========\n", company)

	if all {
		extraRatios = true
		showShares = true
		fleuriet = true
	}

	r := make(map[string]bool)
	r["ExtraRatios"] = extraRatios
	r["ShowShares"] = showShares
	r["Fleuriet"] = fleuriet
	r["PrintSector"] = !omitSector

	parms := Parms{
		Company:   company,
		SpcfctnCd: spcfctnCd,
		Format:    format,
		OutputDir: outputDir,
		YamlFile:  yamlFile,
		Reports:   r,
	}
	err := Report(parms)
	if err != nil {
		fmt.Println("[x]", err)
	}
}

//
// SelectCompany returns the company name compared to the names
// stored in the DB
//
func SelectCompany(company string, scriptMode bool) string {
	db, err := openDatabase()
	if err != nil {
		fmt.Println("[x]", err)
		return ""
	}

	companies, err := reports.ListCompanies(db)
	if err != nil {
		fmt.Println("[x]", err)
		return ""
	}

	// Do a fuzzy match on the company name against
	// all companies listed on the DB
	matches := make([]string, 0, 10)
	for _, c := range companies {
		if fuzzy.MatchNormalizedFold(company, c) {
			matches = append(matches, c)
		}
	}

	// Script mode
	if len(matches) >= 1 && scriptMode {
		rank := fuzzy.RankFindNormalizedFold(company, matches)
		if len(rank) <= 0 {
			return ""
		}
		sort.Sort(rank)
		return rank[0].Target
	}

	// Interactive menu
	if len(matches) >= 1 {
		result := promptUser(matches, "Selecione a Empresa")

		tickers, err := reports.ListTickers(db, result)
		if err != nil {
			fmt.Println("[x] Recuperando lista de tickers ", err)
			return result
		}

		// Interactive menu
		if len(tickers) > 0 {
			ticker := promptUser(tickers, "Selecione o ticker")
			resultWithTicker := fmt.Sprintf("%s@#%s", result, reports.GetSpcfctnCd(db, result, ticker))
			return resultWithTicker
		}

		return result
	}

	return ""
}

//
// Report a company from DB to Excel
//
func Report(p Parms) (err error) {

	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	if p.OutputDir == "" {
		p.OutputDir = outputDir
	}

	file, err := filename(p.OutputDir, p.Company)
	if err != nil {
		return err
	}

	parms := map[string]interface{}{
		"db":        db,
		"dataDir":   dataDir,
		"company":   p.Company,
		"SpcfctnCd": p.SpcfctnCd,
		"format":    p.Format,
		"filename":  file,
		"yamlFile":  p.YamlFile,
		"reports":   p.Reports,
	}

	if p.Format == "stdout" {
		return reports.ReportToStdout(parms)
	}

	return reports.ReportToXlsx(parms)
}
