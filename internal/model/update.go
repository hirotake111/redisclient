package model

import (
	"log"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/util"
)

const (
	expiration = 5 * time.Second
)

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")

	var it command.InfoType
	if id, err := util.NewID(); err == nil {
		it = command.InfoTypeInfo{
			Text:      "Connected to Redis successfully.",
			InfoId:    id,
			ExpiresIn: expiration,
		}
	} else {
		it = command.InfoTypeError{
			Text:      "Failed to generate unique ID for info message.",
			InfoId:    "conn_success_no_id",
			ExpiresIn: expiration,
		}
	}

	return tea.Batch(
		command.GetKeys(m.ctx, m.redis, ""),
		command.SendInfoMsgCmd(it),
	)
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

		if key == "enter" {
			selected := m.keyList.SelectedItem().FilterValue()
			cmds = append(cmds, command.GetValue(m.ctx, m.redis, selected))
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
