package cmd

import (
	"frpok/config"
	"frpok/runner"
	"log"

	"github.com/spf13/cobra"
)

// var cfg config.Config

func init() {
	rootCmd.AddCommand(httpCommand)
}

var httpCommand = &cobra.Command{
	Use:   "http",
	Short: "Expose port for http protocol",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadIni(cfgFile)

		// should have server_addr in common
		if !cfg.Common.HasKey("server_addr") {
			log.Fatal("Config Error: server_addr should in common section!")
		}

		runner.HTTPRun(cfg, args)

		// log.Println(args)

	},
}
