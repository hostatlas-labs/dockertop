// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package docker

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

// Collector periodically fetches container stats from the Docker API.
type Collector struct {
	client  *Client
	mu      sync.RWMutex
	stats   map[string]*ContainerStats
	history map[string][]float64
}

// NewCollector creates a new stats collector.
func NewCollector(client *Client) *Collector {
	return &Collector{
		client:  client,
		stats:   make(map[string]*ContainerStats),
		history: make(map[string][]float64),
	}
}

// Collect performs a single collection cycle: lists containers, fetches stats.
func (c *Collector) Collect(ctx context.Context) (map[string]*ContainerStats, error) {
	containers, err := c.client.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make(map[string]*ContainerStats, len(containers))
	)

	// Limit concurrency to avoid overwhelming the Docker daemon.
	sem := make(chan struct{}, 5)

	for _, container := range containers {
		wg.Add(1)
		go func(ci ContainerInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			cs := &ContainerStats{
				ID:             ci.ID,
				Name:           cleanContainerName(ci.Names),
				Image:          ci.Image,
				State:          ci.State,
				Status:         ci.Status,
				ComposeProject: ci.Labels["com.docker.compose.project"],
				LastUpdated:    time.Now(),
			}

			// Only fetch stats for running containers.
			if ci.State == "running" {
				statsCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				raw, err := c.client.GetStats(statsCtx, ci.ID)
				if err == nil {
					cs.CPUPercent = calculateCPUPercent(raw)
					cs.MemoryUsage = calculateMemUsage(raw)
					cs.MemoryLimit = raw.MemoryStats.Limit
					if cs.MemoryLimit > 0 {
						cs.MemoryPercent = float64(cs.MemoryUsage) / float64(cs.MemoryLimit) * 100
					}
					cs.NetRx, cs.NetTx = calculateNetIO(raw)
				}
			}

			mu.Lock()
			results[ci.ID] = cs
			mu.Unlock()
		}(container)
	}

	wg.Wait()

	// Update history.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove containers that no longer exist.
	activeIDs := make(map[string]bool, len(results))
	for id := range results {
		activeIDs[id] = true
	}
	for id := range c.history {
		if !activeIDs[id] {
			delete(c.history, id)
		}
	}

	// Append CPU history.
	for id, cs := range results {
		hist := c.history[id]
		hist = append(hist, cs.CPUPercent)
		if len(hist) > HistorySize {
			hist = hist[len(hist)-HistorySize:]
		}
		c.history[id] = hist
		cs.CPUHistory = make([]float64, len(hist))
		copy(cs.CPUHistory, hist)
	}

	c.stats = results
	return results, nil
}

// calculateCPUPercent computes CPU usage percentage from raw stats.
func calculateCPUPercent(stats *StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)

	if systemDelta <= 0 || cpuDelta <= 0 {
		return 0.0
	}

	cpus := stats.CPUStats.OnlineCPUs
	if cpus == 0 {
		cpus = 1
	}

	return (cpuDelta / systemDelta) * float64(cpus) * 100.0
}

// calculateMemUsage computes actual memory usage, subtracting cache on Linux.
func calculateMemUsage(stats *StatsResponse) uint64 {
	usage := stats.MemoryStats.Usage
	// On cgroup v1, subtract inactive_file; on v2, subtract cache.
	if stats.MemoryStats.Stats.InactiveFile > 0 {
		if usage > stats.MemoryStats.Stats.InactiveFile {
			usage -= stats.MemoryStats.Stats.InactiveFile
		}
	} else if stats.MemoryStats.Stats.Cache > 0 {
		if usage > stats.MemoryStats.Stats.Cache {
			usage -= stats.MemoryStats.Stats.Cache
		}
	}
	return usage
}

// calculateNetIO sums network rx/tx across all interfaces.
func calculateNetIO(stats *StatsResponse) (rx, tx uint64) {
	for _, net := range stats.Networks {
		rx += net.RxBytes
		tx += net.TxBytes
	}
	return rx, tx
}

// cleanContainerName strips the leading slash from container names.
func cleanContainerName(names []string) string {
	if len(names) == 0 {
		return "unknown"
	}
	name := names[0]
	return strings.TrimPrefix(name, "/")
}

// SortContainers sorts a slice of ContainerStats by the given mode.
func SortContainers(containers []*ContainerStats, mode SortMode) {
	sort.SliceStable(containers, func(i, j int) bool {
		switch mode {
		case SortByMem:
			return containers[i].MemoryUsage > containers[j].MemoryUsage
		case SortByName:
			return containers[i].Name < containers[j].Name
		case SortByNet:
			iNet := containers[i].NetRx + containers[i].NetTx
			jNet := containers[j].NetRx + containers[j].NetTx
			return iNet > jNet
		default: // SortByCPU
			return containers[i].CPUPercent > containers[j].CPUPercent
		}
	})
}

// GroupByCompose groups containers by their Compose project label.
// Containers without a Compose project are grouped under "standalone".
func GroupByCompose(containers []*ContainerStats) map[string][]*ContainerStats {
	groups := make(map[string][]*ContainerStats)
	for _, cs := range containers {
		project := cs.ComposeProject
		if project == "" {
			project = "standalone"
		}
		groups[project] = append(groups[project], cs)
	}
	return groups
}

// SortedGroupNames returns group names sorted alphabetically, with "standalone" last.
func SortedGroupNames(groups map[string][]*ContainerStats) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.SliceStable(names, func(i, j int) bool {
		if names[i] == "standalone" {
			return false
		}
		if names[j] == "standalone" {
			return true
		}
		return names[i] < names[j]
	})
	return names
}
