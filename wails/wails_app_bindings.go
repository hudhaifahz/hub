//go:build bindings
// +build bindings

package wails

// NewBindingsApp returns a zero-value app only for Wails code generation. The
// bindings build never creates a service, opens a wallet, or claims the desktop
// single-instance port.
func NewBindingsApp() *WailsApp {
	return &WailsApp{}
}
