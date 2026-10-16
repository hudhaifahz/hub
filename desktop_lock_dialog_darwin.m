//go:build darwin && wails && !bindings
// +build darwin,wails,!bindings

#import <Cocoa/Cocoa.h>

void showDesktopWalletLockConflict(void) {
  @autoreleasepool {
    [NSApplication sharedApplication];

    NSAlert *alert = [[NSAlert alloc] init];
    alert.alertStyle = NSAlertStyleWarning;
    alert.messageText = @"Your wallet is already open";
    alert.informativeText = @"Close Alby Hub or Grynvault Rebalance, then open Grynvault Rebalance again. Your channels and wallet data are safe. This lock prevents two apps from using the same Lightning node at the same time.";
    [alert addButtonWithTitle:@"OK"];

    [NSApp activateIgnoringOtherApps:YES];
    [alert runModal];
  }
}
