// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	apiVersion    = "v1.43"
	defaultSocket = "/var/run/docker.sock"
)

// Client communicates with the Docker daemon over the unix socket.
// No Docker SDK dependency — raw HTTP over unix socket keeps the binary small.
type Client struct {
	httpClient *http.Client
	baseURL    string
	socketPath string
}

// NewClient creates a new Docker API client using the unix socket.
func NewClient() *Client {
	socketPath := defaultSocket
	if runtime.GOOS == "darwin" {
		// Docker Desktop on macOS sometimes uses this path
		// but /var/run/docker.sock is symlinked, so it works either way.
		socketPath = defaultSocket
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext(ctx, "unix", socketPath)
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		MaxIdleConnsPerHost: 10,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
		baseURL:    fmt.Sprintf("http://localhost/%s", apiVersion),
		socketPath: socketPath,
	}
}

// Ping checks if the Docker daemon is accessible.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/_ping", nil)
	if err != nil {
		return fmt.Errorf("creating ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("docker daemon not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("docker daemon returned status %d", resp.StatusCode)
	}

	return nil
}

// ListContainers returns all containers (running, paused, exited).
func (c *Client) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	url := fmt.Sprintf("%s/containers/json?all=true", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating list request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list containers returned %d: %s", resp.StatusCode, string(body))
	}

	var containers []ContainerInfo
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("decoding container list: %w", err)
	}

	return containers, nil
}

// GetStats fetches a one-shot stats snapshot for a container.
func (c *Client) GetStats(ctx context.Context, containerID string) (*StatsResponse, error) {
	url := fmt.Sprintf("%s/containers/%s/stats?stream=false&one-shot=true", c.baseURL, containerID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating stats request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching stats for %s: %w", containerID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stats returned %d: %s", resp.StatusCode, string(body))
	}

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decoding stats for %s: %w", containerID, err)
	}

	return &stats, nil
}

// Close cleans up the HTTP client transport.
func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}
