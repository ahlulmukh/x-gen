package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"x-gen/internal/menu"
	"x-gen/internal/updater"
	"x-gen/internal/utils"
)

const version = "0.0.1"

func main() {
	if updated := checkForUpdates(); updated {
		return
	}

	app := menu.NewMenuHandler()
	app.ShowMainMenu(version)
}

func checkForUpdates() bool {
	update, err := updater.CheckUpdate(version)
	if err != nil {
		utils.LogMessage("Update check failed: "+err.Error(), "warning")
		return false
	}

	if update != nil {
		utils.LogMessage(
			fmt.Sprintf("Update v%s available! Current: v%s", update.Version, version),
			"info")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Update now? (y/n): ")
		input, _ := reader.ReadString('\n')

		if strings.TrimSpace(input) == "y" {
			performUpdate(update)
			return true
		}
	}
	return false
}

func performUpdate(update *updater.UpdateInfo) {
	var url, checksum string

	switch runtime.GOOS {
	case "windows":
		url = update.Windows.URL
		checksum = update.Windows.Checksum
	case "linux":
		if runtime.GOARCH == "amd64" {
			url = update.Linux.Amd64.URL
			checksum = update.Linux.Amd64.Checksum
		} else if runtime.GOARCH == "arm64" {
			url = update.Linux.Arm64.URL
			checksum = update.Linux.Arm64.Checksum
		} else {
			utils.LogMessage("Unsupported architecture for Linux", "error")
			return
		}
	default:
		utils.LogMessage("Unsupported OS for auto-update", "error")
		return
	}

	utils.LogMessage("Downloading update...", "process")
	tmpFile, err := updater.DownloadUpdate(url)
	if err != nil {
		utils.LogMessage("Download failed: "+err.Error(), "error")
		return
	}

	if !updater.VerifyChecksum(tmpFile, checksum) {
		utils.LogMessage("Checksum verification failed!", "error")
		os.Remove(tmpFile)
		return
	}

	utils.LogMessage("Applying update...", "process")
	if err := updater.ApplyUpdate(tmpFile); err != nil {
		utils.LogMessage("Update failed: "+err.Error(), "error")
		return
	}

	utils.LogMessage("Update successful! Restarting...", "success")
	os.Exit(0)
}
