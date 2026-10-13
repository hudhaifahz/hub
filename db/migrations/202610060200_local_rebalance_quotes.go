package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const localRebalanceQuotesMigration = `
CREATE TABLE local_rebalance_quotes(
	id text PRIMARY KEY,
	state text NOT NULL,
	request_hash text NOT NULL,
	route_fingerprint text NOT NULL,
	amount_msat bigint NOT NULL,
	total_routing_fee_msat bigint NOT NULL,
	max_routing_fee_msat bigint NOT NULL,
	max_total_debit_msat bigint NOT NULL,
	outgoing_channel_id text NOT NULL,
	outgoing_node_pubkey text NOT NULL,
	outgoing_short_channel_id text NOT NULL,
	incoming_channel_id text NOT NULL,
	incoming_node_pubkey text NOT NULL,
	incoming_short_channel_id text NOT NULL,
	outgoing_spendable_snapshot_msat bigint NOT NULL,
	incoming_receivable_snapshot_msat bigint NOT NULL,
	route_json text NOT NULL,
	expires_at {{ .Timestamp }} NOT NULL,
	created_at {{ .Timestamp }} NOT NULL,
	updated_at {{ .Timestamp }} NOT NULL,
	executed_at {{ .Timestamp }},
	failure_reason text NOT NULL DEFAULT ''
);
CREATE INDEX idx_local_rebalance_quotes_state_expires_at ON local_rebalance_quotes(state, expires_at);
CREATE INDEX idx_local_rebalance_quotes_route_fingerprint ON local_rebalance_quotes(route_fingerprint);
`

var localRebalanceQuotesMigrationTmpl = template.Must(template.New("localRebalanceQuotesMigration").Parse(localRebalanceQuotesMigration))

var _202610060200_local_rebalance_quotes = &gormigrate.Migration{
	ID: "202610060200_local_rebalance_quotes",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, localRebalanceQuotesMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("DROP TABLE IF EXISTS local_rebalance_quotes").Error
	},
}
