package help

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component"
)

func Render(width int, height int) string {
	helpWindow := component.HelpWindow()
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, helpWindow)
}
