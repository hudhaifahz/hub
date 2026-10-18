//go:build wails && !bindings && !darwin
// +build wails,!bindings,!darwin

package main

import "errors"

func execReleaseAcceptanceSandboxPlatform(_ []string) error {
	return errors.New("release-acceptance network sandbox is only supported on macOS")
}
