package guidedmigration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func authorityFixture(t *testing.T) (*Quarantine, map[string]string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "albyhub")
	target := filepath.Join(root, "grynvault-rebalance")
	require.NoError(t, createWalletFixture(source, "encrypted-wallet-authority"))
	require.NoError(t, createWalletFixture(target, "encrypted-wallet-authority"))
	hashes, err := HashWalletData(source)
	require.NoError(t, err)
	summary, err := InspectSource(source)
	require.NoError(t, err)
	record := &Quarantine{
		Version:           QuarantineVersion,
		ImportID:          "authority-test",
		SourcePath:        source,
		TargetPath:        target,
		BackupLocation:    filepath.Join(root, "backup"),
		SourceFiles:       hashes,
		WalletFingerprint: summary.WalletFingerprint,
	}
	return record, hashes
}

func TestRetireSourceWalletIsAtomicIdempotentAndRestorable(t *testing.T) {
	record, hashes := authorityFixture(t)
	now := time.Date(2026, time.October, 20, 20, 0, 0, 0, time.UTC)

	require.NoError(t, RetireSourceWallet(record, now))
	require.NoDirExists(t, record.SourcePath)
	require.DirExists(t, record.RetiredSourcePath)
	require.NotEmpty(t, record.SourceRetiredAt)
	retiredHashes, err := HashWalletData(record.RetiredSourcePath)
	require.NoError(t, err)
	require.NoError(t, CompareHashes(hashes, retiredHashes))

	// A crash after the rename but before the next checkpoint can safely retry.
	require.NoError(t, RetireSourceWallet(record, now.Add(time.Minute)))
	require.NoDirExists(t, record.SourcePath)
	require.DirExists(t, record.RetiredSourcePath)

	require.NoError(t, RestoreSourceWallet(record))
	require.DirExists(t, record.SourcePath)
	require.NoDirExists(t, record.RetiredSourcePath)
	require.Empty(t, record.SourceRetiredAt)
	restoredHashes, err := HashWalletData(record.SourcePath)
	require.NoError(t, err)
	require.NoError(t, CompareHashes(hashes, restoredHashes))
}

func TestRetireSourceWalletFailsClosedForAmbiguousPaths(t *testing.T) {
	record, _ := authorityFixture(t)
	record.RetiredSourcePath = RetirementPath(record.SourcePath, record.ImportID)
	require.NoError(t, os.MkdirAll(record.RetiredSourcePath, 0700))

	err := RetireSourceWallet(record, time.Now())
	require.ErrorContains(t, err, "both active and retired paths exist")
	require.DirExists(t, record.SourcePath)
	require.DirExists(t, record.RetiredSourcePath)
}

func TestRetireSourceWalletRejectsChangedSource(t *testing.T) {
	record, _ := authorityFixture(t)
	require.NoError(t, updateFixture(record.SourcePath, "INSERT INTO transactions(id) VALUES (91)"))

	err := RetireSourceWallet(record, time.Now())
	require.ErrorContains(t, err, "source wallet changed before retirement")
	require.DirExists(t, record.SourcePath)
	require.NoDirExists(t, RetirementPath(record.SourcePath, record.ImportID))
}

func TestWalletAuthorityBlocksRecreatedDuplicateSource(t *testing.T) {
	record, _ := authorityFixture(t)
	now := time.Date(2026, time.October, 20, 20, 0, 0, 0, time.UTC)
	require.NoError(t, RetireSourceWallet(record, now))
	require.NoError(t, WriteAuthority(record, now))
	require.NoError(t, EnsureWalletAuthority(record.TargetPath))

	// Recreating the old location with the same wallet must never allow the
	// Grynvault node to start alongside it.
	_, err := CopyWalletData(record.RetiredSourcePath, record.SourcePath)
	require.NoError(t, err)
	err = EnsureWalletAuthority(record.TargetPath)
	require.ErrorContains(t, err, "another runnable copy of this node")
}

func TestWalletAuthorityAllowsDifferentWalletAtFormerSource(t *testing.T) {
	record, _ := authorityFixture(t)
	now := time.Date(2026, time.October, 20, 20, 0, 0, 0, time.UTC)
	require.NoError(t, RetireSourceWallet(record, now))
	require.NoError(t, WriteAuthority(record, now))
	require.NoError(t, createWalletFixture(record.SourcePath, "different-wallet"))

	require.NoError(t, EnsureWalletAuthority(record.TargetPath))
}

func TestWalletAuthorityBlocksRetiredWalletMutation(t *testing.T) {
	record, _ := authorityFixture(t)
	now := time.Date(2026, time.October, 20, 20, 0, 0, 0, time.UTC)
	require.NoError(t, RetireSourceWallet(record, now))
	require.NoError(t, WriteAuthority(record, now))
	require.NoError(t, updateFixture(record.RetiredSourcePath, "INSERT INTO transactions(id) VALUES (93)"))

	err := EnsureWalletAuthority(record.TargetPath)
	require.ErrorContains(t, err, "retired Alby Hub wallet changed after activation")
}

func TestPrepareLegacyAuthorityUpgradeQuarantinesUnchangedSource(t *testing.T) {
	record, _ := authorityFixture(t)
	record.Phase = PhaseActivated
	record.ActivatedBytesVerified = true
	record.BackupLocation = filepath.Join(filepath.Dir(record.TargetPath), "grynvault-rebalance-migration-backups", "20261020-legacy")
	record.EvidencePath = filepath.Join(record.BackupLocation, "import-evidence.json")
	require.NoError(t, os.MkdirAll(record.BackupLocation, 0700))
	require.NoError(t, WriteJSONAtomic(filepath.Join(record.BackupLocation, "migration-acceptance.json"), record))

	upgrade, err := PrepareLegacyAuthorityUpgrade(record.TargetPath, time.Now())
	require.NoError(t, err)
	require.NotNil(t, upgrade)
	require.True(t, upgrade.AuthorityUpgrade)
	require.Equal(t, PhaseVerificationNeeded, upgrade.Phase)
	require.Empty(t, upgrade.Verification)
	require.Equal(t, RetirementPath(record.SourcePath, record.ImportID), upgrade.RetiredSourcePath)
	require.FileExists(t, QuarantinePath(record.TargetPath))
}

func TestPrepareLegacyAuthorityUpgradeBlocksChangedSource(t *testing.T) {
	record, _ := authorityFixture(t)
	record.Phase = PhaseActivated
	record.BackupLocation = filepath.Join(filepath.Dir(record.TargetPath), "grynvault-rebalance-migration-backups", "20261020-legacy")
	require.NoError(t, os.MkdirAll(record.BackupLocation, 0700))
	require.NoError(t, WriteJSONAtomic(filepath.Join(record.BackupLocation, "migration-acceptance.json"), record))
	require.NoError(t, updateFixture(record.SourcePath, "INSERT INTO transactions(id) VALUES (92)"))

	upgrade, err := PrepareLegacyAuthorityUpgrade(record.TargetPath, time.Now())
	require.NoError(t, err)
	require.NotNil(t, upgrade)
	require.Equal(t, PhaseBlocked, upgrade.Phase)
	require.Contains(t, upgrade.LastError, "will not choose an authoritative copy automatically")
}
