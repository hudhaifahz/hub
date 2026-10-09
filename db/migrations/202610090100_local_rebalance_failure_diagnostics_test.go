package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLocalRebalanceFailureDiagnosticsMigrationPreservesExistingRows(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.Exec(`
		CREATE TABLE local_rebalance_quotes (
			id text PRIMARY KEY,
			terminal_evidence_hash text NOT NULL DEFAULT ''
		);
		INSERT INTO local_rebalance_quotes (id, terminal_evidence_hash)
		VALUES ('existing-terminal-record', 'existing-evidence-hash');
	`).Error)

	require.NoError(t, _202610090100_local_rebalance_failure_diagnostics.Migrate(gormDB))

	var migrated struct {
		ID                                string
		TerminalEvidenceHash              string
		TerminalEvidenceVersion           uint32
		FailureCode                       string
		ObservedReceivingChannelIDsJSON   string
		UnidentifiedReceivingChannelCount uint32
	}
	require.NoError(t, gormDB.Raw(`
		SELECT id,
		       terminal_evidence_hash,
		       terminal_evidence_version,
		       failure_code,
		       observed_receiving_channel_ids_json,
		       unidentified_receiving_channel_count
		FROM local_rebalance_quotes
		WHERE id = 'existing-terminal-record'
	`).Scan(&migrated).Error)
	require.Equal(t, "existing-terminal-record", migrated.ID)
	require.Equal(t, "existing-evidence-hash", migrated.TerminalEvidenceHash)
	require.Equal(t, uint32(1), migrated.TerminalEvidenceVersion)
	require.Empty(t, migrated.FailureCode)
	require.JSONEq(t, `[]`, migrated.ObservedReceivingChannelIDsJSON)
	require.Zero(t, migrated.UnidentifiedReceivingChannelCount)
}
