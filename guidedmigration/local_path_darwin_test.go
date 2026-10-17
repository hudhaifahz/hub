//go:build darwin

package guidedmigration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsLocalPathUsesNearestExistingFilesystem(t *testing.T) {
	root := t.TempDir()
	local, err := IsLocalPath(filepath.Join(root, "not-created", "wallet"))
	require.NoError(t, err)
	require.True(t, local)
}
