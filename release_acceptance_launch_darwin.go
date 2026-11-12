//go:build wails && !bindings && darwin
// +build wails,!bindings,darwin

package main

func verifyReleaseAcceptanceBoundaryPlatform() error {
	return verifyReleaseAcceptanceNetworkDenied()
}
