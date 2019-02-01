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
	"os"

	"github.com/dude333/rapina"
	"github.com/spf13/cobra"
)

const yamlFile = "./setores.yml"

var sectors bool

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Baixa os arquivos da CVM e os armazena no bando de dados",
	Long:  `Baixa os arquivos do site da CVM, processa e os armazena no bando de dados.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[√] Coletando dados ===========")
		var err error
		if !sectors { // if -s flag is selected, dowload only the sectors
			err = rapina.FetchCVM()
		}
		if err != nil {
			fmt.Println("[x]", err)
			os.Exit(1)
		}
		err = rapina.FetchSectors(yamlFile)
		if err != nil {
			fmt.Println("[x]", err)
			os.Exit(1)
		}
		fmt.Println("[√] Arquivo salvo: setores.yaml")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVarP(&sectors, "sectors", "s", false, "Baixa a classificação setorial das empresas e fundos negociados na B3")
}
