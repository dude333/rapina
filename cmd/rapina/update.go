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
	"errors"
	"fmt"
	"os"

	"github.com/dude333/rapina"
	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/reports"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sectors bool

// getUpdate represents the get command
var getUpdate = &cobra.Command{
	Use:     "update",
	Aliases: []string{"get"},
	Short:   "Baixa os arquivos da CVM e atualiza o bando de dados",
	Long:    `Baixa os arquivos do site da CVM, processa e os armazena no bando de dados.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := openDatabase()
		if err != nil {
			fmt.Println("[x]", err)
			return
		}

		fmt.Println("[√] Coletando dados ===========")
		err = fetch.Sectors(yamlFile)
		if err != nil && !errors.Is(err, rapina.ErrFileNotUpdated) {
			fmt.Println("[x]", err)
			os.Exit(1)
		}
		if err == nil {
			fmt.Println("[√] Arquivo salvo:", yamlFile)
		}
		//
		fmt.Println()
		//
		if sectors { // skip if -s flag is selected (dowload only the sectors)
			return
		}

		err = fetch.CVM(db, dataDir)
		if err != nil {
			fmt.Println("[x]", err)
			os.Exit(1)
		}

		// Stock codes
		log := reports.NewLogger(os.Stderr)
		stock, err := fetch.NewStock(db, log, viper.GetString("apikey"), dataDir)
		if err != nil {
			log.Error(err.Error())
			return
		}
		_ = stock.UpdateStockCodes()
	},
}

func init() {
	rootCmd.AddCommand(getUpdate)

	getUpdate.Flags().BoolVarP(&sectors, "sectors", "s", false, "Baixa a classificação setorial das empresas e fundos negociados na B3")
}
