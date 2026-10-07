package api

import (
	"math"
	"strings"
	"testing"

	"github.com/getAlby/hub/lnclient"
	"github.com/stretchr/testify/require"
)

func TestFindExactChannelRequiresIdAndPubkey(t *testing.T) {
	pubkey := "02" + strings.Repeat("11", 32)
	channels := []lnclient.Channel{{Id: "channel-1", RemotePubkey: pubkey}}

	channel, err := findExactChannel(channels, "channel-1", pubkey)
	require.NoError(t, err)
	require.Equal(t, "channel-1", channel.Id)

	_, err = findExactChannel(channels, "channel-1", "03-other")
	require.ErrorContains(t, err, "do not match")
}

func TestFindExactChannelRejectsAmbiguousIdentity(t *testing.T) {
	channels := []lnclient.Channel{
		{Id: "channel-1", RemotePubkey: "peer"},
		{Id: "channel-1", RemotePubkey: "peer"},
	}
	_, err := findExactChannel(channels, "channel-1", "peer")
	require.ErrorContains(t, err, "ambiguous")
}

func TestCheckedAddRejectsOverflow(t *testing.T) {
	_, ok := checkedAdd(math.MaxUint64, 1)
	require.False(t, ok)
	value, ok := checkedAdd(10, 20)
	require.True(t, ok)
	require.Equal(t, uint64(30), value)
}

func TestValidateNodePubkey(t *testing.T) {
	require.NoError(t, validateNodePubkey("02"+strings.Repeat("11", 32)))
	require.Error(t, validateNodePubkey("04"+strings.Repeat("11", 32)))
	require.Error(t, validateNodePubkey("02abcd"))
}

func TestRebalanceProviderErrorIncludesStructuredReason(t *testing.T) {
	err := rebalanceProviderError(422, []byte(`{"message":"requested route is unavailable"}`))
	require.EqualError(t, err, "rebalance quote provider returned HTTP 422: requested route is unavailable")
}

func TestRebalanceProviderErrorIncludesNestedErrorMessage(t *testing.T) {
	err := rebalanceProviderError(422, []byte(`{"error":{"name":"Error","message":"no_route_found"}}`))
	require.EqualError(t, err, "rebalance quote provider returned HTTP 422: no_route_found")
}

func TestRebalanceProviderErrorRedactsInvoice(t *testing.T) {
	err := rebalanceProviderError(422, []byte(`{"error":"failed to route lnbc1sensitiveinvoice"}`))
	require.EqualError(t, err, "rebalance quote provider returned HTTP 422: failed to route [redacted invoice]")
}

func TestRebalanceProviderErrorIgnoresUnstructuredBody(t *testing.T) {
	err := rebalanceProviderError(422, []byte(`provider rejected lnbc1sensitiveinvoice`))
	require.EqualError(t, err, "rebalance quote provider returned HTTP 422")
}

func TestLocalRebalanceRouteFingerprintIsDeterministic(t *testing.T) {
	quote := testLocalRebalanceQuoteResponse()

	first, err := hashLocalRebalanceRouteMaterial(quote)
	require.NoError(t, err)
	second, err := hashLocalRebalanceRouteMaterial(quote)
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Len(t, first, 64)
}

func TestLocalRebalanceRouteFingerprintBindsEveryHop(t *testing.T) {
	original := testLocalRebalanceQuoteResponse()
	originalFingerprint, err := hashLocalRebalanceRouteMaterial(original)
	require.NoError(t, err)

	changed := testLocalRebalanceQuoteResponse()
	changed.Paths[0].Hops[1].ShortChannelId = "999"
	changedFingerprint, err := hashLocalRebalanceRouteMaterial(changed)
	require.NoError(t, err)

	require.NotEqual(t, originalFingerprint, changedFingerprint)
}

func TestLocalRebalanceRouteFingerprintBindsExactChannelsAndAmounts(t *testing.T) {
	original := testLocalRebalanceQuoteResponse()
	originalFingerprint, err := hashLocalRebalanceRouteMaterial(original)
	require.NoError(t, err)

	tests := map[string]func(*LocalRebalanceQuoteResponse){
		"amount": func(quote *LocalRebalanceQuoteResponse) { quote.AmountMsat++ },
		"fee":    func(quote *LocalRebalanceQuoteResponse) { quote.TotalRoutingFeeMsat++ },
		"outgoing channel": func(quote *LocalRebalanceQuoteResponse) {
			quote.OutgoingChannelId = "other-outgoing"
		},
		"incoming channel": func(quote *LocalRebalanceQuoteResponse) {
			quote.IncomingChannelId = "other-incoming"
		},
		"final scid": func(quote *LocalRebalanceQuoteResponse) {
			quote.IncomingShortChannelId = "998"
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			changed := testLocalRebalanceQuoteResponse()
			mutate(changed)
			changedFingerprint, err := hashLocalRebalanceRouteMaterial(changed)
			require.NoError(t, err)
			require.NotEqual(t, originalFingerprint, changedFingerprint)
		})
	}
}

func testLocalRebalanceQuoteResponse() *LocalRebalanceQuoteResponse {
	return &LocalRebalanceQuoteResponse{
		AmountMsat:                     20_000_000,
		TotalRoutingFeeMsat:            17_062,
		MaxRoutingFeeMsat:              1_000_000,
		MaxTotalDebitMsat:              20_017_062,
		OutgoingChannelId:              "source-channel",
		OutgoingNodePubkey:             "02source",
		OutgoingShortChannelId:         "101",
		IncomingChannelId:              "kraken-channel",
		IncomingNodePubkey:             "02kraken",
		IncomingShortChannelId:         "202",
		OutgoingSpendableSnapshotMsat:  989_340_000,
		IncomingReceivableSnapshotMsat: 840_710_000,
		Paths: []LocalCircularRoutePath{{
			AmountMsat: 20_000_000,
			FeeMsat:    17_062,
			Hops: []LocalCircularRouteHop{
				{NodePubkey: "02source", ShortChannelId: "101", FeeMsat: 10_000, CltvDelta: 40},
				{NodePubkey: "02middle", ShortChannelId: "150", FeeMsat: 7_062, CltvDelta: 40},
				{NodePubkey: "02kraken", ShortChannelId: "175", CltvDelta: 40},
				{NodePubkey: "02local", ShortChannelId: "202"},
			},
		}},
	}
}
