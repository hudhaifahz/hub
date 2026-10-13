package service

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultWorkdirIsIndependentFromAlbyHub(t *testing.T) {
	dataHome := filepath.Join(string(filepath.Separator), "tmp", "data-home")
	require.Equal(
		t,
		filepath.Join(dataHome, "grynvault-rebalance"),
		defaultWorkdir(dataHome),
	)
	require.NotEqual(t, filepath.Join(dataHome, "albyhub"), defaultWorkdir(dataHome))
}
