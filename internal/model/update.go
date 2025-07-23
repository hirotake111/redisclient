package model

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/cmd"
)

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return cmd.GetKeys(m.ctx, m.redis, m.redisCursor)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case UpdateValueState:
		//
		// UPDATE VALUE STATE
		//
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			return m.UpdateWindowSize(msg.Height, msg.Width), nil
		case tea.KeyMsg:
			key := msg.String()
			if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() || key == tea.KeyTab.String() || key == tea.KeyShiftTab.String() {
				log.Print("Exiting value update state")
				m = m.ToListState()
				return m, nil
			}
			if key == tea.KeyEnter.String() {
				log.Print("Enter key pressed, performing value update")
				return m, cmd.UpdateValue(m.ctx, m.redis, m.currentKey(), m.formValue)
			}
			if key == tea.KeyBackspace.String() {
				log.Print("Backspace key pressed, removing last character from value")
				m = m.removeCharFromFormValue()
				return m, nil
			}
		//TODO: add key to form value

		case cmd.ErrMsg:
			log.Printf("Error occurred: %s", msg.Err)
		}

	// END OF UPDATE VALUE STATE

	case ListState:
		if m.filterHighlighted {
			//
			// FILTER MODE ACTIVATED
			//
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				return m.UpdateWindowSize(msg.Height, msg.Width), nil
			case tea.KeyMsg:
				key := msg.String()
				if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() {
					log.Print("Exiting filter mode")
					m = m.ToggleFilterHighlight()
					return m, nil
				}
				if key == tea.KeyBackspace.String() {
					m = m.removeCharFromFilterValue()
					return m, nil
				}
				if key == tea.KeyEnter.String() {
					m = m.ToggleFilterHighlight()
					log.Printf("Filter applied: %s", m.filterValue)
					return m, nil
				}
				// Handle filter input
				m = m.appendCharToFilterValue(key)
				return m, nil
			}
		}

		// filter mode not activated
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			return m.UpdateWindowSize(msg.Height, msg.Width), nil
		case tea.KeyMsg:
			key := msg.String()
			if m.displayHelp {
				log.Print("Exiting help window")
				m = m.ToggleHelpWindow()
				return m, nil
			}
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
				log.Print("Enter key pressed, open value update window")
				m = m.ToValueUpdateState()
				return m, nil
			}
			if key == "/" {
				log.Print("Filter mode activated")
				m = m.ToggleFilterHighlight()
				m = m.ClarFilterValue()
				return m, nil
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
					return m, cmd.GetKeys(m.ctx, m.redis, m.redisCursor) // Fetch keys for the new tab
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
				return m.ToggleHelpWindow(), nil
			}

		case cmd.ValueMsg:
			return m.UpdateValue(msg), nil

		case cmd.KeysUpdatedMsg:
			log.Printf("Received keys updated message. len: %d. cursor: %d", len(msg.Keys), msg.RedisCursor)
			m = m.UpdateKeyList(msg)
			if len(msg.Keys) == 0 {
				log.Print("No keys found, returning empty value")
				m.value = ""
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
			return m, cmd.GetKeys(m.ctx, m.redis, m.redisCursor)

		case cmd.ErrMsg:
			log.Printf("Error occurred: %s", msg.Err)
		}

	}

	return m, nil
}
