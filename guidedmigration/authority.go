package guidedmigration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const AuthorityVersion = 1

func AuthorityPath(targetWorkdir string) string {
	return filepath.Join(targetWorkdir, AuthorityFileName)
}

func RetirementPath(sourceWorkdir, importID string) string {
	return filepath.Join(
		filepath.Dir(filepath.Clean(sourceWorkdir)),
		filepath.Base(filepath.Clean(sourceWorkdir))+".grynvault-retired-"+importID,
	)
}

func sourceRetirementPaths(record *Quarantine) (string, string, error) {
	if record == nil || record.ImportID == "" || record.SourcePath == "" || record.TargetPath == "" {
		return "", "", errors.New("migration source-retirement record is incomplete")
	}
	source, err := filepath.Abs(record.SourcePath)
	if err != nil {
		return "", "", fmt.Errorf("resolve migration source: %w", err)
	}
	retired := RetirementPath(source, record.ImportID)
	if record.RetiredSourcePath != "" && !samePath(retired, record.RetiredSourcePath) {
		return "", "", errors.New("migration source-retirement path does not match the import")
	}
	if PathsOverlap(source, record.TargetPath) || PathsOverlap(retired, record.TargetPath) {
		return "", "", errors.New("migration source-retirement path overlaps the Grynvault wallet")
	}
	return filepath.Clean(source), filepath.Clean(retired), nil
}

// RetireSourceWallet atomically removes the original Alby Hub wallet from its
// runnable location. The sibling rename stays on the same filesystem and is
// idempotent so an interrupted activation can be recovered without guessing.
func RetireSourceWallet(record *Quarantine, now time.Time) error {
	source, retired, err := sourceRetirementPaths(record)
	if err != nil {
		return err
	}
	record.RetiredSourcePath = retired
	sourceExists, err := pathExists(source)
	if err != nil {
		return fmt.Errorf("inspect source wallet before retirement: %w", err)
	}
	retiredExists, err := pathExists(retired)
	if err != nil {
		return fmt.Errorf("inspect retired source wallet: %w", err)
	}
	if sourceExists && retiredExists {
		return errors.New("source-wallet retirement is ambiguous: both active and retired paths exist")
	}
	if !sourceExists && !retiredExists {
		return errors.New("source-wallet retirement is incomplete: neither active nor retired path exists")
	}

	verifyPath := source
	if retiredExists {
		verifyPath = retired
	}
	processIDs, err := OpenWalletProcessIDs(verifyPath)
	if err != nil {
		return fmt.Errorf("verify source wallet is stopped: %w", err)
	}
	if len(processIDs) > 0 {
		return fmt.Errorf("source wallet is still open by process %s", joinProcessIDs(processIDs))
	}
	hashes, err := HashWalletData(verifyPath)
	if err != nil {
		return fmt.Errorf("verify source wallet before retirement: %w", err)
	}
	if err = CompareHashes(record.SourceFiles, hashes); err != nil {
		return fmt.Errorf("source wallet changed before retirement: %w", err)
	}

	if sourceExists {
		if err = os.Rename(source, retired); err != nil {
			return fmt.Errorf("retire source wallet atomically: %w", err)
		}
		if err = syncDirectory(filepath.Dir(source)); err != nil {
			return fmt.Errorf("persist source-wallet retirement: %w", err)
		}
	}
	retiredHashes, err := HashWalletData(retired)
	if err != nil {
		return fmt.Errorf("verify retired source wallet: %w", err)
	}
	if err = CompareHashes(record.SourceFiles, retiredHashes); err != nil {
		return fmt.Errorf("retired source wallet differs from verified source: %w", err)
	}
	record.SourceRetiredAt = timestamp(now)
	return nil
}

// RestoreSourceWallet reverses a retirement only when the original path is
// absent and the retired bytes still match the verified migration source.
func RestoreSourceWallet(record *Quarantine) error {
	source, retired, err := sourceRetirementPaths(record)
	if err != nil {
		return err
	}
	sourceExists, err := pathExists(source)
	if err != nil {
		return err
	}
	retiredExists, err := pathExists(retired)
	if err != nil {
		return err
	}
	if sourceExists && !retiredExists {
		record.SourceRetiredAt = ""
		return nil
	}
	if sourceExists || !retiredExists {
		return errors.New("source-wallet restore is ambiguous")
	}
	hashes, err := HashWalletData(retired)
	if err != nil {
		return fmt.Errorf("verify retired source before restore: %w", err)
	}
	if err = CompareHashes(record.SourceFiles, hashes); err != nil {
		return fmt.Errorf("retired source changed before restore: %w", err)
	}
	if err = os.Rename(retired, source); err != nil {
		return fmt.Errorf("restore source wallet atomically: %w", err)
	}
	if err = syncDirectory(filepath.Dir(source)); err != nil {
		return fmt.Errorf("persist source-wallet restore: %w", err)
	}
	record.SourceRetiredAt = ""
	return nil
}

func WriteAuthority(record *Quarantine, now time.Time) error {
	if record == nil || record.RetiredSourcePath == "" || record.SourceRetiredAt == "" {
		return errors.New("source wallet must be retired before recording wallet authority")
	}
	authority := Authority{
		Version:           AuthorityVersion,
		ImportID:          record.ImportID,
		TargetPath:        filepath.Clean(record.TargetPath),
		SourcePath:        filepath.Clean(record.SourcePath),
		RetiredSourcePath: filepath.Clean(record.RetiredSourcePath),
		BackupLocation:    filepath.Clean(record.BackupLocation),
		WalletFingerprint: record.WalletFingerprint,
		SourceFiles:       record.SourceFiles,
		ActivatedAt:       timestamp(now),
	}
	return WriteJSONAtomic(AuthorityPath(record.TargetPath), &authority)
}

func LoadAuthority(targetWorkdir string) (*Authority, error) {
	authority := Authority{}
	if err := ReadJSON(AuthorityPath(targetWorkdir), &authority); err != nil {
		return nil, err
	}
	if authority.Version != AuthorityVersion || !samePath(authority.TargetPath, targetWorkdir) {
		return nil, errors.New("wallet authority record is invalid")
	}
	return &authority, nil
}

// EnsureWalletAuthority prevents Grynvault from starting when the original
// Alby Hub location has regained a runnable copy of the migrated node.
func EnsureWalletAuthority(targetWorkdir string) error {
	authority, err := LoadAuthority(targetWorkdir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check wallet authority: %w", err)
	}

	if hasWalletMarkers(authority.SourcePath) {
		source, inspectErr := InspectSource(authority.SourcePath)
		if inspectErr != nil {
			return fmt.Errorf("the retired Alby Hub location contains wallet data that could not be validated; Lightning remains offline: %w", inspectErr)
		}
		if source.WalletFingerprint == authority.WalletFingerprint {
			return errors.New("the original Alby Hub location contains another runnable copy of this node; Lightning remains offline until the duplicate is retired")
		}
	}

	processIDs, err := OpenWalletProcessIDs(authority.RetiredSourcePath)
	if err != nil {
		return fmt.Errorf("check retired Alby Hub wallet process state: %w", err)
	}
	if len(processIDs) > 0 {
		return fmt.Errorf("the retired Alby Hub wallet is open by process %s; Lightning remains offline", joinProcessIDs(processIDs))
	}
	retired, inspectErr := InspectSource(authority.RetiredSourcePath)
	if inspectErr != nil {
		return fmt.Errorf("verify retired Alby Hub wallet: %w", inspectErr)
	}
	if retired.WalletFingerprint != authority.WalletFingerprint {
		return errors.New("retired Alby Hub wallet identity no longer matches the authoritative Grynvault wallet")
	}
	retiredHashes, err := HashWalletData(authority.RetiredSourcePath)
	if err != nil {
		return fmt.Errorf("hash retired Alby Hub wallet: %w", err)
	}
	if err = CompareHashes(authority.SourceFiles, retiredHashes); err != nil {
		return fmt.Errorf("retired Alby Hub wallet changed after activation; Lightning remains offline: %w", err)
	}
	return nil
}

// PrepareLegacyAuthorityUpgrade turns a pre-authority guided migration into a
// one-time offline verification flow. It only proceeds automatically when the
// original source still exactly matches the accepted migration bytes.
func PrepareLegacyAuthorityUpgrade(targetWorkdir string, now time.Time) (*Quarantine, error) {
	if _, err := LoadAuthority(targetWorkdir); err == nil {
		return nil, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	backupParent := filepath.Join(filepath.Dir(filepath.Clean(targetWorkdir)), "grynvault-rebalance-migration-backups")
	entries, err := os.ReadDir(backupParent)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect prior migration evidence: %w", err)
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Name() > entries[right].Name() })
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		acceptancePath := filepath.Join(backupParent, entry.Name(), "migration-acceptance.json")
		record := Quarantine{}
		if readErr := ReadJSON(acceptancePath, &record); errors.Is(readErr, os.ErrNotExist) {
			continue
		} else if readErr != nil {
			return nil, fmt.Errorf("read prior migration acceptance: %w", readErr)
		}
		if record.Phase != PhaseActivated || !samePath(record.TargetPath, targetWorkdir) || record.SourcePath == "" || record.WalletFingerprint == "" {
			continue
		}
		if !hasWalletMarkers(record.SourcePath) {
			return nil, nil
		}
		source, inspectErr := InspectSource(record.SourcePath)
		if inspectErr != nil {
			return nil, fmt.Errorf("inspect source from prior migration acceptance: %w", inspectErr)
		}
		if source.WalletFingerprint != record.WalletFingerprint {
			return nil, nil
		}
		currentHashes, hashErr := HashWalletData(record.SourcePath)
		if hashErr != nil {
			return nil, fmt.Errorf("hash source from prior migration acceptance: %w", hashErr)
		}

		record.AuthorityUpgrade = true
		record.RetiredSourcePath = RetirementPath(record.SourcePath, record.ImportID)
		record.SourceRetiredAt = ""
		record.Verification = nil
		record.ActivationAttemptedAt = ""
		record.UpdatedAt = timestamp(now)
		if compareErr := CompareHashes(record.SourceFiles, currentHashes); compareErr != nil {
			record.Phase = PhaseBlocked
			record.LastError = "the original Alby Hub wallet changed after the previous migration; Grynvault will not choose an authoritative copy automatically"
		} else {
			record.Phase = PhaseVerificationNeeded
			record.LastError = "finish the one-time wallet authority upgrade before reconnecting Lightning"
		}
		if err = WriteJSONAtomic(QuarantinePath(targetWorkdir), &record); err != nil {
			return nil, fmt.Errorf("persist wallet authority upgrade: %w", err)
		}
		return &record, nil
	}
	return nil, nil
}

func AppendEvidenceCheckpoint(record *Quarantine, phase Phase, detail string, now time.Time) error {
	if record == nil || record.EvidencePath == "" {
		return errors.New("migration evidence location is missing")
	}
	evidence := Evidence{}
	if err := ReadJSON(record.EvidencePath, &evidence); err != nil {
		return err
	}
	if evidence.ImportID != record.ImportID || !samePath(evidence.TargetPath, record.TargetPath) {
		return errors.New("migration evidence does not match wallet authority record")
	}
	evidence.Checkpoints = append(evidence.Checkpoints, Checkpoint{
		Phase: phase, CreatedAt: timestamp(now), Detail: detail,
	})
	return WriteJSONAtomic(record.EvidencePath, &evidence)
}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
