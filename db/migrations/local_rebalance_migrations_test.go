package migrations

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRebalanceMigrationsRollback(t *testing.T) {
	t.Run("sqlite", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open("file:rebalance-migration-rollback?mode=memory&cache=shared"), &gorm.Config{})
		require.NoError(t, err)
		testRebalanceMigrationsRollback(t, db)
	})

	if uri := os.Getenv("TEST_REBALANCE_MIGRATIONS_POSTGRES_URI"); uri != "" {
		t.Run("postgres", func(t *testing.T) {
			db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, db.Exec("DROP TABLE IF EXISTS local_rebalance_quotes CASCADE").Error)
			require.NoError(t, db.Exec("DROP TABLE IF EXISTS rebalance_quotes CASCADE").Error)
			t.Cleanup(func() {
				require.NoError(t, db.Exec("DROP TABLE IF EXISTS local_rebalance_quotes CASCADE").Error)
				require.NoError(t, db.Exec("DROP TABLE IF EXISTS rebalance_quotes CASCADE").Error)
			})
			testRebalanceMigrationsRollback(t, db)
		})
	}
}

func testRebalanceMigrationsRollback(t *testing.T, db *gorm.DB) {
	migrations := []*struct {
		name     string
		migrate  func(*gorm.DB) error
		rollback func(*gorm.DB) error
	}{
		{"rebalance quotes", _202610060100_rebalance_quotes.Migrate, _202610060100_rebalance_quotes.Rollback},
		{"local quotes", _202610060200_local_rebalance_quotes.Migrate, _202610060200_local_rebalance_quotes.Rollback},
		{"single execution", _202610070100_single_local_rebalance_execution.Migrate, _202610070100_single_local_rebalance_execution.Rollback},
		{"route bytes", _202610070200_local_rebalance_route_bytes.Migrate, _202610070200_local_rebalance_route_bytes.Rollback},
		{"execution recovery", _202610080100_local_rebalance_execution_recovery.Migrate, _202610080100_local_rebalance_execution_recovery.Rollback},
		{"terminal evidence", _202610080200_local_rebalance_terminal_evidence.Migrate, _202610080200_local_rebalance_terminal_evidence.Rollback},
		{"balance evidence", _202610080300_local_rebalance_balance_evidence.Migrate, _202610080300_local_rebalance_balance_evidence.Rollback},
		{"failure diagnostics", _202610090100_local_rebalance_failure_diagnostics.Migrate, _202610090100_local_rebalance_failure_diagnostics.Rollback},
	}

	for _, migration := range migrations {
		require.NoError(t, migration.migrate(db), migration.name)
	}
	require.True(t, db.Migrator().HasTable("rebalance_quotes"))
	require.True(t, db.Migrator().HasTable("local_rebalance_quotes"))
	require.True(t, db.Migrator().HasIndex("local_rebalance_quotes", "idx_local_rebalance_quotes_single_executing"))
	for _, column := range []string{
		"route_bytes",
		"operation_id",
		"terminal_evidence_hash",
		"outgoing_local_snapshot_msat",
		"failure_code",
	} {
		require.True(t, db.Migrator().HasColumn("local_rebalance_quotes", column), column)
	}

	for index := len(migrations) - 1; index >= 0; index-- {
		require.NoError(t, migrations[index].rollback(db), migrations[index].name)
	}
	require.False(t, db.Migrator().HasTable("local_rebalance_quotes"))
	require.False(t, db.Migrator().HasTable("rebalance_quotes"))
}
