/*
Copyright © 2021 Adriano P <dev@dude333.com>
Distributed under the MIT License.
*/
package main

import (
	"github.com/dude333/rapina/reports"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type serverFlags struct {
	num int
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Inicia o servidor web",
	Long:  `Comando para iniciar o servidor para a exibição dos dados via web browser.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Number of reports
		n := flags.server.num
		if n <= 0 {
			n = 1
		}

		server(args, n, nil)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&flags.server.num,
		Fnum, "n", 1, "número de meses desde o último disponível")
}

func server(codes []string, n int, parms map[string]string) error {

	db, err := openDatabase()
	if err != nil {
		return err
	}

	reports.HTMLServer(
		codes,
		n,
		reports.WithDB(db),
		reports.WithAPIKey(viper.GetString("apikey")),
		reports.WithDataDir(dataDir))

	return nil
}
