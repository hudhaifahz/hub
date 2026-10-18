package wails

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserHomeOutsideSandbox(t *testing.T) {
	tests := []struct {
		name     string
		home     string
		expected string
	}{
		{
			name:     "ordinary home",
			home:     filepath.Join(string(filepath.Separator), "Users", "kode"),
			expected: filepath.Join(string(filepath.Separator), "Users", "kode"),
		},
		{
			name:     "macOS application sandbox",
			home:     filepath.Join(string(filepath.Separator), "Users", "kode", "Library", "Containers", "com.grynvault.rebalance", "Data"),
			expected: filepath.Join(string(filepath.Separator), "Users", "kode"),
		},
		{
			name:     "unrelated data directory",
			home:     filepath.Join(string(filepath.Separator), "tmp", "example", "Data"),
			expected: filepath.Join(string(filepath.Separator), "tmp", "example", "Data"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, userHomeOutsideSandbox(test.home))
		})
	}
}

func TestLegacyAlbyWorkdirFromSelection(t *testing.T) {
	workdir := t.TempDir()
	databasePath := filepath.Join(workdir, "albyhub.db")
	require.NoError(t, os.WriteFile(databasePath, []byte("test"), 0o600))

	selectedDirectory, err := legacyAlbyWorkdirFromSelection(workdir)
	require.NoError(t, err)
	require.Equal(t, workdir, selectedDirectory)

	selectedDatabase, err := legacyAlbyWorkdirFromSelection(databasePath)
	require.NoError(t, err)
	require.Equal(t, workdir, selectedDatabase)
}

func TestLegacyAlbyWorkdirFromSelectionRejectsOtherFiles(t *testing.T) {
	workdir := t.TempDir()
	wrongPath := filepath.Join(workdir, "nwc.db")
	require.NoError(t, os.WriteFile(wrongPath, []byte("test"), 0o600))

	_, err := legacyAlbyWorkdirFromSelection(wrongPath)
	require.EqualError(t, err, "select the albyhub.db file inside the Alby Hub data folder")
}

func TestLegacyAlbyWorkdirFromSelectionRejectsMissingPath(t *testing.T) {
	_, err := legacyAlbyWorkdirFromSelection(filepath.Join(t.TempDir(), "albyhub.db"))
	require.ErrorContains(t, err, "selected Alby Hub path could not be inspected")
}
