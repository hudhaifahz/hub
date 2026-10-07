package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202610070100_single_local_rebalance_execution = &gormigrate.Migration{
	ID: "202610070100_single_local_rebalance_execution",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
CREATE UNIQUE INDEX idx_local_rebalance_quotes_single_executing
ON local_rebalance_quotes(state)
WHERE state = 'executing'
`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("DROP INDEX IF EXISTS idx_local_rebalance_quotes_single_executing").Error
	},
}
