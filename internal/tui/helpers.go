package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hugaojanuario/Paterna/internal/version"
)

func padRight(s string, width int) string {
	if lipgloss.Width(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-lipgloss.Width(s))
}

func alignRow(left, right string, width int) string {
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func displayVersion() string {
	v := version.Version
	if v == "" || v == "dev" {
		return "v0.1.0-alpha"
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
