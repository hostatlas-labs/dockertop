// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/hostatlas-labs/dockertop/internal/docker"
	"github.com/hostatlas-labs/dockertop/internal/tui"
	"github.com/hostatlas-labs/dockertop/internal/updater"
)

// Build-time variables injected via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const banner = `  _   _           _      _   _   _
 | | | | ___  ___| |_   / \ | |_| | __ _ ___
 | |_| |/ _ \/ __| __| / _ \| __| |/ _` + "`" + ` / __|
 |  _  | (_) \__ \ |_ / ___ \ |_| | (_| \__ \
 |_| |_|\___/|___/\__/_/   \_\__|_|\__,_|___/
`

func main() {
	var (
		sortFlag     string
		groupFlag    bool
		noGroupFlag  bool
		intervalFlag int
		updateFlag   bool
	)

	rootCmd := &cobra.Command{
		Use:   "dockertop",
		Short: "A beautiful Docker stats terminal dashboard",
		Long: banner + `
dockertop — a prettier docker stats in the terminal.

Features:
  - Real-time CPU sparklines and memory progress bars
  - Container grouping by Docker Compose project
  - Color-coded status indicators
  - Keyboard-driven sorting and filtering

MIT licensed · © 2026 HostAtlas Technologies LLC
    hello@hostatlas.app`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle --update flag.
			if updateFlag {
				return updater.SelfUpdate(version)
			}

			// Resolve grouping (--group is default true, --no-group overrides).
			grouped := groupFlag
			if noGroupFlag {
				grouped = false
			}

			// Parse interval.
			interval := time.Duration(intervalFlag) * time.Second
			if interval < 1*time.Second {
				interval = 1 * time.Second
			}
			if interval > 60*time.Second {
				interval = 60 * time.Second
			}

			// Check Docker connectivity before launching TUI.
			client := docker.NewClient()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := client.Ping(ctx); err != nil {
				fmt.Print(banner)
				fmt.Println()
				fmt.Println("  Error: Could not connect to Docker daemon.")
				fmt.Println()
				fmt.Println("  Make sure Docker is running and accessible at /var/run/docker.sock")
				fmt.Println()
				fmt.Printf("  Details: %v\n", err)
				fmt.Println()
				os.Exit(1)
			}
			client.Close()

			// Launch TUI.
			cfg := tui.Config{
				SortMode: docker.ParseSortMode(sortFlag),
				Grouped:  grouped,
				Interval: interval,
				Version:  version,
			}

			model := tui.NewModel(cfg)
			p := tea.NewProgram(model, tea.WithAltScreen())

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("running TUI: %w", err)
			}

			return nil
		},
	}

	// Flags.
	rootCmd.Flags().StringVar(&sortFlag, "sort", "cpu", "Sort containers by: cpu, mem, name, net")
	rootCmd.Flags().BoolVar(&groupFlag, "group", true, "Group containers by Compose project")
	rootCmd.Flags().BoolVar(&noGroupFlag, "no-group", false, "Show flat list without grouping")
	rootCmd.Flags().IntVar(&intervalFlag, "interval", 2, "Refresh interval in seconds")
	rootCmd.Flags().BoolVar(&updateFlag, "update", false, "Check for and install updates")

	// Override the default version template to include the banner.
	rootCmd.SetVersionTemplate(fmt.Sprintf("%sdockertop version {{.Version}}\n\n", banner))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
