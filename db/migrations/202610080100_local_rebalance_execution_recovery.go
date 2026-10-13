package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const localRebalanceExecutionRecoveryMigration = `
ALTER TABLE local_rebalance_quotes ADD COLUMN operation_id text NOT NULL DEFAULT '';
ALTER TABLE local_rebalance_quotes ADD COLUMN execution_phase text NOT NULL DEFAULT '';
ALTER TABLE local_rebalance_quotes ADD COLUMN prepared_payment_hash text NOT NULL DEFAULT '';
ALTER TABLE local_rebalance_quotes ADD COLUMN outbound_payment_id text NOT NULL DEFAULT '';
ALTER TABLE local_rebalance_quotes ADD COLUMN prepared_at {{ .Timestamp }};
ALTER TABLE local_rebalance_quotes ADD COLUMN submitted_at {{ .Timestamp }};
CREATE UNIQUE INDEX idx_local_rebalance_quotes_operation_id
ON local_rebalance_quotes(operation_id)
WHERE operation_id <> '';
`

var localRebalanceExecutionRecoveryMigrationTmpl = template.Must(template.New("localRebalanceExecutionRecoveryMigration").Parse(localRebalanceExecutionRecoveryMigration))

var _202610080100_local_rebalance_execution_recovery = &gormigrate.Migration{
	ID: "202610080100_local_rebalance_execution_recovery",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, localRebalanceExecutionRecoveryMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.Exec("DROP INDEX IF EXISTS idx_local_rebalance_quotes_operation_id").Error; err != nil {
			return err
		}
		for _, column := range []string{
			"submitted_at",
			"prepared_at",
			"outbound_payment_id",
			"prepared_payment_hash",
			"execution_phase",
			"operation_id",
		} {
			if err := tx.Exec("ALTER TABLE local_rebalance_quotes DROP COLUMN " + column).Error; err != nil {
				return err
			}
		}
		return nil
	},
}
