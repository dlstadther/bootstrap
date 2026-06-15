package git_test

import (
	"errors"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/git"
	"github.com/dlstadther/bootstrap/cli/internal/version"
)

func TestRepoPath(t *testing.T) {
	t.Run("env override wins", func(t *testing.T) {
		getenv := func(key string) string {
			if key == "BOOTSTRAP_REPO" {
				return "/env/repo"
			}
			return ""
		}
		if got := git.RepoPath(getenv); got != "/env/repo" {
			t.Errorf("got %q want /env/repo", got)
		}
	})

	t.Run("falls back to compiled RepoPath when env empty", func(t *testing.T) {
		orig := version.RepoPath
		version.RepoPath = "/compiled/repo"
		defer func() { version.RepoPath = orig }()

		getenv := func(string) string { return "" }
		if got := git.RepoPath(getenv); got != "/compiled/repo" {
			t.Errorf("got %q want /compiled/repo", got)
		}
	})
}

type fakeExec struct {
	out string
	err error
}

func (f *fakeExec) Run(_ string, _ ...string) (string, error) {
	return f.out, f.err
}

func TestCurrentHash(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		err     error
		want    string
		wantErr bool
	}{
		{"returns trimmed hash", "abc1234\n", nil, "abc1234", false},
		{"propagates error", "", errors.New("not a repo"), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := git.CurrentHash("/some/path", &fakeExec{out: tt.out, err: tt.err})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v got %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestIsDirty(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		err     error
		want    bool
		wantErr bool
	}{
		{"clean repo", "", nil, false, false},
		{"dirty repo", " M some/file\n", nil, true, false},
		{"git error", "", errors.New("git failed"), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := git.IsDirty("/some/path", &fakeExec{out: tt.out, err: tt.err})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v got %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}
