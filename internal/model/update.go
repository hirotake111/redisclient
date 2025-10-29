package model

import (
	"log"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/command"
)

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return command.GetKeys(m.ctx, m.redis, m.mode.FilterForm.Value())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update app state
	m.State, cmd = m.State.Update(msg)
	cmds = append(cmds, cmd)

	if err, ok := msg.(command.ErrMsg); ok {
		log.Printf("Received error message: %s", err.Err)
		return m.UpdateErrorMessage(err.Err), command.TickAndClear(5*time.Second, "error")
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		log.Printf("Received window size message: width=%d, height=%d", msg.Width, msg.Height)
		return m.UpdateWindowSize(msg.Height, msg.Width), nil
	}

	if msg, ok := msg.(command.TimedOutMsg); ok {
		switch msg.Kind {
		case "error":
			return m.ClearErrorMessage(), nil
		}
		return m, nil // No action for other timeout kinds
	}

	m.mode.Viewport, cmd = m.mode.Viewport.Update(msg, m.State)
	cmds = append(cmds, cmd)

	m.mode.KeyList, cmd = m.mode.KeyList.Update(m.ctx, m.redis, msg, m.State)
	cmds = append(cmds, cmd)

	// List mode (defalt)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		log.Printf("KEY HIT: \"%s\"", key)
		if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() || key == "q" {
			if m.mode.KeyList.FilterState() != list.Filtering {
				return m, tea.Quit
			}
		}

		if key == "enter" {
			selected := m.mode.KeyList.SelectedItem().FilterValue()
			cmds = append(cmds, command.GetValue(m.ctx, m.redis, selected))
		}

		if key == "y" {
			if m.mode.Value.Data() == "" {
				// No current key to copy
			} else {
				log.Print("key 'c' pressed, copying value of current key to clipboard")
				cmds = append(cmds, command.CopyValueToClipboard(m.ctx, m.mode.Value.Data()))
			}
		}

		if key == tea.KeyTab.String() {
			m = m.NextTab()
			cmds = append(cmds, command.SwitchTab(m.ctx, m.redis, m.mode.CurrentTab))
		}

		if key == tea.KeyShiftTab.String() {
			m = m.PreviousTab()
			cmds = append(cmds, command.SwitchTab(m.ctx, m.redis, m.mode.CurrentTab))
		}

		if key == "x" {
			log.Print("key 'x' pressed, deleting current key")
			// currentKey := m.currentKey()
			currentKey := m.mode.KeyList.Model.SelectedItem().FilterValue()
			if currentKey == "" {
				log.Print("No current key selected for deletion")
			} else {
				log.Printf("Deleting key: %s", currentKey)
				cmds = append(cmds, command.DeleteKey(m.ctx, m.redis, currentKey))
			}
		}

	case command.ValueUpdatedMsg:
		m = m.UpdateValue(msg)

	case command.NewRedisClientMsg:
		log.Print("Received new Redis client message")
		m = m.UpdateRedisClient(msg)
		cmds = append(cmds, command.GetKeys(m.ctx, m.redis, m.mode.FilterForm.Value())) // Re-fetch keys with the new client

	case command.HighlightedKeyUpdatedMsg:
		log.Printf("Highlighted key updated to: %s", msg.Key)
		return m, command.GetValue(m.ctx, m.redis, msg.Key)
	}

	return m, tea.Batch(cmds...)
}
