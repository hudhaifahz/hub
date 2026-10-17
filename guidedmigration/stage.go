package guidedmigration

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/getAlby/hub/version"
)

type StageOptions struct {
	Now              func() time.Time
	StabilityWait    time.Duration
	OpenFileChecker  func(string) ([]int, error)
	LocalPathChecker func(string) (bool, error)
	HomeDirectory    string
	Progress         func(Progress)
	// AfterFirstHash is a deterministic test seam for simulating a source
	// mutation between the two quiescence snapshots. Production callers leave
	// it nil.
	AfterFirstHash func()
}

func DefaultStageOptions() StageOptions {
	home, _ := os.UserHomeDir()
	return StageOptions{
		Now:              time.Now,
		StabilityWait:    time.Second,
		OpenFileChecker:  OpenWalletProcessIDs,
		LocalPathChecker: IsLocalPath,
		HomeDirectory:    home,
	}
}

func reportProgress(options StageOptions, step, detail string) {
	if options.Progress != nil {
		options.Progress(Progress{Step: step, Detail: detail})
	}
}

func Stage(sourceSelection, targetWorkdir string, options StageOptions) (*StageResult, error) {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.OpenFileChecker == nil {
		options.OpenFileChecker = OpenWalletProcessIDs
	}
	if options.LocalPathChecker == nil {
		options.LocalPathChecker = IsLocalPath
	}
	sourceWorkdir, err := ResolveWorkdir(sourceSelection)
	if err != nil {
		return nil, err
	}
	targetWorkdir, err = filepath.Abs(targetWorkdir)
	if err != nil {
		return nil, fmt.Errorf("resolve Grynvault target folder: %w", err)
	}
	if PathsOverlap(sourceWorkdir, targetWorkdir) {
		return nil, errors.New("source and target wallet folders must be separate")
	}
	for label, path := range map[string]string{"source": sourceWorkdir, "target": targetWorkdir} {
		local, localErr := options.LocalPathChecker(path)
		if localErr != nil {
			return nil, fmt.Errorf("verify %s wallet folder is on a local filesystem: %w", label, localErr)
		}
		if !local {
			return nil, fmt.Errorf("%s wallet folder must be on a local filesystem; network-mounted folders are not supported", label)
		}
	}
	if IsBackupLikePath(sourceWorkdir) {
		return nil, errors.New("the selected folder looks like a backup or staged restore; use the disaster-recovery flow instead")
	}
	if _, err = os.Stat(ControlPath(targetWorkdir)); err == nil {
		return nil, errors.New("a guided import is already pending; reopen Grynvault Rebalance to continue it")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect pending import control: %w", err)
	}

	summary, err := InspectSource(sourceWorkdir)
	if err != nil {
		return nil, err
	}
	reportProgress(options, "source_validated", "Wallet data and embedded databases validated")
	if existingTarget, targetErr := InspectSource(targetWorkdir); targetErr == nil {
		if existingTarget.WalletFingerprint == summary.WalletFingerprint {
			return nil, errors.New("the same wallet is already configured in Grynvault Rebalance; no data was changed")
		}
		return nil, errors.New("a different configured Grynvault wallet already exists; guided import will not replace it")
	} else {
		pristine, pristineErr := IsPristineTarget(targetWorkdir)
		if pristineErr != nil {
			return nil, fmt.Errorf("existing Grynvault target could not be proven pristine: %w", pristineErr)
		}
		if !pristine {
			return nil, errors.New("existing Grynvault target contains wallet state but is not a complete valid wallet; guided import will not replace it")
		}
	}
	if _, restoreErr := os.Stat(filepath.Join(targetWorkdir, "restore")); restoreErr == nil {
		return nil, errors.New("Grynvault has a pending restore; finish or remove it before starting a guided import")
	} else if !errors.Is(restoreErr, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect pending Grynvault restore: %w", restoreErr)
	}
	conflictingPath, err := FindConflictingCopy(summary, KnownCandidatePaths(options.HomeDirectory, targetWorkdir, sourceWorkdir))
	if err != nil {
		return nil, err
	}
	if conflictingPath != "" {
		summary.Ready = false
		summary.ConflictingPath = conflictingPath
		summary.BlockingReason = "another known data folder contains the same wallet; Grynvault cannot prove which copy is current"
		return nil, fmt.Errorf("matching wallet copy found at %s; no data was changed", conflictingPath)
	}

	processIDs, err := options.OpenFileChecker(sourceWorkdir)
	if err != nil {
		return nil, fmt.Errorf("verify Alby Hub is stopped: %w", err)
	}
	if len(processIDs) > 0 {
		return nil, fmt.Errorf("Alby Hub data is still open by process %s; quit it completely before importing", joinProcessIDs(processIDs))
	}

	firstHashes, err := HashWalletData(sourceWorkdir)
	if err != nil {
		return nil, fmt.Errorf("hash source wallet data: %w", err)
	}
	if options.AfterFirstHash != nil {
		options.AfterFirstHash()
	}
	if options.StabilityWait > 0 {
		timer := time.NewTimer(options.StabilityWait)
		<-timer.C
	}
	secondHashes, err := HashWalletData(sourceWorkdir)
	if err != nil {
		return nil, fmt.Errorf("re-hash source wallet data: %w", err)
	}
	if err = CompareHashes(firstHashes, secondHashes); err != nil {
		return nil, fmt.Errorf("Alby Hub data changed during the stability check; quit the source completely and try again: %w", err)
	}
	reportProgress(options, "source_quiescent", "Alby Hub is stopped and wallet files are stable")

	importID, err := NewImportID()
	if err != nil {
		return nil, fmt.Errorf("create import identifier: %w", err)
	}
	now := options.Now()
	backupParent := filepath.Join(filepath.Dir(targetWorkdir), "grynvault-rebalance-migration-backups")
	if err = os.MkdirAll(backupParent, 0700); err != nil {
		return nil, fmt.Errorf("create migration backup folder: %w", err)
	}
	backupDir := filepath.Join(backupParent, now.UTC().Format("20060102T150405.000000000Z")+"-"+importID[:8])
	if err = os.Mkdir(backupDir, 0700); err != nil {
		return nil, fmt.Errorf("create migration backup: %w", err)
	}
	keepBackup := false
	defer func() {
		if !keepBackup {
			_ = os.RemoveAll(backupDir)
		}
	}()

	_, targetReadErr := os.ReadDir(targetWorkdir)
	hadExistingTarget := targetReadErr == nil
	if targetReadErr != nil && !errors.Is(targetReadErr, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect current Grynvault data: %w", targetReadErr)
	}

	sourceSnapshot := filepath.Join(backupDir, "albyhub-original-snapshot")
	backupHashes, err := CopyWalletData(sourceWorkdir, sourceSnapshot)
	if err != nil {
		return nil, fmt.Errorf("back up Alby Hub data: %w", err)
	}
	if err = CompareHashes(firstHashes, backupHashes); err != nil {
		return nil, fmt.Errorf("source changed while creating recovery snapshot: %w", err)
	}
	reportProgress(options, "recovery_verified", "Recovery snapshot created and hash-verified")

	stageDir := StagePath(targetWorkdir, importID)
	if err = os.Mkdir(stageDir, 0700); err != nil {
		return nil, fmt.Errorf("create import staging folder: %w", err)
	}
	keepStage := false
	defer func() {
		if !keepStage {
			_ = os.RemoveAll(stageDir)
		}
	}()
	stagedHashes, err := CopyWalletData(sourceWorkdir, stageDir)
	if err != nil {
		return nil, fmt.Errorf("stage Alby Hub data: %w", err)
	}
	if err = CompareHashes(backupHashes, stagedHashes); err != nil {
		return nil, fmt.Errorf("verify staged wallet copy: %w", err)
	}

	nwcIntegrity, counts, _, fingerprint, ldkIntegrity, err := ValidateWalletDataWithoutMutation(stageDir)
	if err != nil {
		return nil, err
	}
	validatedHashes, err := HashWalletData(stageDir)
	if err != nil {
		return nil, fmt.Errorf("hash validated wallet copy: %w", err)
	}
	if err = CompareHashes(backupHashes, validatedHashes); err != nil {
		return nil, fmt.Errorf("verify validated wallet copy: %w", err)
	}
	reportProgress(options, "staged_verified", "Staged wallet copy validated without changing wallet data")
	finalSourceHashes, err := HashWalletData(sourceWorkdir)
	if err != nil {
		return nil, fmt.Errorf("final source verification: %w", err)
	}
	if err = CompareHashes(firstHashes, finalSourceHashes); err != nil {
		return nil, fmt.Errorf("source changed during import; staged data was not activated: %w", err)
	}
	processIDs, err = options.OpenFileChecker(sourceWorkdir)
	if err != nil {
		return nil, fmt.Errorf("recheck that Alby Hub is stopped: %w", err)
	}
	if len(processIDs) > 0 {
		return nil, fmt.Errorf("Alby Hub data became active during import in process %s; staged data was not activated", joinProcessIDs(processIDs))
	}

	evidencePath := filepath.Join(backupDir, "import-evidence.json")
	evidence := Evidence{
		Version:            EvidenceVersion,
		ImportID:           importID,
		CreatedAt:          timestamp(now),
		SourcePath:         sourceWorkdir,
		TargetPath:         targetWorkdir,
		SourceFiles:        backupHashes,
		SourceCounts:       counts,
		SourceNWCIntegrity: nwcIntegrity,
		SourceLDKIntegrity: ldkIntegrity,
		WalletFingerprint:  fingerprint,
		ReleaseVersion:     version.Tag,
		SourceCommit:       version.SourceCommit,
		LDKLibrarySHA256:   version.LDKLibrarySHA256,
		Checkpoints: []Checkpoint{
			{Phase: PhaseStaged, CreatedAt: timestamp(now), Detail: "source quiescent; recovery snapshot and staged copy hash-verified"},
		},
	}
	if err = WriteJSONAtomic(evidencePath, &evidence); err != nil {
		return nil, fmt.Errorf("write import evidence: %w", err)
	}

	quarantine := Quarantine{
		Version:           QuarantineVersion,
		ImportID:          importID,
		Phase:             PhaseVerificationNeeded,
		CreatedAt:         timestamp(now),
		UpdatedAt:         timestamp(now),
		SourcePath:        sourceWorkdir,
		TargetPath:        targetWorkdir,
		BackupLocation:    backupDir,
		EvidencePath:      evidencePath,
		WalletFingerprint: fingerprint,
		SourceFiles:       backupHashes,
		SourceCounts:      counts,
	}
	if err = WriteJSONAtomic(QuarantinePath(stageDir), &quarantine); err != nil {
		return nil, fmt.Errorf("write migration quarantine: %w", err)
	}

	control := Control{
		Version:           ControlVersion,
		ImportID:          importID,
		Phase:             PhaseStaged,
		CreatedAt:         timestamp(now),
		UpdatedAt:         timestamp(now),
		SourcePath:        sourceWorkdir,
		TargetPath:        targetWorkdir,
		StagePath:         stageDir,
		PreImportPath:     PreImportPath(targetWorkdir, importID),
		HadExistingTarget: hadExistingTarget,
		BackupLocation:    backupDir,
		EvidencePath:      evidencePath,
		SourceFiles:       backupHashes,
		WalletFingerprint: fingerprint,
	}
	if err = WriteJSONAtomic(ControlPath(targetWorkdir), &control); err != nil {
		return nil, fmt.Errorf("write import control: %w", err)
	}

	keepBackup = true
	keepStage = true
	reportProgress(options, "ready_to_restart", "Verified copies are ready; restarting into offline verification")
	summary.Counts = counts
	summary.NWCIntegrity = nwcIntegrity
	summary.LDKIntegrity = ldkIntegrity
	summary.WalletFingerprint = fingerprint
	summary.WalletFingerprintHint = FingerprintHint(fingerprint)
	summary.DurableFileCount = len(backupHashes)
	return &StageResult{RestartRequired: true, BackupLocation: backupDir, Summary: *summary}, nil
}

func OpenWalletProcessIDs(workdir string) ([]int, error) {
	if runtime.GOOS != "darwin" {
		return nil, nil
	}
	paths := []string{
		filepath.Join(workdir, "nwc.db"),
		filepath.Join(workdir, "nwc.db-wal"),
		filepath.Join(workdir, "nwc.db-shm"),
		filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite"),
		filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite-wal"),
		filepath.Join(workdir, "ldk", "storage", "ldk_node_data.sqlite-shm"),
	}
	arguments := []string{"-F", "p"}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			arguments = append(arguments, path)
		}
	}
	if len(arguments) == 2 {
		return nil, errors.New("wallet database files are missing")
	}
	command := exec.CommandContext(context.Background(), "/usr/sbin/lsof", arguments...)
	output, err := command.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) && exitError.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}
	return parseLsofProcessIDs(string(output), os.Getpid()), nil
}

func parseLsofProcessIDs(output string, currentPID int) []int {
	seen := make(map[int]struct{})
	result := []int{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "p") {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimPrefix(line, "p"))
		if err != nil || pid == currentPID {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		result = append(result, pid)
	}
	return result
}

func joinProcessIDs(processIDs []int) string {
	values := make([]string, len(processIDs))
	for index, processID := range processIDs {
		values[index] = strconv.Itoa(processID)
	}
	return strings.Join(values, ", ")
}
