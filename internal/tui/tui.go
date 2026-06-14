package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	p := tea.NewProgram(NewContainersModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
