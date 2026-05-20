// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hostatlas-labs/dockertop/internal/docker"
)

// Config holds TUI configuration options.
type Config struct {
	SortMode docker.SortMode
	Grouped  bool
	Interval time.Duration
	Version  string
}

// Model is the Bubbletea model for the dockertop TUI.
type Model struct {
	// Configuration.
	config Config

	// Docker client and collector.
	client    *docker.Client
	collector *docker.Collector

	// Current container stats.
	containers []*docker.ContainerStats

	// UI state.
	width           int
	height          int
	scrollOffset    int
	filterMode      bool
	filterText      string
	collapsedGroups map[string]bool
	err             error
	quitting        bool

	// Timing.
	lastUpdate time.Time
}

// tickMsg triggers a stats refresh.
type tickMsg struct{}

// statsMsg carries updated container stats.
type statsMsg struct {
	containers []*docker.ContainerStats
	err        error
}

// NewModel creates a new TUI model.
func NewModel(cfg Config) Model {
	client := docker.NewClient()
	collector := docker.NewCollector(client)

	return Model{
		config:          cfg,
		client:          client,
		collector:       collector,
		collapsedGroups: make(map[string]bool),
	}
}

// Init initializes the Bubbletea model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchStats(),
		m.tick(),
	)
}

// tick returns a command that sends a tickMsg after the configured interval.
func (m Model) tick() tea.Cmd {
	return tea.Tick(m.config.Interval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// fetchStats returns a command that collects Docker stats.
func (m Model) fetchStats() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		stats, err := m.collector.Collect(ctx)
		if err != nil {
			return statsMsg{err: err}
		}

		containers := make([]*docker.ContainerStats, 0, len(stats))
		for _, cs := range stats {
			containers = append(containers, cs)
		}

		return statsMsg{containers: containers}
	}
}

// totalCPU returns the sum of CPU percentages across all containers.
func (m Model) totalCPU() float64 {
	total := 0.0
	for _, c := range m.containers {
		total += c.CPUPercent
	}
	return total
}

// totalMemory returns the sum of memory usage across all containers.
func (m Model) totalMemory() uint64 {
	var total uint64
	for _, c := range m.containers {
		total += c.MemoryUsage
	}
	return total
}

// filteredContainers returns containers matching the current filter.
func (m Model) filteredContainers() []*docker.ContainerStats {
	if m.filterText == "" {
		return m.containers
	}

	var filtered []*docker.ContainerStats
	for _, c := range m.containers {
		if containsCI(c.Name, m.filterText) || containsCI(c.Image, m.filterText) || containsCI(c.ComposeProject, m.filterText) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// containsCI performs a case-insensitive substring check.
func containsCI(s, substr string) bool {
	sLower := toLower(s)
	subLower := toLower(substr)
	return len(subLower) == 0 || contains(sLower, subLower)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// maxScroll calculates the maximum scroll offset based on visible content lines.
func (m Model) maxScroll(totalLines int) int {
	// Reserve lines for header (3) + footer (2).
	viewable := m.height - 5
	if viewable < 1 {
		viewable = 1
	}
	max := totalLines - viewable
	if max < 0 {
		max = 0
	}
	return max
}
