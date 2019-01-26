// Copyright © 2019 Adriano P
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
	"github.com/dude333/rapina"
	"github.com/spf13/cobra"
)

// sectorCmd represents the sector command
var sectorCmd = &cobra.Command{
	Hidden: true,
	Use:    "sector nome_empresa",
	Short:  "Mostra todas as empresas do mesmo setor",
	Args:   cobra.MinimumNArgs(1),
	Run:    sector,
}
var setorCmd = &cobra.Command{
	Use:   "setor nome_empresa",
	Short: "Mostra todas as empresas do mesmo setor",
	Args:  cobra.MinimumNArgs(1),
	Run:   sector,
}

func init() {
	rootCmd.AddCommand(sectorCmd)
	rootCmd.AddCommand(setorCmd)
}

func sector(ccmd *cobra.Command, args []string) {
	rapina.ListSector(args[0], yamlFile)
}
