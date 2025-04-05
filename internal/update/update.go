package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

const (
	githubAPI = "https://api.github.com/repos/mguptahub/nanoHttp/releases/latest"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// CheckForUpdates checks if a new version is available
func CheckForUpdates(currentVersion string) (*Release, error) {
	resp, err := http.Get(githubAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check for updates: status code %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release info: %v", err)
	}

	if release.TagName == currentVersion {
		return nil, nil
	}

	return &release, nil
}

// DownloadUpdate downloads the latest version
func DownloadUpdate(release *Release) error {
	// Find the appropriate asset for the current OS and architecture
	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == fmt.Sprintf("nanoHttp-%s-%s", runtime.GOOS, runtime.GOARCH) {
			assetURL = asset.DownloadURL
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no suitable binary found for %s-%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the new binary
	resp, err := http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "nanoHttp-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the new binary to the temporary file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write update file: %v", err)
	}

	// Get the path of the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Create a backup of the current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Move the new binary into place
	if err := os.Rename(tmpFile.Name(), execPath); err != nil {
		// Restore the backup if the move fails
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install update: %v", err)
	}

	// Remove the backup
	os.Remove(backupPath)

	return nil
}

// RollbackUpdate rolls back to the previous version
func RollbackUpdate() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	backupPath := execPath + ".backup"
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("no backup found to rollback to: %v", err)
	}

	// Remove the current binary
	if err := os.Remove(execPath); err != nil {
		return fmt.Errorf("failed to remove current binary: %v", err)
	}

	// Restore the backup
	if err := os.Rename(backupPath, execPath); err != nil {
		return fmt.Errorf("failed to restore backup: %v", err)
	}

	return nil
}
