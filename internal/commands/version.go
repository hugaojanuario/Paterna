package commands

import (
	"fmt"

	"github.com/hugaojanuario/Paterna/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostra a versão do Paterna",
	Run:   showVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("paterna %s\n", version.Version)
	fmt.Printf("commit: %s\n", version.Commit)
	fmt.Printf("date:   %s\n", version.Date)
}
