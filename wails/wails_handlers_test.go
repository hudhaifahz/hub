package wails

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/logger"
)

type mockRouterAPI struct {
	api.API
	paidInvoices []string
}

func (m *mockRouterAPI) SendPayment(ctx context.Context, invoice string, amountMsat *uint64, metadata map[string]interface{}, fromAppId *uint) (*api.SendPaymentResponse, error) {
	m.paidInvoices = append(m.paidInvoices, invoice)
	return &api.SendPaymentResponse{}, nil
}

func (m *mockRouterAPI) LookupSwap(swapId string) (*api.LookupSwapResponse, error) {
	return &api.LookupSwapResponse{}, nil
}

func (m *mockRouterAPI) ListSwaps() (*api.ListSwapsResponse, error) {
	return &api.ListSwapsResponse{}, nil
}

func newTestWailsApp() (*WailsApp, *mockRouterAPI) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockAPI := &mockRouterAPI{}
	return &WailsApp{ctx: context.Background(), api: mockAPI}, mockAPI
}

func TestWailsRequestRouter_PaymentRouteNotMatchedInsideOtherRoute(t *testing.T) {
	for _, method := range []string{"GET", "POST"} {
		t.Run(method, func(t *testing.T) {
			app, mockAPI := newTestWailsApp()
			route := "/api/swaps//api/payments/lnbc1test"

			response := app.WailsRequestRouter(route, method, "")

			assert.Empty(t, mockAPI.paidInvoices)
			assert.Equal(t, "Unhandled route: "+method+" "+route, response.Error)
		})
	}
}

func TestWailsRequestRouter_PaymentRoute(t *testing.T) {
	app, mockAPI := newTestWailsApp()

	response := app.WailsRequestRouter("/api/payments/lnbc1test", "POST", "")

	assert.Equal(t, []string{"lnbc1test"}, mockAPI.paidInvoices)
	assert.Empty(t, response.Error)
}

func TestWailsRequestRouter_PaymentRouteRequiresPost(t *testing.T) {
	app, mockAPI := newTestWailsApp()
	route := "/api/payments/lnbc1test"

	response := app.WailsRequestRouter(route, "GET", "")

	assert.Empty(t, mockAPI.paidInvoices)
	assert.Equal(t, "Unhandled route: GET "+route, response.Error)
}

type rebalanceRouterAPI struct {
	api.API
	quoteRequest        *api.QuoteRebalanceRequest
	localQuoteRequest   *api.QuoteLocalRebalanceRequest
	executeRequest      *api.ExecuteRebalanceRequest
	localExecuteRequest *api.ExecuteLocalRebalanceRequest
	localStatusRequest  *api.ReconcileLocalRebalanceRequest
}

func (mock *rebalanceRouterAPI) QuoteRebalance(_ context.Context, request *api.QuoteRebalanceRequest) (*api.RebalanceQuoteResponse, error) {
	mock.quoteRequest = request
	return &api.RebalanceQuoteResponse{QuoteId: "test-quote"}, nil
}

func (mock *rebalanceRouterAPI) ExecuteRebalance(_ context.Context, request *api.ExecuteRebalanceRequest) (*api.RebalanceChannelResponse, error) {
	mock.executeRequest = request
	return nil, errors.New("execution locked")
}

func (mock *rebalanceRouterAPI) QuoteLocalRebalance(_ context.Context, request *api.QuoteLocalRebalanceRequest) (*api.LocalRebalanceQuoteResponse, error) {
	mock.localQuoteRequest = request
	return &api.LocalRebalanceQuoteResponse{OutgoingShortChannelId: "123", IncomingShortChannelId: "456"}, nil
}

func (mock *rebalanceRouterAPI) ExecuteLocalRebalance(_ context.Context, request *api.ExecuteLocalRebalanceRequest) (*api.LocalRebalanceOperationResponse, error) {
	mock.localExecuteRequest = request
	return &api.LocalRebalanceOperationResponse{State: "executing", Phase: "submitted"}, nil
}

func (mock *rebalanceRouterAPI) ReconcileLocalRebalance(_ context.Context, request *api.ReconcileLocalRebalanceRequest) (*api.LocalRebalanceOperationResponse, error) {
	mock.localStatusRequest = request
	return &api.LocalRebalanceOperationResponse{State: "succeeded", Phase: "succeeded"}, nil
}

func TestWailsRequestRouterDispatchesRebalanceQuote(t *testing.T) {
	mockAPI := &rebalanceRouterAPI{}
	app := &WailsApp{ctx: context.Background(), api: mockAPI}

	response := app.WailsRequestRouter(
		"/api/channels/rebalance/quote",
		"POST",
		`{"outgoingChannelId":"out-channel","outgoingNodePubkey":"out-peer","incomingChannelId":"in-channel","incomingNodePubkey":"in-peer","amountMsat":500000000,"maxProviderFeeMsat":2500000,"maxRoutingFeeMsat":1000000}`,
	)

	require.Empty(t, response.Error)
	require.Equal(t, "test-quote", response.Body.(*api.RebalanceQuoteResponse).QuoteId)
	require.NotNil(t, mockAPI.quoteRequest)
	require.Equal(t, "out-channel", mockAPI.quoteRequest.OutgoingChannelId)
	require.Equal(t, "in-channel", mockAPI.quoteRequest.IncomingChannelId)
	require.Equal(t, uint64(500000000), mockAPI.quoteRequest.AmountMsat)
}

func TestWailsRequestRouterPropagatesLockedRebalanceExecution(t *testing.T) {
	mockAPI := &rebalanceRouterAPI{}
	app := &WailsApp{ctx: context.Background(), api: mockAPI}

	response := app.WailsRequestRouter(
		"/api/channels/rebalance/execute",
		"POST",
		`{"quoteId":"test-quote"}`,
	)

	require.EqualError(t, errors.New(response.Error), "execution locked")
	require.NotNil(t, mockAPI.executeRequest)
	require.Equal(t, "test-quote", mockAPI.executeRequest.QuoteId)
}

func TestWailsRequestRouterDispatchesLocalRebalanceQuote(t *testing.T) {
	mockAPI := &rebalanceRouterAPI{}
	app := &WailsApp{ctx: context.Background(), api: mockAPI}

	response := app.WailsRequestRouter(
		"/api/channels/rebalance/local-quote",
		"POST",
		`{"outgoingChannelId":"out-channel","outgoingNodePubkey":"out-peer","incomingChannelId":"in-channel","incomingNodePubkey":"in-peer","amountMsat":20000000,"maxRoutingFeeMsat":1000000}`,
	)

	require.Empty(t, response.Error)
	require.Equal(t, "123", response.Body.(*api.LocalRebalanceQuoteResponse).OutgoingShortChannelId)
	require.NotNil(t, mockAPI.localQuoteRequest)
	require.Equal(t, uint64(20000000), mockAPI.localQuoteRequest.AmountMsat)
}

func TestWailsRequestRouterDispatchesLocalRebalanceExecutionAndStatus(t *testing.T) {
	mockAPI := &rebalanceRouterAPI{}
	app := &WailsApp{ctx: context.Background(), api: mockAPI}

	executeResponse := app.WailsRequestRouter(
		"/api/channels/rebalance/local-execute",
		"POST",
		`{"quoteId":"test-quote","routeFingerprint":"fingerprint","confirmation":"REBALANCE 20000 SATS st-quote"}`,
	)
	require.Empty(t, executeResponse.Error)
	require.Equal(t, "submitted", executeResponse.Body.(*api.LocalRebalanceOperationResponse).Phase)
	require.NotNil(t, mockAPI.localExecuteRequest)
	require.Equal(t, "fingerprint", mockAPI.localExecuteRequest.RouteFingerprint)
	require.Equal(t, "REBALANCE 20000 SATS st-quote", mockAPI.localExecuteRequest.Confirmation)

	statusResponse := app.WailsRequestRouter(
		"/api/channels/rebalance/local-status",
		"POST",
		`{"quoteId":"test-quote","routeFingerprint":"fingerprint"}`,
	)
	require.Empty(t, statusResponse.Error)
	require.Equal(t, "succeeded", statusResponse.Body.(*api.LocalRebalanceOperationResponse).State)
	require.NotNil(t, mockAPI.localStatusRequest)
	require.Equal(t, "test-quote", mockAPI.localStatusRequest.QuoteId)
}
