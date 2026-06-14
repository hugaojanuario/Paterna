package commands

import (
	"bufio"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"syscall"

	"github.com/hugaojanuario/Paterna/internal/repository"
	"github.com/hugaojanuario/Paterna/pkg/bcrypt"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Configuração inicial: cria conta de administrador",
	Long: `Cria a primeira conta de administrador e inicializa o banco local.

Execute este comando uma vez, na primeira instalação. Após o init:
  - 'paterna' abre o painel TUI
  - 'paterna serve' sobe a API HTTP para a interface web`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Paterna — Configuração inicial")
	fmt.Println()

	email, err := readEmail()
	if err != nil {
		return err
	}

	password, err := readPassword("Senha: ")
	if err != nil {
		return err
	}

	if len(password) < 6 {
		return errors.New("senha precisa ter ao menos 6 caracteres")
	}

	confirm, err := readPassword("Confirme a senha: ")
	if err != nil {
		return err
	}

	if password != confirm {
		return errors.New("senhas não conferem")
	}

	hash, err := bcrypt.Hash(password)
	if err != nil {
		return err
	}

	if err := repository.Create(email, hash); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Conta criada com sucesso.")
	fmt.Println()
	fmt.Println("Próximos passos:")
	fmt.Println("  paterna         abre o painel TUI")
	fmt.Println("  paterna serve   sobe a API HTTP em :8080")
	fmt.Println()

	return nil
}

func readEmail() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Email: ")

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	email := strings.TrimSpace(line)
	if _, err := mail.ParseAddress(email); err != nil {
		return "", errors.New("email inválido")
	}

	return email, nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
