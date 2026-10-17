package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/getAlby/hub/guidedmigration"
)

type MigrationStatusProvider interface {
	GetGuidedMigrationStatus() (*guidedmigration.Status, error)
}

type MigrationGuard interface {
	IsGuidedMigrationQuarantined() bool
}

type MigrationVerifier interface {
	VerifyGuidedMigration(unlockPassword string) (*guidedmigration.VerificationSummary, error)
	ActivateGuidedMigration(unlockPassword, nodeIDConfirmation, activationConfirmation string) error
}

const guidedMigrationVerificationTTL = 5 * time.Minute

func (svc *service) isMigrationQuarantined() bool {
	svc.migrationMu.Lock()
	defer svc.migrationMu.Unlock()
	return svc.migrationQuarantine != nil && svc.migrationQuarantine.Phase != guidedmigration.PhaseActivated
}

func (svc *service) IsGuidedMigrationQuarantined() bool {
	return svc.isMigrationQuarantined()
}

func (svc *service) GetGuidedMigrationStatus() (*guidedmigration.Status, error) {
	return guidedmigration.StatusFor(svc.cfg.GetEnv().Workdir)
}

func (svc *service) VerifyGuidedMigration(unlockPassword string) (*guidedmigration.VerificationSummary, error) {
	svc.migrationMu.Lock()
	defer svc.migrationMu.Unlock()

	record := svc.migrationQuarantine
	if record == nil || record.Phase == guidedmigration.PhaseActivated {
		return nil, errors.New("no guided migration is awaiting verification")
	}
	if record.Phase == guidedmigration.PhaseBlocked || record.Phase == guidedmigration.PhaseAbandoning {
		return nil, errors.New("the blocked migration must be safely archived before another verification attempt")
	}
	if record.Phase == guidedmigration.PhaseActivating {
		return nil, errors.New("the previous activation did not finish; restart Grynvault Rebalance before verifying again")
	}
	summary, err := svc.buildGuidedMigrationVerification(record, unlockPassword)
	if err != nil {
		return nil, err
	}

	record.Phase = guidedmigration.PhaseVerified
	record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	record.Verification = summary
	record.LastError = ""
	workdir := svc.cfg.GetEnv().Workdir
	if err = guidedmigration.WriteJSONAtomic(guidedmigration.QuarantinePath(workdir), record); err != nil {
		return nil, fmt.Errorf("persist migration verification: %w", err)
	}
	return summary, nil
}

func (svc *service) ActivateGuidedMigration(unlockPassword, nodeIDConfirmation, activationConfirmation string) error {
	svc.migrationMu.Lock()
	defer svc.migrationMu.Unlock()

	record := svc.migrationQuarantine
	if !guidedMigrationReadyForActivation(record) {
		return errors.New("complete migration verification before activation")
	}
	now := time.Now()
	if !guidedMigrationVerificationIsFresh(record, now) {
		record.Phase = guidedmigration.PhaseVerificationNeeded
		record.LastError = "offline migration verification expired; verify the wallet again before activation"
		record.UpdatedAt = now.UTC().Format(time.RFC3339Nano)
		if err := guidedmigration.WriteJSONAtomic(
			guidedmigration.QuarantinePath(svc.cfg.GetEnv().Workdir),
			record,
		); err != nil {
			return fmt.Errorf("persist expired migration verification: %w", err)
		}
		return errors.New(record.LastError)
	}
	expectedConfirmation := guidedmigration.ActivationConfirmation(record.Verification.Network, record.WalletFingerprint)
	if activationConfirmation == "" || activationConfirmation != expectedConfirmation {
		return errors.New("activation confirmation does not match the verified wallet and network")
	}
	if nodeIDConfirmation == "" || nodeIDConfirmation != record.Verification.NodeID {
		return errors.New("node identity confirmation does not match the verified wallet")
	}
	freshVerification, err := svc.buildGuidedMigrationVerification(record, unlockPassword)
	if err != nil {
		return fmt.Errorf("final offline continuity check failed: %w", err)
	}
	if !sameGuidedMigrationContinuity(record.Verification, freshVerification) {
		return errors.New("wallet continuity changed after review; activation remains offline")
	}
	processIDs, err := guidedmigration.OpenWalletProcessIDs(record.SourcePath)
	if err != nil {
		return fmt.Errorf("final source-wallet process check failed; activation remains offline: %w", err)
	}
	if len(processIDs) > 0 {
		return fmt.Errorf("the source Alby Hub data became active in process %v; activation remains offline", processIDs)
	}

	workdir := svc.cfg.GetEnv().Workdir
	record.Phase = guidedmigration.PhaseActivating
	record.ActivationAttemptedAt = now.UTC().Format(time.RFC3339Nano)
	record.UpdatedAt = record.ActivationAttemptedAt
	if err = guidedmigration.WriteJSONAtomic(guidedmigration.QuarantinePath(workdir), record); err != nil {
		return err
	}

	err = svc.WithStartLock(func() error {
		return svc.startAppInternal(unlockPassword)
	})
	if err != nil {
		record.Phase = guidedmigration.PhaseActivationFailed
		record.LastError = err.Error()
		record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if writeErr := guidedmigration.WriteJSONAtomic(guidedmigration.QuarantinePath(workdir), record); writeErr != nil {
			return fmt.Errorf("activation failed: %v; additionally failed to preserve quarantine: %w", err, writeErr)
		}
		return err
	}

	activatedRecord := *record
	activatedRecord.Phase = guidedmigration.PhaseActivated
	activatedRecord.LastError = ""
	activatedRecord.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	acceptancePath := filepath.Join(record.BackupLocation, "migration-acceptance.json")
	if err = guidedmigration.WriteJSONAtomic(acceptancePath, &activatedRecord); err != nil {
		svc.stopAppInternal()
		return fmt.Errorf("preserve migration acceptance: %w", err)
	}
	if err = os.Remove(guidedmigration.ControlPath(workdir)); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(acceptancePath)
		svc.stopAppInternal()
		return fmt.Errorf("remove guided import control: %w", err)
	}
	if err = os.Remove(guidedmigration.QuarantinePath(workdir)); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(acceptancePath)
		svc.stopAppInternal()
		return fmt.Errorf("clear migration quarantine: %w", err)
	}
	record.Phase = guidedmigration.PhaseActivated
	svc.migrationQuarantine = nil
	return nil
}

func guidedMigrationReadyForActivation(record *guidedmigration.Quarantine) bool {
	if record == nil || record.Verification == nil {
		return false
	}
	return record.Phase == guidedmigration.PhaseVerified
}

func guidedMigrationVerificationIsFresh(record *guidedmigration.Quarantine, now time.Time) bool {
	if !guidedMigrationReadyForActivation(record) {
		return false
	}
	verifiedAt, err := time.Parse(time.RFC3339Nano, record.Verification.VerifiedAt)
	if err != nil || verifiedAt.After(now.Add(30*time.Second)) || now.Sub(verifiedAt) > guidedMigrationVerificationTTL {
		return false
	}
	if record.ActivationAttemptedAt == "" {
		return true
	}
	attemptedAt, err := time.Parse(time.RFC3339Nano, record.ActivationAttemptedAt)
	return err == nil && verifiedAt.After(attemptedAt)
}

func (svc *service) buildGuidedMigrationVerification(record *guidedmigration.Quarantine, unlockPassword string) (*guidedmigration.VerificationSummary, error) {
	if !svc.cfg.CheckUnlockPassword(unlockPassword) {
		return nil, ErrInvalidPassword
	}
	mnemonic, err := svc.cfg.Get("Mnemonic", unlockPassword)
	if err != nil {
		return nil, fmt.Errorf("decrypt imported wallet identity: %w", err)
	}

	workdir := svc.cfg.GetEnv().Workdir
	_, counts, backendType, fingerprint, err := guidedmigration.ValidateNWCDB(filepath.Join(workdir, "nwc.db"))
	if err != nil {
		return nil, err
	}
	if backendType != "LDK" {
		return nil, errors.New("guided migration verification supports embedded LDK only")
	}
	if fingerprint != record.WalletFingerprint {
		return nil, errors.New("imported wallet identity no longer matches the staged source")
	}
	if counts.Apps != record.SourceCounts.Apps ||
		counts.Permissions != record.SourceCounts.Permissions ||
		counts.Transactions != record.SourceCounts.Transactions ||
		counts.RebalanceRecords != record.SourceCounts.RebalanceRecords {
		return nil, errors.New("imported wallet continuity counts differ from the staged source")
	}

	sourceHashes, err := guidedmigration.HashWalletData(record.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("recheck original Alby Hub data: %w", err)
	}
	if err = guidedmigration.CompareHashes(record.SourceFiles, sourceHashes); err != nil {
		return nil, fmt.Errorf("original Alby Hub data changed after import; activation remains blocked: %w", err)
	}

	summary, err := guidedmigration.InspectLDKOffline(workdir, mnemonic, svc.cfg.GetNetwork())
	if err != nil {
		return nil, err
	}
	summary.Apps = counts.Apps
	summary.Permissions = counts.Permissions
	summary.Transactions = counts.Transactions
	summary.RebalanceRecords = counts.RebalanceRecords
	summary.SourceBytesStillMatch = true
	summary.DestinationBytesStillMatch = record.ActivatedBytesVerified
	if !summary.DestinationBytesStillMatch {
		return nil, errors.New("the activated wallet copy lacks byte-verification evidence")
	}
	return summary, nil
}

func sameGuidedMigrationContinuity(expected, actual *guidedmigration.VerificationSummary) bool {
	if expected == nil || actual == nil {
		return false
	}
	return expected.NodeID == actual.NodeID &&
		expected.Network == actual.Network &&
		expected.ChannelCount == actual.ChannelCount &&
		expected.StoredPaymentCount == actual.StoredPaymentCount &&
		expected.TotalOnchainSats == actual.TotalOnchainSats &&
		expected.SpendableOnchainSats == actual.SpendableOnchainSats &&
		expected.TotalLightningSats == actual.TotalLightningSats &&
		expected.PendingClosureSats == actual.PendingClosureSats &&
		expected.Apps == actual.Apps &&
		expected.Permissions == actual.Permissions &&
		expected.Transactions == actual.Transactions &&
		expected.RebalanceRecords == actual.RebalanceRecords &&
		expected.StaticChannelBackupFiles == actual.StaticChannelBackupFiles &&
		expected.LatestStaticChannelBackup == actual.LatestStaticChannelBackup &&
		expected.LatestStaticBackupSHA256 == actual.LatestStaticBackupSHA256 &&
		reflect.DeepEqual(expected.Channels, actual.Channels)
}
