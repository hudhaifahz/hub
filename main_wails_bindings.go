//go:build wails && bindings
// +build wails,bindings

package main

import (
	"embed"

	hubwails "github.com/getAlby/hub/wails"
)

// The Wails binding generator executes this binary. Keep that process isolated
// from the desktop single-instance port and from all wallet/service startup.

//go:embed all:frontend/dist
var bindingAssets embed.FS

//go:embed appicon.png
var bindingAppIcon []byte

func main() {
	hubwails.LaunchWailsApp(hubwails.NewBindingsApp(), bindingAssets, bindingAppIcon)
}
