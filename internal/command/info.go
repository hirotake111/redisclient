package command

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/domain/infoid"
)

type InfoType interface {
	Type() string
}

// None info type
type InfoTypeNone struct{}

func (i InfoTypeNone) Type() string { return "none" }

// Info info type
type InfoTypeInfo struct {
	InfoId    infoid.InfoID // Unique identifier for the message
	Text      string        // The informational message text
	ExpiresIn time.Duration // Duration after which the message expires
}

func (i InfoTypeInfo) Type() string           { return "info" }
func (i InfoTypeInfo) Expires() time.Duration { return i.ExpiresIn }

// Warning info type
type InfoTypeWarning struct {
	InfoId    infoid.InfoID // Unique identifier for the message
	Text      string        // The warning message text
	ExpiresIn time.Duration // Duration after which the message expires
}

func (i InfoTypeWarning) Type() string           { return "warning" }
func (i InfoTypeWarning) Expires() time.Duration { return i.ExpiresIn }

// Error info type
type InfoTypeError struct {
	InfoId    infoid.InfoID // Unique identifier for the message
	Err       error         // Error
	ExpiresIn time.Duration // Duration after which the message expires
}

func (i InfoTypeError) Type() string           { return "error" }
func (i InfoTypeError) Expires() time.Duration { return i.ExpiresIn }

type InfoMsg struct {
	InfoType InfoType // Type of the informational message
}

func (m InfoMsg) String() string {
	switch it := m.InfoType.(type) {
	case InfoTypeInfo:
		return fmt.Sprintf("InfoMsg[Info]: %s (Id: %s, ExpiresIn: %s)", it.Text, it.InfoId, it.ExpiresIn)
	case InfoTypeWarning:
		return fmt.Sprintf("InfoMsg[Warning]: %s (Id: %s, ExpiresIn: %s)", it.Text, it.InfoId, it.ExpiresIn)
	case InfoTypeError:
		return fmt.Sprintf("InfoMsg[Error]: %s (Id: %s, ExpiresIn: %s)", it.Err.Error(), it.InfoId, it.ExpiresIn)
	default:
		return "InfoMsg[Unknown Type]"
	}
}

// NewInfoMsg creates a command that sends an InfoMsg with the given InfoType.
func NewInfoMsg(id infoid.InfoID, text string, expiresIn time.Duration) tea.Msg {
	return InfoMsg{InfoType: InfoTypeInfo{
		Text:      text,
		InfoId:    id,
		ExpiresIn: expiresIn,
	}}
}

func NewWarningMsg(id infoid.InfoID, text string, expiresIn time.Duration) tea.Msg {
	return InfoMsg{InfoType: InfoTypeWarning{
		Text:      text,
		InfoId:    id,
		ExpiresIn: expiresIn,
	}}
}

func NewErrorMsg(id infoid.InfoID, err error, expiresIn time.Duration) tea.Msg {
	return InfoMsg{InfoType: InfoTypeError{
		Err:       err,
		InfoId:    id,
		ExpiresIn: expiresIn,
	}}
}

type InfoExpiredMsg struct {
	Id infoid.InfoID // Unique identifier for the message that has expired
}

func (i InfoExpiredMsg) String() string {
	return fmt.Sprintf("InfoExpiredMsg:Id=%s", i.Id)
}
