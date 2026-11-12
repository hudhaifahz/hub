package guidedmigration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/getAlby/hub/version"
	"github.com/getAlby/ldk-node-go/ldk_node"
)

func InspectLDKOffline(workdir, mnemonic, network string) (*VerificationSummary, error) {
	if mnemonic == "" {
		return nil, errors.New("wallet mnemonic is empty")
	}
	temp, err := os.MkdirTemp("", "grynvault-offline-ldk-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temp)
	if _, err = CopyTree(workdir, temp, []string{filepath.Join("ldk", "storage")}); err != nil {
		return nil, fmt.Errorf("copy LDK state for offline inspection: %w", err)
	}

	config := ldk_node.DefaultConfig()
	config.StorageDirPath = filepath.Join(temp, "ldk", "storage")
	builder := ldk_node.BuilderFromConfig(config)
	defer builder.Destroy()
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)
	switch network {
	case "regtest":
		builder.SetNetwork(ldk_node.NetworkRegtest)
	case "signet", "testnet":
		builder.SetNetwork(ldk_node.NetworkSignet)
	default:
		builder.SetNetwork(ldk_node.NetworkBitcoin)
	}
	builder.SetStorageDirPath(filepath.Join(temp, "ldk", "storage"))

	node, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("load copied LDK state offline: %w", err)
	}
	defer node.Destroy()

	channels := node.ListChannels()
	channelSummaries := make([]ChannelSummary, 0, len(channels))
	for _, channel := range channels {
		channelSummaries = append(channelSummaries, ChannelSummary{
			ChannelID:        channel.ChannelId,
			UserChannelID:    channel.UserChannelId,
			CounterpartyNode: channel.CounterpartyNodeId,
			CapacitySats:     channel.ChannelValueSats,
			OutboundSats:     channel.OutboundCapacityMsat / 1000,
			InboundSats:      channel.InboundCapacityMsat / 1000,
			Ready:            channel.IsChannelReady,
			Usable:           channel.IsUsable,
		})
	}
	sort.Slice(channelSummaries, func(left, right int) bool {
		return channelSummaries[left].UserChannelID < channelSummaries[right].UserChannelID
	})

	balances := node.ListBalances()
	var pendingClosureSats uint64
	for _, pending := range balances.PendingBalancesFromChannelClosures {
		switch value := pending.(type) {
		case ldk_node.PendingSweepBalancePendingBroadcast:
			pendingClosureSats += value.AmountSatoshis
		case ldk_node.PendingSweepBalanceBroadcastAwaitingConfirmation:
			pendingClosureSats += value.AmountSatoshis
		case ldk_node.PendingSweepBalanceAwaitingThresholdConfirmations:
			pendingClosureSats += value.AmountSatoshis
		}
	}
	backupCount, latestBackup, latestBackupHash, err := staticChannelBackupSummary(workdir)
	if err != nil {
		return nil, err
	}

	return &VerificationSummary{
		VerifiedAt:                time.Now().UTC().Format(time.RFC3339Nano),
		NodeID:                    node.NodeId(),
		Network:                   network,
		Channels:                  channelSummaries,
		ChannelCount:              len(channelSummaries),
		StoredPaymentCount:        len(node.ListPayments()),
		TotalOnchainSats:          balances.TotalOnchainBalanceSats,
		SpendableOnchainSats:      balances.SpendableOnchainBalanceSats,
		TotalLightningSats:        balances.TotalLightningBalanceSats,
		PendingClosureSats:        pendingClosureSats,
		StaticChannelBackupFiles:  backupCount,
		LatestStaticChannelBackup: latestBackup,
		LatestStaticBackupSHA256:  latestBackupHash,
		ReleaseVersion:            version.Tag,
		SourceCommit:              version.SourceCommit,
		LDKLibrarySHA256:          version.LDKLibrarySHA256,
	}, nil
}

func staticChannelBackupSummary(workdir string) (int, string, string, error) {
	backupDirectory := filepath.Join(workdir, "ldk", "static_channel_backups")
	entries, err := os.ReadDir(backupDirectory)
	if errors.Is(err, os.ErrNotExist) {
		return 0, "", "", nil
	}
	if err != nil {
		return 0, "", "", fmt.Errorf("inspect static channel backups: %w", err)
	}
	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil {
			return 0, "", "", fmt.Errorf("inspect static channel backup: %w", infoErr)
		}
		if !info.Mode().IsRegular() {
			return 0, "", "", fmt.Errorf("refusing non-regular static channel backup %q", entry.Name())
		}
		fileNames = append(fileNames, entry.Name())
	}
	if len(fileNames) == 0 {
		return 0, "", "", nil
	}
	sort.Strings(fileNames)
	latest := fileNames[len(fileNames)-1]
	latestHash, err := hashFile(filepath.Join(backupDirectory, latest))
	if err != nil {
		return 0, "", "", fmt.Errorf("hash latest static channel backup: %w", err)
	}
	return len(fileNames), latest, latestHash, nil
}
