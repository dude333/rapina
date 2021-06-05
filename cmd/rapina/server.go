/*
Copyright © 2021 Adriano P <dev@dude333.com>
Distributed under the MIT License.
*/
package main

import (
	"log"

	"github.com/dude333/rapina/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type serverFlags struct {
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Inicia o servidor web",
	Long:  `Comando para iniciar o servidor para a exibição dos dados via web browser.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := serve(nil)
		if err != nil {
			log.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	// serverCmd.Flags().IntVarP(&flags.server.num,
	// 	Fnum, "n", 1, "número de meses desde o último disponível")
}

func serve(parms map[string]string) error {

	db, err := openDatabase()
	if err != nil {
		return err
	}

	server.HTML(
		server.WithDB(db),
		server.WithAPIKey(viper.GetString("apikey")),
		server.WithDataDir(dataDir))

	return nil
}
