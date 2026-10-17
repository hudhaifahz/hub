//go:build regtest

package guidedmigration_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/getAlby/hub/config"
	hubdb "github.com/getAlby/hub/db"
	"github.com/getAlby/hub/guidedmigration"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/stretchr/testify/require"
)

const (
	syntheticMnemonicA = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	syntheticMnemonicB = "legal winner thank year wave sausage worth useful legal winner thank yellow"
	syntheticMnemonicC = "letter advice cage absurd amount doctor acoustic avoid letter advice cage above"
	syntheticPassword  = "synthetic-local-password"
)

type regtestBitcoin struct {
	t           *testing.T
	ctx         context.Context
	cancel      context.CancelFunc
	command     *exec.Cmd
	cliPath     string
	dataDir     string
	rpcPort     int
	rpcUser     string
	rpcPass     string
	logOutput   bytes.Buffer
	mineAddress string
}

type syntheticNode struct {
	t       *testing.T
	node    *ldk_node.Node
	address string
	cancel  context.CancelFunc
	done    chan struct{}
	stopped bool
}

func startRegtestBitcoin(t *testing.T) *regtestBitcoin {
	t.Helper()
	bitcoindPath, err := exec.LookPath("bitcoind")
	require.NoError(t, err, "bitcoind is required for the tagged regtest acceptance")
	cliPath, err := exec.LookPath("bitcoin-cli")
	require.NoError(t, err, "bitcoin-cli is required for the tagged regtest acceptance")
	rpcPort := freeTCPPort(t)
	p2pPort := freeTCPPort(t)
	ctx, cancel := context.WithCancel(context.Background())
	r := &regtestBitcoin{
		t: t, ctx: ctx, cancel: cancel, cliPath: cliPath, dataDir: t.TempDir(),
		rpcPort: rpcPort, rpcUser: "grynvault-regtest", rpcPass: "synthetic-only-password",
	}
	r.command = exec.CommandContext(ctx, bitcoindPath,
		"-regtest", "-server=1", "-listen=1", "-dnsseed=0", "-fixedseeds=0",
		"-datadir="+r.dataDir,
		"-rpcbind=127.0.0.1", "-rpcallowip=127.0.0.1",
		"-rpcport="+strconv.Itoa(rpcPort), "-port="+strconv.Itoa(p2pPort),
		"-rpcuser="+r.rpcUser, "-rpcpassword="+r.rpcPass,
		"-fallbackfee=0.0002", "-txindex=1", "-blockfilterindex=1",
	)
	r.command.Stdout = &r.logOutput
	r.command.Stderr = &r.logOutput
	require.NoError(t, r.command.Start())
	t.Cleanup(func() { r.stop() })
	require.Eventually(t, func() bool {
		_, commandErr := r.cli("getblockchaininfo")
		return commandErr == nil
	}, 30*time.Second, 100*time.Millisecond, "bitcoind did not become ready: %s", r.logOutput.String())
	_, err = r.cli("createwallet", "grynvault-acceptance")
	require.NoError(t, err)
	r.mineAddress = strings.TrimSpace(r.mustCLI("getnewaddress", "", "bech32"))
	r.mine(101)
	return r
}

func (r *regtestBitcoin) cli(arguments ...string) (string, error) {
	base := []string{
		"-regtest", "-datadir=" + r.dataDir, "-rpcport=" + strconv.Itoa(r.rpcPort),
		"-rpcuser=" + r.rpcUser, "-rpcpassword=" + r.rpcPass,
	}
	command := exec.CommandContext(r.ctx, r.cliPath, append(base, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bitcoin-cli %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func (r *regtestBitcoin) mustCLI(arguments ...string) string {
	r.t.Helper()
	output, err := r.cli(arguments...)
	require.NoError(r.t, err)
	return output
}

func (r *regtestBitcoin) mine(blocks int) {
	r.t.Helper()
	r.mustCLI("generatetoaddress", strconv.Itoa(blocks), r.mineAddress)
}

func (r *regtestBitcoin) stop() {
	if r.command == nil || r.command.Process == nil {
		return
	}
	_, _ = r.cli("stop")
	done := make(chan struct{})
	go func() {
		_ = r.command.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		r.cancel()
		<-done
	}
	r.command = nil
}

func TestSyntheticRegtestGuidedMigrationAcceptance(t *testing.T) {
	if os.Getenv("GRYNVAULT_RUN_REGTEST_ACCEPTANCE") != "1" {
		t.Skip("set GRYNVAULT_RUN_REGTEST_ACCEPTANCE=1 to run the disposable regtest acceptance")
	}
	logger.Init("4")
	bitcoin := startRegtestBitcoin(t)
	baseSource := filepath.Join(t.TempDir(), "synthetic-authoritative-albyhub")
	nodeA := startSyntheticNode(t, bitcoin, syntheticMnemonicA, baseSource)
	nodeB := startSyntheticNode(t, bitcoin, syntheticMnemonicB, filepath.Join(t.TempDir(), "peer-b"))
	nodeC := startSyntheticNode(t, bitcoin, syntheticMnemonicC, filepath.Join(t.TempDir(), "peer-c"))
	defer nodeB.stop()
	defer nodeC.stop()

	address, err := nodeA.node.OnchainPayment().NewAddress()
	require.NoError(t, err)
	bitcoin.mustCLI("sendtoaddress", address, "0.03")
	bitcoin.mine(6)
	require.Eventually(t, func() bool {
		if syncErr := nodeA.node.SyncWallets(); syncErr != nil {
			t.Logf("wallet sync: %v", syncErr)
			return false
		}
		return nodeA.node.ListBalances().SpendableOnchainBalanceSats >= 2_900_000
	}, 30*time.Second, time.Second)

	openSyntheticChannel(t, bitcoin, nodeA, nodeB, 500_000, 100_000_000)
	openSyntheticChannel(t, bitcoin, nodeA, nodeC, 500_000, 100_000_000)
	require.Len(t, nodeA.node.ListChannels(), 2)

	invoiceB, err := nodeB.node.Bolt11Payment().Receive(
		10_000_000,
		ldk_node.Bolt11InvoiceDescriptionDirect{Description: "synthetic outbound history"},
		3600,
	)
	require.NoError(t, err)
	paymentToB, err := nodeA.node.Bolt11Payment().Send(invoiceB, nil)
	require.NoError(t, err)
	waitForSyntheticPayment(t, nodeA.node, paymentToB)

	invoiceA, err := nodeA.node.Bolt11Payment().Receive(
		7_000_000,
		ldk_node.Bolt11InvoiceDescriptionDirect{Description: "synthetic inbound history"},
		3600,
	)
	require.NoError(t, err)
	paymentToA, err := nodeC.node.Bolt11Payment().Send(invoiceA, nil)
	require.NoError(t, err)
	waitForSyntheticPayment(t, nodeC.node, paymentToA)
	require.Eventually(t, func() bool {
		payments := nodeA.node.ListPayments()
		if len(payments) < 2 {
			return false
		}
		for _, payment := range payments {
			if payment.Status != ldk_node.PaymentStatusSucceeded {
				return false
			}
		}
		return true
	}, 30*time.Second, 100*time.Millisecond)

	expectedNodeID := nodeA.node.NodeId()
	nodeAAddress := nodeA.address
	nodeA.stop()
	createSyntheticHubDatabase(t, baseSource, syntheticMnemonicA, syntheticPassword)

	baseSummary, err := guidedmigration.InspectLDKOffline(baseSource, syntheticMnemonicA, "regtest")
	require.NoError(t, err)
	require.Equal(t, expectedNodeID, baseSummary.NodeID)
	require.Equal(t, 2, baseSummary.ChannelCount)
	require.GreaterOrEqual(t, baseSummary.StoredPaymentCount, 2)

	t.Run("standard macOS location", func(t *testing.T) {
		home := t.TempDir()
		container := filepath.Join(home, "Library", "Containers", "com.getalby.Alby-Hub")
		source := filepath.Join(container, "Data", "Library", "Application Support", "albyhub")
		copySyntheticWallet(t, baseSource, source)
		target := filepath.Join(home, "Library", "Containers", "com.grynvault.rebalance", "Data", "Library", "Application Support", "grynvault-rebalance")
		runMigrationAcceptance(t, bitcoin, nodeB, nodeC, container, source, target, home, nodeAAddress, expectedNodeID)
	})

	t.Run("custom work directory with spaces and Unicode", func(t *testing.T) {
		home := t.TempDir()
		source := filepath.Join(home, "External Wallets", "Alby 雷", "custom work dir")
		copySyntheticWallet(t, baseSource, source)
		target := filepath.Join(home, "Grynvault Data", "wallet")
		runMigrationAcceptance(t, bitcoin, nodeB, nodeC, source, source, target, home, nodeAAddress, expectedNodeID)
	})

	t.Run("stale matching clone blocks before process or network checks", func(t *testing.T) {
		home := t.TempDir()
		source := filepath.Join(home, "Custom", "current-wallet")
		stale := filepath.Join(home, "Library", "Containers", "com.getalby.Alby-Hub", "Data", "Library", "Application Support", "albyhub")
		copySyntheticWallet(t, baseSource, source)
		copySyntheticWallet(t, baseSource, stale)
		require.NoError(t, os.MkdirAll(filepath.Join(source, "ldk", "static_channel_backups"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(source, "ldk", "static_channel_backups", "newer.synthetic"), []byte("newer synthetic state"), 0o600))
		processChecks := 0
		_, stageErr := guidedmigration.Stage(source, filepath.Join(home, "target"), guidedmigration.StageOptions{
			Now:           time.Now,
			StabilityWait: 10 * time.Millisecond,
			OpenFileChecker: func(string) ([]int, error) {
				processChecks++
				return nil, nil
			},
			HomeDirectory: home,
		})
		require.ErrorContains(t, stageErr, "matching wallet copy found")
		require.Zero(t, processChecks, "duplicate authority must block before any activation-oriented checks")
		require.NoFileExists(t, guidedmigration.ControlPath(filepath.Join(home, "target")))
	})
}

func startSyntheticNode(t *testing.T, bitcoin *regtestBitcoin, mnemonic, workdir string) *syntheticNode {
	t.Helper()
	storage := filepath.Join(workdir, "ldk", "storage")
	require.NoError(t, os.MkdirAll(storage, 0o700))
	address := fmt.Sprintf("127.0.0.1:%d", freeTCPPort(t))
	builder := ldk_node.BuilderFromConfig(ldk_node.DefaultConfig())
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)
	builder.SetNetwork(ldk_node.NetworkRegtest)
	builder.SetChainSourceBitcoindRpc("127.0.0.1", uint16(bitcoin.rpcPort), bitcoin.rpcUser, bitcoin.rpcPass)
	builder.SetStorageDirPath(storage)
	require.NoError(t, builder.SetListeningAddresses([]ldk_node.SocketAddress{address}))
	node, err := builder.Build()
	builder.Destroy()
	require.NoError(t, err)
	require.NoError(t, node.Start())
	ctx, cancel := context.WithCancel(context.Background())
	synthetic := &syntheticNode{t: t, node: node, address: address, cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(synthetic.done)
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for node.NextEvent() != nil {
					_ = node.EventHandled()
				}
			}
		}
	}()
	t.Cleanup(synthetic.stop)
	return synthetic
}

func (n *syntheticNode) stop() {
	if n == nil || n.stopped {
		return
	}
	n.stopped = true
	n.cancel()
	<-n.done
	require.NoError(n.t, n.node.Stop())
	n.node.Destroy()
	return
}

func openSyntheticChannel(t *testing.T, bitcoin *regtestBitcoin, funder, peer *syntheticNode, amountSats, pushMsat uint64) {
	t.Helper()
	_, err := funder.node.OpenChannel(peer.node.NodeId(), peer.address, amountSats, &pushMsat, nil)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return strings.TrimSpace(bitcoin.mustCLI("getrawmempool")) != "[]"
	}, 10*time.Second, 100*time.Millisecond, "channel funding transaction was not broadcast")
	bitcoin.mine(12)
	require.Eventually(t, func() bool {
		_ = funder.node.SyncWallets()
		_ = peer.node.SyncWallets()
		for _, channel := range funder.node.ListChannels() {
			if channel.CounterpartyNodeId == peer.node.NodeId() && channel.IsChannelReady && channel.IsUsable {
				return true
			}
		}
		return false
	}, 90*time.Second, 250*time.Millisecond, "channel did not become ready and usable; funder=%+v peer=%+v", funder.node.ListChannels(), peer.node.ListChannels())
}

func waitForSyntheticPayment(t *testing.T, node *ldk_node.Node, paymentID ldk_node.PaymentId) {
	t.Helper()
	require.Eventually(t, func() bool {
		payment := node.Payment(paymentID)
		return payment != nil && payment.Status == ldk_node.PaymentStatusSucceeded
	}, 30*time.Second, 100*time.Millisecond)
}

func createSyntheticHubDatabase(t *testing.T, workdir, mnemonic, password string) {
	t.Helper()
	database, err := hubdb.NewDB(filepath.Join(workdir, "nwc.db"), false)
	require.NoError(t, err)
	cfg, err := config.NewConfig(&config.AppConfig{Workdir: workdir, Network: "regtest", LDKNetwork: "regtest"}, database)
	require.NoError(t, err)
	require.NoError(t, cfg.SaveUnlockPasswordCheck(password))
	require.NoError(t, cfg.SetUpdate("Mnemonic", mnemonic, password))
	require.NoError(t, cfg.SetUpdate("LNBackendType", config.LDKBackendType, ""))
	require.NoError(t, cfg.SetUpdate("NodeLastStartTime", time.Now().UTC().Format(time.RFC3339), ""))
	require.NoError(t, cfg.SetUpdate("NodeAlias", "Grynvault synthetic migration", ""))
	require.NoError(t, hubdb.Stop(database))
}

func copySyntheticWallet(t *testing.T, source, destination string) {
	t.Helper()
	_, err := guidedmigration.CopyWalletData(source, destination)
	require.NoError(t, err)
}

func runMigrationAcceptance(
	t *testing.T,
	bitcoin *regtestBitcoin,
	peerB, peerC *syntheticNode,
	selection, exactSource, target, home, nodeAddress, expectedNodeID string,
) {
	t.Helper()
	stage, err := guidedmigration.Stage(selection, target, guidedmigration.StageOptions{
		Now:             time.Now,
		StabilityWait:   100 * time.Millisecond,
		OpenFileChecker: guidedmigration.OpenWalletProcessIDs,
		HomeDirectory:   home,
	})
	require.NoError(t, err)
	require.True(t, stage.RestartRequired)
	require.Equal(t, evaluatedPath(t, exactSource), stage.Summary.SourcePath)
	require.NoFileExists(t, filepath.Join(target, "nwc.db"))

	setSyntheticServiceEnvironment(t, bitcoin, target, nodeAddress)
	ctx, cancel := context.WithCancel(context.Background())
	svc, err := service.NewService(ctx)
	require.NoError(t, err)
	shutdown := func() {
		cancel()
		svc.Shutdown()
	}
	defer func() {
		if svc != nil {
			shutdown()
		}
	}()

	require.True(t, svc.IsGuidedMigrationQuarantined())
	require.Nil(t, svc.GetLNClient())
	require.False(t, anyPeerConnectedTo(peerB.node, expectedNodeID))
	require.False(t, anyPeerConnectedTo(peerC.node, expectedNodeID))
	listener, err := net.Listen("tcp", nodeAddress)
	require.NoError(t, err, "the migrated node listener must remain unused in quarantine")
	require.NoError(t, listener.Close())

	verification, err := svc.VerifyGuidedMigration(syntheticPassword)
	require.NoError(t, err)
	require.Equal(t, expectedNodeID, verification.NodeID)
	require.Equal(t, 2, verification.ChannelCount)
	require.GreaterOrEqual(t, verification.StoredPaymentCount, 2)
	require.True(t, verification.SourceBytesStillMatch)
	require.True(t, verification.DestinationBytesStillMatch)
	require.NoError(t, svc.ActivateGuidedMigration(
		syntheticPassword,
		verification.NodeID,
		guidedmigration.ActivationConfirmation(verification.Network, stage.Summary.WalletFingerprint),
	))
	require.False(t, svc.IsGuidedMigrationQuarantined())
	require.NotNil(t, svc.GetLNClient())
	require.NoFileExists(t, guidedmigration.QuarantinePath(target))
	require.NoFileExists(t, guidedmigration.ControlPath(target))
	require.FileExists(t, filepath.Join(stage.BackupLocation, "migration-acceptance.json"))
	require.Eventually(t, func() bool {
		return anyPeerConnectedTo(peerB.node, expectedNodeID) || anyPeerConnectedTo(peerC.node, expectedNodeID)
	}, 30*time.Second, 250*time.Millisecond, "the activated synthetic node did not reconnect to either channel peer")
	secondListener, listenErr := net.Listen("tcp", nodeAddress)
	require.Error(t, listenErr, "the active destination must own the node listener")
	if secondListener != nil {
		_ = secondListener.Close()
	}

	shutdown()
	svc = nil
	require.Eventually(t, func() bool {
		probe, probeErr := net.Listen("tcp", nodeAddress)
		if probeErr != nil {
			return false
		}
		_ = probe.Close()
		return true
	}, 10*time.Second, 100*time.Millisecond, "the node listener was not released after shutdown")
}

func setSyntheticServiceEnvironment(t *testing.T, bitcoin *regtestBitcoin, workdir, nodeAddress string) {
	t.Helper()
	values := map[string]string{
		"WORK_DIR":                  workdir,
		"DATABASE_URI":              "nwc.db",
		"LOG_TO_FILE":               "false",
		"NETWORK":                   "regtest",
		"LDK_NETWORK":               "regtest",
		"LDK_BITCOIND_RPC_HOST":     "127.0.0.1",
		"LDK_BITCOIND_RPC_PORT":     strconv.Itoa(bitcoin.rpcPort),
		"LDK_BITCOIND_RPC_USER":     bitcoin.rpcUser,
		"LDK_BITCOIND_RPC_PASSWORD": bitcoin.rpcPass,
		"LDK_ELECTRUM_SERVER":       "tcp://127.0.0.1:1",
		"LDK_LISTENING_ADDRESSES":   nodeAddress,
		"LDK_GOSSIP_SOURCE":         "",
		"LDK_LSPS2_ADDRESSES":       "",
		"LDK_VSS_URL":               "",
		"RELAY":                     "ws://127.0.0.1:1",
		"MEMPOOL_API":               "http://127.0.0.1:1",
		"BOLTZ_API":                 "http://127.0.0.1:1",
		"SEND_EVENTS_TO_ALBY":       "false",
		"AUTO_LINK_ALBY_ACCOUNT":    "false",
		"HTTP_PROXY":                "http://127.0.0.1:1",
		"HTTPS_PROXY":               "http://127.0.0.1:1",
		"NO_PROXY":                  "127.0.0.1,localhost",
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}

func anyPeerConnectedTo(node *ldk_node.Node, nodeID string) bool {
	for _, peer := range node.ListPeers() {
		if peer.NodeId == nodeID && peer.IsConnected {
			return true
		}
	}
	return false
}

func evaluatedPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)
	return resolved
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
