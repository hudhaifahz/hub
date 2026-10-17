//go:build darwin

package wails

import (
	"fmt"
	"os"
	"os/exec"
)

func revealLocalDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("open recovery folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("recovery evidence path is not a folder")
	}
	if err = exec.Command("/usr/bin/open", path).Start(); err != nil {
		return fmt.Errorf("open recovery folder: %w", err)
	}
	return nil
}
