// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package tui

// sparkBlocks are Unicode block elements for sparkline rendering, ordered
// from lowest to highest.
var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders a sparkline string from a series of values.
// width limits the number of characters; values are scaled to fit the block range.
func Sparkline(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	// Take only the last `width` values.
	start := 0
	if len(values) > width {
		start = len(values) - width
	}
	data := values[start:]

	// Find max for scaling. Use 100 as a reasonable ceiling for CPU percentages,
	// but allow higher values to scale properly.
	maxVal := 100.0
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal <= 0 {
		maxVal = 1
	}

	result := make([]rune, len(data))
	for i, v := range data {
		if v < 0 {
			v = 0
		}
		// Normalize to 0.0 - 1.0.
		normalized := v / maxVal
		if normalized > 1.0 {
			normalized = 1.0
		}

		// Map to block index (0-7).
		idx := int(normalized * float64(len(sparkBlocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkBlocks) {
			idx = len(sparkBlocks) - 1
		}
		result[i] = sparkBlocks[idx]
	}

	// Pad with spaces if fewer values than width.
	padding := ""
	for i := len(result); i < width; i++ {
		padding += " "
	}

	return padding + string(result)
}
