package wails

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	sqlite_wrapper "github.com/getAlby/hub/db/sqlite-wrapper"
	"github.com/stretchr/testify/require"
)

func TestStageLegacyAlbyWorkdir(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	target := filepath.Join(root, "grynvault-rebalance")
	require.NoError(t, createLegacyWalletFixture(source, true))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(source, "log-entry-not-imported.log"), []byte("historical log"), 0600))

	now := time.Date(2026, time.October, 16, 12, 0, 0, 123, time.UTC)
	result, err := stageLegacyAlbyWorkdir(source, target, now)
	require.NoError(t, err)
	require.True(t, result.RestartRequired)
	require.NotEmpty(t, result.BackupLocation)

	restore := filepath.Join(target, "restore")
	require.FileExists(t, filepath.Join(restore, "nwc.db"))
	require.FileExists(t, filepath.Join(restore, "ldk", "storage", "ldk_node_data.sqlite"))
	require.NoFileExists(t, filepath.Join(restore, "nwc.db-wal"))
	require.NoFileExists(t, filepath.Join(restore, "nwc.db-shm"))
	require.NoFileExists(t, filepath.Join(restore, "log-entry-not-imported.log"))
	require.FileExists(t, filepath.Join(result.BackupLocation, "albyhub-original-snapshot", "nwc.db"))
	require.FileExists(t, filepath.Join(source, "nwc.db"))

	evidenceBytes, err := os.ReadFile(filepath.Join(result.BackupLocation, "import-evidence.json"))
	require.NoError(t, err)
	evidence := legacyImportEvidence{}
	require.NoError(t, json.Unmarshal(evidenceBytes, &evidence))
	require.Equal(t, legacyImportEvidenceVersion, evidence.Version)
	require.Equal(t, "ok", evidence.SourceNWCIntegrity)
	require.Equal(t, "ok", evidence.SourceLDKIntegrity)
	require.EqualValues(t, 1, evidence.SourceCounts.Apps)
	require.EqualValues(t, 1, evidence.SourceCounts.Permissions)
	require.EqualValues(t, 1, evidence.SourceCounts.Transactions)
	require.EqualValues(t, 3, evidence.SourceCounts.UserConfigs)
	require.EqualValues(t, 1, evidence.SourceCounts.Migrations)
	require.EqualValues(t, 1, evidence.SourceCounts.RebalanceRecords)
}

func TestStageLegacyAlbyWorkdirBacksUpFreshTargetDatabase(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	target := filepath.Join(root, "grynvault-rebalance")
	require.NoError(t, createLegacyWalletFixture(source, true))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(target, "nwc.db"), []byte("fresh setup database"), 0600))

	result, err := stageLegacyAlbyWorkdir(source, target, time.Now())
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(result.BackupLocation, "grynvault-before-import", "nwc.db"))
	require.FileExists(t, filepath.Join(target, "restore", "nwc.db"))
}

func TestResolveLegacyAlbyWorkdirFromContainer(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "Data", "Library", "Application Support", "albyhub")
	require.NoError(t, createLegacyWalletFixture(workdir, true))

	resolved, err := resolveLegacyAlbyWorkdir(root)
	require.NoError(t, err)
	require.Equal(t, workdir, resolved)
}

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

func TestStageLegacyAlbyWorkdirAcceptsStockDatabaseWithoutRebalanceTable(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	target := filepath.Join(root, "grynvault-rebalance")
	require.NoError(t, createLegacyWalletFixture(source, true))

	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(source, "nwc.db"))
	require.NoError(t, err)
	_, err = database.Exec("DROP TABLE local_rebalance_quotes")
	require.NoError(t, err)
	require.NoError(t, database.Close())

	result, err := stageLegacyAlbyWorkdir(source, target, time.Now())
	require.NoError(t, err)
	evidenceBytes, err := os.ReadFile(filepath.Join(result.BackupLocation, "import-evidence.json"))
	require.NoError(t, err)
	evidence := legacyImportEvidence{}
	require.NoError(t, json.Unmarshal(evidenceBytes, &evidence))
	require.Zero(t, evidence.SourceCounts.RebalanceRecords)
}

func TestStageLegacyAlbyWorkdirRejectsIncompleteWallet(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	target := filepath.Join(root, "grynvault-rebalance")
	require.NoError(t, createLegacyWalletFixture(source, false))

	_, err := stageLegacyAlbyWorkdir(source, target, time.Now())
	require.ErrorContains(t, err, "incomplete")
	require.NoDirExists(t, filepath.Join(target, "restore"))
}

func TestStageLegacyAlbyWorkdirRejectsNonLDKWallet(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	target := filepath.Join(root, "grynvault-rebalance")
	require.NoError(t, createLegacyWalletFixture(source, true))

	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(source, "nwc.db"))
	require.NoError(t, err)
	_, err = database.Exec("UPDATE user_configs SET value = 'LND' WHERE key = 'LNBackendType'")
	require.NoError(t, err)
	require.NoError(t, database.Close())

	_, err = stageLegacyAlbyWorkdir(source, target, time.Now())
	require.ErrorContains(t, err, "not embedded LDK")
	require.NoDirExists(t, filepath.Join(target, "restore"))
}

func TestCopyWalletDataRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	destination := filepath.Join(root, "copy")
	require.NoError(t, createLegacyWalletFixture(source, true))
	require.NoError(t, os.WriteFile(filepath.Join(root, "outside"), []byte("outside"), 0600))
	require.NoError(t, os.Symlink(filepath.Join(root, "outside"), filepath.Join(source, "ldk", "storage", "unsafe")))

	_, err := copyWalletData(source, destination)
	require.ErrorContains(t, err, "symbolic link")
}

func TestCopyWalletDataIgnoresLDKLogsWithSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	destination := filepath.Join(root, "copy")
	require.NoError(t, createLegacyWalletFixture(source, true))
	require.NoError(t, os.MkdirAll(filepath.Join(source, "ldk", "logs"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(source, "ldk", "logs", "ldk_node_2026_10_16.log"), []byte("runtime log"), 0600))
	require.NoError(t, os.Symlink("ldk_node_2026_10_16.log", filepath.Join(source, "ldk", "logs", "ldk_node_latest.log")))

	_, err := copyWalletData(source, destination)
	require.NoError(t, err)
	require.NoDirExists(t, filepath.Join(destination, "ldk", "logs"))
	require.FileExists(t, filepath.Join(destination, "ldk", "storage", "ldk_node_data.sqlite"))
}

func createLegacyWalletFixture(workdir string, complete bool) error {
	if err := os.MkdirAll(filepath.Join(workdir, "ldk", "storage"), 0700); err != nil {
		return err
	}

	nwc, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(workdir, "nwc.db"))
	if err != nil {
		return err
	}
	defer nwc.Close()

	statements := []string{
		"CREATE TABLE user_configs (key TEXT PRIMARY KEY, value TEXT)",
		"CREATE TABLE apps (id INTEGER PRIMARY KEY)",
		"CREATE TABLE app_permissions (id INTEGER PRIMARY KEY)",
		"CREATE TABLE transactions (id INTEGER PRIMARY KEY)",
		"CREATE TABLE migrations (id TEXT PRIMARY KEY)",
		"CREATE TABLE local_rebalance_quotes (id INTEGER PRIMARY KEY)",
		"INSERT INTO apps(id) VALUES (1)",
		"INSERT INTO app_permissions(id) VALUES (1)",
		"INSERT INTO transactions(id) VALUES (1)",
		"INSERT INTO migrations(id) VALUES ('one')",
		"INSERT INTO local_rebalance_quotes(id) VALUES (1)",
	}
	if complete {
		statements = append(statements,
			"INSERT INTO user_configs(key, value) VALUES ('UnlockPasswordCheck', 'encrypted')",
			"INSERT INTO user_configs(key, value) VALUES ('LNBackendType', 'LDK')",
			"INSERT INTO user_configs(key, value) VALUES ('Mnemonic', 'encrypted')",
		)
	}
	for _, statement := range statements {
		if _, err = nwc.Exec(statement); err != nil {
			return err
		}
	}
	if err = nwc.Close(); err != nil {
		return err
	}

	ldk, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite"))
	if err != nil {
		return err
	}
	if _, err = ldk.Exec("CREATE TABLE kv (key TEXT PRIMARY KEY, value BLOB)"); err != nil {
		_ = ldk.Close()
		return err
	}
	return ldk.Close()
}
