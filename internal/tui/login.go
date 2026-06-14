package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hugaojanuario/Paterna/internal/repository"
	"github.com/hugaojanuario/Paterna/pkg/bcrypt"
)

type LoginModel struct {
	emailInput    textinput.Model
	passwordInput textinput.Model
	focused       int
	err           string
}

func NewLoginModel() LoginModel {
	email := textinput.New()
	email.Placeholder = "voce@email.com"
	email.CharLimit = 100
	email.Width = 40
	email.Focus()

	password := textinput.New()
	password.Placeholder = "senha"
	password.CharLimit = 100
	password.Width = 40
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'

	return LoginModel{
		emailInput:    email,
		passwordInput: password,
		focused:       0,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab":
			m.focused = 1 - m.focused
			if m.focused == 0 {
				m.emailInput.Focus()
				m.passwordInput.Blur()
			} else {
				m.passwordInput.Focus()
				m.emailInput.Blur()
			}
			return m, textinput.Blink

		case "enter":
			email := strings.TrimSpace(m.emailInput.Value())
			pw := m.passwordInput.Value()

			if email == "" || pw == "" {
				m.err = "preencha email e senha"
				return m, nil
			}

			user, err := repository.GetByEmail(email)
			if err != nil {
				m.err = "credenciais inválidas"
				return m, nil
			}

			if !bcrypt.CheckHash(pw, user.PasswordHash) {
				m.err = "credenciais inválidas"
				return m, nil
			}

			next := NewContainersModel()
			return next, next.Init()
		}
	}

	var cmd tea.Cmd
	if m.focused == 0 {
		m.emailInput, cmd = m.emailInput.Update(msg)
	} else {
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}
	return m, cmd
}

func (m LoginModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Paterna")

	subtitle := lipgloss.NewStyle().
		Faint(true).
		Render("Faça login para acessar o painel")

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("250"))

	form := fmt.Sprintf(
		"%s\n%s\n\n%s\n%s",
		labelStyle.Render("Email"),
		m.emailInput.View(),
		labelStyle.Render("Senha"),
		m.passwordInput.View(),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Render(form)

	errStr := ""
	if m.err != "" {
		errStr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(m.err)
	}

	footer := lipgloss.NewStyle().
		Faint(true).
		Render("tab: alternar campos  enter: entrar  esc: sair")

	return fmt.Sprintf("\n  %s\n  %s\n\n  %s\n\n  %s\n\n  %s\n",
		title, subtitle, box, errStr, footer)
}
