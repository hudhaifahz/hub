package wails

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGuidedMigrationRouteAllowlist(t *testing.T) {
	allowed := [][2]string{
		{"GET", "/api/info"},
		{"GET", "/api/import-existing-alby/status"},
		{"POST", "/api/import-existing-alby/verify"},
		{"POST", "/api/import-existing-alby/activate"},
		{"POST", "/api/import-existing-alby/open-evidence"},
		{"POST", "/api/import-existing-alby/return-to-source"},
		{"POST", "/api/import-existing-alby/abandon"},
	}
	for _, route := range allowed {
		require.True(t, guidedMigrationRouteAllowed(route[0], route[1]), route)
	}

	blocked := [][2]string{
		{"POST", "/api/start"},
		{"POST", "/api/setup"},
		{"POST", "/api/payments"},
		{"POST", "/api/import-existing-alby/stage"},
		{"GET", "/api/health"},
		{"DELETE", "/api/apps/1"},
	}
	for _, route := range blocked {
		require.False(t, guidedMigrationRouteAllowed(route[0], route[1]), route)
	}
}
