//go:build wails && !bindings && darwin
// +build wails,!bindings,darwin

package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func execReleaseAcceptanceSandboxPlatform(args []string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate Grynvault executable: %w", err)
	}
	sandboxExecutable := "/usr/bin/sandbox-exec"
	sandboxArgs := []string{sandboxExecutable, "-p", releaseAcceptanceSandboxProfile, executable}
	sandboxArgs = append(sandboxArgs, args[1:]...)
	if err = syscall.Exec(sandboxExecutable, sandboxArgs, os.Environ()); err != nil {
		return fmt.Errorf("enter release-acceptance network sandbox: %w", err)
	}
	return errors.New("release-acceptance sandbox returned unexpectedly")
}
