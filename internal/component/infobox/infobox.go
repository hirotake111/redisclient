package infobox

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/command"
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
	if m, ok := msg.(command.InfoExpiredMsg); ok {
		log.Printf("InfoBox received InfoExpiredMsg for Id: %s", m.Id)
		log.Printf("Current InfoType: %+v", i.infoType)
		switch it := i.infoType.(type) {
		case command.InfoTypeInfo:
			if m.Id == it.InfoId {
				log.Print("Expiring InfoTypeInfo")
				i.infoType = command.InfoTypeNone{}
			}
		case command.InfoTypeWarning:
			if m.Id == it.InfoId {
				log.Print("Expiring InfoTypeWarning")
				i.infoType = command.InfoTypeNone{}
			}
		case command.InfoTypeError:
			if m.Id == it.InfoId {
				log.Print("Expiring InfoTypeError")
				i.infoType = command.InfoTypeNone{}
			}
		default:
			// Do nothing
			log.Print("InfoBox received InfoExpiredMsg but no matching InfoType to expire")
		}
		return i, nil
	}

	// if m, ok := msg.(command.KeysUpdatedMsg); ok {
	// 	log.Printf("InfoBox received KeysUpdatedMsg with %d keys", len(m.Keys))
	// 	id, err := infoid.New()
	// 	if err != nil {
	// 		log.Printf("Error generating info ID: %v", err)
	// 		i.infoType = command.InfoTypeError{
	// 			InfoId:    "unknown",
	// 			Err:       err,
	// 			ExpiresIn: 5 * time.Second,
	// 		}
	// 		return i, nil
	// 	}
	// 	i.infoType = command.InfoTypeInfo{
	// 		InfoId:    id,
	// 		Text:      fmt.Sprintf("Fetched %d keys from Redis.", len(m.Keys)),
	// 		ExpiresIn: 5 * time.Second,
	// 	}
	// 	return i, nil
	// }

	if m, ok := msg.(command.InfoMsg); ok {
		log.Printf("InfoBox received InfoMsg: %+v", m)
		i.infoType = m.InfoType
		var id string
		var expiresIn time.Duration

		switch it := m.InfoType.(type) {
		case command.InfoTypeInfo:
			id = it.InfoId
			expiresIn = it.ExpiresIn
		case command.InfoTypeWarning:
			id = it.InfoId
			expiresIn = it.ExpiresIn
		case command.InfoTypeError:
			id = it.InfoId
			expiresIn = it.ExpiresIn
		default: // None
			return i, nil
		}

		return i, func() tea.Msg {
			log.Printf("Message %s will expire in %s", id, expiresIn.String())
			time.Sleep(expiresIn)
			return command.InfoExpiredMsg{Id: id}
		}
	}

	return i, nil
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
