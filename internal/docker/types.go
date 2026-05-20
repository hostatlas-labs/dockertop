// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package docker

import "time"

// ContainerInfo represents a running Docker container with its metadata.
type ContainerInfo struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	State   string            `json:"State"`
	Status  string            `json:"Status"`
	Labels  map[string]string `json:"Labels"`
	Created int64             `json:"Created"`
}

// ContainerListResponse is the raw response from /containers/json.
type ContainerListResponse = []ContainerInfo

// StatsResponse represents the raw stats JSON from Docker API.
type StatsResponse struct {
	Read     time.Time `json:"read"`
	PreRead  time.Time `json:"preread"`
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
		Stats struct {
			InactiveFile uint64 `json:"inactive_file"`
			Cache        uint64 `json:"cache"`
		} `json:"stats"`
	} `json:"memory_stats"`
	Networks map[string]NetworkStats `json:"networks"`
}

// NetworkStats represents network I/O for a single interface.
type NetworkStats struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

// ContainerStats holds the computed stats for a single container.
type ContainerStats struct {
	ID             string
	Name           string
	Image          string
	State          string
	Status         string
	ComposeProject string
	CPUPercent     float64
	MemoryUsage    uint64
	MemoryLimit    uint64
	MemoryPercent  float64
	NetRx          uint64
	NetTx          uint64
	CPUHistory     []float64
	LastUpdated    time.Time
}

// HistorySize is the number of CPU readings to keep per container for sparklines.
const HistorySize = 30

// SortMode defines how containers are sorted.
type SortMode int

const (
	SortByCPU SortMode = iota
	SortByMem
	SortByName
	SortByNet
)

// SortModeNames maps sort modes to their display names.
var SortModeNames = map[SortMode]string{
	SortByCPU:  "cpu",
	SortByMem:  "mem",
	SortByName: "name",
	SortByNet:  "net",
}

// ParseSortMode converts a string to a SortMode.
func ParseSortMode(s string) SortMode {
	switch s {
	case "mem", "memory":
		return SortByMem
	case "name":
		return SortByName
	case "net", "network":
		return SortByNet
	default:
		return SortByCPU
	}
}

// NextSortMode cycles to the next sort mode.
func (s SortMode) Next() SortMode {
	return (s + 1) % 4
}

// String returns the display name.
func (s SortMode) String() string {
	if name, ok := SortModeNames[s]; ok {
		return name
	}
	return "cpu"
}
