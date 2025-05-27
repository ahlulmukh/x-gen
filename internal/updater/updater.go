package updater

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const updateURL = "https://raw.githubusercontent.com/ahlulmukh/x-gen/main/release.json"

type UpdateInfo struct {
	Version string `json:"version"`
	Windows struct {
		URL      string `json:"url"`
		Checksum string `json:"checksum"`
	} `json:"windows"`
	Linux struct {
		Amd64 struct {
			URL      string `json:"url"`
			Checksum string `json:"checksum"`
		} `json:"amd64"`
		Arm64 struct {
			URL      string `json:"url"`
			Checksum string `json:"checksum"`
		} `json:"arm64"`
	} `json:"linux"`
}

func CheckUpdate(currentVersion string) (*UpdateInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(updateURL)
	if err != nil {
		return nil, fmt.Errorf("Failed check update : %v", err)
	}
	defer resp.Body.Close()

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("invalid update info: %v", err)
	}

	if info.Version > currentVersion {
		return &info, nil
	}

	return nil, nil
}

func DownloadUpdate(url string) (string, error) {
	tmpFile := filepath.Join(os.TempDir(), "solix-update-"+filepath.Base(url))

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	size := resp.ContentLength
	if size <= 0 {
		fmt.Println("Download update ...")
		_, err = io.Copy(out, resp.Body)
		return tmpFile, err
	}

	fmt.Println("Download update ...:")
	bar := &ProgressBar{Total: size, Width: 30}
	_, err = io.Copy(out, io.TeeReader(resp.Body, bar))
	fmt.Println()
	return tmpFile, err
}

func VerifyChecksum(filePath, expected string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}

	actual := fmt.Sprintf("%x", hash.Sum(nil))
	return strings.EqualFold(actual, expected)
}

func ApplyUpdate(updateFile string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		batch := fmt.Sprintf(`
		@echo off
		timeout /t 2 /nobreak
		del "%s"
		move /Y "%s" "%s"
		timeout /t 1 /nobreak
		start "" "%s"
		`, exePath, updateFile, exePath, exePath)

		batchFile := filepath.Join(os.TempDir(), "update.bat")
		if err := os.WriteFile(batchFile, []byte(batch), 0755); err != nil {
			return err
		}
		return exec.Command("cmd", "/C", batchFile).Start()
	}

	if err := os.Rename(updateFile, exePath); err != nil {
		return err
	}
	return exec.Command("chmod", "+x", exePath).Start()
}

type ProgressBar struct {
	Total      int64
	Downloaded int64
	Width      int
}

func (pb *ProgressBar) Write(p []byte) (int, error) {
	n := len(p)
	pb.Downloaded += int64(n)
	pb.Print()
	return n, nil
}

func (pb *ProgressBar) Print() {
	percent := float64(pb.Downloaded) / float64(pb.Total)
	done := int(percent * float64(pb.Width))
	remain := pb.Width - done

	bar := strings.Repeat("#", done) + strings.Repeat("-", remain)
	fmt.Printf("\r[%s] %3.0f%%", bar, percent*100)
}
