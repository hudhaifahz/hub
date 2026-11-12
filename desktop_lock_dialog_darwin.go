//go:build darwin && wails && !bindings
// +build darwin,wails,!bindings

package main

/*
#cgo LDFLAGS: -framework Cocoa
void showDesktopWalletLockConflict(void);
*/
import "C"

func showDesktopLockConflict() {
	C.showDesktopWalletLockConflict()
}
