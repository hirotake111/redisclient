package viewport

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/state"
)

var (
	defaultContainer = lipgloss.NewStyle().
				Padding(0, 1).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(color.Primary)

	activeContainer = defaultContainer.BorderStyle(lipgloss.ThickBorder())

	// Styles for various UI components
	titleBarStyle = lipgloss.NewStyle().MarginBottom(1).Padding(0, 1).Background(color.Primary).Foreground(color.White)
)

type Viewport struct {
	model viewport.Model
	ttl   int64
}

func New(width, height int) Viewport {
	return Viewport{
		model: viewport.New(width, height),
		ttl:   0,
	}
}

func (v Viewport) View(width, height int, st state.AppState) string {
	v.model.Width = width - 2
	v.model.Height = height - 2
	title := ValueTitle(v.ttl)
	container := defaultContainer
	if st.ViewportActive() {
		container = activeContainer
	}
	return container.Render(lipgloss.JoinVertical(lipgloss.Left, title, v.model.View()))
}

func (v Viewport) Update(msg tea.Msg, st state.AppState) (Viewport, tea.Cmd) {
	var cmd tea.Cmd
	if msg, ok := msg.(command.ValueUpdatedMsg); ok {
		v.ttl = msg.TTL
		v.model.SetContent(pretty(msg.NewValue))
		return v, nil
	}
	log.Printf("viewport active: %t", st.ViewportActive())
	if !st.ViewportActive() {
		return v, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			return v, state.DeactivateViewportCmd
		}
	}

	v.model, cmd = v.model.Update(msg)
	return v, cmd
}

func ValueTitle(ttl int64) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		titleBarStyle.Render("VALUE"),
		ttlIndicator(ttl),
	)
}

func ttlIndicator(ttl int64) string {
	if ttl < 0 {
		return ""
	}
	if ttl == 0 {
		return ""
	}
	return " (expires in " + strconv.FormatInt(ttl, 10) + " seconds)"
}

func pretty(s string) string {
	var sb strings.Builder
	tabCount := 0
	for _, r := range s {
		switch r {
		case '{', '[':
			tabCount++
			sb.WriteString(fmt.Sprintf("%c\n%s", r, strings.Repeat("  ", tabCount)))
		case '}', ']':
			tabCount--
			sb.WriteString(fmt.Sprintf("\n%s%c", strings.Repeat("  ", tabCount), r))
		case ',':
			sb.WriteString(",\n" + strings.Repeat("  ", tabCount))
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
