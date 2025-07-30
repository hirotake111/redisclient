package mode

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/hirotake111/redisclient/internal/values"
)

// ListMode holds the application state except for context, window size, and redis client.
type ListMode struct {
	ErrorMsg      string
	CurrentKeyIdx int
	RedisCursor   uint64
	Keys          [][]string
	KeyHistoryIdx int
	UpdateForm    *textarea.Model
	FilterForm    *textarea.Model
	Tabs          int
	CurrentTab    int
	Value         values.Value
}

// NewListMode returns a pointer to a new ListMode with zero values.
func NewListMode(
	errorMsg string,
	currentKeyIdx int,
	redisCursor uint64,
	keys [][]string,
	keyHistoryIdx int,
	updateForm *textarea.Model,
	filterForm *textarea.Model,
	tabs int,
	currentTab int,

	value values.Value,
) *ListMode {
	return &ListMode{
		ErrorMsg:      errorMsg,
		CurrentKeyIdx: currentKeyIdx,
		RedisCursor:   redisCursor,
		Keys:          keys,
		KeyHistoryIdx: keyHistoryIdx,
		UpdateForm:    updateForm,
		FilterForm:    filterForm,
		Tabs:          tabs,
		CurrentTab:    currentTab,
		Value:         value,
	}
}
