package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const localRebalanceBalanceEvidenceMigration = `
ALTER TABLE local_rebalance_quotes ADD COLUMN outgoing_local_snapshot_msat bigint NOT NULL DEFAULT 0;
ALTER TABLE local_rebalance_quotes ADD COLUMN outgoing_remote_snapshot_msat bigint NOT NULL DEFAULT 0;
ALTER TABLE local_rebalance_quotes ADD COLUMN outgoing_local_reserve_msat bigint NOT NULL DEFAULT 0;
ALTER TABLE local_rebalance_quotes ADD COLUMN outgoing_remote_reserve_msat bigint NOT NULL DEFAULT 0;
ALTER TABLE local_rebalance_quotes ADD COLUMN incoming_local_snapshot_msat bigint NOT NULL DEFAULT 0;
ALTER TABLE local_rebalance_quotes ADD COLUMN incoming_remote_snapshot_msat bigint NOT NULL DEFAULT 0;
ALTER TABLE local_rebalance_quotes ADD COLUMN incoming_local_reserve_msat bigint NOT NULL DEFAULT 0;
ALTER TABLE local_rebalance_quotes ADD COLUMN incoming_remote_reserve_msat bigint NOT NULL DEFAULT 0;
`

var localRebalanceBalanceEvidenceMigrationTmpl = template.Must(template.New("localRebalanceBalanceEvidenceMigration").Parse(localRebalanceBalanceEvidenceMigration))

var _202610080300_local_rebalance_balance_evidence = &gormigrate.Migration{
	ID: "202610080300_local_rebalance_balance_evidence",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, localRebalanceBalanceEvidenceMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		for _, column := range []string{
			"incoming_remote_reserve_msat",
			"incoming_local_reserve_msat",
			"incoming_remote_snapshot_msat",
			"incoming_local_snapshot_msat",
			"outgoing_remote_reserve_msat",
			"outgoing_local_reserve_msat",
			"outgoing_remote_snapshot_msat",
			"outgoing_local_snapshot_msat",
		} {
			if err := tx.Exec("ALTER TABLE local_rebalance_quotes DROP COLUMN " + column).Error; err != nil {
				return err
			}
		}
		return nil
	},
}
