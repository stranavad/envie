package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	githubRepo      = "stranavad/envie"
	githubAPILatest = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Envie CLI to the latest version",
	Long: `Check for updates and install the latest version of Envie CLI.

This command will:
1. Check the latest version on GitHub
2. Download the new binary if an update is available
3. Replace the current binary

Examples:
  envie update          # Update to latest
  envie update --check  # Just check for updates`,
	RunE: runUpdate,
}

var (
	updateCheck bool
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Only check for updates, don't install")
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Printf("Current version: %s\n", version)
	fmt.Println("Checking for updates...")

	// Fetch latest release info
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Parse version from tag (remove "cli-" prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "cli-")

	// Compare versions
	if latestVersion == version {
		fmt.Println("You're already on the latest version!")
		return nil
	}

	fmt.Printf("New version available: %s\n", latestVersion)
	fmt.Printf("Release URL: %s\n", release.HTMLURL)

	if updateCheck {
		fmt.Println("\nRun 'envie update' to install the update.")
		return nil
	}

	// Find the right asset for this platform
	assetName := getAssetName()
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Download new binary
	newBinary, err := downloadBinary(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(newBinary)

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Replace current binary
	fmt.Printf("Installing to %s...\n", execPath)
	if err := replaceBinary(newBinary, execPath); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	fmt.Printf("✓ Successfully updated to %s\n", latestVersion)
	return nil
}

func getLatestRelease() (*githubRelease, error) {
	resp, err := http.Get(githubAPILatest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getAssetName() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("cli-envie-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)
}

func downloadBinary(url string) (string, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "envie-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Download
	resp, err := http.Get(url)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Write to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func replaceBinary(newPath, oldPath string) error {
	// On Unix, we can just move the file
	// On Windows, we need to rename the old file first

	if runtime.GOOS == "windows" {
		// Rename old binary
		oldBackup := oldPath + ".old"
		os.Remove(oldBackup) // Remove any existing backup
		if err := os.Rename(oldPath, oldBackup); err != nil {
			return err
		}

		// Move new binary
		if err := os.Rename(newPath, oldPath); err != nil {
			// Try to restore old binary
			os.Rename(oldBackup, oldPath)
			return err
		}

		// Schedule old binary for deletion (will be deleted on next boot or can be manually deleted)
		os.Remove(oldBackup)
		return nil
	}

	// Unix: check if we have write permission
	info, err := os.Stat(oldPath)
	if err != nil {
		return err
	}

	// Check if we need sudo
	if err := os.Rename(newPath, oldPath); err != nil {
		// Try with sudo on Unix
		fmt.Println("Requesting sudo access to update...")
		cmd := exec.Command("sudo", "mv", newPath, oldPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to move binary (try running with sudo): %w", err)
		}

		// Restore permissions
		exec.Command("sudo", "chmod", fmt.Sprintf("%o", info.Mode()), oldPath).Run()
	}

	return nil
}
