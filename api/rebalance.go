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

var errPinnedRebalanceExecutionDisabled = errors.New("pinned rebalance execution is disabled until incoming-channel atomicity is proven")
var errLocalRebalanceExecutionDisabled = errors.New("local circular-route execution is not implemented or authorized")

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
// first- and last-hop channels. It is deliberately side-effect free: no provider request, invoice,
// database row, probe, HTLC, or payment is created.
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

	return &LocalRebalanceQuoteResponse{
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
		ExecutionEnabled:               false,
		BlockedReason:                  errLocalRebalanceExecutionDisabled.Error(),
	}, nil
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
