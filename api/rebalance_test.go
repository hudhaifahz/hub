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
