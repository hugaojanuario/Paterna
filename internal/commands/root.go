package commands

import (
	"os"

	"github.com/hugaojanuario/Paterna/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paterna",
	Short: "Paterna — monitor de sistema e containers Docker no terminal",
	Long: `Paterna é um monitor de sistema estilo btop direto no terminal.
Mostra CPU, memória, disco, rede, processos e seus containers Docker.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
