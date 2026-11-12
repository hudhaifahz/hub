//go:build wails && !bindings
// +build wails,!bindings

package main

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

type trackingListener struct {
	closed bool
}

func (listener *trackingListener) Accept() (net.Conn, error) { return nil, errors.New("unused") }
func (listener *trackingListener) Addr() net.Addr            { return &net.TCPAddr{} }
func (listener *trackingListener) Close() error {
	listener.closed = true
	return nil
}

func TestAcquireDesktopLocksClosesEarlierLockOnConflict(t *testing.T) {
	first := &trackingListener{}
	calls := 0
	listeners, err := acquireDesktopLocks(func(_, _ string) (net.Listener, error) {
		calls++
		if calls == 1 {
			return first, nil
		}
		return nil, errors.New("already in use")
	})

	require.Error(t, err)
	require.Nil(t, listeners)
	require.True(t, first.closed)
}

func TestAcquireDesktopLocksHoldsBothPorts(t *testing.T) {
	created := []*trackingListener{}
	listeners, err := acquireDesktopLocks(func(_, _ string) (net.Listener, error) {
		listener := &trackingListener{}
		created = append(created, listener)
		return listener, nil
	})

	require.NoError(t, err)
	require.Len(t, listeners, 2)
	closeDesktopLocks(listeners)
	require.True(t, created[0].closed)
	require.True(t, created[1].closed)
}
