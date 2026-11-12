package swaps

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/BoltzExchange/boltz-client/v2/pkg/boltz"
	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/stretchr/testify/assert"
)

// testConfig overrides only the config methods used by the swap provider
// functions (tests/mocks cannot be used here as it imports this package)
type testConfig struct {
	config.Config
	env     *config.AppConfig
	network string
}

func (c *testConfig) GetEnv() *config.AppConfig { return c.env }
func (c *testConfig) GetNetwork() string        { return c.network }

func TestFetchSwapProviderAPIURL(t *testing.T) {
	t.Run("mainnet maps the provider from Alby info", func(t *testing.T) {
		cfg := &testConfig{network: "bitcoin"}
		apiURL, err := fetchSwapProviderAPIURL(context.Background(), &testAlbyService{swapProvider: "satsrouting"}, cfg)
		assert.NoError(t, err)
		assert.Equal(t, "https://api.satsrouting.exchange", apiURL)
	})

	t.Run("off mainnet does not fetch Alby info", func(t *testing.T) {
		cfg := &testConfig{network: "regtest"}
		apiURL, err := fetchSwapProviderAPIURL(context.Background(), &testAlbyService{err: errors.New("should not be called")}, cfg)
		assert.NoError(t, err)
		assert.Equal(t, "", apiURL)
	})

	t.Run("Alby info unavailable", func(t *testing.T) {
		cfg := &testConfig{network: "bitcoin"}
		_, err := fetchSwapProviderAPIURL(context.Background(), &testAlbyService{err: errors.New("timeout")}, cfg)
		assert.Error(t, err)
	})
}

func TestMapSwapProviderToAPIURL(t *testing.T) {
	assert.Equal(t, "https://api.boltz.exchange", mapSwapProviderToAPIURL("boltz"))
	assert.Equal(t, coinosSwapAPIURL, mapSwapProviderToAPIURL("coinos"))
	assert.Equal(t, coinosSwapAPIURL, mapSwapProviderToAPIURL(" Coinos "))
	assert.Equal(t, "https://api.satsrouting.exchange", mapSwapProviderToAPIURL("satsrouting"))
	assert.Equal(t, "https://api.satsrouting.exchange", mapSwapProviderToAPIURL(" SatsRouting "))
	assert.Equal(t, "", mapSwapProviderToAPIURL(""))
	assert.Equal(t, "", mapSwapProviderToAPIURL("unknown"))
}

// TestCoinosBoltzV2Compatibility is deliberately opt-in because it calls a
// third-party mainnet API. It performs discovery only: no swap, invoice, HTLC,
// or payment is created.
func TestCoinosBoltzV2Compatibility(t *testing.T) {
	if os.Getenv("COINOS_READ_ONLY_INTEGRATION") != "1" {
		t.Skip("set COINOS_READ_ONLY_INTEGRATION=1 to query Coinos read-only pair data")
	}

	api := &boltz.Api{
		URL: coinosSwapAPIURL,
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
	reversePairs, err := api.GetReversePairs()
	assert.NoError(t, err)

	pairInfo, err := boltz.FindPair(
		boltz.Pair{From: boltz.CurrencyBtc, To: boltz.CurrencyBtc},
		reversePairs,
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, pairInfo.Hash)
	assert.Positive(t, pairInfo.Limits.Minimal)
	assert.GreaterOrEqual(t, pairInfo.Limits.Maximal, pairInfo.Limits.Minimal)
}

func TestCheckSwapApiUrl(t *testing.T) {
	svc := &swapsService{boltzApi: &boltz.Api{URL: "https://api.satsrouting.exchange"}}

	assert.NoError(t, svc.checkSwapApiUrl(&db.Swap{ApiUrl: "https://api.satsrouting.exchange"}))

	err := svc.checkSwapApiUrl(&db.Swap{ApiUrl: "https://api.boltz.exchange"})
	assert.EqualError(t, err, "swap was created with swap API https://api.boltz.exchange but the current swap API is https://api.satsrouting.exchange, set BOLTZ_API=https://api.boltz.exchange to process it")
}

// testAlbyService overrides only GetInfo, returning the given swap provider
type testAlbyService struct {
	alby.AlbyService
	swapProvider string
	err          error
}

func (s *testAlbyService) GetInfo(ctx context.Context) (*alby.AlbyInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &alby.AlbyInfo{Hub: alby.AlbyInfoHub{SwapProvider: s.swapProvider}}, nil
}

func TestCheckSwapProvider(t *testing.T) {
	tests := []struct {
		name         string
		boltzApi     string
		swapProvider string
		albyErr      error
		expectedErr  error
	}{
		{"provider unchanged", "", "boltz", nil, nil},
		{"provider changed", "", "satsrouting", nil, ErrSwapProviderChanged},
		{"provider empty", "", "", nil, ErrSwapProviderUnavailable},
		{"override skips check", "https://api.boltz.exchange", "satsrouting", errors.New("unreachable"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &swapsService{
				ctx: context.Background(),
				cfg: &testConfig{
					env:     &config.AppConfig{BoltzApi: tt.boltzApi},
					network: "bitcoin",
				},
				albySvc:  &testAlbyService{swapProvider: tt.swapProvider, err: tt.albyErr},
				boltzApi: &boltz.Api{URL: "https://api.boltz.exchange"},
			}
			assert.ErrorIs(t, svc.checkSwapProvider(), tt.expectedErr)
		})
	}

	t.Run("alby info unavailable", func(t *testing.T) {
		svc := &swapsService{
			ctx:      context.Background(),
			cfg:      &testConfig{env: &config.AppConfig{}, network: "bitcoin"},
			albySvc:  &testAlbyService{err: errors.New("timeout")},
			boltzApi: &boltz.Api{URL: "https://api.boltz.exchange"},
		}
		assert.Error(t, svc.checkSwapProvider())
	})
}
