//go:build wails && !bindings && !darwin
// +build wails,!bindings,!darwin

package main

import "errors"

func verifyReleaseAcceptanceBoundaryPlatform() error {
	return errors.New("release-acceptance network boundary is only supported on macOS")
}
