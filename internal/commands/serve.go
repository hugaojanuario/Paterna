package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hugaojanuario/Paterna/internal/api"
	"github.com/spf13/cobra"
)

var servePort string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Sobe a API HTTP do Paterna",
	Long:  "Sobe o servidor HTTP que expõe a API REST do Paterna. A interface web em www/ consome esta API.",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&servePort, "port", "p", "8080", "Porta do servidor HTTP")
}

func runServe(cmd *cobra.Command, args []string) error {
	if envPort := os.Getenv("API_PORT"); envPort != "" {
		servePort = envPort
	}

	fmt.Println("Paterna API rodando em :" + servePort)
	return http.ListenAndServe(":"+servePort, api.Router())
}
