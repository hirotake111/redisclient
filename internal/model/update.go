package model

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/cmd"
)

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return cmd.GetKeys(m.ctx, m.redis, m.mode.FilterForm.Value())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	var cmds []tea.Cmd

	if err, ok := msg.(cmd.ErrMsg); ok {
		log.Printf("Received error message: %s", err.Err)
		return m.UpdateErrorMessage(err.Err), cmd.TickAndClear(5*time.Second, "error")
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		log.Printf("Received window size message: width=%d, height=%d", msg.Width, msg.Height)
		return m.UpdateWindowSize(msg.Height, msg.Width), nil
	}

	if msg, ok := msg.(cmd.TimedOutMsg); ok {
		switch msg.Kind {
		case "error":
			return m.ClearErrorMessage(), nil
		}
		return m, nil // No action for other timeout kinds
	}

	var c tea.Cmd
	m.mode.KeyList, c = m.mode.KeyList.Update(msg) // Update the key list component
	cmds = append(cmds, c)

	switch m.State {
	case ListState:
		if m.mode.UpdateForm.Focused() {
			//
			// UPDATE VALUE FORM ACTIVATED
			//
			log.Print("UPDATE VALUE FORM ACTIVATED")
			switch msg := msg.(type) {
			case tea.KeyMsg:
				key := msg.String()
				if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() {
					log.Print("Exiting update value form")
					m.mode.UpdateForm.Blur()
				}
				if key == tea.KeyEnter.String() {
					log.Printf("Updating value for key: %s", m.currentKey())
					m.mode.UpdateForm.Blur()
					cmds = append(cmds, cmd.UpdateValue(m.ctx, m.redis, m.currentKey(), m.mode.UpdateForm.Value()))
				}
				// Handle form input
				log.Printf("Appending character \"%s\" to update form value for key \"%s\"", msg, m.currentKey())
				newForm, cmd := m.mode.UpdateForm.Update(msg)
				m.mode.UpdateForm = &newForm
				cmds = append(cmds, cmd)
			}
			return m, nil

		}
		if m.mode.FilterForm.Focused() {
			//
			// FILTER MODE ACTIVATED
			//
			log.Print("FILTER MODE ACTIVATED")
			switch msg := msg.(type) {
			case tea.KeyMsg:
				key := msg.String()
				if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() {
					log.Print("Exiting filter mode")
					m.mode.FilterForm.Blur()
				}
				if key == tea.KeyEnter.String() {
					log.Printf("Applyig filter keyword: \"%s\"", m.mode.FilterForm.Value())
					m.mode.FilterForm.Blur()
					m = m.ResetKeyIndex()
					cmds = append(cmds, cmd.GetKeys(m.ctx, m.redis, m.mode.FilterForm.Value()))
				}
				// Handle filter input
				log.Printf("Appending character '%s' to form value", key)
				newForm, cmd := m.mode.FilterForm.Update(msg)
				m.mode.FilterForm = &newForm
				cmds = append(cmds, cmd)
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
				if m, err = m.MoveCursorDown(); err != nil {
					log.Printf("m.MoveCursorDown: %s", err)
				} else {
					cmds = append(cmds, cmd.GetValue(m.ctx, m.redis, m.currentKey()))
				}
			}

			if key == "k" || key == tea.KeyUp.String() {
				log.Print("Moving cursor up")
				if m, err = m.MoveCursorUp(); err != nil {
					log.Printf("Error on m.MoveCursorUp: %s", err)
				} else {
					cmds = append(cmds, cmd.GetValue(m.ctx, m.redis, m.currentKey()))
				}
			}

			if key == tea.KeyEnter.String() {
				log.Print("Actrivating update value form")
				cmds = append(cmds, m.mode.UpdateForm.Focus())
			}

			if key == "/" {
				log.Print("Activating filter mode")
				cmds = append(cmds, m.mode.FilterForm.Focus())
			}

			if key == "y" {
				if m.mode.Value.Data() == "" {
					// No current key to copy
				} else {
					log.Print("key 'c' pressed, copying value of current key to clipboard")
					cmds = append(cmds, cmd.CopyValueToClipboard(m.ctx, m.mode.Value.Data()))
				}
			}

			if key == tea.KeyTab.String() {
				m = m.NextTab()
				cmds = append(cmds, cmd.SwitchTab(m.ctx, m.redis, m.mode.CurrentTab))
			}

			if key == tea.KeyShiftTab.String() {
				m = m.PreviousTab()
				cmds = append(cmds, cmd.SwitchTab(m.ctx, m.redis, m.mode.CurrentTab))
			}

			if key == "d" {
				log.Print("key 'd' pressed, deleting current key")
				currentKey := m.currentKey()
				if currentKey == "" {
					log.Print("No current key selected for deletion")
					return m, tea.Batch(cmds...)
				}
				log.Printf("Deleting key: %s", currentKey)
				cmds = append(cmds, cmd.DeleteKey(m.ctx, m.redis, currentKey))
			}

		case cmd.ValueUpdatedMsg:
			m = m.UpdateValue(msg)

		case cmd.KeysUpdatedMsg:
			log.Printf("Received keys updated message. The message has %d keys", len(msg.Keys))
			m = m.UpdateKeyList(msg)
			if len(msg.Keys) == 0 {
				log.Print("No keys found, returning empty value")
				m.EmptyValue()
				cmds = append(cmds, cmd.DisplayEmptyValue)
				return m, tea.Batch(cmds...)
			}
			cmds = append(cmds, cmd.GetValue(m.ctx, m.redis, m.currentKey())) // Fetch value for the first key

		case cmd.KeyDeletedMsg:
			log.Printf("Received key deleted message for key: %s", msg.Key)
			m = m.DeleteKeyFromList(msg.Key)

		case cmd.NewRedisClientMsg:
			log.Print("Received new Redis client message")
			m = m.UpdateRedisClient(msg).ResetKeyIndex()
			cmds = append(cmds, cmd.GetKeys(m.ctx, m.redis, m.mode.FilterForm.Value())) // Re-fetch keys with the new client
		}

	}

	return m, tea.Batch(cmds...)
}
