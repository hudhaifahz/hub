package guidedmigration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FinishResult struct {
	Activated bool
	Blocked   bool
	Reason    string
}

func FinishPending(targetWorkdir string, now time.Time) (*FinishResult, error) {
	controlPath := ControlPath(targetWorkdir)
	control := Control{}
	if err := ReadJSON(controlPath, &control); errors.Is(err, os.ErrNotExist) {
		return &FinishResult{}, nil
	} else if err != nil {
		return nil, err
	}
	if control.Version != ControlVersion || control.TargetPath != filepath.Clean(targetWorkdir) {
		return nil, errors.New("pending guided import control is invalid")
	}
	if control.Phase == PhaseAbandoning {
		if err := finishAbandon(&control, controlPath, now); err != nil {
			return nil, fmt.Errorf("finish abandoning guided import: %w", err)
		}
		return &FinishResult{}, nil
	}
	if control.Phase == PhaseBlocked {
		return &FinishResult{Blocked: true, Reason: control.BlockedReason}, nil
	}
	if control.Phase == PhaseVerificationNeeded || control.Phase == PhaseVerified || control.Phase == PhaseActivationFailed || control.Phase == PhaseActivating || control.Phase == PhaseActivated {
		return &FinishResult{Activated: true}, nil
	}

	if control.Phase == PhaseStaged {
		processIDs, err := OpenWalletProcessIDs(control.SourcePath)
		if err != nil {
			return persistBlocked(&control, controlPath, now, fmt.Sprintf("source wallet process state could not be rechecked: %v", err))
		}
		if len(processIDs) > 0 {
			return persistBlocked(&control, controlPath, now, fmt.Sprintf("source wallet became active after staging in process %s", joinProcessIDs(processIDs)))
		}

		stageHashes, err := HashWalletData(control.StagePath)
		if err != nil {
			return persistBlocked(&control, controlPath, now, fmt.Sprintf("staged wallet validation failed: %v", err))
		}
		if err = CompareHashes(control.SourceFiles, stageHashes); err != nil {
			return persistBlocked(&control, controlPath, now, fmt.Sprintf("staged wallet bytes changed: %v", err))
		}
		sourceHashes, err := HashWalletData(control.SourcePath)
		if err != nil {
			return persistBlocked(&control, controlPath, now, fmt.Sprintf("source wallet could not be rechecked: %v", err))
		}
		if err = CompareHashes(control.SourceFiles, sourceHashes); err != nil {
			return persistBlocked(&control, controlPath, now, fmt.Sprintf("source wallet changed after staging: %v", err))
		}

		_, preImportErr := os.Stat(control.PreImportPath)
		_, targetErr := os.Stat(targetWorkdir)
		if preImportErr == nil && errors.Is(targetErr, os.ErrNotExist) {
			// The prior process completed the target rename but stopped before it
			// durably advanced the control record. Resume from that checkpoint.
			control.Phase = PhaseTargetMoved
			control.UpdatedAt = timestamp(now)
			if err = WriteJSONAtomic(controlPath, &control); err != nil {
				return nil, err
			}
		} else if preImportErr == nil && targetErr == nil {
			return persistBlocked(&control, controlPath, now, "guided import recovery is ambiguous: both target and pre-import paths exist")
		} else if preImportErr != nil && !errors.Is(preImportErr, os.ErrNotExist) {
			return nil, preImportErr
		} else if targetErr != nil && !errors.Is(targetErr, os.ErrNotExist) {
			return nil, targetErr
		} else if targetErr == nil {
			if err = os.Rename(targetWorkdir, control.PreImportPath); err != nil {
				return nil, fmt.Errorf("preserve pre-import target: %w", err)
			}
			control.Phase = PhaseTargetMoved
			control.UpdatedAt = timestamp(now)
			if err = WriteJSONAtomic(controlPath, &control); err != nil {
				_ = os.Rename(control.PreImportPath, targetWorkdir)
				return nil, err
			}
		} else if control.HadExistingTarget {
			return persistBlocked(&control, controlPath, now, "pre-import Grynvault target disappeared before activation")
		} else {
			control.Phase = PhaseTargetMoved
			control.UpdatedAt = timestamp(now)
			if err = WriteJSONAtomic(controlPath, &control); err != nil {
				return nil, err
			}
		}
	}

	if control.Phase == PhaseTargetMoved {
		if _, err := os.Stat(targetWorkdir); errors.Is(err, os.ErrNotExist) {
			if err = os.Rename(control.StagePath, targetWorkdir); err != nil {
				return nil, fmt.Errorf("activate staged wallet: %w", err)
			}
		} else if err != nil {
			return nil, err
		} else if _, stageErr := os.Stat(control.StagePath); stageErr == nil {
			return nil, errors.New("guided import activation is ambiguous: both target and stage exist")
		}
		quarantine := Quarantine{}
		if err := ReadJSON(QuarantinePath(targetWorkdir), &quarantine); err != nil {
			return nil, fmt.Errorf("read activated migration quarantine: %w", err)
		}
		if quarantine.ImportID != control.ImportID || quarantine.WalletFingerprint != control.WalletFingerprint {
			return nil, errors.New("activated migration quarantine does not match import control")
		}
		activatedHashes, err := HashWalletData(targetWorkdir)
		if err != nil {
			return nil, err
		}
		if err = CompareHashes(control.SourceFiles, activatedHashes); err != nil {
			return nil, fmt.Errorf("activated wallet bytes differ: %w", err)
		}
		quarantine.ActivatedBytesVerified = true
		quarantine.UpdatedAt = timestamp(now)
		if err = WriteJSONAtomic(QuarantinePath(targetWorkdir), &quarantine); err != nil {
			return nil, fmt.Errorf("record activated byte verification: %w", err)
		}
		preImportBackup := filepath.Join(control.BackupLocation, "grynvault-before-import")
		_, preImportErr := os.Stat(control.PreImportPath)
		_, backupErr := os.Stat(preImportBackup)
		if preImportErr == nil && backupErr == nil {
			return nil, errors.New("both pre-import recovery locations exist")
		}
		if preImportErr == nil {
			if !errors.Is(backupErr, os.ErrNotExist) {
				return nil, backupErr
			}
			if err = os.Rename(control.PreImportPath, preImportBackup); err != nil {
				return nil, fmt.Errorf("preserve pre-import Grynvault data: %w", err)
			}
		} else if !errors.Is(preImportErr, os.ErrNotExist) {
			return nil, preImportErr
		} else if control.HadExistingTarget && errors.Is(backupErr, os.ErrNotExist) {
			return nil, errors.New("pre-import Grynvault recovery data is missing")
		} else if backupErr != nil && !errors.Is(backupErr, os.ErrNotExist) {
			return nil, backupErr
		}
		control.Phase = PhaseVerificationNeeded
		control.UpdatedAt = timestamp(now)
		if err = WriteJSONAtomic(controlPath, &control); err != nil {
			return nil, err
		}
		return &FinishResult{Activated: true}, nil
	}

	return nil, fmt.Errorf("unsupported guided import phase %q", control.Phase)
}

func RequestAbandon(targetWorkdir string, now time.Time) (string, error) {
	controlPath := ControlPath(targetWorkdir)
	control := Control{}
	if err := ReadJSON(controlPath, &control); err != nil {
		return "", err
	}
	if control.Version != ControlVersion || control.TargetPath != filepath.Clean(targetWorkdir) {
		return "", errors.New("pending guided import control is invalid")
	}
	if control.Phase == PhaseActivating || control.Phase == PhaseActivated {
		return "", errors.New("an activating or activated wallet cannot be abandoned")
	}
	control.Phase = PhaseAbandoning
	control.UpdatedAt = timestamp(now)
	if err := WriteJSONAtomic(controlPath, &control); err != nil {
		return "", err
	}
	return control.BackupLocation, nil
}

func finishAbandon(control *Control, controlPath string, now time.Time) error {
	if err := os.MkdirAll(control.BackupLocation, 0700); err != nil {
		return err
	}
	abandonedStage := filepath.Join(control.BackupLocation, "abandoned-staged-wallet")
	if err := moveIfPresent(control.StagePath, abandonedStage); err != nil {
		return fmt.Errorf("archive staged wallet: %w", err)
	}

	quarantine := Quarantine{}
	quarantineErr := ReadJSON(QuarantinePath(control.TargetPath), &quarantine)
	if quarantineErr == nil {
		if quarantine.ImportID != control.ImportID {
			return errors.New("target quarantine does not match the import being abandoned")
		}
		abandonedTarget := filepath.Join(control.BackupLocation, "abandoned-imported-target")
		if err := moveIfPresent(control.TargetPath, abandonedTarget); err != nil {
			return fmt.Errorf("archive quarantined target: %w", err)
		}
	} else if !errors.Is(quarantineErr, os.ErrNotExist) {
		return quarantineErr
	} else if _, preImportErr := os.Stat(control.PreImportPath); preImportErr == nil {
		if _, targetErr := os.Stat(control.TargetPath); targetErr == nil {
			return errors.New("cannot safely abandon an ambiguous target and pre-import pair")
		}
	}

	preImportBackup := filepath.Join(control.BackupLocation, "grynvault-before-import")
	if _, err := os.Stat(control.PreImportPath); err == nil {
		if _, backupErr := os.Stat(preImportBackup); errors.Is(backupErr, os.ErrNotExist) {
			if err = os.Rename(control.PreImportPath, preImportBackup); err != nil {
				return fmt.Errorf("preserve pre-import target: %w", err)
			}
		} else if backupErr != nil {
			return backupErr
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if _, targetErr := os.Stat(control.TargetPath); errors.Is(targetErr, os.ErrNotExist) {
		if _, backupErr := os.Stat(preImportBackup); backupErr == nil {
			if err := os.Rename(preImportBackup, control.TargetPath); err != nil {
				return fmt.Errorf("restore pre-import Grynvault target: %w", err)
			}
		} else if !errors.Is(backupErr, os.ErrNotExist) {
			return backupErr
		}
	} else if targetErr != nil {
		return targetErr
	}

	evidence := Evidence{}
	if err := ReadJSON(control.EvidencePath, &evidence); err != nil {
		return fmt.Errorf("read migration evidence: %w", err)
	}
	evidence.Checkpoints = append(evidence.Checkpoints, Checkpoint{
		Phase: PhaseAbandoned, CreatedAt: timestamp(now),
		Detail: "failed or cancelled guided import archived locally without starting Lightning",
	})
	if err := WriteJSONAtomic(control.EvidencePath, &evidence); err != nil {
		return fmt.Errorf("record abandoned import: %w", err)
	}
	control.Phase = PhaseAbandoned
	control.UpdatedAt = timestamp(now)
	if err := WriteJSONAtomic(filepath.Join(control.BackupLocation, "abandoned-import-control.json"), control); err != nil {
		return fmt.Errorf("archive import control: %w", err)
	}
	if err := os.Remove(controlPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func moveIfPresent(source, destination string) error {
	if _, err := os.Stat(source); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(source, destination)
}

func persistBlocked(control *Control, controlPath string, now time.Time, reason string) (*FinishResult, error) {
	control.Phase = PhaseBlocked
	control.BlockedReason = reason
	control.UpdatedAt = timestamp(now)
	if err := WriteJSONAtomic(controlPath, control); err != nil {
		return nil, fmt.Errorf("%s; additionally failed to persist blocked state: %w", reason, err)
	}
	return &FinishResult{Blocked: true, Reason: reason}, nil
}

func LoadQuarantine(workdir string) (*Quarantine, error) {
	record := Quarantine{}
	if err := ReadJSON(QuarantinePath(workdir), &record); err != nil {
		return nil, err
	}
	if record.Version != QuarantineVersion {
		return nil, fmt.Errorf("unsupported migration quarantine version %d", record.Version)
	}
	return &record, nil
}

func LoadControl(workdir string) (*Control, error) {
	control := Control{}
	if err := ReadJSON(ControlPath(workdir), &control); err != nil {
		return nil, err
	}
	return &control, nil
}

func StatusFor(workdir string) (*Status, error) {
	record, err := LoadQuarantine(workdir)
	if err == nil {
		return &Status{
			Active:           record.Phase != PhaseActivated,
			Phase:            record.Phase,
			SourcePath:       record.SourcePath,
			TargetPath:       record.TargetPath,
			BackupLocation:   record.BackupLocation,
			FingerprintHint:  FingerprintHint(record.WalletFingerprint),
			BlockingReason:   record.LastError,
			Verification:     record.Verification,
			LightningOffline: record.Phase != PhaseActivated,
		}, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	control, controlErr := LoadControl(workdir)
	if controlErr == nil {
		return &Status{
			Active:           true,
			Phase:            control.Phase,
			SourcePath:       control.SourcePath,
			TargetPath:       control.TargetPath,
			BackupLocation:   control.BackupLocation,
			FingerprintHint:  FingerprintHint(control.WalletFingerprint),
			BlockingReason:   control.BlockedReason,
			LightningOffline: true,
		}, nil
	}
	if errors.Is(controlErr, os.ErrNotExist) {
		return &Status{}, nil
	}
	return nil, controlErr
}
