//go:build wails && !bindings
// +build wails,!bindings

package main

import (
	"context"
	"embed"
	"net"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/getAlby/hub/wails"
	log "github.com/sirupsen/logrus"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed appicon.png
var appIcon []byte

func main() {
	// Hold Grynvault's lock and Alby Hub's desktop lock for the complete app
	// lifetime. This prevents both apps from starting the same copied Lightning
	// node at once after an import.
	listeners, err := acquireDesktopLocks(net.Listen)
	if err != nil {
		log.Println("Alby Hub or Grynvault Rebalance is already running.")
		showDesktopLockConflict()
		return
	}
	defer closeDesktopLocks(listeners)

	log.Info("Grynvault Rebalance starting in WAILS mode")
	ctx, cancel := context.WithCancel(context.Background())
	svc, err := service.NewService(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to create service")
		return
	}

	app := wails.NewApp(svc)
	wails.LaunchWailsApp(app, assets, appIcon)
	logger.Logger.Info("Wails app exited")

	logger.Logger.Info("Cancelling service context...")
	// cancel the service context
	cancel()
	svc.Shutdown()
	logger.Logger.Info("Service exited")
	logger.Logger.Info("Grynvault Rebalance needs to stay online to send and receive transactions. Channels may be closed if your hub stays offline for an extended period of time.")
}
