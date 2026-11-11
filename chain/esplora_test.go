package chain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEsploraGetTransactionHex(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	known := strings.Repeat("a", 64)
	unknown := strings.Repeat("b", 64)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/tx/"+known+"/hex" {
			_, _ = w.Write([]byte("0200000001abcd\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Transaction not found"))
	}))
	defer server.Close()

	lookup := newEsploraLookup(server.URL)
	defer lookup.Close()

	txHex, err := lookup.GetTransactionHex(context.Background(), known)
	require.NoError(t, err)
	assert.Equal(t, "0200000001abcd", txHex)

	_, err = lookup.GetTransactionHex(context.Background(), unknown)
	assert.ErrorContains(t, err, "HTTP 404")

	_, err = lookup.GetTransactionHex(context.Background(), "invalid-id")
	require.ErrorContains(t, err, "invalid transaction ID")
	assert.EqualValues(t, 2, requests.Load())
}
