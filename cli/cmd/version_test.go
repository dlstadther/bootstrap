package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/version"
)

// versionExec answers git rev-parse / status by inspecting the args.
type versionExec struct {
	hash      string
	hashErr   error
	status    string
	statusErr error
}

func (f versionExec) Run(_ string, args ...string) (string, error) {
	for _, a := range args {
		if a == "rev-parse" {
			return f.hash, f.hashErr
		}
		if a == "status" {
			return f.status, f.statusErr
		}
	}
	return "", nil
}

func TestRunVersion(t *testing.T) {
	origHash := version.CommitHash
	defer func() { version.CommitHash = origHash }()
	version.CommitHash = "abc1234"

	tests := []struct {
		name       string
		repoPath   string
		ex         versionExec
		wantOut    []string
		wantErrOut []string
	}{
		{
			name:     "no repo path",
			repoPath: "",
			wantOut:  []string{"compiled: abc1234", "repo:     unknown"},
		},
		{
			name:     "clean and in sync",
			repoPath: "/repo",
			ex:       versionExec{hash: "abc1234", status: ""},
			wantOut:  []string{"repo:     abc1234"},
		},
		{
			name:     "dirty repo",
			repoPath: "/repo",
			ex:       versionExec{hash: "abc1234", status: " M file"},
			wantOut:  []string{"(dirty)"},
		},
		{
			name:     "drift",
			repoPath: "/repo",
			ex:       versionExec{hash: "deadbeef", status: ""},
			wantOut:  []string{"(drift"},
		},
		{
			name:       "hash read error goes to stderr",
			repoPath:   "/repo",
			ex:         versionExec{hashErr: errors.New("not a repo")},
			wantErrOut: []string{"error reading repo hash"},
		},
		{
			name:     "dirty check error",
			repoPath: "/repo",
			ex:       versionExec{hash: "abc1234", statusErr: errors.New("boom")},
			wantOut:  []string{"error checking dirty state"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := runVersion(&out, &errOut, tt.repoPath, tt.ex); err != nil {
				t.Fatalf("runVersion() error: %v", err)
			}
			for _, want := range tt.wantOut {
				if !strings.Contains(out.String(), want) {
					t.Errorf("stdout missing %q, got:\n%s", want, out.String())
				}
			}
			for _, want := range tt.wantErrOut {
				if !strings.Contains(errOut.String(), want) {
					t.Errorf("stderr missing %q, got:\n%s", want, errOut.String())
				}
			}
		})
	}
}
