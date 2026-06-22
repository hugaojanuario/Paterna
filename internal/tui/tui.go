package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hugaojanuario/Paterna/internal/client"
)

var apiClient *client.Client

func Run() error {
	email, password, err := client.LoadCredentials()
	if err != nil {
		return fmt.Errorf("carregar credenciais: %w", err)
	}

	baseURL := client.BaseURL()
	token, err := client.Login(baseURL, email, password)
	if err != nil {
		return fmt.Errorf("autenticar no daemon (%s): %w", baseURL, err)
	}

	apiClient = client.New(baseURL, token)

	p := tea.NewProgram(NewWelcomeModel(), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
