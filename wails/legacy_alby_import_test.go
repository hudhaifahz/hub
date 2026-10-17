package wails

import (
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
