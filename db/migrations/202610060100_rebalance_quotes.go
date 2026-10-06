package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const rebalanceQuotesMigration = `
CREATE TABLE rebalance_quotes(
	id text PRIMARY KEY,
	state text NOT NULL,
	request_hash text NOT NULL,
	order_id text NOT NULL,
	receive_payment_request text NOT NULL,
	receive_payment_hash text NOT NULL,
	payment_request text NOT NULL,
	payment_hash text NOT NULL,
	amount_msat bigint NOT NULL,
	provider_fee_msat bigint NOT NULL,
	max_provider_fee_msat bigint NOT NULL,
	max_routing_fee_msat bigint NOT NULL,
	outgoing_channel_id text NOT NULL,
	outgoing_node_pubkey text NOT NULL,
	incoming_channel_id text NOT NULL,
	incoming_node_pubkey text NOT NULL,
	outgoing_spendable_snapshot_msat bigint NOT NULL,
	incoming_receivable_snapshot_msat bigint NOT NULL,
	expires_at {{ .Timestamp }} NOT NULL,
	created_at {{ .Timestamp }} NOT NULL,
	updated_at {{ .Timestamp }} NOT NULL,
	executed_at {{ .Timestamp }},
	failure_reason text NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX idx_rebalance_quotes_order_id ON rebalance_quotes(order_id);
CREATE INDEX idx_rebalance_quotes_state_expires_at ON rebalance_quotes(state, expires_at);
`

var rebalanceQuotesMigrationTmpl = template.Must(template.New("rebalanceQuotesMigration").Parse(rebalanceQuotesMigration))

var _202610060100_rebalance_quotes = &gormigrate.Migration{
	ID: "202610060100_rebalance_quotes",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, rebalanceQuotesMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("DROP TABLE IF EXISTS rebalance_quotes").Error
	},
}
