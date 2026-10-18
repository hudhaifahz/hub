//go:build wails && !bindings
// +build wails,!bindings

package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	releaseAcceptanceArgPrefix        = "--grynvault-release-acceptance-"
	releaseAcceptanceWorkdirArg       = "--grynvault-release-acceptance-work-dir="
	releaseAcceptanceNetworkDarkArg   = "--grynvault-release-acceptance-network-dark"
	releaseAcceptanceWorkdirEnv       = "GRYNVAULT_RELEASE_ACCEPTANCE_WORK_DIR"
	releaseAcceptanceActiveEnv        = "GRYNVAULT_RELEASE_ACCEPTANCE_ACTIVE"
	releaseAcceptanceAuthoritySuffix  = ".grynvault-release-acceptance"
	releaseAcceptanceAuthorityVersion = "grynvault-release-acceptance-v1"
)

var verifyReleaseAcceptanceBoundary = verifyReleaseAcceptanceBoundaryPlatform

func prepareReleaseAcceptanceLaunch(args []string) error {
	workdir, requested, err := parseReleaseAcceptanceArgs(args)
	if err != nil {
		return err
	}

	active := os.Getenv(releaseAcceptanceActiveEnv) == "1"
	if !requested && !active {
		return nil
	}
	if active && workdir == "" {
		workdir = os.Getenv(releaseAcceptanceWorkdirEnv)
	}
	if workdir == "" {
		return errors.New("release-acceptance work directory is required")
	}

	workdir, err = validateReleaseAcceptanceWorkdir(workdir)
	if err != nil {
		return err
	}
	if err = setReleaseAcceptanceEnvironment(workdir); err != nil {
		return err
	}

	return verifyReleaseAcceptanceBoundary()
}

func parseReleaseAcceptanceArgs(args []string) (string, bool, error) {
	var workdir string
	workdirProvided := false
	networkDark := false
	for _, argument := range args[1:] {
		switch {
		case strings.HasPrefix(argument, releaseAcceptanceWorkdirArg):
			if workdirProvided {
				return "", false, errors.New("release-acceptance work directory was provided more than once")
			}
			workdirProvided = true
			workdir = strings.TrimPrefix(argument, releaseAcceptanceWorkdirArg)
			if workdir == "" {
				return "", false, errors.New("release-acceptance work directory cannot be empty")
			}
		case argument == releaseAcceptanceNetworkDarkArg:
			if networkDark {
				return "", false, errors.New("release-acceptance network-dark flag was provided more than once")
			}
			networkDark = true
		case strings.HasPrefix(argument, releaseAcceptanceArgPrefix):
			return "", false, fmt.Errorf("unrecognized release-acceptance argument %q", argument)
		}
	}

	requested := workdirProvided || networkDark
	if requested && (!workdirProvided || !networkDark) {
		return "", false, errors.New("release acceptance requires both the bound work directory and network-dark flag")
	}
	return workdir, requested, nil
}

func validateReleaseAcceptanceWorkdir(workdir string) (string, error) {
	if !filepath.IsAbs(workdir) {
		return "", errors.New("release-acceptance work directory must be absolute")
	}
	cleaned := filepath.Clean(workdir)
	pathInfo, err := os.Lstat(cleaned)
	if err != nil {
		return "", fmt.Errorf("inspect release-acceptance work directory path: %w", err)
	}
	if pathInfo.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("release-acceptance work directory cannot be a symlink")
	}
	resolved, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		return "", fmt.Errorf("resolve release-acceptance work directory: %w", err)
	}
	cleaned = resolved
	info, err := os.Stat(cleaned)
	if err != nil {
		return "", fmt.Errorf("inspect release-acceptance work directory: %w", err)
	}
	if !info.IsDir() {
		return "", errors.New("release-acceptance work directory is not a directory")
	}
	if info.Mode().Perm()&0077 != 0 {
		return "", errors.New("release-acceptance work directory must not be accessible by group or other users")
	}

	authorityPath := cleaned + releaseAcceptanceAuthoritySuffix
	authorityInfo, err := os.Lstat(authorityPath)
	if err != nil {
		return "", fmt.Errorf("inspect release-acceptance authority file: %w", err)
	}
	if !authorityInfo.Mode().IsRegular() || authorityInfo.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("release-acceptance authority must be a regular file")
	}
	if authorityInfo.Mode().Perm()&0077 != 0 {
		return "", errors.New("release-acceptance authority must not be accessible by group or other users")
	}
	authority, err := os.ReadFile(authorityPath)
	if err != nil {
		return "", fmt.Errorf("read release-acceptance authority: %w", err)
	}
	expectedAuthority := releaseAcceptanceAuthorityVersion + "\n" + cleaned + "\n"
	if string(authority) != expectedAuthority {
		return "", errors.New("release-acceptance authority is not bound to the selected work directory")
	}
	return cleaned, nil
}

func setReleaseAcceptanceEnvironment(workdir string) error {
	values := map[string]string{
		"WORK_DIR":                  workdir,
		"LOG_TO_FILE":               "false",
		"SEND_EVENTS_TO_ALBY":       "false",
		"AUTO_LINK_ALBY_ACCOUNT":    "false",
		releaseAcceptanceWorkdirEnv: workdir,
		releaseAcceptanceActiveEnv:  "1",
	}
	for key, value := range values {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set release-acceptance environment %s: %w", key, err)
		}
	}
	return nil
}

func verifyReleaseAcceptanceNetworkDenied() error {
	dialer := net.Dialer{Timeout: 250 * time.Millisecond}
	connection, err := dialer.Dial("tcp", "192.0.2.1:9")
	if connection != nil {
		_ = connection.Close()
		return errors.New("release-acceptance sandbox allowed an external connection")
	}
	if !errors.Is(err, syscall.EPERM) && !errors.Is(err, syscall.EACCES) {
		return fmt.Errorf("release-acceptance external-network denial is not proven: %w", err)
	}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if listener != nil {
		_ = listener.Close()
		return errors.New("release-acceptance boundary allowed a non-loopback bind")
	}
	if !errors.Is(err, syscall.EPERM) && !errors.Is(err, syscall.EACCES) {
		return fmt.Errorf("release-acceptance non-loopback bind denial is not proven: %w", err)
	}

	return nil
}
