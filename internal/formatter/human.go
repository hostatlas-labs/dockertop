// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package formatter

import "fmt"

// Bytes formats a byte count into a human-readable string (e.g., 1.2G, 256M, 45K).
func Bytes(b uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case b >= TB:
		return fmt.Sprintf("%.1fT", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.1fG", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.0fM", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.0fK", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// BytesPrecise formats bytes with one decimal for values >= 1GB.
func BytesPrecise(b uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case b >= TB:
		return fmt.Sprintf("%.1fT", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.1fG", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%dM", b/MB)
	case b >= KB:
		return fmt.Sprintf("%dK", b/KB)
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// CPU formats a CPU percentage with one decimal place.
func CPU(percent float64) string {
	if percent < 0.1 {
		return "0.0%"
	}
	return fmt.Sprintf("%.1f%%", percent)
}

// NetIO formats network I/O with directional arrows.
func NetIO(tx, rx uint64) string {
	return fmt.Sprintf("↑%s ↓%s", Bytes(tx), Bytes(rx))
}

// MemoryBar renders a progress bar for memory usage.
// width is the total number of bar characters.
// Returns the bar string and a color indicator:
//
//	"green" (<70%), "yellow" (70-90%), "red" (>90%)
func MemoryBar(used, limit uint64, width int) (string, string) {
	if limit == 0 || width <= 0 {
		return "", "green"
	}

	percent := float64(used) / float64(limit)
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	color := "green"
	if percent >= 0.9 {
		color = "red"
	} else if percent >= 0.7 {
		color = "yellow"
	}

	return bar, color
}
