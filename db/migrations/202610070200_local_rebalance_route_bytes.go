package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const localRebalanceRouteBytesMigration = `
ALTER TABLE local_rebalance_quotes ADD COLUMN route_bytes {{ .Binary }};
`

var localRebalanceRouteBytesMigrationTmpl = template.Must(template.New("localRebalanceRouteBytesMigration").Parse(localRebalanceRouteBytesMigration))

var _202610070200_local_rebalance_route_bytes = &gormigrate.Migration{
	ID: "202610070200_local_rebalance_route_bytes",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, localRebalanceRouteBytesMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("ALTER TABLE local_rebalance_quotes DROP COLUMN route_bytes").Error
	},
}
