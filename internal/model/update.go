package model

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/domain/infoid"
	"github.com/hirotake111/redisclient/internal/util"
)

const (
	expiration   = 3 * time.Second
	tickDuration = 3 * time.Second
)

func doTick() tea.Cmd {
	return tea.Tick(tickDuration, func(t time.Time) tea.Msg {
		return command.TickMsg{Time: t}
	})
}

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return tea.Batch(
		command.GetKeys(m.ctx, m.redis, ""),
		command.NewInfoInfoCmd(infoid.New(), "Connected to Redis successfully.", expiration),
		doTick(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	util.LogMsg("Update()", msg)

	// Update app state
	m.State, cmd = m.State.Update(msg)
	cmds = append(cmds, cmd)

	// Update info box
	m.infoBox, cmd = m.infoBox.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg, m.State)
	cmds = append(cmds, cmd)

	// Update key list
	m.keyList, cmd = m.keyList.Update(m.ctx, m.redis, msg, m.State)
	cmds = append(cmds, cmd)
	for _, c := range cmds {
		if c != nil {
			log.Printf("After keyList.Update(): %v", c)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Printf("KEY HIT: \"%s\"", msg.String())
		m, _cmds := m.updateWithKey(msg.String())
		cmds = append(cmds, _cmds...)
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg: // Handle window resize
		log.Printf("Received window size message: width=%d, height=%d", msg.Width, msg.Height)
		return m.UpdateWindowSize(msg.Height, msg.Width), tea.Batch(cmds...)

	case command.TickMsg:
		log.Print("Received tick message")
		cmds = append(cmds, doTick())
		if m.keyList.IsBeingUnfiltered() {
			cmds = append(cmds, command.GetKeys(m.ctx, m.redis, ""))
		}
		return m, tea.Batch(cmds...)

	case command.NewRedisClientMsg:
		log.Print("Received new Redis client message")
		m = m.UpdateRedisClient(msg)
		cmds = append(cmds, command.GetKeys(m.ctx, m.redis, "")) // Re-fetch keys with the new client
		return m, tea.Batch(cmds...)

	case command.HighlightedKeyUpdatedMsg:
		log.Printf("Highlighted key updated to: %s", msg.Key)
		return m, command.GetValue(m.ctx, m.redis, msg.Key)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateWithKey(key string) (Model, []tea.Cmd) {
	var cmds []tea.Cmd
	switch key {
	case tea.KeyEsc.String(), tea.KeyCtrlC.String(), "q":
		if m.keyList.IsFitering() {
			return m, []tea.Cmd{tea.Quit}
		}

	case tea.KeyTab.String():
		m = m.NextTab()
		cmds = append(cmds, command.SwitchTab(m.ctx, m.redis, m.currentTab))
		return m, cmds

	case tea.KeyShiftTab.String():
		m = m.PreviousTab()
		cmds = append(cmds, command.SwitchTab(m.ctx, m.redis, m.currentTab))
		return m, cmds
	}

	return m, cmds
}
