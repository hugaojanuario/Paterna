package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tentaculum-dev/go-sdk/validate"
	"github.com/google/uuid"
	"github.com/hugaojanuario/Paterna/internal/repository"
	"github.com/hugaojanuario/Paterna/pkg/bcrypt"
	"github.com/spf13/cobra"
)

// coisoCmd represents the coiso command
var authentication = &cobra.Command{
	Use:   "auth",
	Short: "Autenticação de usuário para acessar a aplicação",
	Long: `A autenticação é necessária para acessar a aplicação.
	Use --login para fazer login e obter um token de acesso.
	Use --register para se registrar e criar uma conta para usar a aplicação.
	O token de acesso é armazenado em um arquivo no diretório de configuração do usuário e tem um tempo de validade de 1 mês.`,
	Run: auth,
}

var loginFlag bool
var registerFlag bool

func init() {
	rootCmd.AddCommand(authentication)

	authentication.Flags().BoolVar(&loginFlag, "login", false, "Fazer login")
	authentication.Flags().BoolVar(&registerFlag, "register", false, "Registrar usuário")
}

// primeiro o usuario precisa fazer a autenticação para conseguir utilizar
// a aplicação.
// se for --login, o usuário pode se autenticar e obter um token de acesso para usar a aplicação
// se for --register, o usuário pode se registrar e criar uma conta para usar a aplicação
// o token de acesso é armazenado em um arquivo no diretório de configuração do usuário e tem um tempo de validade de 1 mês
func auth(cmd *cobra.Command, args []string) {
	if !loginFlag && !registerFlag {
		fmt.Println("Use --login ou --register")
		os.Exit(1)
	}

	if loginFlag {
		login()
		return
	}

	if registerFlag {
		register()
		return
	}
}

func validateToken() bool {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Erro ao acessar diretório de configuração:", err)
		return false
	}

	appDir := filepath.Join(configDir, "Paterna")

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		fmt.Println("Nenhum token de acesso encontrado.")
		return false
	}

	files, err := os.ReadDir(appDir)
	if err != nil {
		fmt.Println("Erro ao ler diretório de configuração:", err)
		return false
	}

	var tokenFile string

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".paterna" {
			tokenFile = file.Name()
			break
		}
	}

	if tokenFile == "" {
		fmt.Println("Nenhum token de acesso encontrado.")
		return false
	}

	tokenPath := filepath.Join(appDir, tokenFile)

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		fmt.Println("Erro ao ler token:", err)
		return false
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "valid_until:") {
			continue
		}

		value := strings.TrimSpace(
			strings.TrimPrefix(line, "valid_until:"),
		)

		validUntil, err := time.Parse(
			"2006-01-02 15:04:05",
			value,
		)

		if err != nil {
			fmt.Println("Token inválido: data de expiração mal formatada.")
			return false
		}

		if time.Now().After(validUntil) {
			fmt.Println("Token expirado. Removendo...")

			removeToken(tokenPath)

			return false
		}

		return true
	}

	fmt.Println("Token inválido: campo valid_until não encontrado.")
	return false
}

func removeToken(tokenPath string) {
	if err := os.Remove(tokenPath); err != nil {
		fmt.Println("Erro ao remover token:", err)
		return
	}

	fmt.Println("Token removido com sucesso.")
}

func login() {
	var hash string
	var email string
	var password string
	var err error

	fmt.Println("Digite seu email:")
	fmt.Scanln(&email)
	fmt.Println("Digite sua senha:")
	fmt.Scanln(&password)

	repositoryUser, err := repository.GetByEmail(email)
	if err != nil {
		fmt.Println("Erro ao buscar usuário:", err)
		return
	}

	if repositoryUser == nil {
		fmt.Println("Usuário não encontrado. Por favor, registre-se para continuar.")
		return
	}

	if !bcrypt.CheckHash(password, repositoryUser.PasswordHash) {
		fmt.Println("Senha incorreta. Por favor, tente novamente.")
		return
	}

	//refresh do token de acesso
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Erro ao acessar diretório de configuração:", err)
		return
	}

	appDir := filepath.Join(configDir, "Paterna")

	files, err := os.ReadDir(appDir)
	if err != nil {
		fmt.Println("Erro ao ler diretório de configuração:", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".paterna" {
			err = os.Remove(filepath.Join(appDir, file.Name()))
			if err != nil {
				fmt.Println("Erro ao remover token de acesso antigo:", err)
				return
			}
		}
	}

	hash, err = bcrypt.Hash(time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		fmt.Println("Erro ao criar usuário:", err)
		return
	}

	filePath := filepath.Join(appDir, hash+".paterna")
	userLogged, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer userLogged.Close()

	values := fmt.Sprintf("email: %s\npassword: %s\ngenerated_at: %s\n valid_until: %s\n, hash: %s\n",
		email, password, time.Now().Format("2006-01-02 15:04:05"), time.Now().Add(730*time.Hour).Format("2006-01-02 15:04:05"), hash)
	_, err = userLogged.WriteString(values)
	if err != nil {
		panic(err)
	}

	fmt.Println("Login bem-sucedido! Bem-vindo de volta!")
}

func register() {
	var hash string
	var email string
	var password string
	var err error

	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	appDir := filepath.Join(configDir, "Paterna")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		panic(err)
	}

	// pegando o email e a senha do usuário para autenticação
	fmt.Println("Digite seu email:")
	fmt.Scanln(&email)
	fmt.Println("Digite sua senha:")
	fmt.Scanln(&password)

	err = validate.Mail(email)
	if err != nil {
		fmt.Println("Email inválido:", err)
		return
	}

	passwordHash, err := bcrypt.Hash(password)
	if err != nil {
		fmt.Println("Erro ao criar usuário:", err)
		return
	}

	err = repository.Create(email, passwordHash)
	if err != nil {
		fmt.Println("Erro ao criar usuário:", err)
		return
	}

	// criando um hash para o usuário logado
	hash = uuid.NewString()
	filePath := filepath.Join(appDir, hash+".paterna")
	userLogged, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer userLogged.Close()

	values := fmt.Sprintf("email: %s\npassword: %s\ngenerated_at: %s\n valid_until: %s\n, hash: %s\n",
		email, password, time.Now().Format("2006-01-02 15:04:05"), time.Now().Add(730*time.Hour).Format("2006-01-02 15:04:05"), hash)
	_, err = userLogged.WriteString(values)
	if err != nil {
		panic(err)
	}

	fmt.Println("Usuário registrado com sucesso!")
}
