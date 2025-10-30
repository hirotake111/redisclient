package infobox

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/util"
)

var (
	defaultContainerStyle = lipgloss.NewStyle().
				Padding(0, 1).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(color.Grey).
				Foreground(color.Grey)

	defaultTitleStyle = lipgloss.NewStyle().
				MarginBottom(1).
				Padding(0, 1).
				Foreground(color.Secondary)

	titleInfoStyle = defaultTitleStyle.
			Background(color.Primary).
			Foreground(color.White)

	titleWarnStyle = defaultTitleStyle.
			Background(color.Warning).
			Foreground(color.Black)

	titleErrorStyle = defaultTitleStyle.
			Background(color.Error).
			Foreground(color.White)
)

type InfoBox struct {
	infoType command.InfoType // Type of the informational message
}

func New() InfoBox {
	return InfoBox{
		infoType: command.InfoTypeNone{},
	}
}

func (i InfoBox) Update(msg tea.Msg) (InfoBox, tea.Cmd) {
	util.LogMsg("InfoBox received a message", msg)

	if m, ok := msg.(command.InfoExpiredMsg); ok {
		log.Printf("InfoBox received InfoExpiredMsg for Id: %s", m.Id)
		switch it := i.infoType.(type) {
		case command.InfoTypeInfo:
			if m.Id == it.InfoId {
				i.infoType = command.InfoTypeNone{}
			}
		case command.InfoTypeWarning:
			if m.Id == it.InfoId {
				i.infoType = command.InfoTypeNone{}
			}
		case command.InfoTypeError:
			if m.Id == it.InfoId {
				i.infoType = command.InfoTypeNone{}
			}
		default:
			// Do nothing
		}
		return i, nil
	}

	m, ok := msg.(command.InfoMsg)
	if !ok {
		return i, nil
	}

	i.infoType = m.InfoType

	log.Printf("InfoBox sending expiration command for InfoType: %+v", m.InfoType.Type())
	return i, func() tea.Msg {
		switch it := m.InfoType.(type) {
		case command.InfoTypeInfo:
			log.Printf("Message %s will expire in %s", it.InfoId, it.ExpiresIn.String())
			time.Sleep(it.ExpiresIn)
			log.Printf("Message %s expired", it.InfoId)
			return command.InfoExpiredMsg{Id: it.InfoId}
		case command.InfoTypeWarning:
			log.Printf("Message %s will expire in %s", it.InfoId, it.ExpiresIn.String())
			time.Sleep(it.ExpiresIn)
			return command.InfoExpiredMsg{Id: it.InfoId}
		case command.InfoTypeError:
			log.Printf("Message %s will expire in %s", it.InfoId, it.ExpiresIn.String())
			time.Sleep(it.ExpiresIn)
			return command.InfoExpiredMsg{Id: it.InfoId}
		default: // None
			return nil
		}
	}
}

func (i InfoBox) View(width, height int) string {
	var title, text string
	container := defaultContainerStyle.Width(width).Height(height)

	switch it := i.infoType.(type) {
	case command.InfoTypeInfo:
		container = container.BorderForeground(color.Primary).Foreground(color.Primary)
		title = InfoInfoTitle()
		text = it.Text

	case command.InfoTypeWarning:
		container = container.BorderForeground(color.Warning).Foreground(color.Warning)
		title = InfoWarnTitle()
		text = it.Text

	case command.InfoTypeError:
		container = container.BorderForeground(color.Error).Foreground(color.Error)
		title = ErrorTitle()
		text = it.Err.Error()

	default:
		// No info to show
	}

	return container.Render(lipgloss.JoinVertical(lipgloss.Left, title, text))
}

func InfoNoneTitle() string {
	return defaultTitleStyle.Render("")
}

func InfoInfoTitle() string {
	return titleInfoStyle.Render("INFO")
}

func InfoWarnTitle() string {
	return titleWarnStyle.Render("WARN")
}

func ErrorTitle() string {
	return titleErrorStyle.Render("ERROR")
}
