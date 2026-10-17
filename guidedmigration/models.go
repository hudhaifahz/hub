package guidedmigration

import (
	"fmt"
	"strings"
	"time"
)

const (
	EvidenceVersion   = 2
	ControlVersion    = 1
	QuarantineVersion = 1

	QuarantineFileName = ".grynvault-migration-quarantine.json"
)

type Phase string

const (
	PhaseStaged             Phase = "staged"
	PhaseTargetMoved        Phase = "target_moved"
	PhaseVerificationNeeded Phase = "verification_needed"
	PhaseVerified           Phase = "verified"
	PhaseActivating         Phase = "activating"
	PhaseActivationFailed   Phase = "activation_failed"
	PhaseAbandoning         Phase = "abandoning"
	PhaseAbandoned          Phase = "abandoned"
	PhaseActivated          Phase = "activated"
	PhaseBlocked            Phase = "blocked"
)

type Counts struct {
	Apps             int64 `json:"apps"`
	Permissions      int64 `json:"permissions"`
	Transactions     int64 `json:"transactions"`
	UserConfigs      int64 `json:"userConfigs"`
	Migrations       int64 `json:"migrations"`
	RebalanceRecords int64 `json:"rebalanceRecords"`
}

type SourceSummary struct {
	SourcePath            string `json:"sourcePath"`
	BackendType           string `json:"backendType"`
	WalletFingerprint     string `json:"walletFingerprint"`
	WalletFingerprintHint string `json:"walletFingerprintHint"`
	NWCIntegrity          string `json:"nwcIntegrity"`
	LDKIntegrity          string `json:"ldkIntegrity"`
	Counts                Counts `json:"counts"`
	DurableFileCount      int    `json:"durableFileCount"`
	LatestDurableChange   string `json:"latestDurableChange,omitempty"`
	DefaultLocation       bool   `json:"defaultLocation"`
	Ready                 bool   `json:"ready"`
	BlockingReason        string `json:"blockingReason,omitempty"`
	ConflictingPath       string `json:"conflictingPath,omitempty"`
}

type Evidence struct {
	Version            int               `json:"version"`
	ImportID           string            `json:"importId"`
	CreatedAt          string            `json:"createdAt"`
	SourcePath         string            `json:"sourcePath"`
	TargetPath         string            `json:"targetPath"`
	SourceFiles        map[string]string `json:"sourceFiles"`
	SourceCounts       Counts            `json:"sourceCounts"`
	SourceNWCIntegrity string            `json:"sourceNwcIntegrity"`
	SourceLDKIntegrity string            `json:"sourceLdkIntegrity"`
	WalletFingerprint  string            `json:"walletFingerprint"`
	ReleaseVersion     string            `json:"releaseVersion,omitempty"`
	SourceCommit       string            `json:"sourceCommit,omitempty"`
	LDKLibrarySHA256   string            `json:"ldkLibrarySha256,omitempty"`
	Checkpoints        []Checkpoint      `json:"checkpoints"`
}

type Checkpoint struct {
	Phase     Phase  `json:"phase"`
	CreatedAt string `json:"createdAt"`
	Detail    string `json:"detail,omitempty"`
}

type Control struct {
	Version           int               `json:"version"`
	ImportID          string            `json:"importId"`
	Phase             Phase             `json:"phase"`
	CreatedAt         string            `json:"createdAt"`
	UpdatedAt         string            `json:"updatedAt"`
	SourcePath        string            `json:"sourcePath"`
	TargetPath        string            `json:"targetPath"`
	StagePath         string            `json:"stagePath"`
	PreImportPath     string            `json:"preImportPath"`
	HadExistingTarget bool              `json:"hadExistingTarget"`
	BackupLocation    string            `json:"backupLocation"`
	EvidencePath      string            `json:"evidencePath"`
	SourceFiles       map[string]string `json:"sourceFiles"`
	WalletFingerprint string            `json:"walletFingerprint"`
	BlockedReason     string            `json:"blockedReason,omitempty"`
}

type ChannelSummary struct {
	ChannelID        string `json:"channelId"`
	UserChannelID    string `json:"userChannelId"`
	CounterpartyNode string `json:"counterpartyNodeId"`
	CapacitySats     uint64 `json:"capacitySats"`
	OutboundSats     uint64 `json:"outboundSats"`
	InboundSats      uint64 `json:"inboundSats"`
	Ready            bool   `json:"ready"`
	Usable           bool   `json:"usable"`
}

type VerificationSummary struct {
	VerifiedAt                 string           `json:"verifiedAt"`
	NodeID                     string           `json:"nodeId"`
	Network                    string           `json:"network"`
	Channels                   []ChannelSummary `json:"channels"`
	ChannelCount               int              `json:"channelCount"`
	StoredPaymentCount         int              `json:"storedPaymentCount"`
	TotalOnchainSats           uint64           `json:"totalOnchainSats"`
	SpendableOnchainSats       uint64           `json:"spendableOnchainSats"`
	TotalLightningSats         uint64           `json:"totalLightningSats"`
	PendingClosureSats         uint64           `json:"pendingClosureSats"`
	Apps                       int64            `json:"apps"`
	Permissions                int64            `json:"permissions"`
	Transactions               int64            `json:"transactions"`
	RebalanceRecords           int64            `json:"rebalanceRecords"`
	StaticChannelBackupFiles   int              `json:"staticChannelBackupFiles"`
	LatestStaticChannelBackup  string           `json:"latestStaticChannelBackup,omitempty"`
	LatestStaticBackupSHA256   string           `json:"latestStaticBackupSha256,omitempty"`
	ReleaseVersion             string           `json:"releaseVersion,omitempty"`
	SourceCommit               string           `json:"sourceCommit,omitempty"`
	LDKLibrarySHA256           string           `json:"ldkLibrarySha256,omitempty"`
	SourceBytesStillMatch      bool             `json:"sourceBytesStillMatch"`
	DestinationBytesStillMatch bool             `json:"destinationBytesStillMatch"`
}

type Quarantine struct {
	Version                int                  `json:"version"`
	ImportID               string               `json:"importId"`
	Phase                  Phase                `json:"phase"`
	CreatedAt              string               `json:"createdAt"`
	UpdatedAt              string               `json:"updatedAt"`
	SourcePath             string               `json:"sourcePath"`
	TargetPath             string               `json:"targetPath"`
	BackupLocation         string               `json:"backupLocation"`
	EvidencePath           string               `json:"evidencePath"`
	WalletFingerprint      string               `json:"walletFingerprint"`
	SourceFiles            map[string]string    `json:"sourceFiles"`
	SourceCounts           Counts               `json:"sourceCounts"`
	ActivatedBytesVerified bool                 `json:"activatedBytesVerified"`
	Verification           *VerificationSummary `json:"verification,omitempty"`
	LastError              string               `json:"lastError,omitempty"`
	ActivationAttemptedAt  string               `json:"activationAttemptedAt,omitempty"`
}

type StageResult struct {
	Cancelled       bool          `json:"cancelled"`
	RestartRequired bool          `json:"restartRequired"`
	BackupLocation  string        `json:"backupLocation,omitempty"`
	Summary         SourceSummary `json:"summary"`
}

type Progress struct {
	Step   string `json:"step"`
	Detail string `json:"detail"`
}

type Status struct {
	Active                 bool                 `json:"active"`
	Phase                  Phase                `json:"phase,omitempty"`
	SourcePath             string               `json:"sourcePath,omitempty"`
	TargetPath             string               `json:"targetPath,omitempty"`
	BackupLocation         string               `json:"backupLocation,omitempty"`
	FingerprintHint        string               `json:"fingerprintHint,omitempty"`
	ActivationConfirmation string               `json:"activationConfirmation,omitempty"`
	BlockingReason         string               `json:"blockingReason,omitempty"`
	Verification           *VerificationSummary `json:"verification,omitempty"`
	LightningOffline       bool                 `json:"lightningOffline"`
}

func timestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339Nano)
}

func ActivationConfirmation(network, walletFingerprint string) string {
	return fmt.Sprintf(
		"CONNECT %s %s",
		strings.ToUpper(strings.TrimSpace(network)),
		FingerprintHint(walletFingerprint),
	)
}
