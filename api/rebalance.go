package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/version"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const rebalanceQuoteLifetime = 5 * time.Minute

const (
	localRebalancePhaseAcquired  = "acquired"
	localRebalancePhasePrepared  = "prepared"
	localRebalancePhaseSubmitted = "submitted"
	localRebalancePhaseSucceeded = "succeeded"
	localRebalancePhaseFailed    = "failed"
)

var errPinnedRebalanceExecutionDisabled = errors.New("pinned rebalance execution is disabled until incoming-channel atomicity is proven")
var errLocalRebalanceExecutionDisabled = errors.New("local circular-route execution is not implemented or authorized")

type localRebalanceExecutionResult struct {
	OperationId       string
	PaymentHash       string
	OutboundPaymentId string
	Phase             string
}

type localRebalanceTerminalResult struct {
	State                string
	OperationId          string
	PaymentHash          string
	OutboundPaymentId    string
	ActualRoutingFeeMsat *uint64
	LightningTerminalAt  *time.Time
	ReconciledAt         *time.Time
	TerminalEvidenceHash string
	FailureReason        string
}

type localRebalanceTerminalEvidenceMaterial struct {
	State                 string  `json:"state"`
	OperationId           string  `json:"operation_id"`
	PaymentHash           string  `json:"payment_hash"`
	OutboundPaymentId     string  `json:"outbound_payment_id"`
	AmountMsat            uint64  `json:"amount_msat"`
	ActualRoutingFeeMsat  *uint64 `json:"actual_routing_fee_msat"`
	LatestUpdateTimestamp uint64  `json:"latest_update_timestamp"`
	InboundPaymentStatus  string  `json:"inbound_payment_status"`
	OutboundPaymentStatus string  `json:"outbound_payment_status"`
}

type rspRebalanceCreateOrderResponse struct {
	OrderId    string `json:"order_id"`
	PayRequest string `json:"pay_request"`
}

// RebalanceChannel deliberately rejects the legacy one-step flow. It created and paid a provider
// invoice in one request and could not enforce either exact local channel.
func (api *api) RebalanceChannel(_ context.Context, _ *RebalanceChannelRequest) (*RebalanceChannelResponse, error) {
	return nil, errors.New("direct rebalance execution is disabled; create and review an exact-channel quote")
}

// QuoteRebalance validates exact local channel identities, obtains a provider invoice without
// paying it, and persists a short-lived single-use review record.
func (api *api) QuoteRebalance(ctx context.Context, request *QuoteRebalanceRequest) (*RebalanceQuoteResponse, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return nil, ErrLNClientNotStarted
	}
	if request == nil || request.AmountMsat == 0 {
		return nil, errors.New("rebalance amount must be positive")
	}
	if request.OutgoingChannelId == "" || request.IncomingChannelId == "" {
		return nil, errors.New("exact outgoing and incoming channel IDs are required")
	}
	if request.OutgoingChannelId == request.IncomingChannelId {
		return nil, errors.New("outgoing and incoming channels must be different")
	}
	if err := validateNodePubkey(request.OutgoingNodePubkey); err != nil {
		return nil, fmt.Errorf("invalid outgoing node pubkey: %w", err)
	}
	if err := validateNodePubkey(request.IncomingNodePubkey); err != nil {
		return nil, fmt.Errorf("invalid incoming node pubkey: %w", err)
	}

	maxInvoiceAmountMsat, ok := checkedAdd(request.AmountMsat, request.MaxProviderFeeMsat)
	if !ok {
		return nil, errors.New("rebalance amount and provider-fee limit overflow")
	}
	maxTotalDebitMsat, ok := checkedAdd(maxInvoiceAmountMsat, request.MaxRoutingFeeMsat)
	if !ok || maxTotalDebitMsat > math.MaxInt64 {
		return nil, errors.New("rebalance maximum debit is too large")
	}

	channels, err := lnClient.ListChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	outgoing, err := findExactChannel(channels, request.OutgoingChannelId, request.OutgoingNodePubkey)
	if err != nil {
		return nil, fmt.Errorf("outgoing channel: %w", err)
	}
	incoming, err := findExactChannel(channels, request.IncomingChannelId, request.IncomingNodePubkey)
	if err != nil {
		return nil, fmt.Errorf("incoming channel: %w", err)
	}
	if !outgoing.Active {
		return nil, errors.New("outgoing channel is not online and usable")
	}
	if !incoming.Active {
		return nil, errors.New("incoming channel is not online and usable")
	}
	if outgoing.LocalSpendableBalanceMsat < 0 || uint64(outgoing.LocalSpendableBalanceMsat) < maxTotalDebitMsat {
		return nil, errors.New("outgoing channel has insufficient spendable balance for the maximum debit")
	}
	if incoming.RemoteBalanceMsat < 0 || uint64(incoming.RemoteBalanceMsat) < request.AmountMsat {
		return nil, errors.New("incoming channel has insufficient receiving capacity")
	}

	receiveMetadata := map[string]interface{}{
		"incoming_channel_id":  request.IncomingChannelId,
		"incoming_node_pubkey": request.IncomingNodePubkey,
	}
	receiveInvoice, err := api.svc.GetTransactionsService().MakeInvoice(
		ctx,
		request.AmountMsat,
		"Alby Hub exact-channel rebalance quote",
		"",
		uint64(rebalanceQuoteLifetime/time.Second),
		receiveMetadata,
		lnClient,
		nil,
		nil,
		&request.IncomingNodePubkey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rebalance receive invoice: %w", err)
	}

	order, err := api.createRebalanceOrder(ctx, receiveInvoice.PaymentRequest, request.IncomingNodePubkey)
	if err != nil {
		return nil, err
	}
	receivePaymentRequest, err := decodepay.Decodepay(receiveInvoice.PaymentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode receive invoice: %w", err)
	}
	paymentRequest, err := decodepay.Decodepay(order.PayRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode provider invoice: %w", err)
	}
	if paymentRequest.Currency != receivePaymentRequest.Currency {
		return nil, errors.New("provider invoice network does not match the node invoice network")
	}
	if paymentRequest.MSatoshi < 0 || uint64(paymentRequest.MSatoshi) < request.AmountMsat {
		return nil, errors.New("provider invoice amount is lower than the requested principal")
	}
	providerFeeMsat := uint64(paymentRequest.MSatoshi) - request.AmountMsat
	if providerFeeMsat > request.MaxProviderFeeMsat {
		return nil, fmt.Errorf("provider fee %d msat exceeds limit %d msat", providerFeeMsat, request.MaxProviderFeeMsat)
	}

	providerExpiry := time.Unix(int64(paymentRequest.CreatedAt), 0).Add(time.Duration(paymentRequest.Expiry) * time.Second)
	expiresAt := time.Now().UTC().Add(rebalanceQuoteLifetime)
	if providerExpiry.Before(expiresAt) {
		expiresAt = providerExpiry.UTC()
	}
	if time.Until(expiresAt) < 30*time.Second {
		return nil, errors.New("provider invoice expires too soon to review safely")
	}

	quoteId, err := randomQuoteId()
	if err != nil {
		return nil, fmt.Errorf("failed to create quote ID: %w", err)
	}
	requestHash, err := hashQuoteMaterial(request, order.OrderId, receivePaymentRequest.PaymentHash, paymentRequest.PaymentHash)
	if err != nil {
		return nil, err
	}
	quote := db.RebalanceQuote{
		ID:                             quoteId,
		State:                          "quoted",
		RequestHash:                    requestHash,
		OrderId:                        order.OrderId,
		ReceivePaymentRequest:          receiveInvoice.PaymentRequest,
		ReceivePaymentHash:             receivePaymentRequest.PaymentHash,
		PaymentRequest:                 order.PayRequest,
		PaymentHash:                    paymentRequest.PaymentHash,
		AmountMsat:                     request.AmountMsat,
		ProviderFeeMsat:                providerFeeMsat,
		MaxProviderFeeMsat:             request.MaxProviderFeeMsat,
		MaxRoutingFeeMsat:              request.MaxRoutingFeeMsat,
		OutgoingChannelId:              request.OutgoingChannelId,
		OutgoingNodePubkey:             request.OutgoingNodePubkey,
		IncomingChannelId:              request.IncomingChannelId,
		IncomingNodePubkey:             request.IncomingNodePubkey,
		OutgoingSpendableSnapshotMsat:  uint64(outgoing.LocalSpendableBalanceMsat),
		IncomingReceivableSnapshotMsat: uint64(incoming.RemoteBalanceMsat),
		ExpiresAt:                      expiresAt,
	}
	if err := api.db.Create(&quote).Error; err != nil {
		return nil, fmt.Errorf("failed to persist rebalance quote: %w", err)
	}

	actualMaxDebit, _ := checkedAdd(request.AmountMsat+providerFeeMsat, request.MaxRoutingFeeMsat)
	return &RebalanceQuoteResponse{
		QuoteId:                        quote.ID,
		AmountMsat:                     quote.AmountMsat,
		ProviderFeeMsat:                quote.ProviderFeeMsat,
		MaxProviderFeeMsat:             quote.MaxProviderFeeMsat,
		MaxRoutingFeeMsat:              quote.MaxRoutingFeeMsat,
		MaxTotalDebitMsat:              actualMaxDebit,
		OutgoingChannelId:              quote.OutgoingChannelId,
		OutgoingNodePubkey:             quote.OutgoingNodePubkey,
		IncomingChannelId:              quote.IncomingChannelId,
		IncomingNodePubkey:             quote.IncomingNodePubkey,
		OutgoingSpendableSnapshotMsat:  quote.OutgoingSpendableSnapshotMsat,
		IncomingReceivableSnapshotMsat: quote.IncomingReceivableSnapshotMsat,
		ExpiresAt:                      quote.ExpiresAt,
		ExecutionEnabled:               false,
		BlockedReason:                  errPinnedRebalanceExecutionDisabled.Error(),
	}, nil
}

// QuoteLocalRebalance constructs and validates a candidate circular route using exact local
// first- and last-hop channels. It persists only an expiring review record: no provider request,
// invoice, probe, HTLC, or payment is created.
func (api *api) QuoteLocalRebalance(ctx context.Context, request *QuoteLocalRebalanceRequest) (*LocalRebalanceQuoteResponse, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return nil, ErrLNClientNotStarted
	}
	quoter, ok := lnClient.(lnclient.CircularRouteQuoter)
	if !ok {
		return nil, errors.New("the active Lightning backend does not support local circular route quotes")
	}
	if request == nil || request.AmountMsat == 0 {
		return nil, errors.New("rebalance amount must be positive")
	}
	if request.OutgoingChannelId == "" || request.IncomingChannelId == "" {
		return nil, errors.New("exact outgoing and incoming channel IDs are required")
	}
	if request.OutgoingChannelId == request.IncomingChannelId {
		return nil, errors.New("outgoing and incoming channels must be different")
	}
	if err := validateNodePubkey(request.OutgoingNodePubkey); err != nil {
		return nil, fmt.Errorf("invalid outgoing node pubkey: %w", err)
	}
	if err := validateNodePubkey(request.IncomingNodePubkey); err != nil {
		return nil, fmt.Errorf("invalid incoming node pubkey: %w", err)
	}
	maxTotalDebitMsat, ok := checkedAdd(request.AmountMsat, request.MaxRoutingFeeMsat)
	if !ok || maxTotalDebitMsat > math.MaxInt64 {
		return nil, errors.New("rebalance maximum debit is too large")
	}

	channels, err := lnClient.ListChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}
	outgoing, err := findExactChannel(channels, request.OutgoingChannelId, request.OutgoingNodePubkey)
	if err != nil {
		return nil, fmt.Errorf("outgoing channel: %w", err)
	}
	incoming, err := findExactChannel(channels, request.IncomingChannelId, request.IncomingNodePubkey)
	if err != nil {
		return nil, fmt.Errorf("incoming channel: %w", err)
	}
	if !outgoing.Active {
		return nil, errors.New("outgoing channel is not online and usable")
	}
	if !incoming.Active {
		return nil, errors.New("incoming channel is not online and usable")
	}
	if outgoing.LocalSpendableBalanceMsat < 0 || uint64(outgoing.LocalSpendableBalanceMsat) < maxTotalDebitMsat {
		return nil, errors.New("outgoing channel has insufficient spendable balance for the maximum debit")
	}
	if incoming.RemoteBalanceMsat < 0 || uint64(incoming.RemoteBalanceMsat) < request.AmountMsat {
		return nil, errors.New("incoming channel has insufficient receiving capacity")
	}

	quote, err := quoter.QuoteCircularRoute(
		request.AmountMsat,
		request.OutgoingChannelId,
		request.IncomingChannelId,
		request.MaxRoutingFeeMsat,
	)
	if err != nil {
		return nil, err
	}
	if quote.AmountMsat != request.AmountMsat {
		return nil, errors.New("local route quote amount does not match the request")
	}
	if quote.FirstHopChannelId != request.OutgoingChannelId || quote.LastHopChannelId != request.IncomingChannelId {
		return nil, errors.New("local route quote did not preserve the exact selected channel IDs")
	}
	if quote.TotalRoutingFeeMsat > request.MaxRoutingFeeMsat {
		return nil, fmt.Errorf("local route fee %d msat exceeds limit %d msat", quote.TotalRoutingFeeMsat, request.MaxRoutingFeeMsat)
	}
	if len(quote.Paths) == 0 {
		return nil, errors.New("local route quote returned no paths")
	}
	if len(quote.RouteBytes) == 0 {
		return nil, errors.New("local route quote returned no executable route bytes")
	}

	paths := make([]LocalCircularRoutePath, 0, len(quote.Paths))
	var quotedAmountMsat uint64
	var quotedFeeMsat uint64
	for _, path := range quote.Paths {
		if len(path.Hops) < 2 {
			return nil, errors.New("local circular route path is too short")
		}
		firstHop := path.Hops[0]
		lastHop := path.Hops[len(path.Hops)-1]
		penultimateHop := path.Hops[len(path.Hops)-2]
		if firstHop.NodeId != request.OutgoingNodePubkey || firstHop.ShortChannelId != quote.FirstHopShortChannelId {
			return nil, errors.New("local route path did not preserve the exact outgoing first hop")
		}
		if penultimateHop.NodeId != request.IncomingNodePubkey || lastHop.ShortChannelId != quote.LastHopShortChannelId {
			return nil, errors.New("local route path did not preserve the exact incoming final hop")
		}
		if lastHop.NodeId != lnClient.GetPubkey() {
			return nil, errors.New("local circular route did not terminate at this node")
		}
		quotedAmountMsat, ok = checkedAdd(quotedAmountMsat, path.AmountMsat)
		if !ok {
			return nil, errors.New("local route path amounts overflow")
		}
		quotedFeeMsat, ok = checkedAdd(quotedFeeMsat, path.FeeMsat)
		if !ok {
			return nil, errors.New("local route path fees overflow")
		}

		hops := make([]LocalCircularRouteHop, 0, len(path.Hops))
		for _, hop := range path.Hops {
			hops = append(hops, LocalCircularRouteHop{
				NodePubkey:     hop.NodeId,
				ShortChannelId: strconv.FormatUint(hop.ShortChannelId, 10),
				FeeMsat:        hop.FeeMsat,
				CltvDelta:      hop.CltvExpiryDelta,
			})
		}
		paths = append(paths, LocalCircularRoutePath{
			Hops:       hops,
			AmountMsat: path.AmountMsat,
			FeeMsat:    path.FeeMsat,
		})
	}
	if quotedAmountMsat != quote.AmountMsat || quotedFeeMsat != quote.TotalRoutingFeeMsat {
		return nil, errors.New("local route path totals do not match the quote")
	}
	actualTotalDebitMsat, ok := checkedAdd(quote.AmountMsat, quote.TotalRoutingFeeMsat)
	if !ok {
		return nil, errors.New("local route quote debit overflow")
	}

	expiresAt := time.Now().UTC().Add(rebalanceQuoteLifetime)
	response := &LocalRebalanceQuoteResponse{
		AmountMsat:                     quote.AmountMsat,
		TotalRoutingFeeMsat:            quote.TotalRoutingFeeMsat,
		MaxRoutingFeeMsat:              request.MaxRoutingFeeMsat,
		MaxTotalDebitMsat:              actualTotalDebitMsat,
		OutgoingChannelId:              request.OutgoingChannelId,
		OutgoingNodePubkey:             request.OutgoingNodePubkey,
		OutgoingShortChannelId:         strconv.FormatUint(quote.FirstHopShortChannelId, 10),
		IncomingChannelId:              request.IncomingChannelId,
		IncomingNodePubkey:             request.IncomingNodePubkey,
		IncomingShortChannelId:         strconv.FormatUint(quote.LastHopShortChannelId, 10),
		OutgoingSpendableSnapshotMsat:  uint64(outgoing.LocalSpendableBalanceMsat),
		IncomingReceivableSnapshotMsat: uint64(incoming.RemoteBalanceMsat),
		Paths:                          paths,
		ExpiresAt:                      expiresAt,
		ExecutionEnabled:               false,
		BlockedReason:                  errLocalRebalanceExecutionDisabled.Error(),
		routeBytes:                     append([]byte(nil), quote.RouteBytes...),
	}
	quoteId, err := randomQuoteId()
	if err != nil {
		return nil, fmt.Errorf("failed to create local quote ID: %w", err)
	}
	requestHash, err := hashLocalRebalanceRequest(request)
	if err != nil {
		return nil, err
	}
	routeFingerprint, err := hashLocalRebalanceRouteMaterial(response)
	if err != nil {
		return nil, err
	}
	routeJSON, err := json.Marshal(paths)
	if err != nil {
		return nil, fmt.Errorf("failed to encode local rebalance route: %w", err)
	}
	response.QuoteId = quoteId
	response.RouteFingerprint = routeFingerprint

	localQuote := db.LocalRebalanceQuote{
		ID:                             quoteId,
		State:                          "quoted",
		RequestHash:                    requestHash,
		RouteFingerprint:               routeFingerprint,
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
	}
	if err := api.db.Create(&localQuote).Error; err != nil {
		return nil, fmt.Errorf("failed to persist local rebalance quote: %w", err)
	}

	return response, nil
}

func deriveLocalRebalanceOperationId(quoteId string, routeFingerprint string) string {
	hasher := sha256.New()
	hasher.Write([]byte("alby-hub local rebalance operation id v1"))
	hasher.Write([]byte{0})
	hasher.Write([]byte(quoteId))
	hasher.Write([]byte{0})
	hasher.Write([]byte(routeFingerprint))
	return hex.EncodeToString(hasher.Sum(nil))
}

func deriveCircularOutboundPaymentId(operationId string) (string, error) {
	decodedOperationId, err := hex.DecodeString(operationId)
	if err != nil || len(decodedOperationId) != sha256.Size {
		return "", errors.New("local rebalance operation ID is invalid")
	}
	material := append([]byte("ldk-node circular outbound payment id v1"), decodedOperationId...)
	hash := sha256.Sum256(material)
	return hex.EncodeToString(hash[:]), nil
}

func localRebalanceRouteFromRecord(quote *db.LocalRebalanceQuote) (*lnclient.CircularRouteQuote, error) {
	if quote == nil || len(quote.RouteBytes) == 0 {
		return nil, errors.New("local rebalance quote has no executable route bytes")
	}
	firstHopShortChannelId, err := strconv.ParseUint(quote.OutgoingShortChannelId, 10, 64)
	if err != nil {
		return nil, errors.New("local rebalance quote has an invalid outgoing short channel ID")
	}
	lastHopShortChannelId, err := strconv.ParseUint(quote.IncomingShortChannelId, 10, 64)
	if err != nil {
		return nil, errors.New("local rebalance quote has an invalid incoming short channel ID")
	}
	var storedPaths []LocalCircularRoutePath
	if err := json.Unmarshal([]byte(quote.RouteJson), &storedPaths); err != nil {
		return nil, errors.New("local rebalance quote route is invalid")
	}
	recomputedFingerprint, err := hashLocalRebalanceRouteMaterial(&LocalRebalanceQuoteResponse{
		AmountMsat:                     quote.AmountMsat,
		TotalRoutingFeeMsat:            quote.TotalRoutingFeeMsat,
		MaxRoutingFeeMsat:              quote.MaxRoutingFeeMsat,
		MaxTotalDebitMsat:              quote.MaxTotalDebitMsat,
		OutgoingChannelId:              quote.OutgoingChannelId,
		OutgoingNodePubkey:             quote.OutgoingNodePubkey,
		OutgoingShortChannelId:         quote.OutgoingShortChannelId,
		IncomingChannelId:              quote.IncomingChannelId,
		IncomingNodePubkey:             quote.IncomingNodePubkey,
		IncomingShortChannelId:         quote.IncomingShortChannelId,
		OutgoingSpendableSnapshotMsat:  quote.OutgoingSpendableSnapshotMsat,
		IncomingReceivableSnapshotMsat: quote.IncomingReceivableSnapshotMsat,
		Paths:                          storedPaths,
		routeBytes:                     append([]byte(nil), quote.RouteBytes...),
	})
	if err != nil || recomputedFingerprint != quote.RouteFingerprint {
		return nil, errors.New("local rebalance quote fingerprint does not match its persisted route material")
	}

	paths := make([]lnclient.CircularRoutePath, 0, len(storedPaths))
	for _, storedPath := range storedPaths {
		hops := make([]lnclient.CircularRouteHop, 0, len(storedPath.Hops))
		for _, storedHop := range storedPath.Hops {
			shortChannelId, err := strconv.ParseUint(storedHop.ShortChannelId, 10, 64)
			if err != nil {
				return nil, errors.New("local rebalance quote contains an invalid hop short channel ID")
			}
			hops = append(hops, lnclient.CircularRouteHop{
				NodeId:          storedHop.NodePubkey,
				ShortChannelId:  shortChannelId,
				FeeMsat:         storedHop.FeeMsat,
				CltvExpiryDelta: storedHop.CltvDelta,
			})
		}
		paths = append(paths, lnclient.CircularRoutePath{
			Hops:       hops,
			AmountMsat: storedPath.AmountMsat,
			FeeMsat:    storedPath.FeeMsat,
		})
	}

	return &lnclient.CircularRouteQuote{
		AmountMsat:             quote.AmountMsat,
		TotalRoutingFeeMsat:    quote.TotalRoutingFeeMsat,
		FirstHopChannelId:      quote.OutgoingChannelId,
		FirstHopShortChannelId: firstHopShortChannelId,
		LastHopChannelId:       quote.IncomingChannelId,
		LastHopShortChannelId:  lastHopShortChannelId,
		Paths:                  paths,
		RouteBytes:             append([]byte(nil), quote.RouteBytes...),
	}, nil
}

// acquireLocalRebalanceQuoteForExecution performs the durable compare-and-set that must precede
// any future value-moving backend call. The quote's complete persisted route material is
// re-fingerprinted first, then only the exact still-quoted, unexpired record can transition to
// executing. A crash after this transition is fail-closed: the quote cannot be acquired again.
func (api *api) acquireLocalRebalanceQuoteForExecution(
	quoteId string, expectedRouteFingerprint string, now time.Time,
) (*db.LocalRebalanceQuote, error) {
	if quoteId == "" {
		return nil, errors.New("quote ID is required")
	}
	if expectedRouteFingerprint == "" {
		return nil, errors.New("route fingerprint is required")
	}
	now = now.UTC()

	var quote db.LocalRebalanceQuote
	if err := api.db.First(&quote, "id = ?", quoteId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("local rebalance quote was not found")
		}
		return nil, fmt.Errorf("failed to load local rebalance quote: %w", err)
	}
	if quote.State != "quoted" {
		return nil, fmt.Errorf("local rebalance quote is not executable in state %s", quote.State)
	}
	if !now.Before(quote.ExpiresAt) {
		result := api.db.Model(&db.LocalRebalanceQuote{}).
			Where("id = ? AND state = ?", quote.ID, "quoted").
			Updates(map[string]interface{}{
				"state":          "expired",
				"failure_reason": "quote expired",
				"updated_at":     now,
			})
		if result.Error != nil {
			return nil, fmt.Errorf("failed to expire local rebalance quote: %w", result.Error)
		}
		return nil, errors.New("local rebalance quote has expired")
	}
	if quote.RouteFingerprint != expectedRouteFingerprint {
		return nil, errors.New("local rebalance quote fingerprint does not match the reviewed route")
	}
	if _, err := localRebalanceRouteFromRecord(&quote); err != nil {
		return nil, err
	}
	operationId := deriveLocalRebalanceOperationId(quote.ID, quote.RouteFingerprint)

	result := api.db.Model(&db.LocalRebalanceQuote{}).
		Where("id = ? AND state = ? AND route_fingerprint = ? AND expires_at > ?", quote.ID, "quoted", expectedRouteFingerprint, now).
		Updates(map[string]interface{}{
			"state":           "executing",
			"operation_id":    operationId,
			"execution_phase": localRebalancePhaseAcquired,
			"updated_at":      now,
		})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to acquire local rebalance quote: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return nil, errors.New("local rebalance quote was already acquired or is no longer executable")
	}
	if err := api.db.First(&quote, "id = ?", quote.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload acquired local rebalance quote: %w", err)
	}
	if quote.State != "executing" || quote.OperationId != operationId || quote.ExecutionPhase != localRebalancePhaseAcquired {
		return nil, errors.New("local rebalance quote acquisition did not persist")
	}
	return &quote, nil
}

func (api *api) loadOrAcquireLocalRebalanceExecution(
	quoteId string, expectedRouteFingerprint string, now time.Time,
) (*db.LocalRebalanceQuote, error) {
	var quote db.LocalRebalanceQuote
	if err := api.db.First(&quote, "id = ?", quoteId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("local rebalance quote was not found")
		}
		return nil, fmt.Errorf("failed to load local rebalance quote: %w", err)
	}
	if quote.State == "quoted" {
		return api.acquireLocalRebalanceQuoteForExecution(quoteId, expectedRouteFingerprint, now)
	}
	if quote.State != "executing" {
		return nil, fmt.Errorf("local rebalance quote is not recoverable in state %s", quote.State)
	}
	if quote.RouteFingerprint != expectedRouteFingerprint {
		return nil, errors.New("local rebalance quote fingerprint does not match the reviewed route")
	}
	if quote.OperationId != deriveLocalRebalanceOperationId(quote.ID, quote.RouteFingerprint) {
		return nil, errors.New("local rebalance operation ID does not match the reviewed quote")
	}
	switch quote.ExecutionPhase {
	case localRebalancePhaseAcquired, localRebalancePhasePrepared, localRebalancePhaseSubmitted:
	default:
		return nil, errors.New("local rebalance quote has an invalid execution phase")
	}
	if _, err := localRebalanceRouteFromRecord(&quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

func validatePreparedCircularPayment(quote *db.LocalRebalanceQuote, route *lnclient.CircularRouteQuote, prepared *lnclient.PreparedCircularPayment) error {
	if prepared == nil {
		return errors.New("circular payment preparation returned no result")
	}
	if prepared.OperationID != quote.OperationId || prepared.AmountMsat != quote.AmountMsat || prepared.MaxRoutingFeeMsat != quote.MaxRoutingFeeMsat {
		return errors.New("circular payment preparation does not match the acquired quote")
	}
	if prepared.FirstHopChannelID != quote.OutgoingChannelId || prepared.FirstHopShortChannelID != route.FirstHopShortChannelId {
		return errors.New("circular payment preparation does not match the exact outgoing channel")
	}
	if prepared.LastHopChannelID != quote.IncomingChannelId || prepared.LastHopShortChannelID != route.LastHopShortChannelId {
		return errors.New("circular payment preparation does not match the exact incoming channel")
	}
	for _, value := range []string{prepared.PaymentHash, prepared.OutboundPaymentID} {
		decoded, err := hex.DecodeString(value)
		if err != nil || len(decoded) != sha256.Size {
			return errors.New("circular payment preparation returned an invalid payment identifier")
		}
	}
	expectedOutboundPaymentId, err := deriveCircularOutboundPaymentId(quote.OperationId)
	if err != nil || prepared.OutboundPaymentID != expectedOutboundPaymentId {
		return errors.New("circular payment preparation returned an unexpected outbound payment ID")
	}
	return nil
}

func (api *api) persistLocalRebalancePreparation(
	quote *db.LocalRebalanceQuote, prepared *lnclient.PreparedCircularPayment, now time.Time,
) (*db.LocalRebalanceQuote, error) {
	if quote.ExecutionPhase == localRebalancePhaseAcquired {
		preparedAt := now.UTC()
		result := api.db.Model(&db.LocalRebalanceQuote{}).
			Where("id = ? AND state = ? AND operation_id = ? AND execution_phase = ?", quote.ID, "executing", quote.OperationId, localRebalancePhaseAcquired).
			Updates(map[string]interface{}{
				"execution_phase":       localRebalancePhasePrepared,
				"prepared_payment_hash": prepared.PaymentHash,
				"outbound_payment_id":   prepared.OutboundPaymentID,
				"prepared_at":           preparedAt,
				"updated_at":            preparedAt,
			})
		if result.Error != nil {
			return nil, fmt.Errorf("failed to persist local rebalance preparation: %w", result.Error)
		}
	}

	var persisted db.LocalRebalanceQuote
	if err := api.db.First(&persisted, "id = ?", quote.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload local rebalance preparation: %w", err)
	}
	if persisted.State != "executing" || persisted.OperationId != quote.OperationId {
		return nil, errors.New("local rebalance preparation lost its execution binding")
	}
	if persisted.ExecutionPhase != localRebalancePhasePrepared && persisted.ExecutionPhase != localRebalancePhaseSubmitted {
		return nil, errors.New("local rebalance preparation did not persist")
	}
	if persisted.PreparedPaymentHash != prepared.PaymentHash || persisted.OutboundPaymentId != prepared.OutboundPaymentID {
		return nil, errors.New("local rebalance preparation identifiers do not match persisted state")
	}
	return &persisted, nil
}

func (api *api) persistLocalRebalanceSubmission(
	quote *db.LocalRebalanceQuote, outboundPaymentId string, now time.Time,
) (*db.LocalRebalanceQuote, error) {
	if quote.OutboundPaymentId != outboundPaymentId {
		return nil, errors.New("submitted payment ID does not match the prepared operation")
	}
	if quote.ExecutionPhase == localRebalancePhasePrepared {
		submittedAt := now.UTC()
		result := api.db.Model(&db.LocalRebalanceQuote{}).
			Where("id = ? AND state = ? AND operation_id = ? AND execution_phase = ? AND outbound_payment_id = ?", quote.ID, "executing", quote.OperationId, localRebalancePhasePrepared, outboundPaymentId).
			Updates(map[string]interface{}{
				"execution_phase": localRebalancePhaseSubmitted,
				"submitted_at":    submittedAt,
				"updated_at":      submittedAt,
			})
		if result.Error != nil {
			return nil, fmt.Errorf("failed to persist local rebalance submission: %w", result.Error)
		}
	}

	var persisted db.LocalRebalanceQuote
	if err := api.db.First(&persisted, "id = ?", quote.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload local rebalance submission: %w", err)
	}
	if persisted.State != "executing" || persisted.ExecutionPhase != localRebalancePhaseSubmitted || persisted.OperationId != quote.OperationId || persisted.OutboundPaymentId != outboundPaymentId {
		return nil, errors.New("local rebalance submission did not persist with the exact operation binding")
	}
	return &persisted, nil
}

// runLocalRebalanceExecution is deliberately not connected to an HTTP, Wails, or UI entrypoint.
// It exists so the full durable transition can be tested before any value-moving feature is
// authorized. A caller must still provide the exact reviewed fingerprint and an explicit backend
// implementing crash-safe prepared circular payments.
func (api *api) runLocalRebalanceExecution(
	quoteId string, expectedRouteFingerprint string, now time.Time,
	executor lnclient.PreparedCircularPaymentClient,
) (*localRebalanceExecutionResult, error) {
	if executor == nil {
		return nil, errors.New("prepared circular payment backend is required")
	}
	quote, err := api.loadOrAcquireLocalRebalanceExecution(quoteId, expectedRouteFingerprint, now)
	if err != nil {
		return nil, err
	}
	route, err := localRebalanceRouteFromRecord(quote)
	if err != nil {
		return nil, err
	}
	if quote.ExecutionPhase == localRebalancePhaseSubmitted {
		expectedOutboundPaymentId, err := deriveCircularOutboundPaymentId(quote.OperationId)
		paymentHash, paymentHashErr := hex.DecodeString(quote.PreparedPaymentHash)
		if err != nil || paymentHashErr != nil || len(paymentHash) != sha256.Size || quote.OutboundPaymentId != expectedOutboundPaymentId {
			return nil, errors.New("submitted local rebalance is missing persisted payment identifiers")
		}
		return &localRebalanceExecutionResult{
			OperationId:       quote.OperationId,
			PaymentHash:       quote.PreparedPaymentHash,
			OutboundPaymentId: quote.OutboundPaymentId,
			Phase:             quote.ExecutionPhase,
		}, nil
	}

	expirySeconds := uint32(0)
	if now.Before(quote.ExpiresAt) {
		remaining := quote.ExpiresAt.Sub(now)
		expirySeconds = uint32((remaining + time.Second - 1) / time.Second)
		if expirySeconds > uint32(rebalanceQuoteLifetime/time.Second) {
			expirySeconds = uint32(rebalanceQuoteLifetime / time.Second)
		}
	}
	prepared, err := executor.PrepareCircularPayment(
		quote.AmountMsat,
		expirySeconds,
		quote.OperationId,
		quote.OutgoingChannelId,
		quote.IncomingChannelId,
		quote.MaxRoutingFeeMsat,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare local rebalance: %w", err)
	}
	if err := validatePreparedCircularPayment(quote, route, prepared); err != nil {
		return nil, err
	}
	quote, err = api.persistLocalRebalancePreparation(quote, prepared, now)
	if err != nil {
		return nil, err
	}
	if quote.ExecutionPhase == localRebalancePhaseSubmitted {
		return &localRebalanceExecutionResult{
			OperationId:       quote.OperationId,
			PaymentHash:       quote.PreparedPaymentHash,
			OutboundPaymentId: quote.OutboundPaymentId,
			Phase:             quote.ExecutionPhase,
		}, nil
	}

	outboundPaymentId, err := executor.SendPreparedCircularPayment(quote.OperationId, route)
	if err != nil {
		return nil, fmt.Errorf("failed to submit local rebalance: %w", err)
	}
	quote, err = api.persistLocalRebalanceSubmission(quote, outboundPaymentId, now)
	if err != nil {
		return nil, err
	}
	return &localRebalanceExecutionResult{
		OperationId:       quote.OperationId,
		PaymentHash:       quote.PreparedPaymentHash,
		OutboundPaymentId: quote.OutboundPaymentId,
		Phase:             quote.ExecutionPhase,
	}, nil
}

func hashLocalRebalanceTerminalEvidence(evidence *lnclient.CircularPaymentReconciliation) (string, error) {
	if evidence == nil {
		return "", errors.New("local rebalance terminal evidence is required")
	}
	encoded, err := json.Marshal(localRebalanceTerminalEvidenceMaterial{
		State:                 evidence.State,
		OperationId:           evidence.OperationID,
		PaymentHash:           evidence.PaymentHash,
		OutboundPaymentId:     evidence.OutboundPaymentID,
		AmountMsat:            evidence.AmountMsat,
		ActualRoutingFeeMsat:  evidence.ActualRoutingFeeMsat,
		LatestUpdateTimestamp: evidence.LatestUpdateTimestamp,
		InboundPaymentStatus:  evidence.InboundPaymentStatus,
		OutboundPaymentStatus: evidence.OutboundPaymentStatus,
	})
	if err != nil {
		return "", fmt.Errorf("failed to encode local rebalance terminal evidence: %w", err)
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}

func validateLocalRebalanceReconciliation(
	quote *db.LocalRebalanceQuote,
	evidence *lnclient.CircularPaymentReconciliation,
) error {
	if quote == nil || evidence == nil {
		return errors.New("local rebalance reconciliation evidence is required")
	}
	if evidence.OperationID != quote.OperationId || evidence.PaymentHash != quote.PreparedPaymentHash || evidence.OutboundPaymentID != quote.OutboundPaymentId || evidence.AmountMsat != quote.AmountMsat {
		return errors.New("local rebalance reconciliation does not match the exact prepared operation")
	}
	switch evidence.State {
	case lnclient.CircularPaymentStatePending:
		if evidence.ActualRoutingFeeMsat != nil {
			return errors.New("pending local rebalance unexpectedly reports a routing fee")
		}
	case lnclient.CircularPaymentStateSucceeded:
		if evidence.InboundPaymentStatus != lnclient.CircularPaymentStateSucceeded || evidence.OutboundPaymentStatus != lnclient.CircularPaymentStateSucceeded || evidence.ActualRoutingFeeMsat == nil || *evidence.ActualRoutingFeeMsat != quote.TotalRoutingFeeMsat || *evidence.ActualRoutingFeeMsat > quote.MaxRoutingFeeMsat || evidence.LatestUpdateTimestamp == 0 || evidence.LatestUpdateTimestamp > math.MaxInt64 {
			return errors.New("successful local rebalance lacks exact two-leg terminal evidence")
		}
	case lnclient.CircularPaymentStateFailed:
		if evidence.InboundPaymentStatus != lnclient.CircularPaymentStateFailed || evidence.OutboundPaymentStatus != lnclient.CircularPaymentStateFailed || evidence.ActualRoutingFeeMsat != nil || evidence.LatestUpdateTimestamp == 0 || evidence.LatestUpdateTimestamp > math.MaxInt64 {
			return errors.New("failed local rebalance lacks exact two-leg terminal evidence")
		}
	default:
		return errors.New("local rebalance reconciliation returned an unknown state")
	}
	return nil
}

func localRebalanceTerminalResultFromQuote(quote *db.LocalRebalanceQuote) (*localRebalanceTerminalResult, error) {
	if quote == nil || (quote.State != localRebalancePhaseSucceeded && quote.State != localRebalancePhaseFailed) || quote.ExecutionPhase != quote.State || quote.OperationId != deriveLocalRebalanceOperationId(quote.ID, quote.RouteFingerprint) || quote.PreparedPaymentHash == "" || quote.OutboundPaymentId == "" || quote.TerminalEvidenceHash == "" || quote.LightningTerminalAt == nil || quote.ReconciledAt == nil {
		return nil, errors.New("persisted local rebalance terminal record is incomplete")
	}
	if quote.State == localRebalancePhaseSucceeded {
		if quote.ActualRoutingFeeMsat == nil || *quote.ActualRoutingFeeMsat != quote.TotalRoutingFeeMsat || *quote.ActualRoutingFeeMsat > quote.MaxRoutingFeeMsat || quote.ExecutedAt == nil || quote.FailureReason != "" {
			return nil, errors.New("persisted successful local rebalance evidence is invalid")
		}
	} else if quote.ActualRoutingFeeMsat != nil || quote.ExecutedAt != nil || quote.FailureReason == "" {
		return nil, errors.New("persisted failed local rebalance evidence is invalid")
	}
	evidence := &lnclient.CircularPaymentReconciliation{
		State:                 quote.State,
		OperationID:           quote.OperationId,
		PaymentHash:           quote.PreparedPaymentHash,
		OutboundPaymentID:     quote.OutboundPaymentId,
		AmountMsat:            quote.AmountMsat,
		ActualRoutingFeeMsat:  quote.ActualRoutingFeeMsat,
		LatestUpdateTimestamp: uint64(quote.LightningTerminalAt.Unix()),
		InboundPaymentStatus:  quote.State,
		OutboundPaymentStatus: quote.State,
	}
	evidenceHash, err := hashLocalRebalanceTerminalEvidence(evidence)
	if err != nil || evidenceHash != quote.TerminalEvidenceHash {
		return nil, errors.New("persisted local rebalance terminal evidence hash does not match")
	}
	return &localRebalanceTerminalResult{
		State:                quote.State,
		OperationId:          quote.OperationId,
		PaymentHash:          quote.PreparedPaymentHash,
		OutboundPaymentId:    quote.OutboundPaymentId,
		ActualRoutingFeeMsat: quote.ActualRoutingFeeMsat,
		LightningTerminalAt:  quote.LightningTerminalAt,
		ReconciledAt:         quote.ReconciledAt,
		TerminalEvidenceHash: quote.TerminalEvidenceHash,
		FailureReason:        quote.FailureReason,
	}, nil
}

func (api *api) persistLocalRebalanceTerminalEvidence(
	quote *db.LocalRebalanceQuote,
	evidence *lnclient.CircularPaymentReconciliation,
	now time.Time,
) (*localRebalanceTerminalResult, error) {
	if evidence.State == lnclient.CircularPaymentStatePending {
		return &localRebalanceTerminalResult{
			State:             evidence.State,
			OperationId:       quote.OperationId,
			PaymentHash:       quote.PreparedPaymentHash,
			OutboundPaymentId: quote.OutboundPaymentId,
		}, nil
	}
	evidenceHash, err := hashLocalRebalanceTerminalEvidence(evidence)
	if err != nil {
		return nil, err
	}
	lightningTerminalAt := time.Unix(int64(evidence.LatestUpdateTimestamp), 0).UTC()
	reconciledAt := now.UTC()
	updates := map[string]interface{}{
		"state":                   evidence.State,
		"execution_phase":         evidence.State,
		"actual_routing_fee_msat": evidence.ActualRoutingFeeMsat,
		"lightning_terminal_at":   lightningTerminalAt,
		"reconciled_at":           reconciledAt,
		"terminal_evidence_hash":  evidenceHash,
		"updated_at":              reconciledAt,
	}
	if evidence.State == lnclient.CircularPaymentStateSucceeded {
		updates["executed_at"] = lightningTerminalAt
		updates["failure_reason"] = ""
	} else {
		updates["executed_at"] = nil
		updates["failure_reason"] = "both exact-bound Lightning payment legs failed"
	}
	result := api.db.Model(&db.LocalRebalanceQuote{}).
		Where("id = ? AND state = ? AND operation_id = ? AND execution_phase IN ? AND prepared_payment_hash = ? AND outbound_payment_id = ?", quote.ID, "executing", quote.OperationId, []string{localRebalancePhasePrepared, localRebalancePhaseSubmitted}, quote.PreparedPaymentHash, quote.OutboundPaymentId).
		Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to persist local rebalance terminal evidence: %w", result.Error)
	}
	var persisted db.LocalRebalanceQuote
	if err := api.db.First(&persisted, "id = ?", quote.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload local rebalance terminal evidence: %w", err)
	}
	terminalResult, err := localRebalanceTerminalResultFromQuote(&persisted)
	if err != nil {
		return nil, err
	}
	if terminalResult.TerminalEvidenceHash != evidenceHash || terminalResult.State != evidence.State {
		return nil, errors.New("persisted local rebalance terminal evidence does not match the reconciled operation")
	}
	return terminalResult, nil
}

// reconcileLocalRebalanceTerminal is deliberately unreachable from HTTP, Wails, and UI routes.
// It reads the two exact durable Lightning payment records and releases the single-operation lock
// only after an idempotent terminal compare-and-set.
func (api *api) reconcileLocalRebalanceTerminal(
	quoteId string,
	expectedRouteFingerprint string,
	now time.Time,
	executor lnclient.PreparedCircularPaymentClient,
) (*localRebalanceTerminalResult, error) {
	if executor == nil {
		return nil, errors.New("prepared circular payment backend is required")
	}
	var quote db.LocalRebalanceQuote
	if err := api.db.First(&quote, "id = ?", quoteId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("local rebalance quote was not found")
		}
		return nil, fmt.Errorf("failed to load local rebalance quote: %w", err)
	}
	if quote.RouteFingerprint != expectedRouteFingerprint || quote.OperationId != deriveLocalRebalanceOperationId(quote.ID, quote.RouteFingerprint) {
		return nil, errors.New("local rebalance terminal reconciliation does not match the reviewed operation")
	}
	if _, err := localRebalanceRouteFromRecord(&quote); err != nil {
		return nil, err
	}
	if quote.State == localRebalancePhaseSucceeded || quote.State == localRebalancePhaseFailed {
		return localRebalanceTerminalResultFromQuote(&quote)
	}
	if quote.State != "executing" || (quote.ExecutionPhase != localRebalancePhasePrepared && quote.ExecutionPhase != localRebalancePhaseSubmitted) || quote.PreparedPaymentHash == "" || quote.OutboundPaymentId == "" {
		return nil, errors.New("local rebalance is not ready for terminal reconciliation")
	}
	evidence, err := executor.ReconcilePreparedCircularPayment(
		quote.OperationId,
		quote.PreparedPaymentHash,
		quote.OutboundPaymentId,
		quote.AmountMsat,
		quote.TotalRoutingFeeMsat,
		quote.MaxRoutingFeeMsat,
		quote.OutgoingChannelId,
		quote.IncomingChannelId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile local rebalance payment records: %w", err)
	}
	if err := validateLocalRebalanceReconciliation(&quote, evidence); err != nil {
		return nil, err
	}
	return api.persistLocalRebalanceTerminalEvidence(&quote, evidence, now)
}

// ExecuteRebalance is intentionally fail-closed. Keeping the endpoint and persisted quote contract
// separate lets the owner review the complete action packet, but no payment can start until the
// provider protocol has been proven atomic with pre-claim incoming-channel enforcement.
func (api *api) ExecuteRebalance(_ context.Context, request *ExecuteRebalanceRequest) (*RebalanceChannelResponse, error) {
	if request == nil || request.QuoteId == "" {
		return nil, errors.New("quote ID is required")
	}
	var quote db.RebalanceQuote
	if err := api.db.First(&quote, "id = ?", request.QuoteId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("rebalance quote was not found")
		}
		return nil, fmt.Errorf("failed to load rebalance quote: %w", err)
	}
	if quote.State != "quoted" {
		return nil, fmt.Errorf("rebalance quote is not executable in state %s", quote.State)
	}
	if !time.Now().UTC().Before(quote.ExpiresAt) {
		api.db.Model(&quote).Updates(map[string]interface{}{"state": "expired", "failure_reason": "quote expired"})
		return nil, errors.New("rebalance quote has expired")
	}
	return nil, errPinnedRebalanceExecutionDisabled
}

func (api *api) createRebalanceOrder(ctx context.Context, payRequest string, incomingNodePubkey string) (*rspRebalanceCreateOrderResponse, error) {
	payloadBytes, err := json.Marshal(struct {
		Token                   string `json:"token"`
		PayRequest              string `json:"pay_request"`
		PayThroughThisPublicKey string `json:"pay_through_this_public_key"`
	}{
		Token:                   "alby-hub",
		PayRequest:              payRequest,
		PayThroughThisPublicKey: incomingNodePubkey,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.cfg.GetEnv().RebalanceServiceUrl+"/api/rebalance/v1/create_order", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlbyHub/"+version.Tag)
	res, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request rebalance quote: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, errors.New("failed to read rebalance quote response")
	}
	if res.StatusCode != http.StatusOK {
		providerErr := rebalanceProviderError(res.StatusCode, body)
		logger.Logger.WithFields(logrus.Fields{"statusCode": res.StatusCode}).WithError(providerErr).Error("rebalance create_order endpoint returned non-success code")
		return nil, providerErr
	}
	var order rspRebalanceCreateOrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, errors.New("failed to decode rebalance quote response")
	}
	if order.OrderId == "" || order.PayRequest == "" {
		return nil, errors.New("rebalance quote response is missing required fields")
	}
	return &order, nil
}

func rebalanceProviderError(statusCode int, body []byte) error {
	reason := safeRebalanceProviderReason(body)
	if reason == "" {
		return fmt.Errorf("rebalance quote provider returned HTTP %d", statusCode)
	}
	return fmt.Errorf("rebalance quote provider returned HTTP %d: %s", statusCode, reason)
}

func safeRebalanceProviderReason(body []byte) string {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	for _, key := range []string{"message", "error", "detail", "reason"} {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		var reason string
		if err := json.Unmarshal(raw, &reason); err != nil {
			if key != "error" {
				continue
			}
			var nested map[string]json.RawMessage
			if err := json.Unmarshal(raw, &nested); err != nil {
				continue
			}
			for _, nestedKey := range []string{"message", "detail", "reason"} {
				nestedRaw, ok := nested[nestedKey]
				if !ok || json.Unmarshal(nestedRaw, &reason) != nil || reason == "" {
					continue
				}
				break
			}
		}
		reason = strings.TrimSpace(strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == '\t' {
				return ' '
			}
			if r < 0x20 || r == 0x7f {
				return -1
			}
			return r
		}, reason))
		if reason == "" {
			continue
		}
		lowerReason := strings.ToLower(reason)
		for _, invoicePrefix := range []string{"lnbcrt", "lnbc", "lntb"} {
			if invoiceIndex := strings.Index(lowerReason, invoicePrefix); invoiceIndex >= 0 {
				reason = strings.TrimSpace(reason[:invoiceIndex]) + " [redacted invoice]"
				break
			}
		}
		runes := []rune(reason)
		if len(runes) > 240 {
			reason = string(runes[:240]) + "…"
		}
		return reason
	}
	return ""
}

func findExactChannel(channels []lnclient.Channel, channelId string, nodePubkey string) (*lnclient.Channel, error) {
	var match *lnclient.Channel
	for i := range channels {
		if channels[i].Id != channelId || channels[i].RemotePubkey != nodePubkey {
			continue
		}
		if match != nil {
			return nil, errors.New("channel identity is ambiguous")
		}
		match = &channels[i]
	}
	if match == nil {
		return nil, errors.New("channel ID and counterparty pubkey do not match a current channel")
	}
	return match, nil
}

func validateNodePubkey(pubkey string) error {
	decoded, err := hex.DecodeString(pubkey)
	if err != nil || len(decoded) != 33 || (decoded[0] != 2 && decoded[0] != 3) {
		return errors.New("expected a full compressed secp256k1 public key")
	}
	return nil
}

func checkedAdd(a uint64, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, false
	}
	return a + b, true
}

func randomQuoteId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashQuoteMaterial(request *QuoteRebalanceRequest, orderId string, receivePaymentHash string, paymentHash string) (string, error) {
	encoded, err := json.Marshal(struct {
		Request            *QuoteRebalanceRequest `json:"request"`
		OrderId            string                 `json:"orderId"`
		ReceivePaymentHash string                 `json:"receivePaymentHash"`
		PaymentHash        string                 `json:"paymentHash"`
	}{request, orderId, receivePaymentHash, paymentHash})
	if err != nil {
		return "", fmt.Errorf("failed to hash rebalance quote: %w", err)
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}

func hashLocalRebalanceRequest(request *QuoteLocalRebalanceRequest) (string, error) {
	encoded, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to hash local rebalance request: %w", err)
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}

// hashLocalRebalanceRouteMaterial fingerprints every field that could affect a later fixed-route
// execution. Quote identity, expiry, and UI-only lock text are intentionally separate record data.
func hashLocalRebalanceRouteMaterial(quote *LocalRebalanceQuoteResponse) (string, error) {
	encoded, err := json.Marshal(struct {
		AmountMsat                     uint64                   `json:"amountMsat"`
		TotalRoutingFeeMsat            uint64                   `json:"totalRoutingFeeMsat"`
		MaxRoutingFeeMsat              uint64                   `json:"maxRoutingFeeMsat"`
		MaxTotalDebitMsat              uint64                   `json:"maxTotalDebitMsat"`
		OutgoingChannelId              string                   `json:"outgoingChannelId"`
		OutgoingNodePubkey             string                   `json:"outgoingNodePubkey"`
		OutgoingShortChannelId         string                   `json:"outgoingShortChannelId"`
		IncomingChannelId              string                   `json:"incomingChannelId"`
		IncomingNodePubkey             string                   `json:"incomingNodePubkey"`
		IncomingShortChannelId         string                   `json:"incomingShortChannelId"`
		OutgoingSpendableSnapshotMsat  uint64                   `json:"outgoingSpendableSnapshotMsat"`
		IncomingReceivableSnapshotMsat uint64                   `json:"incomingReceivableSnapshotMsat"`
		Paths                          []LocalCircularRoutePath `json:"paths"`
		RouteBytes                     []byte                   `json:"routeBytes"`
	}{
		AmountMsat:                     quote.AmountMsat,
		TotalRoutingFeeMsat:            quote.TotalRoutingFeeMsat,
		MaxRoutingFeeMsat:              quote.MaxRoutingFeeMsat,
		MaxTotalDebitMsat:              quote.MaxTotalDebitMsat,
		OutgoingChannelId:              quote.OutgoingChannelId,
		OutgoingNodePubkey:             quote.OutgoingNodePubkey,
		OutgoingShortChannelId:         quote.OutgoingShortChannelId,
		IncomingChannelId:              quote.IncomingChannelId,
		IncomingNodePubkey:             quote.IncomingNodePubkey,
		IncomingShortChannelId:         quote.IncomingShortChannelId,
		OutgoingSpendableSnapshotMsat:  quote.OutgoingSpendableSnapshotMsat,
		IncomingReceivableSnapshotMsat: quote.IncomingReceivableSnapshotMsat,
		Paths:                          quote.Paths,
		RouteBytes:                     quote.routeBytes,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fingerprint local rebalance route: %w", err)
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}
