package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var coisoCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicialização do Paterna: cria conta admin e prepara o ambiente",
	Long: `O comando init é a primeira coisa que você deve rodar após instalar o Paterna.
Ele pede um email e senha para criar sua conta admin, gera um token de sessão
e habilita o uso do Paterna no seu computador. Sem rodar 'paterna init', o
comando 'paterna' fica bloqueado por segurança.`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(coisoCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	if isInitialized() {
		fmt.Println("Paterna já está inicializado neste computador.")
		fmt.Println("Se quiser reautenticar use: paterna auth --login")
		return
	}

	fmt.Println("Bem-vindo ao Paterna!")
	fmt.Println("Vamos criar sua conta admin para liberar o uso da ferramenta.")
	fmt.Println()

	// register vive em auth.go (mesmo pacote): cria usuário + token
	register()

	// só marca como inicializado se realmente existe token válido
	if !validateToken() {
		fmt.Println()
		fmt.Println("Inicialização não concluída. Rode 'paterna init' novamente.")
		return
	}

	if err := markInitialized(); err != nil {
		fmt.Fprintln(os.Stderr, "erro ao marcar inicialização:", err)
		return
	}

	fmt.Println()
	fmt.Println("Paterna inicializado com sucesso!")
	fmt.Println("Agora rode: paterna")
}
