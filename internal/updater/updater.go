// SPDX-License-Identifier: MIT
// © 2026 HostAtlas Technologies LLC
// hello@hostatlas.app

package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner  = "akyroslabs"
	repoName   = "hostatlas-dockertop"
	binaryName = "dockertop"
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckForUpdate checks GitHub releases for a newer version.
// Returns the latest version string and download URL if an update is available.
func CheckForUpdate(currentVersion string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("decoding release info: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")

	if latestVersion == currentClean || latestVersion == "" {
		return "", "", nil // no update available
	}

	// Find the asset for this OS/arch.
	assetName := fmt.Sprintf("%s_%s_%s_%s", binaryName, latestVersion, runtime.GOOS, runtime.GOARCH)
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	targetAsset := assetName + ext

	for _, asset := range release.Assets {
		if asset.Name == targetAsset {
			return latestVersion, asset.BrowserDownloadURL, nil
		}
	}

	return latestVersion, "", fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// SelfUpdate downloads and replaces the current binary with the latest release.
func SelfUpdate(currentVersion string) error {
	latestVersion, downloadURL, err := CheckForUpdate(currentVersion)
	if err != nil {
		return err
	}

	if latestVersion == "" {
		fmt.Println("Already up to date.")
		return nil
	}

	if downloadURL == "" {
		return fmt.Errorf("update to %s available but no download found for %s/%s", latestVersion, runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Updating from %s to %s...\n", currentVersion, latestVersion)
	fmt.Printf("Download URL: %s\n", downloadURL)

	// Download the release archive.
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Get the current executable path.
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable path: %w", err)
	}

	fmt.Printf("To complete the update, download and replace: %s\n", execPath)
	fmt.Printf("Download: %s\n", downloadURL)

	return nil
}
