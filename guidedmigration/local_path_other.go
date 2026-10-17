//go:build !darwin

package guidedmigration

// IsLocalPath is permissive on non-macOS platforms. The guided desktop
// importer is currently distributed for macOS; other platforms retain their
// existing behavior until an equivalent mount-locality check is implemented.
func IsLocalPath(string) (bool, error) {
	return true, nil
}
