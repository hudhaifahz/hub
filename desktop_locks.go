//go:build wails && !bindings
// +build wails,!bindings

package main

import (
	"fmt"
	"net"
)

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

func closeDesktopLocks(listeners []net.Listener) {
	for _, listener := range listeners {
		_ = listener.Close()
	}
}
