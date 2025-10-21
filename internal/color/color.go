package color

import "github.com/charmbracelet/lipgloss"

var (
	// 🌿 Primary brand color — grounded, natural green tone
	Primary = lipgloss.Color("#0F828C") // OliveDrab

	// 🍃 Secondary accent — gentle green-grey, pairs well with Primary
	Secondary = lipgloss.Color("#065084") // VerdantGrey

	// ☁️ Background — soft, calm neutral base
	Background = lipgloss.Color("#A8BCA1") // SageMist

	// ✨ Highlight / Accent — for hover states or selections
	Accent = lipgloss.Color("#DDDDDD") // PaleSprout

	// 🪵 Text / Foreground — deep contrasting green
	Foreground = lipgloss.Color("#36493E") // CharcoalGreen

	// 🚨 Error — muted coral red, stands out but still fits the vibe
	Error = lipgloss.Color("#D86B6B")

	// 🩶 Neutral grey — perfect for borders, inactive text, or separators
	Grey = lipgloss.Color("#9CA3AF")
)
