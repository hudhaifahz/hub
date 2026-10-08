package api

import (
	"encoding/json"
	"errors"
	"fmt"
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

func TestLocalRebalanceRouteFingerprintBindsExecutableRouteBytes(t *testing.T) {
	original := testLocalRebalanceQuoteResponse()
	originalFingerprint, err := hashLocalRebalanceRouteMaterial(original)
	require.NoError(t, err)

	changed := testLocalRebalanceQuoteResponse()
	changed.routeBytes[1] ^= 0xff
	changedFingerprint, err := hashLocalRebalanceRouteMaterial(changed)
	require.NoError(t, err)

	require.NotEqual(t, originalFingerprint, changedFingerprint)
}

func TestLocalRebalanceQuoteResponseDoesNotExposeExecutableRouteBytes(t *testing.T) {
	encoded, err := json.Marshal(testLocalRebalanceQuoteResponse())
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "routeBytes")
	require.NotContains(t, string(encoded), "AQIDBA==")
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
	require.Equal(t, localRebalancePhaseAcquired, persisted.ExecutionPhase)
	require.Equal(t, deriveLocalRebalanceOperationId(quote.ID, fingerprint), persisted.OperationId)
	require.Nil(t, persisted.ExecutedAt)
}

func TestAcquireLocalRebalanceQuoteForExecutionAllowsOnlyOneExecutingQuote(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)

	const contenders = 8
	quotes := make([]db.LocalRebalanceQuote, 0, contenders)
	fingerprints := make([]string, 0, contenders)
	for i := 0; i < contenders; i++ {
		quote, fingerprint := testPersistedLocalRebalanceQuote(
			t,
			fmt.Sprintf("distinct-quote-%d", i),
			now.Add(time.Minute),
		)
		require.NoError(t, gormDB.Create(&quote).Error)
		quotes = append(quotes, quote)
		fingerprints = append(fingerprints, fingerprint)
	}

	start := make(chan struct{})
	results := make(chan error, contenders)
	var waitGroup sync.WaitGroup
	for i := 0; i < contenders; i++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			<-start
			_, acquireErr := theAPI.acquireLocalRebalanceQuoteForExecution(
				quotes[index].ID,
				fingerprints[index],
				now,
			)
			results <- acquireErr
		}(i)
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
	var executingCount int64
	require.NoError(t, gormDB.Model(&db.LocalRebalanceQuote{}).Where("state = ?", "executing").Count(&executingCount).Error)
	require.Equal(t, int64(1), executingCount)
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

	tamperedRoute, tamperedRouteFingerprint := testPersistedLocalRebalanceQuote(t, "tampered-route", now.Add(time.Minute))
	tamperedRoute.RouteBytes[0] ^= 0xff
	require.NoError(t, gormDB.Create(&tamperedRoute).Error)
	_, err = theAPI.acquireLocalRebalanceQuoteForExecution(tamperedRoute.ID, tamperedRouteFingerprint, now)
	require.ErrorContains(t, err, "persisted route material")
	require.NoError(t, gormDB.First(&tamperedRoute, "id = ?", tamperedRoute.ID).Error)
	require.Equal(t, "quoted", tamperedRoute.State)
}

type recordingCircularPaymentExecutor struct {
	prepareCalls   int
	sendCalls      int
	reconcileCalls int
	lastExpiry     uint32
	lastOperation  string
	prepared       *lnclient.PreparedCircularPayment
	reconciliation *lnclient.CircularPaymentReconciliation
	prepareErr     error
	sendErr        error
	reconcileErr   error
}

func (executor *recordingCircularPaymentExecutor) PrepareCircularPayment(
	_ uint64, expirySeconds uint32, operationID string, _ string, _ string, _ uint64,
) (*lnclient.PreparedCircularPayment, error) {
	executor.prepareCalls++
	executor.lastExpiry = expirySeconds
	executor.lastOperation = operationID
	if executor.prepareErr != nil {
		return nil, executor.prepareErr
	}
	prepared := *executor.prepared
	prepared.OperationID = operationID
	prepared.OutboundPaymentID, _ = deriveCircularOutboundPaymentId(operationID)
	executor.prepared.OutboundPaymentID = prepared.OutboundPaymentID
	return &prepared, nil
}

func (executor *recordingCircularPaymentExecutor) SendPreparedCircularPayment(
	operationID string, _ *lnclient.CircularRouteQuote,
) (string, error) {
	executor.sendCalls++
	executor.lastOperation = operationID
	if executor.sendErr != nil {
		return "", executor.sendErr
	}
	return executor.prepared.OutboundPaymentID, nil
}

func (executor *recordingCircularPaymentExecutor) ReconcilePreparedCircularPayment(
	operationID string, _ string, _ string, _ uint64, _ uint64, _ uint64, _ string, _ string,
) (*lnclient.CircularPaymentReconciliation, error) {
	executor.reconcileCalls++
	executor.lastOperation = operationID
	if executor.reconcileErr != nil {
		return nil, executor.reconcileErr
	}
	if executor.reconciliation == nil {
		return nil, errors.New("no reconciliation configured")
	}
	result := *executor.reconciliation
	return &result, nil
}

func TestRunLocalRebalanceExecutionPersistsExactPreparedAndSubmittedState(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, "execute-quote", now.Add(time.Minute))
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)

	result, err := theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now, executor)
	require.NoError(t, err)
	require.Equal(t, localRebalancePhaseSubmitted, result.Phase)
	require.Equal(t, deriveLocalRebalanceOperationId(quote.ID, fingerprint), result.OperationId)
	require.Equal(t, strings.Repeat("a", 64), result.PaymentHash)
	expectedOutboundPaymentId, err := deriveCircularOutboundPaymentId(result.OperationId)
	require.NoError(t, err)
	require.Equal(t, expectedOutboundPaymentId, result.OutboundPaymentId)
	require.Equal(t, 1, executor.prepareCalls)
	require.Equal(t, 1, executor.sendCalls)
	require.Equal(t, uint32(60), executor.lastExpiry)

	var persisted db.LocalRebalanceQuote
	require.NoError(t, gormDB.First(&persisted, "id = ?", quote.ID).Error)
	require.Equal(t, "executing", persisted.State)
	require.Equal(t, localRebalancePhaseSubmitted, persisted.ExecutionPhase)
	require.Equal(t, result.OperationId, persisted.OperationId)
	require.Equal(t, result.PaymentHash, persisted.PreparedPaymentHash)
	require.Equal(t, result.OutboundPaymentId, persisted.OutboundPaymentId)
	require.NotNil(t, persisted.PreparedAt)
	require.NotNil(t, persisted.SubmittedAt)
	require.Nil(t, persisted.ExecutedAt)

	recovered, err := theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now.Add(time.Second), executor)
	require.NoError(t, err)
	require.Equal(t, result, recovered)
	require.Equal(t, 1, executor.prepareCalls)
	require.Equal(t, 1, executor.sendCalls)
}

func TestRunLocalRebalanceExecutionRetriesOnlySameOperationAfterSubmissionUncertainty(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, "recovery-quote", now.Add(time.Minute))
	quote.State = "executing"
	quote.OperationId = deriveLocalRebalanceOperationId(quote.ID, fingerprint)
	quote.ExecutionPhase = localRebalancePhasePrepared
	quote.PreparedPaymentHash = strings.Repeat("a", 64)
	quote.OutboundPaymentId, err = deriveCircularOutboundPaymentId(quote.OperationId)
	require.NoError(t, err)
	quote.PreparedAt = &now
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)

	result, err := theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now, executor)
	require.NoError(t, err)
	require.Equal(t, quote.OperationId, result.OperationId)
	require.Equal(t, quote.OutboundPaymentId, result.OutboundPaymentId)
	require.Equal(t, 1, executor.prepareCalls)
	require.Equal(t, 1, executor.sendCalls)
	require.Equal(t, quote.OperationId, executor.lastOperation)

	// Model a crash after LDK accepted the exact send but before Hub recorded the submitted phase.
	require.NoError(t, gormDB.Model(&db.LocalRebalanceQuote{}).Where("id = ?", quote.ID).Updates(map[string]interface{}{
		"execution_phase": localRebalancePhasePrepared,
		"submitted_at":    nil,
	}).Error)
	recovered, err := theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now.Add(time.Second), executor)
	require.NoError(t, err)
	require.Equal(t, result, recovered)
	require.Equal(t, 2, executor.prepareCalls)
	require.Equal(t, 2, executor.sendCalls)
	require.Equal(t, quote.OperationId, executor.lastOperation)
}

func TestRunLocalRebalanceExecutionUsesExpiredQuoteOnlyForRecovery(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, "expired-recovery", now.Add(-time.Second))
	quote.State = "executing"
	quote.OperationId = deriveLocalRebalanceOperationId(quote.ID, fingerprint)
	quote.ExecutionPhase = localRebalancePhasePrepared
	quote.PreparedPaymentHash = strings.Repeat("a", 64)
	quote.OutboundPaymentId, err = deriveCircularOutboundPaymentId(quote.OperationId)
	require.NoError(t, err)
	quote.PreparedAt = &now
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)

	_, err = theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now, executor)
	require.NoError(t, err)
	require.Equal(t, uint32(0), executor.lastExpiry)
	require.Equal(t, 1, executor.prepareCalls)
	require.Equal(t, 1, executor.sendCalls)
}

func TestRunLocalRebalanceExecutionDoesNotCreateExpiredPreparation(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, "expired-acquired", now.Add(-time.Second))
	quote.State = "executing"
	quote.OperationId = deriveLocalRebalanceOperationId(quote.ID, fingerprint)
	quote.ExecutionPhase = localRebalancePhaseAcquired
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)
	executor.prepareErr = errors.New("no prepared operation exists")

	_, err = theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now, executor)
	require.ErrorContains(t, err, "failed to prepare local rebalance")
	require.Equal(t, uint32(0), executor.lastExpiry)
	require.Equal(t, 1, executor.prepareCalls)
	require.Zero(t, executor.sendCalls)

	var persisted db.LocalRebalanceQuote
	require.NoError(t, gormDB.First(&persisted, "id = ?", quote.ID).Error)
	require.Equal(t, localRebalancePhaseAcquired, persisted.ExecutionPhase)
	require.Empty(t, persisted.PreparedPaymentHash)
	require.Empty(t, persisted.OutboundPaymentId)
}

func TestRunLocalRebalanceExecutionRejectsTamperedSubmittedIdentifiers(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, "tampered-submitted", now.Add(time.Minute))
	quote.State = "executing"
	quote.OperationId = deriveLocalRebalanceOperationId(quote.ID, fingerprint)
	quote.ExecutionPhase = localRebalancePhaseSubmitted
	quote.PreparedPaymentHash = "not-a-payment-hash"
	quote.OutboundPaymentId, err = deriveCircularOutboundPaymentId(quote.OperationId)
	require.NoError(t, err)
	quote.PreparedAt = &now
	quote.SubmittedAt = &now
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)

	_, err = theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now, executor)
	require.ErrorContains(t, err, "missing persisted payment identifiers")
	require.Zero(t, executor.prepareCalls)
	require.Zero(t, executor.sendCalls)
}

func TestReconcileLocalRebalanceTerminalPersistsSuccessAndReplaysWithoutBackend(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Second)
	quote, fingerprint := testPreparedLocalRebalanceQuote(t, "terminal-success", now, localRebalancePhaseSubmitted)
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)
	fee := quote.TotalRoutingFeeMsat
	executor.reconciliation = &lnclient.CircularPaymentReconciliation{
		State:                 lnclient.CircularPaymentStateSucceeded,
		OperationID:           quote.OperationId,
		PaymentHash:           quote.PreparedPaymentHash,
		OutboundPaymentID:     quote.OutboundPaymentId,
		AmountMsat:            quote.AmountMsat,
		ActualRoutingFeeMsat:  &fee,
		LatestUpdateTimestamp: uint64(now.Add(2 * time.Second).Unix()),
		InboundPaymentStatus:  lnclient.CircularPaymentStateSucceeded,
		OutboundPaymentStatus: lnclient.CircularPaymentStateSucceeded,
	}

	result, err := theAPI.reconcileLocalRebalanceTerminal(quote.ID, fingerprint, now.Add(3*time.Second), executor)
	require.NoError(t, err)
	require.Equal(t, localRebalancePhaseSucceeded, result.State)
	require.Equal(t, quote.OperationId, result.OperationId)
	require.Equal(t, quote.PreparedPaymentHash, result.PaymentHash)
	require.Equal(t, quote.OutboundPaymentId, result.OutboundPaymentId)
	require.Equal(t, fee, *result.ActualRoutingFeeMsat)
	require.NotEmpty(t, result.TerminalEvidenceHash)
	require.Empty(t, result.FailureReason)
	require.Equal(t, 1, executor.reconcileCalls)

	var persisted db.LocalRebalanceQuote
	require.NoError(t, gormDB.First(&persisted, "id = ?", quote.ID).Error)
	require.Equal(t, localRebalancePhaseSucceeded, persisted.State)
	require.Equal(t, localRebalancePhaseSucceeded, persisted.ExecutionPhase)
	require.NotNil(t, persisted.ExecutedAt)
	require.NotNil(t, persisted.LightningTerminalAt)
	require.NotNil(t, persisted.ReconciledAt)
	require.Equal(t, result.TerminalEvidenceHash, persisted.TerminalEvidenceHash)

	replayed, err := theAPI.reconcileLocalRebalanceTerminal(quote.ID, fingerprint, now.Add(4*time.Second), executor)
	require.NoError(t, err)
	require.Equal(t, result, replayed)
	require.Equal(t, 1, executor.reconcileCalls)

	nextQuote, nextFingerprint := testPersistedLocalRebalanceQuote(t, "after-terminal-success", now.Add(time.Minute))
	require.NoError(t, gormDB.Create(&nextQuote).Error)
	_, err = theAPI.acquireLocalRebalanceQuoteForExecution(nextQuote.ID, nextFingerprint, now)
	require.NoError(t, err)

	require.NoError(t, gormDB.Model(&db.LocalRebalanceQuote{}).Where("id = ?", quote.ID).Update("terminal_evidence_hash", "tampered").Error)
	_, err = theAPI.reconcileLocalRebalanceTerminal(quote.ID, fingerprint, now.Add(5*time.Second), executor)
	require.ErrorContains(t, err, "evidence hash does not match")
	require.Equal(t, 1, executor.reconcileCalls)
}

func TestReconcileLocalRebalanceTerminalPersistsTwoLegFailure(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Second)
	quote, fingerprint := testPreparedLocalRebalanceQuote(t, "terminal-failure", now, localRebalancePhasePrepared)
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)
	executor.reconciliation = &lnclient.CircularPaymentReconciliation{
		State:                 lnclient.CircularPaymentStateFailed,
		OperationID:           quote.OperationId,
		PaymentHash:           quote.PreparedPaymentHash,
		OutboundPaymentID:     quote.OutboundPaymentId,
		AmountMsat:            quote.AmountMsat,
		LatestUpdateTimestamp: uint64(now.Add(2 * time.Second).Unix()),
		InboundPaymentStatus:  lnclient.CircularPaymentStateFailed,
		OutboundPaymentStatus: lnclient.CircularPaymentStateFailed,
	}

	result, err := theAPI.reconcileLocalRebalanceTerminal(quote.ID, fingerprint, now.Add(3*time.Second), executor)
	require.NoError(t, err)
	require.Equal(t, localRebalancePhaseFailed, result.State)
	require.Nil(t, result.ActualRoutingFeeMsat)
	require.Equal(t, "both exact-bound Lightning payment legs failed", result.FailureReason)
	require.NotEmpty(t, result.TerminalEvidenceHash)

	var persisted db.LocalRebalanceQuote
	require.NoError(t, gormDB.First(&persisted, "id = ?", quote.ID).Error)
	require.Equal(t, localRebalancePhaseFailed, persisted.State)
	require.Equal(t, localRebalancePhaseFailed, persisted.ExecutionPhase)
	require.Nil(t, persisted.ExecutedAt)
	require.Nil(t, persisted.ActualRoutingFeeMsat)
	require.NotNil(t, persisted.LightningTerminalAt)
	require.NotNil(t, persisted.ReconciledAt)
}

func TestReconcileLocalRebalanceTerminalKeepsPendingAndContradictoryOperationsLocked(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Second)
	quote, fingerprint := testPreparedLocalRebalanceQuote(t, "terminal-pending", now, localRebalancePhaseSubmitted)
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)
	executor.reconciliation = &lnclient.CircularPaymentReconciliation{
		State:                 lnclient.CircularPaymentStatePending,
		OperationID:           quote.OperationId,
		PaymentHash:           quote.PreparedPaymentHash,
		OutboundPaymentID:     quote.OutboundPaymentId,
		AmountMsat:            quote.AmountMsat,
		LatestUpdateTimestamp: uint64(now.Unix()),
		InboundPaymentStatus:  lnclient.CircularPaymentStateFailed,
		OutboundPaymentStatus: lnclient.CircularPaymentStatePending,
	}

	result, err := theAPI.reconcileLocalRebalanceTerminal(quote.ID, fingerprint, now, executor)
	require.NoError(t, err)
	require.Equal(t, lnclient.CircularPaymentStatePending, result.State)
	var persisted db.LocalRebalanceQuote
	require.NoError(t, gormDB.First(&persisted, "id = ?", quote.ID).Error)
	require.Equal(t, "executing", persisted.State)
	require.Equal(t, localRebalancePhaseSubmitted, persisted.ExecutionPhase)
	require.Empty(t, persisted.TerminalEvidenceHash)

	fee := quote.TotalRoutingFeeMsat
	executor.reconciliation = &lnclient.CircularPaymentReconciliation{
		State:                 lnclient.CircularPaymentStateSucceeded,
		OperationID:           quote.OperationId,
		PaymentHash:           quote.PreparedPaymentHash,
		OutboundPaymentID:     quote.OutboundPaymentId,
		AmountMsat:            quote.AmountMsat,
		ActualRoutingFeeMsat:  &fee,
		LatestUpdateTimestamp: uint64(now.Add(time.Second).Unix()),
		InboundPaymentStatus:  lnclient.CircularPaymentStateSucceeded,
		OutboundPaymentStatus: lnclient.CircularPaymentStateFailed,
	}
	_, err = theAPI.reconcileLocalRebalanceTerminal(quote.ID, fingerprint, now.Add(time.Second), executor)
	require.ErrorContains(t, err, "two-leg terminal evidence")
	require.NoError(t, gormDB.First(&persisted, "id = ?", quote.ID).Error)
	require.Equal(t, "executing", persisted.State)
	require.Empty(t, persisted.TerminalEvidenceHash)

	blockedQuote, blockedFingerprint := testPersistedLocalRebalanceQuote(t, "blocked-by-pending", now.Add(time.Minute))
	require.NoError(t, gormDB.Create(&blockedQuote).Error)
	_, err = theAPI.acquireLocalRebalanceQuoteForExecution(blockedQuote.ID, blockedFingerprint, now)
	require.Error(t, err)
}

func testPreparedLocalRebalanceQuote(
	t *testing.T, quoteId string, now time.Time, phase string,
) (db.LocalRebalanceQuote, string) {
	t.Helper()
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, quoteId, now.Add(time.Minute))
	quote.State = "executing"
	quote.OperationId = deriveLocalRebalanceOperationId(quote.ID, fingerprint)
	quote.ExecutionPhase = phase
	quote.PreparedPaymentHash = strings.Repeat("a", 64)
	var err error
	quote.OutboundPaymentId, err = deriveCircularOutboundPaymentId(quote.OperationId)
	require.NoError(t, err)
	preparedAt := now
	quote.PreparedAt = &preparedAt
	if phase == localRebalancePhaseSubmitted {
		submittedAt := now.Add(time.Second)
		quote.SubmittedAt = &submittedAt
	}
	return quote, fingerprint
}

func TestRunLocalRebalanceExecutionRejectsMismatchedPreparationBeforeSend(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { test_db.CloseDB(gormDB) })
	theAPI := &api{db: gormDB}
	now := time.Now().UTC().Truncate(time.Microsecond)
	quote, fingerprint := testPersistedLocalRebalanceQuote(t, "mismatch-execution", now.Add(time.Minute))
	require.NoError(t, gormDB.Create(&quote).Error)
	executor := testCircularPaymentExecutor(&quote)
	executor.prepared.LastHopChannelID = "wrong-channel"

	_, err = theAPI.runLocalRebalanceExecution(quote.ID, fingerprint, now, executor)
	require.ErrorContains(t, err, "exact incoming channel")
	require.Equal(t, 1, executor.prepareCalls)
	require.Zero(t, executor.sendCalls)
}

func testCircularPaymentExecutor(quote *db.LocalRebalanceQuote) *recordingCircularPaymentExecutor {
	firstHopShortChannelId, _ := strconv.ParseUint(quote.OutgoingShortChannelId, 10, 64)
	lastHopShortChannelId, _ := strconv.ParseUint(quote.IncomingShortChannelId, 10, 64)
	return &recordingCircularPaymentExecutor{prepared: &lnclient.PreparedCircularPayment{
		PaymentHash:            strings.Repeat("a", 64),
		OutboundPaymentID:      strings.Repeat("b", 64),
		AmountMsat:             quote.AmountMsat,
		MaxRoutingFeeMsat:      quote.MaxRoutingFeeMsat,
		FirstHopChannelID:      quote.OutgoingChannelId,
		FirstHopShortChannelID: firstHopShortChannelId,
		LastHopChannelID:       quote.IncomingChannelId,
		LastHopShortChannelID:  lastHopShortChannelId,
	}}
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
		RouteBytes:                     append([]byte(nil), response.routeBytes...),
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
		routeBytes:                     []byte{0x01, 0x02, 0x03, 0x04},
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
