package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const localRebalanceTerminalEvidenceMigration = `
ALTER TABLE local_rebalance_quotes ADD COLUMN actual_routing_fee_msat bigint;
ALTER TABLE local_rebalance_quotes ADD COLUMN lightning_terminal_at {{ .Timestamp }};
ALTER TABLE local_rebalance_quotes ADD COLUMN reconciled_at {{ .Timestamp }};
ALTER TABLE local_rebalance_quotes ADD COLUMN terminal_evidence_hash text NOT NULL DEFAULT '';
`

var localRebalanceTerminalEvidenceMigrationTmpl = template.Must(template.New("localRebalanceTerminalEvidenceMigration").Parse(localRebalanceTerminalEvidenceMigration))

var _202610080200_local_rebalance_terminal_evidence = &gormigrate.Migration{
	ID: "202610080200_local_rebalance_terminal_evidence",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, localRebalanceTerminalEvidenceMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		for _, column := range []string{
			"terminal_evidence_hash",
			"reconciled_at",
			"lightning_terminal_at",
			"actual_routing_fee_msat",
		} {
			if err := tx.Exec("ALTER TABLE local_rebalance_quotes DROP COLUMN " + column).Error; err != nil {
				return err
			}
		}
		return nil
	},
}
