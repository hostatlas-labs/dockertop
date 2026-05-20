// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/hostatlas-labs/dockertop/internal/docker"
	"github.com/hostatlas-labs/dockertop/internal/formatter"
)

// View renders the TUI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header.
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Error message.
	if m.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Content.
	filtered := m.filteredContainers()
	docker.SortContainers(filtered, m.config.SortMode)

	if len(filtered) == 0 {
		if len(m.containers) == 0 {
			b.WriteString(EmptyStyle.Render("No containers found. Is Docker running?"))
		} else {
			b.WriteString(EmptyStyle.Render("No containers match the filter."))
		}
		b.WriteString("\n")
	} else if m.config.Grouped {
		b.WriteString(m.renderGrouped(filtered))
	} else {
		b.WriteString(m.renderFlat(filtered))
	}

	// Filter bar (replaces footer when active).
	if m.filterMode {
		b.WriteString(m.renderFilterBar())
	} else {
		b.WriteString(m.renderFooter())
	}

	// Frame the entire output.
	content := b.String()

	return content
}

// renderHeader renders the top status bar.
func (m Model) renderHeader() string {
	left := HeaderStyle.Render(fmt.Sprintf("  dockertop %s",
		HeaderVersionStyle.Render("v"+m.config.Version)))

	totalCPU := formatter.CPU(m.totalCPU())
	totalMem := formatter.Bytes(m.totalMemory())
	containerCount := len(m.containers)

	right := HeaderStatsStyle.Render(fmt.Sprintf(
		"Containers: %d  CPU: %s  MEM: %s  ",
		containerCount, totalCPU, totalMem))

	// Calculate gap to fill the header bar.
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := m.width - leftWidth - rightWidth
	if gap < 1 {
		gap = 1
	}

	headerBg := lipgloss.NewStyle().Background(ColorPrimary)

	return left + headerBg.Render(strings.Repeat(" ", gap)) + HeaderStyle.Render(right)
}

// renderGrouped renders containers grouped by compose project.
func (m Model) renderGrouped(containers []*docker.ContainerStats) string {
	groups := docker.GroupByCompose(containers)
	groupNames := docker.SortedGroupNames(groups)

	var lines []string

	for _, groupName := range groupNames {
		members := groups[groupName]
		docker.SortContainers(members, m.config.SortMode)

		collapsed := m.collapsedGroups[groupName]
		arrow := "▸"
		if !collapsed {
			arrow = "▾"
		}

		// Group header.
		header := fmt.Sprintf("  %s %s %s",
			GroupHeaderStyle.Render(arrow+" "+groupName),
			GroupCountStyle.Render(fmt.Sprintf("(%d containers)", len(members))),
			"",
		)
		lines = append(lines, header)

		if !collapsed {
			// Table header.
			lines = append(lines, m.renderTableHeader())

			// Container rows.
			for _, cs := range members {
				lines = append(lines, m.renderContainerRow(cs))
			}
		}

		lines = append(lines, "") // blank separator
	}

	// Apply scrolling.
	return m.applyScroll(lines)
}

// renderFlat renders all containers in a single flat list.
func (m Model) renderFlat(containers []*docker.ContainerStats) string {
	var lines []string

	lines = append(lines, m.renderTableHeader())

	for _, cs := range containers {
		lines = append(lines, m.renderContainerRow(cs))
	}

	return m.applyScroll(lines)
}

// renderTableHeader renders the column headers for the container table.
func (m Model) renderTableHeader() string {
	nameW, stateW, cpuW, memW, netW := m.columnWidths()

	name := padRight("  Container", nameW)
	state := padRight("State", stateW)
	cpu := padRight("CPU", cpuW)
	mem := padRight("Memory", memW)
	net := padRight("Net I/O", netW)

	row := fmt.Sprintf("%s %s %s %s %s",
		name, state, cpu, mem, net)

	return TableHeaderStyle.Render(row)
}

// renderContainerRow renders a single container's stats row.
func (m Model) renderContainerRow(cs *docker.ContainerStats) string {
	nameW, stateW, cpuW, memW, netW := m.columnWidths()

	// Container name.
	name := ContainerNameStyle.Render(padRight("  "+truncate(cs.Name, nameW-2), nameW))

	// State.
	state := padRight(StatusLabel(cs.State), stateW+3) // +3 for ANSI color codes in dot

	// CPU sparkline + percentage.
	sparkWidth := 8
	spark := SparklineStyle.Render(Sparkline(cs.CPUHistory, sparkWidth))
	cpuPct := CPUStyle.Render(fmt.Sprintf("%5s", formatter.CPU(cs.CPUPercent)))
	cpuCol := padRight(spark+" "+cpuPct, cpuW+12) // extra for ANSI codes

	// Memory bar + usage.
	barWidth := 8
	bar, color := formatter.MemoryBar(cs.MemoryUsage, cs.MemoryLimit, barWidth)

	var memBar string
	if bar != "" {
		// Split bar into filled and empty portions.
		filled := ""
		empty := ""
		for _, r := range bar {
			if r == '█' {
				filled += string(r)
			} else {
				empty += string(r)
			}
		}
		memBar = MemBarColor(color).Render(filled) + MemBarEmpty.Render(empty)
	}

	memUsage := formatter.Bytes(cs.MemoryUsage)
	memLimit := formatter.Bytes(cs.MemoryLimit)
	memText := fmt.Sprintf("%s/%s", memUsage, memLimit)
	memCol := fmt.Sprintf("%s %s", memBar, memText)
	// Don't pad memCol with padRight since it has ANSI codes; just leave it.

	// Network I/O.
	netUp := NetUpStyle.Render("↑" + formatter.Bytes(cs.NetTx))
	netDn := NetDnStyle.Render("↓" + formatter.Bytes(cs.NetRx))
	netCol := padRight(netUp+" "+netDn, netW+12) // extra for ANSI codes

	_ = memW
	_ = netW

	return fmt.Sprintf("%s %s %s %s %s",
		name, state, cpuCol, memCol, netCol)
}

// columnWidths calculates column widths based on terminal width.
func (m Model) columnWidths() (name, state, cpu, mem, net int) {
	w := m.width
	if w < 80 {
		w = 80
	}

	// Fixed-ish proportions.
	name = w * 22 / 100
	if name < 16 {
		name = 16
	}
	state = 10
	cpu = 18
	mem = 22
	net = 16

	// Reclaim leftover space for the name column.
	used := name + state + cpu + mem + net + 5 // 5 for separators
	if used < w {
		name += w - used
	}

	return
}

// renderFooter renders the bottom help bar.
func (m Model) renderFooter() string {
	groupStatus := "on"
	if !m.config.Grouped {
		groupStatus = "off"
	}

	filter := ""
	if m.filterText != "" {
		filter = fmt.Sprintf("  filter: %s", FilterInputStyle.Render(m.filterText))
	}

	items := []struct{ key, desc string }{
		{"q", "quit"},
		{"s", "sort: " + m.config.SortMode.String()},
		{"g", "group: " + groupStatus},
		{"/", "filter"},
		{"↑↓", "scroll"},
	}

	var parts []string
	for _, item := range items {
		parts = append(parts,
			FooterKeyStyle.Render("["+item.key+"]")+FooterDescStyle.Render(item.desc))
	}

	help := "  " + strings.Join(parts, "  ")
	return "\n" + FooterStyle.Render(help+filter)
}

// renderFilterBar renders the filter input bar.
func (m Model) renderFilterBar() string {
	prompt := FilterPromptStyle.Render("  / Filter: ")
	input := FilterInputStyle.Render(m.filterText)
	cursor := FilterPromptStyle.Render("█")
	return "\n" + prompt + input + cursor
}

// applyScroll applies vertical scrolling to a slice of content lines.
func (m Model) applyScroll(lines []string) string {
	// Calculate viewable area (total height minus header and footer).
	viewable := m.height - 5
	if viewable < 1 {
		viewable = 1
	}

	// Clamp scroll offset.
	maxScroll := len(lines) - viewable
	if maxScroll < 0 {
		maxScroll = 0
	}
	offset := m.scrollOffset
	if offset > maxScroll {
		offset = maxScroll
	}
	if offset < 0 {
		offset = 0
	}

	// Slice visible lines.
	end := offset + viewable
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[offset:end]
	return strings.Join(visible, "\n") + "\n"
}

// padRight pads a string to width with spaces. If s is longer, it is not truncated
// (use truncate separately before calling this if needed).
func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// truncate truncates a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 3 {
		maxLen = 4
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
