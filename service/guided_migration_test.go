package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/getAlby/hub/config"
	hubdb "github.com/getAlby/hub/db"
	"github.com/getAlby/hub/guidedmigration"
	"github.com/getAlby/hub/logger"
	"github.com/stretchr/testify/require"
)

func TestStartAppFailsClosedWhileMigrationIsQuarantined(t *testing.T) {
	svc := &service{
		migrationQuarantine: &guidedmigration.Quarantine{
			Phase: guidedmigration.PhaseVerificationNeeded,
		},
	}
	require.ErrorIs(t, svc.StartApp("never-read"), ErrMigrationQuarantined)
}

func TestActivatedMigrationDoesNotBlockOrdinaryStartGuard(t *testing.T) {
	svc := &service{
		migrationQuarantine: &guidedmigration.Quarantine{
			Phase: guidedmigration.PhaseActivated,
		},
	}
	require.False(t, svc.isMigrationQuarantined())
}

func TestGuidedMigrationActivationRetryRequiresPreservedVerification(t *testing.T) {
	verification := &guidedmigration.VerificationSummary{NodeID: "node-a"}
	require.True(t, guidedMigrationReadyForActivation(&guidedmigration.Quarantine{
		Phase:        guidedmigration.PhaseVerified,
		Verification: verification,
	}))
	require.True(t, guidedMigrationReadyForActivation(&guidedmigration.Quarantine{
		Phase:        guidedmigration.PhaseActivationFailed,
		Verification: verification,
	}))
	require.False(t, guidedMigrationReadyForActivation(&guidedmigration.Quarantine{
		Phase: guidedmigration.PhaseActivationFailed,
	}))
	require.False(t, guidedMigrationReadyForActivation(&guidedmigration.Quarantine{
		Phase:        guidedmigration.PhaseVerificationNeeded,
		Verification: verification,
	}))
	require.False(t, guidedMigrationReadyForActivation(nil))
}

func TestGuidedMigrationContinuityComparison(t *testing.T) {
	expected := &guidedmigration.VerificationSummary{
		NodeID:             "node-a",
		Network:            "regtest",
		ChannelCount:       1,
		StoredPaymentCount: 2,
		Channels: []guidedmigration.ChannelSummary{{
			UserChannelID: "channel-a",
			CapacitySats:  1000,
		}},
		LatestStaticBackupSHA256: "backup-hash",
	}
	actual := *expected
	actual.VerifiedAt = time.Now().UTC().Format(time.RFC3339Nano)
	require.True(t, sameGuidedMigrationContinuity(expected, &actual))
	actual.Channels = append([]guidedmigration.ChannelSummary(nil), expected.Channels...)
	actual.Channels[0].CapacitySats++
	require.False(t, sameGuidedMigrationContinuity(expected, &actual))
}

func TestNewServiceQuarantinePreventsAutoUnlockAndNetworkInitialization(t *testing.T) {
	logger.Init("4")
	workdir, password := prepareMigrationServiceWorkdir(t)

	require.NoError(t, guidedmigration.WriteJSONAtomic(guidedmigration.QuarantinePath(workdir), &guidedmigration.Quarantine{
		Version:    guidedmigration.QuarantineVersion,
		ImportID:   "test-quarantine",
		Phase:      guidedmigration.PhaseVerificationNeeded,
		TargetPath: workdir,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	svc, err := NewService(ctx)
	require.NoError(t, err)
	require.True(t, svc.IsGuidedMigrationQuarantined())
	require.Nil(t, svc.GetLNClient())
	require.Empty(t, svc.GetRelayStatuses())
	require.ErrorIs(t, svc.StartApp(password), ErrMigrationQuarantined)
	_, err = svc.VerifyGuidedMigration("wrong-password")
	require.ErrorIs(t, err, ErrInvalidPassword)
	_, err = svc.VerifyGuidedMigration(password)
	require.ErrorContains(t, err, "identity no longer matches")
	require.Nil(t, svc.GetLNClient())
	cancel()
	svc.Shutdown()
}

func TestNewServiceRecoversInterruptedActivationIntoOfflineFailure(t *testing.T) {
	workdir, _ := prepareMigrationServiceWorkdir(t)
	require.NoError(t, guidedmigration.WriteJSONAtomic(guidedmigration.QuarantinePath(workdir), &guidedmigration.Quarantine{
		Version:    guidedmigration.QuarantineVersion,
		ImportID:   "interrupted-activation",
		Phase:      guidedmigration.PhaseActivating,
		TargetPath: workdir,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	svc, err := NewService(ctx)
	require.NoError(t, err)
	require.True(t, svc.IsGuidedMigrationQuarantined())
	require.Nil(t, svc.GetLNClient())
	require.Equal(t, guidedmigration.PhaseActivationFailed, svc.migrationQuarantine.Phase)
	require.Contains(t, svc.migrationQuarantine.LastError, "interrupted")
	persisted, err := guidedmigration.LoadQuarantine(workdir)
	require.NoError(t, err)
	require.Equal(t, guidedmigration.PhaseActivationFailed, persisted.Phase)
	cancel()
	svc.Shutdown()
}

func TestNewServiceKeepsControlOnlyBlockedImportOffline(t *testing.T) {
	workdir, _ := prepareMigrationServiceWorkdir(t)
	require.NoError(t, guidedmigration.WriteJSONAtomic(guidedmigration.ControlPath(workdir), &guidedmigration.Control{
		Version:       guidedmigration.ControlVersion,
		ImportID:      "blocked-control-only",
		Phase:         guidedmigration.PhaseBlocked,
		TargetPath:    workdir,
		BlockedReason: "source wallet changed after staging",
	}))

	ctx, cancel := context.WithCancel(context.Background())
	svc, err := NewService(ctx)
	require.NoError(t, err)
	require.True(t, svc.IsGuidedMigrationQuarantined())
	require.Nil(t, svc.GetLNClient())
	require.Equal(t, guidedmigration.PhaseBlocked, svc.migrationQuarantine.Phase)
	require.Equal(t, "source wallet changed after staging", svc.migrationQuarantine.LastError)
	require.ErrorIs(t, svc.StartApp("local-test-password"), ErrMigrationQuarantined)
	cancel()
	svc.Shutdown()
}

func prepareMigrationServiceWorkdir(t *testing.T) (string, string) {
	t.Helper()
	workdir := t.TempDir()
	password := "local-test-password"
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	database, err := hubdb.NewDB(filepath.Join(workdir, "nwc.db"), false)
	require.NoError(t, err)
	cfg, err := config.NewConfig(&config.AppConfig{Workdir: workdir, Network: "regtest"}, database)
	require.NoError(t, err)
	require.NoError(t, cfg.SaveUnlockPasswordCheck(password))
	require.NoError(t, cfg.SetUpdate("Mnemonic", mnemonic, password))
	require.NoError(t, cfg.SetUpdate("LNBackendType", config.LDKBackendType, ""))
	require.NoError(t, cfg.SetUpdate("NodeLastStartTime", time.Now().UTC().Format(time.RFC3339), ""))
	require.NoError(t, hubdb.Stop(database))

	t.Setenv("WORK_DIR", workdir)
	t.Setenv("DATABASE_URI", "nwc.db")
	t.Setenv("LOG_TO_FILE", "false")
	t.Setenv("NETWORK", "regtest")
	t.Setenv("LDK_NETWORK", "regtest")
	t.Setenv("AUTO_UNLOCK_PASSWORD", password)
	return workdir, password
}
