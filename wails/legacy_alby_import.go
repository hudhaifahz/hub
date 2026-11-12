package wails

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	releaseAcceptanceActiveEnvironment  = "GRYNVAULT_RELEASE_ACCEPTANCE_ACTIVE"
	releaseAcceptanceWorkdirEnvironment = "GRYNVAULT_RELEASE_ACCEPTANCE_WORK_DIR"
)

func canonicalLegacyAlbyWorkdir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	home = userHomeOutsideSandbox(home)
	return filepath.Join(home, "Library", "Containers", "com.getalby.Alby-Hub", "Data", "Library", "Application Support", "albyhub")
}

func userHomeOutsideSandbox(home string) string {
	home = filepath.Clean(home)
	if filepath.Base(home) != "Data" {
		return home
	}
	containerDirectory := filepath.Dir(home)
	containersDirectory := filepath.Dir(containerDirectory)
	libraryDirectory := filepath.Dir(containersDirectory)
	if filepath.Base(containersDirectory) != "Containers" || filepath.Base(libraryDirectory) != "Library" {
		return home
	}
	return filepath.Dir(libraryDirectory)
}

func legacyAlbyWorkdirFromSelection(selection string) (string, error) {
	selection = filepath.Clean(selection)
	info, err := os.Stat(selection)
	if err != nil {
		return "", fmt.Errorf("selected Alby Hub path could not be inspected: %w", err)
	}
	if info.IsDir() {
		return selection, nil
	}
	if filepath.Base(selection) != "albyhub.db" {
		return "", fmt.Errorf("select the albyhub.db file inside the Alby Hub data folder")
	}
	return filepath.Dir(selection), nil
}

func releaseAcceptanceStageHomeDirectory(targetWorkdir string) string {
	if os.Getenv(releaseAcceptanceActiveEnvironment) != "1" {
		return ""
	}
	boundWorkdir := os.Getenv(releaseAcceptanceWorkdirEnvironment)
	if boundWorkdir == "" {
		return ""
	}
	boundAbsolute, boundErr := filepath.Abs(boundWorkdir)
	targetAbsolute, targetErr := filepath.Abs(targetWorkdir)
	if boundErr != nil || targetErr != nil || filepath.Clean(boundAbsolute) != filepath.Clean(targetAbsolute) {
		return ""
	}
	return filepath.Join(filepath.Dir(filepath.Clean(targetAbsolute)), "acceptance-home")
}
