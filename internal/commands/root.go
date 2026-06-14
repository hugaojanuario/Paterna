package commands

import (
	"os"

	"github.com/hugaojanuario/Paterna/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paterna",
	Short: "Paterna — observabilidade e gerenciamento de containers Docker",
	Long:  "Paterna abre o TUI quando executado sem argumentos. Use subcomandos para auth e init.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
