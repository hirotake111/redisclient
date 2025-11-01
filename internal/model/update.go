package model

import (
	"log"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/domain/infoid"
)

const (
	expiration = 5 * time.Second
)

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")

	id, err := infoid.New()
	if err != nil {
		return func() tea.Msg {
			return command.NewErrorMsg("unknown", err, expiration)
		}
	}

	cmd := func() tea.Msg {
		return command.NewInfoMsg(
			id,
			"Connected to Redis successfully.",
			expiration,
		)
	}
	return tea.Batch(command.GetKeys(m.ctx, m.redis, ""), cmd)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update app state
	m.State, cmd = m.State.Update(msg)
	cmds = append(cmds, cmd)

	// Update info box
	m.infoBox, cmd = m.infoBox.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
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

	m.viewport, cmd = m.viewport.Update(msg, m.State)
	cmds = append(cmds, cmd)

	m.keyList, cmd = m.keyList.Update(m.ctx, m.redis, msg, m.State)
	cmds = append(cmds, cmd)

	// List mode (defalt)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		log.Printf("KEY HIT: \"%s\"", key)
		if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() || key == "q" {
			if m.keyList.FilterState() != list.Filtering {
				return m, tea.Quit
			}
		}

		// TODO: Copy value to clipboard
		// if key == "y" {
		// 	if m.mode.Value.Data() == "" {
		// 		// No current key to copy
		// 	} else {
		// 		log.Print("key 'c' pressed, copying value of current key to clipboard")
		// 		cmds = append(cmds, command.CopyValueToClipboard(m.ctx, m.mode.Value.Data()))
		// 	}
		// }

		if key == tea.KeyTab.String() {
			m = m.NextTab()
			cmds = append(cmds, command.SwitchTab(m.ctx, m.redis, m.currentTab))
		}

		if key == tea.KeyShiftTab.String() {
			m = m.PreviousTab()
			cmds = append(cmds, command.SwitchTab(m.ctx, m.redis, m.currentTab))
		}

	case command.NewRedisClientMsg:
		log.Print("Received new Redis client message")
		m = m.UpdateRedisClient(msg)
		cmds = append(cmds, command.GetKeys(m.ctx, m.redis, "")) // Re-fetch keys with the new client

	case command.HighlightedKeyUpdatedMsg:
		log.Printf("Highlighted key updated to: %s", msg.Key)
		return m, command.GetValue(m.ctx, m.redis, msg.Key)
	}

	return m, tea.Batch(cmds...)
}
