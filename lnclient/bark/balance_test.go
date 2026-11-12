//go:build (darwin && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (windows && amd64)

package bark

import (
	"testing"

	bark "gitlab.com/ark-bitcoin/bark-ffi-bindings/golang/bark"
)

func TestBalancesExcludeFundsAwaitingRefresh(t *testing.T) {
	for _, spendable := range []uint64{0, 100} {
		balance := balancesResponse(bark.Balance{
			SpendableSats:       spendable,
			NeedsRefreshSats:    800,
			PendingInRoundSats:  200,
			PendingBoardSats:    300,
			PendingOffboardSats: 400,
		})
		want := int64(spendable) * 1000
		if balance.Lightning.TotalSpendableMsat != want ||
			balance.Lightning.NextMaxSpendableMsat != want ||
			balance.Lightning.NextMaxSpendableMPPMsat != want {
			t.Fatalf("refresh/pending funds must not enter spendable balances: got %+v, want %d", balance.Lightning, want)
		}
	}
}
