package wails

import (
	"os"
	"path/filepath"
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
