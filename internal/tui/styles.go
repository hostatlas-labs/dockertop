// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package tui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#A78BFA") // Light purple
	ColorAccent    = lipgloss.Color("#06B6D4") // Cyan
	ColorGreen     = lipgloss.Color("#10B981") // Emerald
	ColorYellow    = lipgloss.Color("#F59E0B") // Amber
	ColorRed       = lipgloss.Color("#EF4444") // Red
	ColorBlue      = lipgloss.Color("#3B82F6") // Blue
	ColorDim       = lipgloss.Color("#6B7280") // Gray-500
	ColorMuted     = lipgloss.Color("#9CA3AF") // Gray-400
	ColorText      = lipgloss.Color("#F9FAFB") // Gray-50
	ColorSubtle    = lipgloss.Color("#374151") // Gray-700
	ColorBorder    = lipgloss.Color("#4B5563") // Gray-600
	ColorBg        = lipgloss.Color("#111827") // Gray-900
)

// Styles used throughout the TUI.
var (
	// Header bar.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 1)

	HeaderVersionStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	HeaderStatsStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Bold(true)

	// Group header (compose project name).
	GroupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent).
				Padding(0, 0, 0, 1)

	GroupCountStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Table styles.
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorMuted).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorSubtle)

	ContainerNameStyle = lipgloss.NewStyle().
				Foreground(ColorText)

	// Status dot styles.
	StatusRunning    = lipgloss.NewStyle().Foreground(ColorGreen)
	StatusPaused     = lipgloss.NewStyle().Foreground(ColorYellow)
	StatusExited     = lipgloss.NewStyle().Foreground(ColorRed)
	StatusRestarting = lipgloss.NewStyle().Foreground(ColorBlue)
	StatusOther      = lipgloss.NewStyle().Foreground(ColorDim)

	// Memory bar colors.
	MemBarGreen  = lipgloss.NewStyle().Foreground(ColorGreen)
	MemBarYellow = lipgloss.NewStyle().Foreground(ColorYellow)
	MemBarRed    = lipgloss.NewStyle().Foreground(ColorRed)
	MemBarEmpty  = lipgloss.NewStyle().Foreground(ColorSubtle)

	// Sparkline.
	SparklineStyle = lipgloss.NewStyle().Foreground(ColorAccent)

	// CPU percentage.
	CPUStyle = lipgloss.NewStyle().Foreground(ColorText)

	// Network I/O.
	NetStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	NetUpStyle = lipgloss.NewStyle().Foreground(ColorGreen)
	NetDnStyle = lipgloss.NewStyle().Foreground(ColorBlue)

	// Footer / help bar.
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(0, 1)

	FooterKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	FooterDescStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Filter input.
	FilterPromptStyle = lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true)

	FilterInputStyle = lipgloss.NewStyle().
				Foreground(ColorText)

	// Border for the outer frame.
	OuterBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder)

	// "No containers" message.
	EmptyStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true).
			Padding(1, 2)

	// Error message.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true).
			Padding(1, 2)
)

// StatusDot returns a colored dot based on container state.
func StatusDot(state string) string {
	switch state {
	case "running":
		return StatusRunning.Render("●")
	case "paused":
		return StatusPaused.Render("●")
	case "exited", "dead":
		return StatusExited.Render("●")
	case "restarting":
		return StatusRestarting.Render("●")
	default:
		return StatusOther.Render("●")
	}
}

// StatusLabel returns a colored short label for the state.
func StatusLabel(state string) string {
	dot := StatusDot(state)
	switch state {
	case "running":
		return dot + " Up"
	case "paused":
		return dot + " Paused"
	case "exited":
		return dot + " Exit"
	case "dead":
		return dot + " Dead"
	case "restarting":
		return dot + " Restart"
	case "created":
		return dot + " Created"
	default:
		return dot + " " + state
	}
}

// MemBarColor returns the lipgloss style for a memory bar color name.
func MemBarColor(color string) lipgloss.Style {
	switch color {
	case "yellow":
		return MemBarYellow
	case "red":
		return MemBarRed
	default:
		return MemBarGreen
	}
}
