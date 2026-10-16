package wails

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	sqlite_wrapper "github.com/getAlby/hub/db/sqlite-wrapper"
)

const legacyImportEvidenceVersion = 1

type LegacyAlbyImportResult struct {
	Cancelled       bool   `json:"cancelled"`
	RestartRequired bool   `json:"restartRequired"`
	BackupLocation  string `json:"backupLocation,omitempty"`
}

type legacyImportCounts struct {
	Apps             int64 `json:"apps"`
	Permissions      int64 `json:"permissions"`
	Transactions     int64 `json:"transactions"`
	UserConfigs      int64 `json:"userConfigs"`
	Migrations       int64 `json:"migrations"`
	RebalanceRecords int64 `json:"rebalanceRecords"`
}

type legacyImportEvidence struct {
	Version            int                `json:"version"`
	CreatedAt          string             `json:"createdAt"`
	SourceFiles        map[string]string  `json:"sourceFiles"`
	SourceCounts       legacyImportCounts `json:"sourceCounts"`
	SourceNWCIntegrity string             `json:"sourceNwcIntegrity"`
	SourceLDKIntegrity string             `json:"sourceLdkIntegrity"`
}

func canonicalLegacyAlbyWorkdir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	home = userHomeOutsideSandbox(home)
	return filepath.Join(home, "Library", "Containers", "com.getalby.Alby-Hub", "Data", "Library", "Application Support", "albyhub")
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

func resolveLegacyAlbyWorkdir(selected string) (string, error) {
	if selected == "" {
		return "", errors.New("no Alby Hub data folder was selected")
	}

	selected, err := filepath.Abs(selected)
	if err != nil {
		return "", fmt.Errorf("resolve selected folder: %w", err)
	}

	candidates := []string{
		selected,
		filepath.Join(selected, "Data", "Library", "Application Support", "albyhub"),
		filepath.Join(selected, "Library", "Application Support", "albyhub"),
		filepath.Join(selected, "albyhub"),
	}

	for _, candidate := range candidates {
		if isRegularFile(filepath.Join(candidate, "nwc.db")) && isRegularFile(filepath.Join(candidate, "ldk", "storage", "ldk_node_data.sqlite")) {
			return candidate, nil
		}
	}

	return "", errors.New("the selected folder is not an Alby Hub LDK wallet data folder")
}

func isRegularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular()
}

func stageLegacyAlbyWorkdir(sourceWorkdir, targetWorkdir string, now time.Time) (*LegacyAlbyImportResult, error) {
	sourceWorkdir, err := filepath.Abs(sourceWorkdir)
	if err != nil {
		return nil, fmt.Errorf("resolve Alby Hub source folder: %w", err)
	}
	targetWorkdir, err = filepath.Abs(targetWorkdir)
	if err != nil {
		return nil, fmt.Errorf("resolve Grynvault target folder: %w", err)
	}
	if sourceWorkdir == targetWorkdir || strings.HasPrefix(targetWorkdir, sourceWorkdir+string(os.PathSeparator)) || strings.HasPrefix(sourceWorkdir, targetWorkdir+string(os.PathSeparator)) {
		return nil, errors.New("source and target wallet folders must be separate")
	}

	resolvedSource, err := resolveLegacyAlbyWorkdir(sourceWorkdir)
	if err != nil {
		return nil, err
	}
	sourceWorkdir = resolvedSource
	if sourceWorkdir == targetWorkdir || strings.HasPrefix(targetWorkdir, sourceWorkdir+string(os.PathSeparator)) || strings.HasPrefix(sourceWorkdir, targetWorkdir+string(os.PathSeparator)) {
		return nil, errors.New("source and target wallet folders must be separate")
	}

	if err := os.MkdirAll(targetWorkdir, 0700); err != nil {
		return nil, fmt.Errorf("create Grynvault data folder: %w", err)
	}
	restoreDir := filepath.Join(targetWorkdir, "restore")
	if _, err := os.Stat(restoreDir); err == nil {
		return nil, errors.New("an import is already staged; restart Grynvault Rebalance")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect staged import: %w", err)
	}

	backupParent := filepath.Join(filepath.Dir(targetWorkdir), "grynvault-rebalance-migration-backups")
	if err := os.MkdirAll(backupParent, 0700); err != nil {
		return nil, fmt.Errorf("create migration backup folder: %w", err)
	}
	backupDir := filepath.Join(backupParent, now.UTC().Format("20060102T150405.000000000Z"))
	if err := os.Mkdir(backupDir, 0700); err != nil {
		return nil, fmt.Errorf("create migration backup: %w", err)
	}
	keepBackup := false
	defer func() {
		if !keepBackup {
			_ = os.RemoveAll(backupDir)
		}
	}()

	targetSnapshot := filepath.Join(backupDir, "grynvault-before-import")
	if targetEntries, err := os.ReadDir(targetWorkdir); err != nil {
		return nil, fmt.Errorf("inspect current Grynvault data: %w", err)
	} else if len(targetEntries) > 0 {
		if _, err = copyDataTree(targetWorkdir, targetSnapshot, nil); err != nil {
			return nil, fmt.Errorf("back up current Grynvault data: %w", err)
		}
	}

	sourceSnapshot := filepath.Join(backupDir, "albyhub-original-snapshot")
	backupHashes, err := copyWalletData(sourceWorkdir, sourceSnapshot)
	if err != nil {
		return nil, fmt.Errorf("back up Alby Hub data: %w", err)
	}

	stagingDir, err := os.MkdirTemp(filepath.Dir(targetWorkdir), ".grynvault-import-")
	if err != nil {
		return nil, fmt.Errorf("create import staging folder: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	stagedHashes, err := copyWalletData(sourceWorkdir, stagingDir)
	if err != nil {
		return nil, fmt.Errorf("stage Alby Hub data: %w", err)
	}
	if err = compareFileHashes(backupHashes, stagedHashes); err != nil {
		return nil, fmt.Errorf("verify staged wallet copy: %w", err)
	}

	nwcIntegrity, counts, err := validateNWCDB(filepath.Join(stagingDir, "nwc.db"))
	if err != nil {
		return nil, err
	}
	ldkIntegrity, err := validateSQLite(filepath.Join(stagingDir, "ldk", "storage", "ldk_node_data.sqlite"))
	if err != nil {
		return nil, fmt.Errorf("validate copied LDK database: %w", err)
	}
	if err = removeGeneratedSQLiteSidecars(stagingDir, stagedHashes); err != nil {
		return nil, fmt.Errorf("clean up validation sidecars: %w", err)
	}
	validatedHashes, err := hashWalletData(stagingDir)
	if err != nil {
		return nil, fmt.Errorf("hash validated wallet copy: %w", err)
	}
	if err = compareFileHashes(backupHashes, validatedHashes); err != nil {
		return nil, fmt.Errorf("verify validated wallet copy: %w", err)
	}

	evidence := legacyImportEvidence{
		Version:            legacyImportEvidenceVersion,
		CreatedAt:          now.UTC().Format(time.RFC3339Nano),
		SourceFiles:        backupHashes,
		SourceCounts:       counts,
		SourceNWCIntegrity: nwcIntegrity,
		SourceLDKIntegrity: ldkIntegrity,
	}
	evidenceBytes, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode import evidence: %w", err)
	}
	if err = os.WriteFile(filepath.Join(backupDir, "import-evidence.json"), evidenceBytes, 0600); err != nil {
		return nil, fmt.Errorf("write import evidence: %w", err)
	}

	if err = os.Rename(stagingDir, restoreDir); err != nil {
		return nil, fmt.Errorf("activate staged import: %w", err)
	}
	keepBackup = true

	return &LegacyAlbyImportResult{
		RestartRequired: true,
		BackupLocation:  backupDir,
	}, nil
}

func copyWalletData(sourceWorkdir, destination string) (map[string]string, error) {
	// Copy durable wallet state and static channel backups, but not runtime logs.
	// Stock Alby Hub can leave a rotating `ldk_node_latest.log` symlink in the
	// logs directory; logs are neither required for recovery nor appropriate to
	// carry into a newly imported wallet.
	roots := []string{
		"nwc.db",
		"nwc.db-shm",
		"nwc.db-wal",
		filepath.Join("ldk", "storage"),
		filepath.Join("ldk", "static_channel_backups"),
	}
	hashes, err := copyDataTree(sourceWorkdir, destination, roots)
	if err != nil {
		return nil, err
	}

	if _, ok := hashes["nwc.db"]; !ok {
		return nil, errors.New("Alby Hub database is missing")
	}
	if _, ok := hashes["ldk/storage/ldk_node_data.sqlite"]; !ok {
		return nil, errors.New("Alby Hub LDK database is missing")
	}
	return hashes, nil
}

func removeGeneratedSQLiteSidecars(stagingDir string, sourceHashes map[string]string) error {
	// Read-only SQLite validation can still create empty WAL/SHM sidecars when
	// the copied database uses WAL journaling. They are validation artifacts,
	// not source wallet data, so remove them only when the source copy did not
	// contain the same path. Source sidecars are retained and covered by the
	// final hash comparison below.
	sidecars := []string{
		"nwc.db-shm",
		"nwc.db-wal",
		filepath.Join("ldk", "storage", "ldk_node_data.sqlite-shm"),
		filepath.Join("ldk", "storage", "ldk_node_data.sqlite-wal"),
	}
	for _, relativePath := range sidecars {
		if _, exists := sourceHashes[filepath.ToSlash(relativePath)]; exists {
			continue
		}
		if err := os.Remove(filepath.Join(stagingDir, relativePath)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func hashWalletData(sourceWorkdir string) (map[string]string, error) {
	roots := []string{
		"nwc.db",
		"nwc.db-shm",
		"nwc.db-wal",
		filepath.Join("ldk", "storage"),
		filepath.Join("ldk", "static_channel_backups"),
	}
	hashes := make(map[string]string)
	for _, root := range roots {
		sourcePath := filepath.Join(sourceWorkdir, root)
		info, err := os.Lstat(sourcePath)
		if errors.Is(err, os.ErrNotExist) && root != "nwc.db" && root != filepath.Join("ldk", "storage") {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect %s: %w", root, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("refusing symbolic link %q", root)
		}

		if info.IsDir() {
			err = filepath.WalkDir(sourcePath, func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				rel, relErr := filepath.Rel(sourceWorkdir, path)
				if relErr != nil || !filepath.IsLocal(rel) {
					return fmt.Errorf("invalid wallet data path %q", path)
				}
				if entry.Type()&os.ModeSymlink != 0 {
					return fmt.Errorf("refusing symbolic link %q", rel)
				}
				if entry.IsDir() {
					return nil
				}
				if !entry.Type().IsRegular() {
					return fmt.Errorf("refusing non-regular wallet file %q", rel)
				}
				hash, hashErr := hashFile(path)
				if hashErr != nil {
					return hashErr
				}
				hashes[filepath.ToSlash(rel)] = hash
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("refusing non-regular wallet file %q", root)
		}
		hash, err := hashFile(sourcePath)
		if err != nil {
			return nil, err
		}
		hashes[filepath.ToSlash(root)] = hash
	}
	return hashes, nil
}

func hashFile(path string) (string, error) {
	input, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer input.Close()
	hash := sha256.New()
	if _, err = io.Copy(hash, input); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyDataTree(source, destination string, roots []string) (map[string]string, error) {
	if err := os.MkdirAll(destination, 0700); err != nil {
		return nil, err
	}
	if roots == nil {
		entries, err := os.ReadDir(source)
		if err != nil {
			return nil, err
		}
		roots = make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.Name() != "restore" {
				roots = append(roots, entry.Name())
			}
		}
	}

	hashes := make(map[string]string)
	for _, root := range roots {
		if !filepath.IsLocal(root) {
			return nil, fmt.Errorf("invalid data path %q", root)
		}
		sourcePath := filepath.Join(source, root)
		info, err := os.Lstat(sourcePath)
		if errors.Is(err, os.ErrNotExist) && root != "nwc.db" && root != filepath.Join("ldk", "storage") {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect %s: %w", root, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("refusing symbolic link %q", root)
		}

		if info.IsDir() {
			err = filepath.WalkDir(sourcePath, func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				rel, err := filepath.Rel(source, path)
				if err != nil || !filepath.IsLocal(rel) {
					return fmt.Errorf("invalid wallet data path %q", path)
				}
				if entry.Type()&os.ModeSymlink != 0 {
					return fmt.Errorf("refusing symbolic link %q", rel)
				}
				destinationPath := filepath.Join(destination, rel)
				if entry.IsDir() {
					return os.MkdirAll(destinationPath, 0700)
				}
				if !entry.Type().IsRegular() {
					return fmt.Errorf("refusing non-regular wallet file %q", rel)
				}
				hash, err := copyAndHashFile(path, destinationPath)
				if err != nil {
					return err
				}
				hashes[filepath.ToSlash(rel)] = hash
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("refusing non-regular wallet file %q", root)
		}
		hash, err := copyAndHashFile(sourcePath, filepath.Join(destination, root))
		if err != nil {
			return nil, err
		}
		hashes[root] = hash
	}
	return hashes, nil
}

func copyAndHashFile(source, destination string) (string, error) {
	input, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer input.Close()

	if err = os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return "", err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(output, hash), input)
	syncErr := output.Sync()
	closeErr := output.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if syncErr != nil {
		return "", syncErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func compareFileHashes(expected, actual map[string]string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("file count differs: expected %d, copied %d", len(expected), len(actual))
	}
	keys := make([]string, 0, len(expected))
	for key := range expected {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if actual[key] != expected[key] {
			return fmt.Errorf("hash mismatch for %s", key)
		}
	}
	return nil
}

func validateSQLite(path string) (string, error) {
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

func validateNWCDB(path string) (string, legacyImportCounts, error) {
	integrity, err := validateSQLite(path)
	if err != nil {
		return "", legacyImportCounts{}, fmt.Errorf("validate copied Alby Hub database: %w", err)
	}

	dsn := (&url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}).String()
	database, err := sql.Open(sqlite_wrapper.Sqlite3WrapperDriverName, dsn)
	if err != nil {
		return "", legacyImportCounts{}, err
	}
	defer database.Close()

	var setupKeys int64
	if err = database.QueryRow("SELECT count(*) FROM user_configs WHERE key IN ('UnlockPasswordCheck', 'LNBackendType', 'Mnemonic')").Scan(&setupKeys); err != nil {
		return "", legacyImportCounts{}, fmt.Errorf("inspect wallet configuration: %w", err)
	}
	if setupKeys != 3 {
		return "", legacyImportCounts{}, errors.New("selected Alby Hub data is incomplete")
	}
	var backendType string
	if err = database.QueryRow("SELECT value FROM user_configs WHERE key = 'LNBackendType'").Scan(&backendType); err != nil {
		return "", legacyImportCounts{}, fmt.Errorf("inspect wallet backend: %w", err)
	}
	if backendType != "LDK" {
		return "", legacyImportCounts{}, fmt.Errorf("selected Alby Hub wallet uses %s, not embedded LDK", backendType)
	}

	counts := legacyImportCounts{}
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
		return "", legacyImportCounts{}, fmt.Errorf("read wallet continuity counts: %w", err)
	}
	var rebalanceTableCount int64
	if err = database.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'local_rebalance_quotes'").Scan(&rebalanceTableCount); err != nil {
		return "", legacyImportCounts{}, fmt.Errorf("inspect rebalance records: %w", err)
	}
	if rebalanceTableCount == 1 {
		if err = database.QueryRow("SELECT count(*) FROM local_rebalance_quotes").Scan(&counts.RebalanceRecords); err != nil {
			return "", legacyImportCounts{}, fmt.Errorf("read rebalance records: %w", err)
		}
	}

	return integrity, counts, nil
}
