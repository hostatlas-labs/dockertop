// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hostatlas-labs/dockertop/internal/docker"
)

// Update handles messages for the Bubbletea model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, tea.Batch(
			m.fetchStats(),
			m.tick(),
		)

	case statsMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.containers = msg.containers
		}
		return m, nil
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In filter mode, handle typing.
	if m.filterMode {
		return m.handleFilterKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		m.client.Close()
		return m, tea.Quit

	case "s":
		m.config.SortMode = m.config.SortMode.Next()
		return m, nil

	case "g":
		m.config.Grouped = !m.config.Grouped
		m.scrollOffset = 0
		return m, nil

	case "/":
		m.filterMode = true
		m.filterText = ""
		return m, nil

	case "up", "k":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil

	case "down", "j":
		m.scrollOffset++
		return m, nil

	case "pgup":
		m.scrollOffset -= 10
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		return m, nil

	case "pgdown":
		m.scrollOffset += 10
		return m, nil

	case "home":
		m.scrollOffset = 0
		return m, nil

	case "end":
		m.scrollOffset = 9999
		return m, nil

	case "enter":
		// Toggle collapse of the group under the cursor.
		// Find which group is at the current scroll position.
		if m.config.Grouped {
			groups := m.getVisibleGroups()
			lineCount := 0
			for _, gn := range groups {
				if lineCount == m.scrollOffset {
					m.collapsedGroups[gn] = !m.collapsedGroups[gn]
					break
				}
				lineCount++ // group header
				if !m.collapsedGroups[gn] {
					lineCount++ // table header
					filtered := m.filteredContainers()
					grouped := docker.GroupByCompose(filtered)
					lineCount += len(grouped[gn])
				}
				lineCount++ // blank line
			}
		}
		return m, nil
	}

	return m, nil
}

// handleFilterKey processes keyboard input while in filter mode.
func (m Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.filterMode = false
		if msg.String() == "esc" {
			m.filterText = ""
		}
		return m, nil

	case "backspace":
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
		}
		return m, nil

	case "ctrl+c":
		m.quitting = true
		m.client.Close()
		return m, tea.Quit

	default:
		// Append printable characters.
		if len(msg.String()) == 1 {
			m.filterText += msg.String()
		}
		return m, nil
	}
}

// getVisibleGroups returns the sorted group names for the current filtered containers.
func (m Model) getVisibleGroups() []string {
	filtered := m.filteredContainers()
	docker.SortContainers(filtered, m.config.SortMode)
	groups := docker.GroupByCompose(filtered)
	return docker.SortedGroupNames(groups)
}
