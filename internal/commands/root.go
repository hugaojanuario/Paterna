package commands

import (
	"fmt"
	"os"

	"github.com/hugaojanuario/Paterna/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paterna",
	Short: "Paterna — observabilidade e gerenciamento de containers Docker",
	Long: `Paterna abre o TUI quando executado sem argumentos.
Na primeira vez rode 'paterna init' para criar sua conta admin.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isInitialized() {
			fmt.Println("Paterna ainda não foi inicializado neste computador.")
			fmt.Println("Rode primeiro: paterna init")
			return nil
		}

		if !validateToken() {
			fmt.Println("Você não está autenticado ou sua sessão expirou.")
			fmt.Println("Rode: paterna auth --login")
			return nil
		}

		return tui.Run()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
