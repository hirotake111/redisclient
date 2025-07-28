package model

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/cmd"
)

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return cmd.GetKeys(m.ctx, m.redis, m.redisCursor, m.filterForm.Value())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if err, ok := msg.(cmd.ErrMsg); ok {
		log.Printf("Received error message: %s", err.Err)
		return m.UpdateErrorMessage(err.Err), cmd.TickAndClear(5*time.Second, "error")
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		log.Printf("Received window size message: height=%d, width=%d", msg.Height, msg.Width)
		return m.UpdateWindowSize(msg.Height, msg.Width), nil
	}

	if msg, ok := msg.(cmd.TimedOutMsg); ok {
		switch msg.Kind {
		case "error":
			return m.ClearErrorMessage(), nil
		}
		return m, nil // No action for other timeout kinds
	}

	switch m.state {
	case HelpWindowState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			if key == tea.KeyCtrlC.String() {
				log.Print("Exiting app...")
				return m, tea.Quit
			}
			log.Print("Closing help window")
			return m.ToListState(), nil
		}

	// END OF UPDATE VALUE STATE

	case ListState:
		if m.updateForm.Focused() {
			//
			// UPDATE VALUE FORM ACTIVATED
			//
			log.Print("UPDATE VALUE FORM ACTIVATED")
			switch msg := msg.(type) {
			case tea.KeyMsg:
				key := msg.String()
				if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() {
					log.Print("Exiting update value form")
					m.updateForm.Blur()
					return m, nil
				}
				if key == tea.KeyEnter.String() {
					log.Printf("Updating value for key: %s", m.currentKey())
					m.updateForm.Blur()
					return m, cmd.UpdateValue(m.ctx, m.redis, m.currentKey(), m.updateForm.Value())
				}
				// Handle form input
				log.Printf("Appending character \"%s\" to update form value for key \"%s\"", msg, m.currentKey())
				newForm, cmd := m.updateForm.Update(msg)
				m.updateForm = &newForm
				return m, cmd
			}
			return m, nil

		}
		if m.filterForm.Focused() {
			//
			// FILTER MODE ACTIVATED
			//
			log.Print("FILTER MODE ACTIVATED")
			switch msg := msg.(type) {
			case tea.KeyMsg:
				key := msg.String()
				if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() {
					log.Print("Exiting filter mode")
					m.filterForm.Blur()
					return m, nil
				}
				if key == tea.KeyEnter.String() {
					log.Printf("Applyig filter keyword: \"%s\"", m.filterForm.Value())
					m.filterForm.Blur()
					m = m.ClearCurrentKeyIdx().ClearKeyHistory().ClearRedisCursor()
					return m, cmd.GetKeys(m.ctx, m.redis, m.redisCursor, m.filterForm.Value()) // Re-fetch keys with the filter applied
				}
				// Handle filter input
				log.Printf("Appending character '%s' to form value", key)
				newForm, cmd := m.filterForm.Update(msg)
				m.filterForm = &newForm
				return m, cmd
			}
		}

		// List mode (defalt)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() || key == "q" {
				return m, tea.Quit
			}
			if key == "j" || key == tea.KeyDown.String() {
				log.Print("Moving cursor down")
				m = m.MoveCursorDown()
				return m, cmd.GetValue(m.ctx, m.redis, m.currentKey())
			}
			if key == "k" || key == tea.KeyUp.String() {
				log.Print("Moving cursor up")
				m = m.MoveCursorUp()
				return m, cmd.GetValue(m.ctx, m.redis, m.currentKey())
			}
			if key == tea.KeyEnter.String() {
				log.Print("Actrivating update value form")
				return m, m.updateForm.Focus()
			}
			if key == "/" {
				log.Print("Activating filter mode")
				return m, m.filterForm.Focus()
			}
			if key == "n" {
				log.Print("key 'n' pressed, moving to next key list")
				if m.HasNextHistory() {
					log.Print("Next history exists, moving to next key list")
					m = m.NextHistory()
					return m, cmd.GetValue(m.ctx, m.redis, m.currentKey()) // Fetch value for the current key
				}
				if m.HasMoreKeysOnServer() {
					log.Print("Fetching a next key list from server")
					return m, cmd.GetKeys(m.ctx, m.redis, m.redisCursor, m.filterForm.Value()) // Fetch keys for the new tab
				}
				log.Print("No more keys to fetch")
				return m, nil
			}
			if key == "p" {
				log.Print("key 'p' pressed, moving to previous key list")
				if m.HasPreviousKeys() {
					log.Print("Moving to previous key list")
					m = m.PreviousHistory()
				} else {
					log.Print("No previous keys to fetch")
				}
				return m, cmd.GetValue(m.ctx, m.redis, m.currentKey()) // Fetch value for the current key
			}
			if key == "y" {
				if m.value.Data() == "" {
					return m, nil // No current key to copy
				}
				log.Print("key 'c' pressed, copying value of current key to clipboard")
				return m, cmd.CopyValueToClipboard(m.ctx, m.value.Data())
			}
			if key == tea.KeyTab.String() {
				m = m.NextTab()
				return m, cmd.UpdateDatabase(m.ctx, m.redis, m.currentTab)
			}
			if key == tea.KeyShiftTab.String() {
				m = m.PreviousTab()
				return m, cmd.UpdateDatabase(m.ctx, m.redis, m.currentTab)
			}
			if key == "d" {
				log.Print("key 'd' pressed, deleting current key")
				currentKey := m.currentKey()
				if currentKey == "" {
					log.Print("No current key selected for deletion")
					return m, nil
				}
				log.Printf("Deleting key: %s", currentKey)
				return m, cmd.DeleteKey(m.ctx, m.redis, currentKey)
			}
			if key == "?" {
				log.Print("key '?' pressed, showing help")
				return m.toHelpWindowState(), nil
			}

		case cmd.ValueUpdatedMsg:
			return m.UpdateValue(msg), nil

		case cmd.KeysUpdatedMsg:
			log.Printf("Received keys updated message. len: %d. cursor: %d", len(msg.Keys), msg.RedisCursor)
			m = m.UpdateKeyList(msg)
			if len(msg.Keys) == 0 {
				log.Print("No keys found, returning empty value")
				m.EmptyValue()
				return m, cmd.DisplayEmptyValue
			}
			return m, cmd.GetValue(m.ctx, m.redis, m.currentKey()) // Fetch value for the first key

		case cmd.KeyDeletedMsg:
			log.Printf("Received key deleted message for key: %s", msg.Key)
			m = m.DeleteKeyFromList(msg.Key)
			return m, nil

		case cmd.NewRedisClientMsg:
			log.Print("Received new Redis client message")
			m = m.UpdateRedisClient(msg).ClearCurrentKeyIdx().ClearKeyHistory().ClearRedisCursor()
			return m, cmd.GetKeys(m.ctx, m.redis, m.redisCursor, m.filterForm.Value()) // Re-fetch keys with the new client
		}

	}

	return m, nil
}
