//go:build wails && !bindings
// +build wails,!bindings

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseReleaseAcceptanceArgs(t *testing.T) {
	workdir := "/tmp/grynvault-acceptance"
	tests := []struct {
		name      string
		args      []string
		workdir   string
		requested bool
		wantErr   string
	}{
		{name: "normal launch", args: []string{"grynvault"}},
		{
			name:      "complete acceptance launch",
			args:      []string{"grynvault", releaseAcceptanceWorkdirArg + workdir, releaseAcceptanceNetworkDarkArg},
			workdir:   workdir,
			requested: true,
		},
		{
			name:    "workdir without network denial",
			args:    []string{"grynvault", releaseAcceptanceWorkdirArg + workdir},
			wantErr: "requires both",
		},
		{
			name:    "network denial without workdir",
			args:    []string{"grynvault", releaseAcceptanceNetworkDarkArg},
			wantErr: "requires both",
		},
		{
			name:    "empty workdir",
			args:    []string{"grynvault", releaseAcceptanceWorkdirArg, releaseAcceptanceNetworkDarkArg},
			wantErr: "cannot be empty",
		},
		{
			name:    "mistyped acceptance flag",
			args:    []string{"grynvault", "--grynvault-release-acceptance-workdir=" + workdir, releaseAcceptanceNetworkDarkArg},
			wantErr: "unrecognized",
		},
		{
			name:    "duplicate workdir",
			args:    []string{"grynvault", releaseAcceptanceWorkdirArg + workdir, releaseAcceptanceWorkdirArg + workdir, releaseAcceptanceNetworkDarkArg},
			wantErr: "more than once",
		},
		{
			name:    "duplicate network denial",
			args:    []string{"grynvault", releaseAcceptanceWorkdirArg + workdir, releaseAcceptanceNetworkDarkArg, releaseAcceptanceNetworkDarkArg},
			wantErr: "more than once",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualWorkdir, requested, err := parseReleaseAcceptanceArgs(test.args)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse acceptance args: %v", err)
			}
			if actualWorkdir != test.workdir || requested != test.requested {
				t.Fatalf("got workdir %q requested %v", actualWorkdir, requested)
			}
		})
	}
}

func TestValidateReleaseAcceptanceWorkdir(t *testing.T) {
	workdir := createReleaseAcceptanceTarget(t)
	actual, err := validateReleaseAcceptanceWorkdir(workdir)
	if err != nil {
		t.Fatalf("validate acceptance workdir: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(workdir)
	if err != nil {
		t.Fatalf("resolve expected workdir: %v", err)
	}
	if actual != resolved {
		t.Fatalf("expected %q, got %q", resolved, actual)
	}

	if err = os.WriteFile(workdir+releaseAcceptanceAuthoritySuffix, []byte("wrong\n"), 0600); err != nil {
		t.Fatalf("replace authority: %v", err)
	}
	if _, err = validateReleaseAcceptanceWorkdir(workdir); err == nil || !strings.Contains(err.Error(), "not bound") {
		t.Fatalf("expected binding error, got %v", err)
	}
}

func TestValidateReleaseAcceptanceWorkdirRejectsUnsafeTargets(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(t *testing.T, workdir string) string
		wantErr string
	}{
		{
			name: "relative path",
			mutate: func(t *testing.T, workdir string) string {
				t.Helper()
				return filepath.Base(workdir)
			},
			wantErr: "must be absolute",
		},
		{
			name: "workdir symlink",
			mutate: func(t *testing.T, workdir string) string {
				t.Helper()
				link := filepath.Join(filepath.Dir(workdir), "workdir-link")
				if err := os.Symlink(workdir, link); err != nil {
					t.Fatalf("create workdir symlink: %v", err)
				}
				return link
			},
			wantErr: "cannot be a symlink",
		},
		{
			name: "workdir permissions",
			mutate: func(t *testing.T, workdir string) string {
				t.Helper()
				if err := os.Chmod(workdir, 0755); err != nil {
					t.Fatalf("change workdir mode: %v", err)
				}
				return workdir
			},
			wantErr: "must not be accessible",
		},
		{
			name: "authority symlink",
			mutate: func(t *testing.T, workdir string) string {
				t.Helper()
				authorityPath := workdir + releaseAcceptanceAuthoritySuffix
				target := authorityPath + ".target"
				authority, err := os.ReadFile(authorityPath)
				if err != nil {
					t.Fatalf("read authority: %v", err)
				}
				if err = os.WriteFile(target, authority, 0600); err != nil {
					t.Fatalf("write authority target: %v", err)
				}
				if err = os.Remove(authorityPath); err != nil {
					t.Fatalf("remove authority: %v", err)
				}
				if err = os.Symlink(target, authorityPath); err != nil {
					t.Fatalf("create authority symlink: %v", err)
				}
				return workdir
			},
			wantErr: "must be a regular file",
		},
		{
			name: "authority permissions",
			mutate: func(t *testing.T, workdir string) string {
				t.Helper()
				if err := os.Chmod(workdir+releaseAcceptanceAuthoritySuffix, 0644); err != nil {
					t.Fatalf("change authority mode: %v", err)
				}
				return workdir
			},
			wantErr: "must not be accessible",
		},
		{
			name: "authority binding",
			mutate: func(t *testing.T, workdir string) string {
				t.Helper()
				if err := os.WriteFile(workdir+releaseAcceptanceAuthoritySuffix, []byte("wrong\n"), 0600); err != nil {
					t.Fatalf("replace authority: %v", err)
				}
				return workdir
			},
			wantErr: "not bound",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workdir := createReleaseAcceptanceTarget(t)
			workdir = test.mutate(t, workdir)
			if _, err := validateReleaseAcceptanceWorkdir(workdir); err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
			}
		})
	}
}

func TestPrepareReleaseAcceptanceLaunchVerifiesBoundaryBeforeService(t *testing.T) {
	workdir := createReleaseAcceptanceTarget(t)
	t.Setenv(releaseAcceptanceActiveEnv, "")
	t.Setenv(releaseAcceptanceWorkdirEnv, "")
	t.Setenv("WORK_DIR", "")
	t.Setenv("LOG_TO_FILE", "")
	t.Setenv("SEND_EVENTS_TO_ALBY", "")
	t.Setenv("AUTO_LINK_ALBY_ACCOUNT", "")

	expected := errors.New("boundary checked")
	originalVerify := verifyReleaseAcceptanceBoundary
	verifyReleaseAcceptanceBoundary = func() error {
		return expected
	}
	t.Cleanup(func() { verifyReleaseAcceptanceBoundary = originalVerify })

	err := prepareReleaseAcceptanceLaunch([]string{
		"grynvault",
		releaseAcceptanceWorkdirArg + workdir,
		releaseAcceptanceNetworkDarkArg,
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected boundary result, got %v", err)
	}
	resolved, err := filepath.EvalSymlinks(workdir)
	if err != nil {
		t.Fatalf("resolve expected workdir: %v", err)
	}
	if os.Getenv("WORK_DIR") != resolved {
		t.Fatalf("WORK_DIR not bound to acceptance target")
	}
	if os.Getenv(releaseAcceptanceActiveEnv) != "1" {
		t.Fatalf("acceptance marker not set before boundary verification")
	}
}

func TestReleaseAcceptanceSandboxHelper(t *testing.T) {
	if os.Getenv("GRYNVAULT_RELEASE_ACCEPTANCE_TEST_HELPER") != "1" {
		t.Skip("helper subprocess only")
	}
	if err := prepareReleaseAcceptanceLaunch([]string{os.Args[0]}); err != nil {
		t.Fatalf("prepare sandboxed acceptance launch: %v", err)
	}
}

func TestReleaseAcceptanceNetworkSandboxIntegration(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS sandbox integration")
	}
	workdir := createReleaseAcceptanceTarget(t)
	command := exec.Command(
		"/usr/bin/sandbox-exec",
		"-p",
		`(version 1)
(allow default)
(deny network-outbound (remote ip "*:*"))`,
		os.Args[0],
		"-test.run=^TestReleaseAcceptanceSandboxHelper$",
	)
	command.Env = append(os.Environ(),
		"GRYNVAULT_RELEASE_ACCEPTANCE_TEST_HELPER=1",
		releaseAcceptanceActiveEnv+"=1",
		releaseAcceptanceWorkdirEnv+"="+workdir,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("sandbox helper failed: %v\n%s", err, output)
	}
}

func createReleaseAcceptanceTarget(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	workdir := filepath.Join(root, "isolated-workdir")
	if err := os.Mkdir(workdir, 0700); err != nil {
		t.Fatalf("create workdir: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(workdir)
	if err != nil {
		t.Fatalf("resolve workdir: %v", err)
	}
	authority := releaseAcceptanceAuthorityVersion + "\n" + resolved + "\n"
	if err := os.WriteFile(workdir+releaseAcceptanceAuthoritySuffix, []byte(authority), 0600); err != nil {
		t.Fatalf("create authority: %v", err)
	}
	return workdir
}
