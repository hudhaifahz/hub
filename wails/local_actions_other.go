//go:build !darwin

package wails

import "errors"

func revealLocalDirectory(string) error {
	return errors.New("opening the recovery folder is currently supported in the macOS desktop app")
}
