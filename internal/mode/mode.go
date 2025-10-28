package mode

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/hirotake111/redisclient/internal/component/list"
	"github.com/hirotake111/redisclient/internal/component/viewport"
	"github.com/hirotake111/redisclient/internal/values"
)

const (
	defaultKeyListWIdth   = 30
	defaultKeyListHeight  = 20
	defaultViewportWidth  = 50
	defaultViewportHeight = 20
)

// ListMode holds the application state except for context, window size, and redis client.
type ListMode struct {
	ErrorMsg      string
	CurrentKeyIdx int
	Keys          []string
	UpdateForm    *textarea.Model
	FilterForm    *textarea.Model
	Tabs          int
	CurrentTab    int // Also an index for Redis database
	Value         values.Value
	KeyList       list.CustomKeyList
	Viewport      viewport.Viewport
}

// NewListMode returns a pointer to a new ListMode with zero values.
func NewListMode(
	errorMsg string,
	currentKeyIdx int,
	keys []string,
	updateForm *textarea.Model,
	filterForm *textarea.Model,
	tabs int,
	currentTab int,
	value values.Value,
) *ListMode {
	return &ListMode{
		ErrorMsg:      errorMsg,
		CurrentKeyIdx: currentKeyIdx,
		Keys:          keys,
		UpdateForm:    updateForm,
		FilterForm:    filterForm,
		Tabs:          tabs,
		CurrentTab:    currentTab,
		Value:         value,
		KeyList:       list.New(keys, defaultKeyListWIdth, defaultKeyListHeight),
		Viewport:      viewport.New(defaultViewportWidth, defaultViewportHeight),
	}
}
