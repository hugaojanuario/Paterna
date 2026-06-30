package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run abre a TUI no dashboard estilo btop.
func Run() error {
	p := tea.NewProgram(NewDashboardModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
