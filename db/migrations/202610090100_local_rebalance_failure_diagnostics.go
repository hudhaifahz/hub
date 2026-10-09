package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const localRebalanceFailureDiagnosticsMigration = `
ALTER TABLE local_rebalance_quotes ADD COLUMN terminal_evidence_version bigint NOT NULL DEFAULT 1;
ALTER TABLE local_rebalance_quotes ADD COLUMN failure_code text NOT NULL DEFAULT '';
ALTER TABLE local_rebalance_quotes ADD COLUMN observed_receiving_channel_ids_json text NOT NULL DEFAULT '[]';
ALTER TABLE local_rebalance_quotes ADD COLUMN unidentified_receiving_channel_count bigint NOT NULL DEFAULT 0;
`

var localRebalanceFailureDiagnosticsMigrationTmpl = template.Must(template.New("localRebalanceFailureDiagnosticsMigration").Parse(localRebalanceFailureDiagnosticsMigration))

var _202610090100_local_rebalance_failure_diagnostics = &gormigrate.Migration{
	ID: "202610090100_local_rebalance_failure_diagnostics",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, localRebalanceFailureDiagnosticsMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		for _, column := range []string{
			"unidentified_receiving_channel_count",
			"observed_receiving_channel_ids_json",
			"failure_code",
			"terminal_evidence_version",
		} {
			if err := tx.Exec("ALTER TABLE local_rebalance_quotes DROP COLUMN " + column).Error; err != nil {
				return err
			}
		}
		return nil
	},
}
