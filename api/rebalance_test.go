package api

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	test_db "github.com/getAlby/hub/tests/db"
	"github.com/sirupsen/logrus"
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

func TestAcquireLocalRebalanceQuoteForExecutionIsAtomicAndSingleUse(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, "atomic-quote", now.Add(time.Minute))
	require.NoError(t, gormDB.Create(&quote).Error)

	const contenders = 8
	start := make(chan struct{})
	results := make(chan error, contenders)
	var waitGroup sync.WaitGroup
	for i := 0; i < contenders; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, acquireErr := theAPI.acquireLocalRebalanceQuoteForExecution(quote.ID, fingerprint, now)
			results <- acquireErr
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	for acquireErr := range results {
		if acquireErr == nil {
			successes++
		}
	}
	require.Equal(t, 1, successes)
	var persisted db.LocalRebalanceQuote
	require.NoError(t, gormDB.First(&persisted, "id = ?", quote.ID).Error)
	require.Equal(t, "executing", persisted.State)
	require.Nil(t, persisted.ExecutedAt)
}

func TestAcquireLocalRebalanceQuoteForExecutionRejectsMismatchExpiryAndTampering(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)

	mismatched, _ := testPersistedLocalRebalanceQuote(t, "mismatched-quote", now.Add(time.Minute))
	require.NoError(t, gormDB.Create(&mismatched).Error)
	_, err = theAPI.acquireLocalRebalanceQuoteForExecution(mismatched.ID, strings.Repeat("0", 64), now)
	require.ErrorContains(t, err, "fingerprint does not match the reviewed route")
	require.NoError(t, gormDB.First(&mismatched, "id = ?", mismatched.ID).Error)
	require.Equal(t, "quoted", mismatched.State)

	expired, expiredFingerprint := testPersistedLocalRebalanceQuote(t, "expired-quote", now)
	require.NoError(t, gormDB.Create(&expired).Error)
	_, err = theAPI.acquireLocalRebalanceQuoteForExecution(expired.ID, expiredFingerprint, now)
	require.ErrorContains(t, err, "expired")
	require.NoError(t, gormDB.First(&expired, "id = ?", expired.ID).Error)
	require.Equal(t, "expired", expired.State)

	tampered, tamperedFingerprint := testPersistedLocalRebalanceQuote(t, "tampered-quote", now.Add(time.Minute))
	tampered.AmountMsat++
	require.NoError(t, gormDB.Create(&tampered).Error)
	_, err = theAPI.acquireLocalRebalanceQuoteForExecution(tampered.ID, tamperedFingerprint, now)
	require.ErrorContains(t, err, "persisted route material")
	require.NoError(t, gormDB.First(&tampered, "id = ?", tampered.ID).Error)
	require.Equal(t, "quoted", tampered.State)
}

func testPersistedLocalRebalanceQuote(
	t *testing.T, quoteId string, expiresAt time.Time,
) (db.LocalRebalanceQuote, string) {
	t.Helper()
	response := testLocalRebalanceQuoteResponse()
	fingerprint, err := hashLocalRebalanceRouteMaterial(response)
	require.NoError(t, err)
	routeJSON, err := json.Marshal(response.Paths)
	require.NoError(t, err)
	return db.LocalRebalanceQuote{
		ID:                             quoteId,
		State:                          "quoted",
		RequestHash:                    strings.Repeat("1", 64),
		RouteFingerprint:               fingerprint,
		AmountMsat:                     response.AmountMsat,
		TotalRoutingFeeMsat:            response.TotalRoutingFeeMsat,
		MaxRoutingFeeMsat:              response.MaxRoutingFeeMsat,
		MaxTotalDebitMsat:              response.MaxTotalDebitMsat,
		OutgoingChannelId:              response.OutgoingChannelId,
		OutgoingNodePubkey:             response.OutgoingNodePubkey,
		OutgoingShortChannelId:         response.OutgoingShortChannelId,
		IncomingChannelId:              response.IncomingChannelId,
		IncomingNodePubkey:             response.IncomingNodePubkey,
		IncomingShortChannelId:         response.IncomingShortChannelId,
		OutgoingSpendableSnapshotMsat:  response.OutgoingSpendableSnapshotMsat,
		IncomingReceivableSnapshotMsat: response.IncomingReceivableSnapshotMsat,
		RouteJson:                      string(routeJSON),
		ExpiresAt:                      expiresAt,
	}, fingerprint
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
