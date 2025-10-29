package command

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type InfoType interface {
	Type() string
}

// None info type
type InfoTypeNone struct{}

func (i InfoTypeNone) Type() string { return "none" }

// Info info type
type InfoTypeInfo struct {
	Text      string        // The informational message text
	InfoId    string        // Unique identifier for the message
	ExpiresIn time.Duration // Duration after which the message expires
}

func (i InfoTypeInfo) Type() string           { return "info" }
func (i InfoTypeInfo) Expires() time.Duration { return i.ExpiresIn }

// Warning info type
type InfoTypeWarning struct {
	Text      string        // The warning message text
	InfoId    string        // Unique identifier for the message
	ExpiresIn time.Duration // Duration after which the message expires
}

func (i InfoTypeWarning) Type() string           { return "warning" }
func (i InfoTypeWarning) Expires() time.Duration { return i.ExpiresIn }

// Error info type
type InfoTypeError struct {
	Text      string        // The error message text
	InfoId    string        // Unique identifier for the message
	ExpiresIn time.Duration // Duration after which the message expires
}

func (i InfoTypeError) Type() string           { return "error" }
func (i InfoTypeError) Expires() time.Duration { return i.ExpiresIn }

type InfoMsg struct {
	InfoType InfoType // Type of the informational message
}

func (InfoMsg) Kind() string {
	return "info"
}

// SendInfoMsgCmd creates a command that sends an InfoMsg with the given InfoType.
func SendInfoMsgCmd(infoType InfoType) tea.Cmd {
	return func() tea.Msg {
		return InfoMsg{InfoType: infoType}
	}
}

type InfoExpiredMsg struct {
	Id string // Unique identifier for the message that has expired
}

func (InfoExpiredMsg) Kind() string {
	return "info_expired"
}
