package guidedmigration

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/stretchr/testify/require"

	sqlite_wrapper "github.com/getAlby/hub/db/sqlite-wrapper"
)

func TestResolveWorkdirSupportsContainerAndCustomPaths(t *testing.T) {
	root := t.TempDir()
	containerRoot := filepath.Join(root, "Custom Container With Spaces")
	workdir := filepath.Join(containerRoot, "Data", "Library", "Application Support", "albyhub")
	require.NoError(t, createWalletFixture(workdir, "encrypted-wallet-a"))

	resolved, err := ResolveWorkdir(containerRoot)
	require.NoError(t, err)
	require.Equal(t, evaluatedPath(t, workdir), resolved)

	resolved, err = ResolveWorkdir(workdir)
	require.NoError(t, err)
	require.Equal(t, evaluatedPath(t, workdir), resolved)

	for name, selected := range map[string]string{
		"data directory":                filepath.Join(containerRoot, "Data"),
		"application support directory": filepath.Join(containerRoot, "Data", "Library", "Application Support"),
		"albyhub parent":                filepath.Dir(workdir),
	} {
		t.Run(name, func(t *testing.T) {
			resolvedPath, resolveErr := ResolveWorkdir(selected)
			require.NoError(t, resolveErr)
			require.Equal(t, evaluatedPath(t, workdir), resolvedPath)
		})
	}

	unicodeWorkdir := filepath.Join(root, "External Volume", "Wallets", "Alby 雷", "albyhub")
	require.NoError(t, createWalletFixture(unicodeWorkdir, "encrypted-wallet-unicode"))
	resolved, err = ResolveWorkdir(unicodeWorkdir)
	require.NoError(t, err)
	require.Equal(t, evaluatedPath(t, unicodeWorkdir), resolved)
}

func TestInspectSourceRejectsInvalidWalletData(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		_, err := InspectSource(t.TempDir())
		require.ErrorContains(t, err, "not an Alby Hub")
	})

	t.Run("corrupt database", func(t *testing.T) {
		workdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(workdir, "ldk", "storage"), 0700))
		require.NoError(t, os.WriteFile(filepath.Join(workdir, "nwc.db"), []byte("not sqlite"), 0600))
		require.NoError(t, createSQLiteFixture(filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite")))
		_, err := InspectSource(workdir)
		require.Error(t, err)
	})

	t.Run("incomplete", func(t *testing.T) {
		workdir := t.TempDir()
		require.NoError(t, createWalletFixture(workdir, "encrypted-wallet-a"))
		require.NoError(t, updateFixture(workdir, "DELETE FROM user_configs WHERE key = 'Mnemonic'"))
		_, err := InspectSource(workdir)
		require.ErrorContains(t, err, "incomplete")
	})

	t.Run("non LDK", func(t *testing.T) {
		workdir := t.TempDir()
		require.NoError(t, createWalletFixture(workdir, "encrypted-wallet-a"))
		require.NoError(t, updateFixture(workdir, "UPDATE user_configs SET value = 'LND' WHERE key = 'LNBackendType'"))
		_, err := InspectSource(workdir)
		require.ErrorContains(t, err, "not embedded LDK")
	})

	t.Run("symlinked database", func(t *testing.T) {
		workdir := t.TempDir()
		require.NoError(t, createWalletFixture(workdir, "encrypted-wallet-a"))
		external := filepath.Join(t.TempDir(), "nwc.db")
		require.NoError(t, os.Rename(filepath.Join(workdir, "nwc.db"), external))
		require.NoError(t, os.Symlink(external, filepath.Join(workdir, "nwc.db")))
		_, err := InspectSource(workdir)
		require.ErrorContains(t, err, "not an Alby Hub")
	})

	t.Run("permission denied", func(t *testing.T) {
		workdir := t.TempDir()
		require.NoError(t, createWalletFixture(workdir, "encrypted-wallet-a"))
		databasePath := filepath.Join(workdir, "nwc.db")
		require.NoError(t, os.Chmod(databasePath, 0000))
		_, err := InspectSource(workdir)
		require.NoError(t, os.Chmod(databasePath, 0600))
		if err == nil {
			t.Skip("this platform or test user can read mode-000 files")
		}
		require.Error(t, err)
	})
}

func TestValidationDoesNotCreateSidecarsInSource(t *testing.T) {
	workdir := t.TempDir()
	require.NoError(t, createWalletFixture(workdir, "encrypted-wallet-a"))
	before, err := HashWalletData(workdir)
	require.NoError(t, err)

	_, _, _, _, _, err = ValidateWalletDataWithoutMutation(workdir)
	require.NoError(t, err)
	after, err := HashWalletData(workdir)
	require.NoError(t, err)
	require.NoError(t, CompareHashes(before, after))
	require.NoFileExists(t, filepath.Join(workdir, "nwc.db-wal"))
	require.NoFileExists(t, filepath.Join(workdir, "nwc.db-shm"))
	require.NoFileExists(t, filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite-wal"))
	require.NoFileExists(t, filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite-shm"))
}

func TestStageAndFinishPendingPreserveSourceAndActivateQuarantine(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source", "albyhub")
	target := filepath.Join(root, "target", "grynvault-rebalance")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(target, "fresh-target.txt"), []byte("preserve me"), 0600))

	now := time.Date(2026, time.October, 16, 12, 0, 0, 123, time.UTC)
	result, err := Stage(source, target, testStageOptions(root, now))
	require.NoError(t, err)
	require.True(t, result.RestartRequired)
	require.FileExists(t, filepath.Join(source, "nwc.db"))
	require.FileExists(t, filepath.Join(target, "fresh-target.txt"))

	control, err := LoadControl(target)
	require.NoError(t, err)
	require.Equal(t, PhaseStaged, control.Phase)
	require.DirExists(t, control.StagePath)
	require.FileExists(t, QuarantinePath(control.StagePath))
	require.FileExists(t, filepath.Join(result.BackupLocation, "albyhub-original-snapshot", "nwc.db"))
	require.NoFileExists(t, filepath.Join(result.BackupLocation, "grynvault-before-import", "fresh-target.txt"))

	finish, err := FinishPending(target, now.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, finish.Activated)
	require.NoFileExists(t, filepath.Join(target, "fresh-target.txt"))
	require.FileExists(t, filepath.Join(target, "nwc.db"))
	require.FileExists(t, QuarantinePath(target))
	require.FileExists(t, filepath.Join(result.BackupLocation, "grynvault-before-import", "fresh-target.txt"))

	quarantine, err := LoadQuarantine(target)
	require.NoError(t, err)
	require.Equal(t, PhaseVerificationNeeded, quarantine.Phase)
	require.True(t, quarantine.ActivatedBytesVerified)
	targetHashes, err := HashWalletData(target)
	require.NoError(t, err)
	sourceHashes, err := HashWalletData(source)
	require.NoError(t, err)
	require.NoError(t, CompareHashes(sourceHashes, targetHashes))
}

func TestFinishPendingBlocksWhenSourceChanges(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(target, "untouched"), []byte("yes"), 0600))

	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	_, err := Stage(source, target, testStageOptions(root, now))
	require.NoError(t, err)
	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(source, "nwc.db"))
	require.NoError(t, err)
	_, err = database.Exec("INSERT INTO transactions(id) VALUES (2)")
	require.NoError(t, err)
	require.NoError(t, database.Close())

	finish, err := FinishPending(target, now.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, finish.Blocked)
	require.Contains(t, finish.Reason, "source wallet changed")
	require.FileExists(t, filepath.Join(target, "untouched"))
	require.NoFileExists(t, QuarantinePath(target))
	control, err := LoadControl(target)
	require.NoError(t, err)
	require.Equal(t, PhaseBlocked, control.Phase)
}

func TestStageRejectsKnownMatchingWalletCopy(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "custom", "albyhub")
	target := filepath.Join(root, "target")
	knownCopy := filepath.Join(root, "Library", "Containers", "com.grynvault.rebalance", "Data", "Library", "Application Support", "grynvault-rebalance")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, createWalletFixture(knownCopy, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))

	_, err := Stage(source, target, testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "matching wallet copy found")
	require.NoFileExists(t, ControlPath(target))
}

func TestStageAllowsKnownUnrelatedWalletCopy(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "custom", "albyhub")
	target := filepath.Join(root, "target")
	unrelated := filepath.Join(root, "Library", "Containers", "com.grynvault.rebalance", "Data", "Library", "Application Support", "grynvault-rebalance")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, createWalletFixture(unrelated, "encrypted-wallet-b"))
	require.NoError(t, os.MkdirAll(target, 0700))

	result, err := Stage(source, target, testStageOptions(root, time.Now()))
	require.NoError(t, err)
	require.True(t, result.RestartRequired)
	require.FileExists(t, ControlPath(target))
}

func TestStageRejectsNonLocalSourceOrTarget(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))

	for _, rejected := range []string{source, target} {
		options := testStageOptions(root, time.Now())
		options.LocalPathChecker = func(path string) (bool, error) {
			return filepath.Base(path) != filepath.Base(rejected), nil
		}
		_, err := Stage(source, target, options)
		require.ErrorContains(t, err, "local filesystem")
		require.NoFileExists(t, ControlPath(target))
	}
}

func TestStageReportsAuthoritativeProgress(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	options := testStageOptions(root, time.Now())
	var steps []string
	options.Progress = func(progress Progress) {
		steps = append(steps, progress.Step)
		require.NotEmpty(t, progress.Detail)
	}

	_, err := Stage(source, target, options)
	require.NoError(t, err)
	require.Equal(t, []string{
		"source_validated",
		"source_quiescent",
		"recovery_verified",
		"staged_verified",
		"ready_to_restart",
	}, steps)
}

func TestStageRejectsUnreadableOrCorruptKnownWalletCandidate(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "custom", "albyhub")
	target := filepath.Join(root, "target")
	knownCandidate := filepath.Join(root, "Library", "Containers", "com.grynvault.rebalance", "Data", "Library", "Application Support", "grynvault-rebalance")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(filepath.Join(knownCandidate, "ldk", "storage"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(knownCandidate, "nwc.db"), []byte("corrupt"), 0600))
	require.NoError(t, createSQLiteFixture(filepath.Join(knownCandidate, "ldk", "storage", "ldk_node_data.sqlite")))
	require.NoError(t, os.MkdirAll(target, 0700))

	_, err := Stage(source, target, testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "possible competing wallet")
	require.NoFileExists(t, ControlPath(target))
}

func TestStageRejectsOpenSourceProcess(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	options := testStageOptions(root, time.Now())
	options.OpenFileChecker = func(string) ([]int, error) { return []int{4242}, nil }

	_, err := Stage(source, target, options)
	require.ErrorContains(t, err, "process 4242")
	require.NoFileExists(t, ControlPath(target))
}

func TestStageRejectsSourceThatBecomesActiveDuringCopy(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	options := testStageOptions(root, time.Now())
	checks := 0
	options.OpenFileChecker = func(string) ([]int, error) {
		checks++
		if checks == 2 {
			return []int{4243}, nil
		}
		return nil, nil
	}

	_, err := Stage(source, target, options)
	require.ErrorContains(t, err, "became active")
	require.NoFileExists(t, ControlPath(target))
	require.NoFileExists(t, filepath.Join(target, "nwc.db"))
}

func TestStageRejectsChangingSource(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	options := testStageOptions(root, time.Now())
	options.AfterFirstHash = func() {
		require.NoError(t, updateFixture(source, "INSERT INTO transactions(id) VALUES (99)"))
	}

	_, err := Stage(source, target, options)
	require.ErrorContains(t, err, "changed during the stability check")
	require.NoFileExists(t, ControlPath(target))
}

func TestStageRejectsOverlappingAndExistingWalletTargets(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))

	_, err := Stage(source, source, testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "must be separate")
	_, err = Stage(source, filepath.Join(source, "nested-target"), testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "must be separate")

	target := filepath.Join(root, "configured-target")
	require.NoError(t, createWalletFixture(target, "different-wallet"))
	_, err = Stage(source, target, testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "different configured Grynvault wallet")
}

func TestStageAllowsFreshServiceDatabaseTarget(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "Library", "Containers", "com.grynvault.rebalance", "Data", "Library", "Application Support", "grynvault-rebalance")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, createFreshTargetFixture(target))

	result, err := Stage(source, target, testStageOptions(root, time.Now()))
	require.NoError(t, err)
	require.True(t, result.RestartRequired)
	require.FileExists(t, ControlPath(target))
}

func TestStageRejectsPartialOrCorruptTargetState(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))

	t.Run("setup marker without LDK storage", func(t *testing.T) {
		target := filepath.Join(root, "partial-target")
		require.NoError(t, createFreshTargetFixture(target))
		require.NoError(t, updateFixture(target, "INSERT INTO user_configs(key, value) VALUES ('Mnemonic', 'encrypted')"))
		_, err := Stage(source, target, testStageOptions(root, time.Now()))
		require.ErrorContains(t, err, "contains wallet state")
		require.NoFileExists(t, ControlPath(target))
	})

	t.Run("wallet activity without setup marker", func(t *testing.T) {
		target := filepath.Join(root, "active-target")
		require.NoError(t, createFreshTargetFixture(target))
		require.NoError(t, updateFixture(target, "INSERT INTO apps(id) VALUES (1)"))
		_, err := Stage(source, target, testStageOptions(root, time.Now()))
		require.ErrorContains(t, err, "contains wallet state")
		require.NoFileExists(t, ControlPath(target))
	})

	t.Run("partial LDK directory", func(t *testing.T) {
		target := filepath.Join(root, "ldk-target")
		require.NoError(t, createFreshTargetFixture(target))
		require.NoError(t, os.MkdirAll(filepath.Join(target, "ldk"), 0700))
		_, err := Stage(source, target, testStageOptions(root, time.Now()))
		require.ErrorContains(t, err, "contains wallet state")
		require.NoFileExists(t, ControlPath(target))
	})

	t.Run("corrupt startup database", func(t *testing.T) {
		target := filepath.Join(root, "corrupt-target")
		require.NoError(t, os.MkdirAll(target, 0700))
		require.NoError(t, os.WriteFile(filepath.Join(target, "nwc.db"), []byte("not sqlite"), 0600))
		_, err := Stage(source, target, testStageOptions(root, time.Now()))
		require.ErrorContains(t, err, "could not be proven pristine")
		require.NoFileExists(t, ControlPath(target))
	})
}

func TestStageRejectsPendingImportRestoreAndBackupSource(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(filepath.Join(target, "restore"), 0700))
	_, err := Stage(source, target, testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "pending restore")

	require.NoError(t, os.RemoveAll(filepath.Join(target, "restore")))
	require.NoError(t, WriteJSONAtomic(ControlPath(target), &Control{Version: ControlVersion}))
	_, err = Stage(source, target, testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "already pending")

	backupSource := filepath.Join(root, "grynvault-rebalance-migration-backups", "old", "albyhub")
	require.NoError(t, createWalletFixture(backupSource, "backup-wallet"))
	_, err = Stage(backupSource, filepath.Join(root, "another-target"), testStageOptions(root, time.Now()))
	require.ErrorContains(t, err, "backup or staged restore")
}

func TestFinishPendingResumesAfterTargetMove(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(target, "fresh"), []byte("target"), 0600))
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	_, err := Stage(source, target, testStageOptions(root, now))
	require.NoError(t, err)

	control, err := LoadControl(target)
	require.NoError(t, err)
	require.NoError(t, os.Rename(target, control.PreImportPath))

	finish, err := FinishPending(target, now.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, finish.Activated)
	require.FileExists(t, QuarantinePath(target))
}

func TestFinishPendingResumesAfterStagedWalletWasMovedIntoPlace(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(target, "before.txt"), []byte("preserved"), 0600))
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	_, err := Stage(source, target, testStageOptions(root, now))
	require.NoError(t, err)
	control, err := LoadControl(target)
	require.NoError(t, err)
	require.NoError(t, os.Rename(target, control.PreImportPath))
	require.NoError(t, os.Rename(control.StagePath, target))
	control.Phase = PhaseTargetMoved
	require.NoError(t, WriteJSONAtomic(ControlPath(target), control))

	finish, err := FinishPending(target, now.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, finish.Activated)
	require.FileExists(t, filepath.Join(control.BackupLocation, "grynvault-before-import", "before.txt"))
	require.FileExists(t, QuarantinePath(target))
}

func TestFinishPendingPreservesPostMoveRecoveryPhases(t *testing.T) {
	for _, phase := range []Phase{
		PhaseVerificationNeeded,
		PhaseVerified,
		PhaseActivationFailed,
		PhaseActivating,
		PhaseActivated,
	} {
		t.Run(string(phase), func(t *testing.T) {
			target := t.TempDir()
			control := &Control{
				Version:    ControlVersion,
				ImportID:   "checkpoint-" + string(phase),
				Phase:      phase,
				TargetPath: target,
			}
			require.NoError(t, WriteJSONAtomic(ControlPath(target), control))

			result, err := FinishPending(target, time.Now())
			require.NoError(t, err)
			require.True(t, result.Activated)
			persisted, err := LoadControl(target)
			require.NoError(t, err)
			require.Equal(t, phase, persisted.Phase)
		})
	}
}

func TestAbandonBlockedImportArchivesStageAndAllowsRetry(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(target, "original.txt"), []byte("preserved"), 0600))
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	result, err := Stage(source, target, testStageOptions(root, now))
	require.NoError(t, err)
	require.NoError(t, updateFixture(source, "INSERT INTO transactions(id) VALUES (44)"))
	blocked, err := FinishPending(target, now.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, blocked.Blocked)

	_, err = RequestAbandon(target, now.Add(2*time.Minute))
	require.NoError(t, err)
	finished, err := FinishPending(target, now.Add(3*time.Minute))
	require.NoError(t, err)
	require.False(t, finished.Activated)
	require.FileExists(t, filepath.Join(target, "original.txt"))
	require.NoFileExists(t, ControlPath(target))
	require.DirExists(t, filepath.Join(result.BackupLocation, "abandoned-staged-wallet"))
	require.FileExists(t, filepath.Join(result.BackupLocation, "abandoned-import-control.json"))
	evidence := Evidence{}
	require.NoError(t, ReadJSON(filepath.Join(result.BackupLocation, "import-evidence.json"), &evidence))
	require.Equal(t, PhaseAbandoned, evidence.Checkpoints[len(evidence.Checkpoints)-1].Phase)
}

func TestAbandonQuarantinedImportRestoresPreviousTarget(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.MkdirAll(target, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(target, "original.txt"), []byte("preserved"), 0600))
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	result, err := Stage(source, target, testStageOptions(root, now))
	require.NoError(t, err)
	_, err = FinishPending(target, now.Add(time.Minute))
	require.NoError(t, err)
	require.FileExists(t, QuarantinePath(target))

	_, err = RequestAbandon(target, now.Add(2*time.Minute))
	require.NoError(t, err)
	_, err = FinishPending(target, now.Add(3*time.Minute))
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(target, "original.txt"))
	require.NoFileExists(t, QuarantinePath(target))
	require.FileExists(t, filepath.Join(result.BackupLocation, "abandoned-imported-target", "nwc.db"))
	require.FileExists(t, filepath.Join(result.BackupLocation, "abandoned-import-control.json"))
}

func TestStagePreservesSQLiteSidecarsAndEvidenceIsRedacted(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-secret-value"))
	require.NoError(t, os.MkdirAll(target, 0700))

	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(source, "nwc.db"))
	require.NoError(t, err)
	_, err = database.Exec("PRAGMA journal_mode=WAL")
	require.NoError(t, err)
	_, err = database.Exec("INSERT INTO transactions(id) VALUES (8)")
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(source, "nwc.db-wal"))

	result, err := Stage(source, target, testStageOptions(root, time.Now()))
	require.NoError(t, err)
	require.NoError(t, database.Close())
	control, err := LoadControl(target)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(control.StagePath, "nwc.db-wal"))
	evidenceBytes, err := os.ReadFile(filepath.Join(result.BackupLocation, "import-evidence.json"))
	require.NoError(t, err)
	require.NotContains(t, string(evidenceBytes), "encrypted-wallet-secret-value")
	require.NotContains(t, strings.ToLower(string(evidenceBytes)), "password")

	evidence := Evidence{}
	require.NoError(t, json.Unmarshal(evidenceBytes, &evidence))
	require.Equal(t, EvidenceVersion, evidence.Version)
	require.Contains(t, evidence.SourceFiles, "nwc.db-wal")
}

func TestStaticChannelBackupSummary(t *testing.T) {
	workdir := t.TempDir()
	backupDir := filepath.Join(workdir, "ldk", "static_channel_backups")
	require.NoError(t, os.MkdirAll(backupDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(backupDir, "2026-01.json"), []byte("first"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(backupDir, "2026-02.json"), []byte("second"), 0600))
	count, latest, digest, err := staticChannelBackupSummary(workdir)
	require.NoError(t, err)
	require.Equal(t, 2, count)
	require.Equal(t, "2026-02.json", latest)
	require.Len(t, digest, 64)
}

func TestCopyWalletDataRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "destination")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-a"))
	require.NoError(t, os.WriteFile(filepath.Join(root, "outside"), []byte("outside"), 0600))
	require.NoError(t, os.Symlink(filepath.Join(root, "outside"), filepath.Join(source, "ldk", "storage", "unsafe")))

	_, err := CopyWalletData(source, destination)
	require.ErrorContains(t, err, "symbolic link")
}

func TestParseLsofProcessIDsExcludesCurrentAndDuplicates(t *testing.T) {
	require.Equal(t, []int{42, 77}, parseLsofProcessIDs("p42\np99\np42\np77\n", 99))
}

func TestInspectLDKOfflineDoesNotStartNode(t *testing.T) {
	workdir := t.TempDir()
	storage := filepath.Join(workdir, "ldk", "storage")
	require.NoError(t, os.MkdirAll(storage, 0700))
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	config := ldk_node.DefaultConfig()
	builder := ldk_node.BuilderFromConfig(config)
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)
	builder.SetNetwork(ldk_node.NetworkRegtest)
	builder.SetStorageDirPath(storage)
	node, err := builder.Build()
	require.NoError(t, err)
	expectedNodeID := node.NodeId()
	require.False(t, node.Status().IsRunning)
	node.Destroy()
	builder.Destroy()

	summary, err := InspectLDKOffline(workdir, mnemonic, "regtest")
	require.NoError(t, err)
	require.Equal(t, expectedNodeID, summary.NodeID)
	require.Zero(t, summary.ChannelCount)
	require.Zero(t, summary.StoredPaymentCount)
}

func testStageOptions(home string, now time.Time) StageOptions {
	return StageOptions{
		Now:              func() time.Time { return now },
		StabilityWait:    0,
		OpenFileChecker:  func(string) ([]int, error) { return nil, nil },
		LocalPathChecker: func(string) (bool, error) { return true, nil },
		HomeDirectory:    home,
	}
}

func evaluatedPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)
	return resolved
}

func createWalletFixture(workdir, encryptedMnemonic string) error {
	if err := os.MkdirAll(filepath.Join(workdir, "ldk", "storage"), 0700); err != nil {
		return err
	}
	nwc, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(workdir, "nwc.db"))
	if err != nil {
		return err
	}
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
		"INSERT INTO user_configs(key, value) VALUES ('UnlockPasswordCheck', 'encrypted')",
		"INSERT INTO user_configs(key, value) VALUES ('LNBackendType', 'LDK')",
		"INSERT INTO user_configs(key, value) VALUES ('Mnemonic', '" + encryptedMnemonic + "')",
	}
	for _, statement := range statements {
		if _, err = nwc.Exec(statement); err != nil {
			_ = nwc.Close()
			return err
		}
	}
	if err = nwc.Close(); err != nil {
		return err
	}
	return createSQLiteFixture(filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite"))
}

func createFreshTargetFixture(workdir string) error {
	if err := os.MkdirAll(workdir, 0700); err != nil {
		return err
	}
	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(workdir, "nwc.db"))
	if err != nil {
		return err
	}
	defer database.Close()
	for _, statement := range []string{
		"CREATE TABLE user_configs (key TEXT PRIMARY KEY, value TEXT)",
		"CREATE TABLE apps (id INTEGER PRIMARY KEY)",
		"CREATE TABLE app_permissions (id INTEGER PRIMARY KEY)",
		"CREATE TABLE transactions (id INTEGER PRIMARY KEY)",
		"CREATE TABLE swaps (id INTEGER PRIMARY KEY)",
		"CREATE TABLE forwards (id INTEGER PRIMARY KEY)",
		"CREATE TABLE request_events (id INTEGER PRIMARY KEY)",
		"CREATE TABLE response_events (id INTEGER PRIMARY KEY)",
		"CREATE TABLE local_rebalance_quotes (id INTEGER PRIMARY KEY)",
		"CREATE TABLE rebalance_quotes (id INTEGER PRIMARY KEY)",
		"INSERT INTO user_configs(key, value) VALUES ('Relay', 'wss://relay.example')",
	} {
		if _, err = database.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func createSQLiteFixture(path string) error {
	ldk, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, path)
	if err != nil {
		return err
	}
	if _, err = ldk.Exec("CREATE TABLE kv (key TEXT PRIMARY KEY, value BLOB)"); err != nil {
		_ = ldk.Close()
		return err
	}
	return ldk.Close()
}

func updateFixture(workdir, statement string) error {
	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, filepath.Join(workdir, "nwc.db"))
	if err != nil {
		return err
	}
	defer database.Close()
	_, err = database.Exec(statement)
	return err
}
