package guidedmigration

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var durableWalletRoots = []string{
	"nwc.db",
	"nwc.db-shm",
	"nwc.db-wal",
	filepath.Join("ldk", "storage"),
	filepath.Join("ldk", "static_channel_backups"),
}

func NewImportID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ControlPath(targetWorkdir string) string {
	return filepath.Clean(targetWorkdir) + ".grynvault-import-control.json"
}

func QuarantinePath(workdir string) string {
	return filepath.Join(workdir, QuarantineFileName)
}

func StagePath(targetWorkdir, importID string) string {
	return filepath.Join(filepath.Dir(targetWorkdir), "."+filepath.Base(targetWorkdir)+".grynvault-import-"+importID)
}

func PreImportPath(targetWorkdir, importID string) string {
	return filepath.Join(filepath.Dir(targetWorkdir), "."+filepath.Base(targetWorkdir)+".grynvault-preimport-"+importID)
}

func WriteJSONAtomic(path string, value any) error {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".grynvault-json-")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err = temp.Chmod(0600); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err = temp.Write(bytes); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = os.Rename(tempPath, path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}

func ReadJSON(path string, value any) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, value); err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func HashWalletData(sourceWorkdir string) (map[string]string, error) {
	return hashTree(sourceWorkdir, durableWalletRoots)
}

func CopyWalletData(sourceWorkdir, destination string) (map[string]string, error) {
	hashes, err := copyTree(sourceWorkdir, destination, durableWalletRoots)
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

func CopyTree(source, destination string, roots []string) (map[string]string, error) {
	return copyTree(source, destination, roots)
}

func copyTree(source, destination string, roots []string) (map[string]string, error) {
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
			roots = append(roots, entry.Name())
		}
	}

	hashes := make(map[string]string)
	for _, root := range roots {
		if !filepath.IsLocal(root) {
			return nil, fmt.Errorf("invalid data path %q", root)
		}
		sourcePath := filepath.Join(source, root)
		info, err := os.Lstat(sourcePath)
		if errors.Is(err, os.ErrNotExist) && isOptionalWalletRoot(root) {
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
				relative, relErr := filepath.Rel(source, path)
				if relErr != nil || !filepath.IsLocal(relative) {
					return fmt.Errorf("invalid wallet data path %q", path)
				}
				if entry.Type()&os.ModeSymlink != 0 {
					return fmt.Errorf("refusing symbolic link %q", relative)
				}
				destinationPath := filepath.Join(destination, relative)
				if entry.IsDir() {
					return os.MkdirAll(destinationPath, 0700)
				}
				if !entry.Type().IsRegular() {
					return fmt.Errorf("refusing non-regular wallet file %q", relative)
				}
				hash, copyErr := copyAndHashFile(path, destinationPath)
				if copyErr != nil {
					return copyErr
				}
				hashes[filepath.ToSlash(relative)] = hash
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
		hashes[filepath.ToSlash(root)] = hash
	}
	return hashes, nil
}

func hashTree(source string, roots []string) (map[string]string, error) {
	hashes := make(map[string]string)
	for _, root := range roots {
		if !filepath.IsLocal(root) {
			return nil, fmt.Errorf("invalid data path %q", root)
		}
		sourcePath := filepath.Join(source, root)
		info, err := os.Lstat(sourcePath)
		if errors.Is(err, os.ErrNotExist) && isOptionalWalletRoot(root) {
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
				relative, relErr := filepath.Rel(source, path)
				if relErr != nil || !filepath.IsLocal(relative) {
					return fmt.Errorf("invalid wallet data path %q", path)
				}
				if entry.Type()&os.ModeSymlink != 0 {
					return fmt.Errorf("refusing symbolic link %q", relative)
				}
				if entry.IsDir() {
					return nil
				}
				if !entry.Type().IsRegular() {
					return fmt.Errorf("refusing non-regular wallet file %q", relative)
				}
				hash, hashErr := hashFile(path)
				if hashErr != nil {
					return hashErr
				}
				hashes[filepath.ToSlash(relative)] = hash
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

func isOptionalWalletRoot(root string) bool {
	return root != "nwc.db" && root != filepath.Join("ldk", "storage")
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

func CompareHashes(expected, actual map[string]string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("file count differs: expected %d, found %d", len(expected), len(actual))
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

func RemoveGeneratedSQLiteSidecars(stagingDir string, sourceHashes map[string]string) error {
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

func IsBackupLikePath(path string) bool {
	for _, component := range strings.Split(filepath.ToSlash(filepath.Clean(path)), "/") {
		lower := strings.ToLower(component)
		if lower == "restore" || lower == "grynvault-rebalance-migration-backups" || strings.HasPrefix(lower, ".grynvault-import-") || strings.HasPrefix(lower, ".grynvault-preimport-") {
			return true
		}
	}
	return false
}
