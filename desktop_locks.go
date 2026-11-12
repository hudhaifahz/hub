//go:build wails && !bindings
// +build wails,!bindings

package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const guidedMigrationRelaunchEnv = "GRYNVAULT_GUIDED_MIGRATION_RELAUNCH"

var desktopLockAddresses = []string{
	"127.0.0.1:21421", // Grynvault Rebalance
	"127.0.0.1:21420", // Alby Hub desktop
}

func acquireDesktopLocks(listen func(network, address string) (net.Listener, error)) ([]net.Listener, error) {
	listeners := make([]net.Listener, 0, len(desktopLockAddresses))
	for _, address := range desktopLockAddresses {
		listener, err := listen("tcp", address)
		if err != nil {
			closeDesktopLocks(listeners)
			return nil, fmt.Errorf("acquire desktop wallet lock %s: %w", address, err)
		}
		listeners = append(listeners, listener)
	}
	return listeners, nil
}

func acquireDesktopLocksForLaunch(listen func(network, address string) (net.Listener, error)) ([]net.Listener, error) {
	if os.Getenv(guidedMigrationRelaunchEnv) != "1" {
		return acquireDesktopLocks(listen)
	}

	// The replacement process is created before the current process begins its
	// orderly shutdown. Wait until both wallet-ownership locks are released so
	// the replacement cannot race the old service or silently exit.
	deadline := time.Now().Add(2 * time.Minute)
	for {
		listeners, err := acquireDesktopLocks(listen)
		if err == nil {
			_ = os.Unsetenv(guidedMigrationRelaunchEnv)
			return listeners, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("wait for previous Grynvault process to release wallet locks: %w", err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func closeDesktopLocks(listeners []net.Listener) {
	for _, listener := range listeners {
		_ = listener.Close()
	}
}
