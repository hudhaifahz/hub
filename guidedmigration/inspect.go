package guidedmigration

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	sqlite_wrapper "github.com/getAlby/hub/db/sqlite-wrapper"
)

func ResolveWorkdir(selected string) (string, error) {
	if selected == "" {
		return "", errors.New("no Alby Hub data folder was selected")
	}
	selected, err := filepath.Abs(selected)
	if err != nil {
		return "", fmt.Errorf("resolve selected folder: %w", err)
	}
	candidates := workdirCandidates(selected)
	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if isRegularFile(filepath.Join(candidate, "nwc.db")) && isRegularFile(filepath.Join(candidate, "ldk", "storage", "ldk_node_data.sqlite")) {
			resolved, resolveErr := filepath.EvalSymlinks(candidate)
			if resolveErr != nil {
				return "", fmt.Errorf("resolve selected wallet folder: %w", resolveErr)
			}
			return filepath.Clean(resolved), nil
		}
	}
	return "", errors.New("the selected folder is not an Alby Hub embedded-LDK wallet data folder")
}

func workdirCandidates(selected string) []string {
	return []string{
		selected,
		filepath.Join(selected, "Data", "Library", "Application Support", "albyhub"),
		filepath.Join(selected, "Library", "Application Support", "albyhub"),
		filepath.Join(selected, "Application Support", "albyhub"),
		filepath.Join(selected, "albyhub"),
	}
}

func isRegularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular()
}

func InspectSource(selected string) (*SourceSummary, error) {
	resolved, err := ResolveWorkdir(selected)
	if err != nil {
		return nil, err
	}
	if IsBackupLikePath(resolved) {
		return nil, errors.New("the selected folder looks like a backup or staged restore; use the disaster-recovery flow instead")
	}
	temp, err := os.MkdirTemp("", "grynvault-import-inspect-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temp)
	hashes, err := CopyWalletData(resolved, temp)
	if err != nil {
		return nil, fmt.Errorf("snapshot selected Alby Hub data: %w", err)
	}
	nwcIntegrity, counts, backend, fingerprint, ldkIntegrity, err := ValidateWalletDataWithoutMutation(temp)
	if err != nil {
		return nil, err
	}
	validatedHashes, err := HashWalletData(temp)
	if err != nil {
		return nil, err
	}
	if err = CompareHashes(hashes, validatedHashes); err != nil {
		return nil, fmt.Errorf("verify inspected wallet snapshot: %w", err)
	}
	latest, err := latestDurableChange(resolved, hashes)
	if err != nil {
		return nil, err
	}
	return &SourceSummary{
		SourcePath:            resolved,
		BackendType:           backend,
		WalletFingerprint:     fingerprint,
		WalletFingerprintHint: FingerprintHint(fingerprint),
		NWCIntegrity:          nwcIntegrity,
		LDKIntegrity:          ldkIntegrity,
		Counts:                counts,
		DurableFileCount:      len(hashes),
		LatestDurableChange:   latest,
		Ready:                 true,
	}, nil
}

func ValidateWalletDataWithoutMutation(workdir string) (string, Counts, string, string, string, error) {
	validationCopy, err := os.MkdirTemp("", "grynvault-wallet-validation-")
	if err != nil {
		return "", Counts{}, "", "", "", err
	}
	defer os.RemoveAll(validationCopy)
	sourceHashes, err := CopyWalletData(workdir, validationCopy)
	if err != nil {
		return "", Counts{}, "", "", "", fmt.Errorf("create disposable validation copy: %w", err)
	}
	nwcIntegrity, counts, backend, fingerprint, err := ValidateNWCDB(filepath.Join(validationCopy, "nwc.db"))
	if err != nil {
		return "", Counts{}, "", "", "", err
	}
	ldkIntegrity, err := ValidateSQLite(filepath.Join(validationCopy, "ldk", "storage", "ldk_node_data.sqlite"))
	if err != nil {
		return "", Counts{}, "", "", "", fmt.Errorf("validate copied LDK database: %w", err)
	}
	if err = RemoveGeneratedSQLiteSidecars(validationCopy, sourceHashes); err != nil {
		return "", Counts{}, "", "", "", fmt.Errorf("clean disposable validation sidecars: %w", err)
	}
	return nwcIntegrity, counts, backend, fingerprint, ldkIntegrity, nil
}

func ValidateSQLite(path string) (string, error) {
	dsn := (&url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}).String()
	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, dsn)
	if err != nil {
		return "", err
	}
	defer database.Close()
	var result string
	if err = database.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return "", err
	}
	if result != "ok" {
		return result, fmt.Errorf("SQLite integrity check returned %q", result)
	}
	return result, nil
}

func ValidateNWCDB(path string) (string, Counts, string, string, error) {
	integrity, err := ValidateSQLite(path)
	if err != nil {
		return "", Counts{}, "", "", fmt.Errorf("validate copied Alby Hub database: %w", err)
	}
	dsn := (&url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}).String()
	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, dsn)
	if err != nil {
		return "", Counts{}, "", "", err
	}
	defer database.Close()

	var setupKeys int64
	if err = database.QueryRow("SELECT count(*) FROM user_configs WHERE key IN ('UnlockPasswordCheck', 'LNBackendType', 'Mnemonic')").Scan(&setupKeys); err != nil {
		return "", Counts{}, "", "", fmt.Errorf("inspect wallet configuration: %w", err)
	}
	if setupKeys != 3 {
		return "", Counts{}, "", "", errors.New("selected Alby Hub data is incomplete")
	}
	var backendType string
	if err = database.QueryRow("SELECT value FROM user_configs WHERE key = 'LNBackendType'").Scan(&backendType); err != nil {
		return "", Counts{}, "", "", fmt.Errorf("inspect wallet backend: %w", err)
	}
	if backendType != "LDK" {
		return "", Counts{}, "", "", fmt.Errorf("selected Alby Hub wallet uses %s, not embedded LDK", backendType)
	}
	var encryptedMnemonic string
	if err = database.QueryRow("SELECT value FROM user_configs WHERE key = 'Mnemonic'").Scan(&encryptedMnemonic); err != nil {
		return "", Counts{}, "", "", fmt.Errorf("inspect encrypted wallet identity: %w", err)
	}
	if encryptedMnemonic == "" {
		return "", Counts{}, "", "", errors.New("selected Alby Hub wallet has no encrypted wallet identity")
	}
	fingerprintHash := sha256.Sum256([]byte("grynvault-guided-import-wallet-v1\x00" + encryptedMnemonic))
	fingerprint := hex.EncodeToString(fingerprintHash[:])

	counts := Counts{}
	query := `SELECT
		(SELECT count(*) FROM apps),
		(SELECT count(*) FROM app_permissions),
		(SELECT count(*) FROM transactions),
		(SELECT count(*) FROM user_configs),
		(SELECT count(*) FROM migrations)`
	if err = database.QueryRow(query).Scan(
		&counts.Apps,
		&counts.Permissions,
		&counts.Transactions,
		&counts.UserConfigs,
		&counts.Migrations,
	); err != nil {
		return "", Counts{}, "", "", fmt.Errorf("read wallet continuity counts: %w", err)
	}
	var rebalanceTableCount int64
	if err = database.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'local_rebalance_quotes'").Scan(&rebalanceTableCount); err != nil {
		return "", Counts{}, "", "", fmt.Errorf("inspect rebalance records: %w", err)
	}
	if rebalanceTableCount == 1 {
		if err = database.QueryRow("SELECT count(*) FROM local_rebalance_quotes").Scan(&counts.RebalanceRecords); err != nil {
			return "", Counts{}, "", "", fmt.Errorf("read rebalance records: %w", err)
		}
	}
	return integrity, counts, backendType, fingerprint, nil
}

func FingerprintHint(fingerprint string) string {
	if len(fingerprint) <= 12 {
		return fingerprint
	}
	return fingerprint[:12]
}

func latestDurableChange(root string, hashes map[string]string) (string, error) {
	var latest time.Time
	for relative := range hashes {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return "", err
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	if latest.IsZero() {
		return "", nil
	}
	return latest.UTC().Format(time.RFC3339), nil
}

func KnownCandidatePaths(home, target, selected string) []string {
	home = userHomeOutsideSandbox(home)
	candidates := []string{
		filepath.Join(home, "Library", "Containers", "com.getalby.Alby-Hub", "Data", "Library", "Application Support", "albyhub"),
		filepath.Join(home, "Library", "Application Support", "albyhub"),
		filepath.Join(home, ".local", "share", "albyhub"),
		filepath.Join(home, "Library", "Containers", "com.grynvault.rebalance", "Data", "Library", "Application Support", "grynvault-rebalance"),
		target,
		selected,
	}
	seen := make(map[string]struct{})
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		absolute = filepath.Clean(absolute)
		if _, ok := seen[absolute]; ok {
			continue
		}
		seen[absolute] = struct{}{}
		result = append(result, absolute)
	}
	sort.Strings(result)
	return result
}

func FindConflictingCopy(summary *SourceSummary, candidates []string) (string, error) {
	for _, candidate := range candidates {
		resolved, err := ResolveWorkdir(candidate)
		if err != nil {
			if hasWalletMarkers(candidate) {
				return "", fmt.Errorf("possible competing wallet at %s could not be validated: %w", candidate, err)
			}
			continue
		}
		if samePath(resolved, summary.SourcePath) {
			continue
		}
		other, err := InspectSource(resolved)
		if err != nil {
			return "", fmt.Errorf("possible competing wallet at %s could not be validated: %w", resolved, err)
		}
		if other.WalletFingerprint == summary.WalletFingerprint {
			return other.SourcePath, nil
		}
	}
	return "", nil
}

func hasWalletMarkers(selected string) bool {
	for _, candidate := range workdirCandidates(selected) {
		if _, err := os.Lstat(filepath.Join(candidate, "nwc.db")); err == nil {
			return true
		}
		if _, err := os.Lstat(filepath.Join(candidate, "ldk", "storage", "ldk_node_data.sqlite")); err == nil {
			return true
		}
	}
	return false
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	return leftErr == nil && rightErr == nil && filepath.Clean(leftAbs) == filepath.Clean(rightAbs)
}

func userHomeOutsideSandbox(home string) string {
	home = filepath.Clean(home)
	if filepath.Base(home) != "Data" {
		return home
	}
	containerDir := filepath.Dir(home)
	containersDir := filepath.Dir(containerDir)
	libraryDir := filepath.Dir(containersDir)
	if filepath.Base(containersDir) != "Containers" || filepath.Base(libraryDir) != "Library" {
		return home
	}
	return filepath.Dir(libraryDir)
}

func PathsOverlap(left, right string) bool {
	left, leftErr := canonicalComparisonPath(left)
	right, rightErr := canonicalComparisonPath(right)
	if leftErr != nil || rightErr != nil {
		return true
	}
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	separator := string(os.PathSeparator)
	return left == right || strings.HasPrefix(left, right+separator) || strings.HasPrefix(right, left+separator)
}

func canonicalComparisonPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	current := filepath.Clean(absolute)
	suffix := []string{}
	for {
		resolved, resolveErr := filepath.EvalSymlinks(current)
		if resolveErr == nil {
			parts := append([]string{resolved}, suffix...)
			return filepath.Clean(filepath.Join(parts...)), nil
		}
		if !errors.Is(resolveErr, os.ErrNotExist) {
			return "", resolveErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", resolveErr
		}
		suffix = append([]string{filepath.Base(current)}, suffix...)
		current = parent
	}
}
