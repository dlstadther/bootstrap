package audit

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubExec returns canned output for every call and records the args.
type stubExec struct {
	out  string
	err  error
	args [][]string
}

func (s *stubExec) Run(_ string, args ...string) (string, error) {
	s.args = append(s.args, args)
	// Mimic `brew bundle dump --file=<path>` writing the live state to disk,
	// which checkBrew then reads back.
	if s.err == nil {
		for _, a := range args {
			if strings.HasPrefix(a, "--file=") {
				path := strings.TrimPrefix(a, "--file=")
				if err := os.WriteFile(path, []byte(s.out), 0o644); err != nil {
					return "", err
				}
			}
		}
	}
	return s.out, s.err
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what
// was written, so tests can assert on the audit report.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestRun_Seams(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()

	// Shared dotfile + a host-specific dotfile under the injected hostname.
	dotfiles := filepath.Join(repo, "dotfiles")
	if err := os.MkdirAll(dotfiles, 0o755); err != nil {
		t.Fatal(err)
	}
	zshrc := filepath.Join(dotfiles, ".zshrc")
	if err := os.WriteFile(zshrc, []byte("# zshrc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Repo Brewfile so checkBrew runs the dump path.
	if err := os.WriteFile(filepath.Join(dotfiles, ".Brewfile"), []byte("brew \"git\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	hostDir := filepath.Join(repo, "hosts", "testhost")
	if err := os.MkdirAll(hostDir, 0o755); err != nil {
		t.Fatal(err)
	}
	hostFile := filepath.Join(hostDir, ".gitconfig")
	if err := os.WriteFile(hostFile, []byte("[user]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// home/.zshrc correctly symlinked; home/.gitconfig missing.
	if err := os.Symlink(zshrc, filepath.Join(home, ".zshrc")); err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	// brew dump writes the live state; report it as having an extra package.
	exec := &stubExec{out: "brew \"git\"\nbrew \"jq\"\n"}

	opts := Options{
		All:      true,
		RepoPath: repo,
		Hostname: func() (string, error) { return "testhost.local", nil },
		Home:     home,
		TempDir:  tmpDir,
	}

	var runErr error
	out := captureStdout(t, func() { runErr = Run(opts, exec) })
	if runErr != nil {
		t.Fatalf("Run() error: %v", runErr)
	}

	if !strings.Contains(out, "OK        .zshrc") {
		t.Errorf("expected OK .zshrc line, got:\n%s", out)
	}
	if !strings.Contains(out, "MISSING   .gitconfig") {
		t.Errorf("expected MISSING .gitconfig line, got:\n%s", out)
	}
	if !strings.Contains(out, "hosts/testhost") {
		t.Errorf("expected host section for testhost, got:\n%s", out)
	}
	// jq is installed but not in repo; git is in both.
	_, after, found := strings.Cut(out, "Installed on machine but NOT in repo:")
	if !found || !strings.Contains(after, "brew \"jq\"") {
		t.Errorf("expected jq in machine-only drift, got:\n%s", out)
	}

	if len(exec.args) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(exec.args))
	}
	got := strings.Join(exec.args[0], " ")
	if !strings.HasPrefix(got, "bundle dump") {
		t.Errorf("expected brew bundle dump, got %q", got)
	}
}

func TestRun_HostnameError(t *testing.T) {
	err := Run(Options{
		RepoPath: t.TempDir(),
		Hostname: func() (string, error) { return "", errors.New("boom") },
	}, &stubExec{})
	if err == nil {
		t.Fatal("expected error when hostname fails")
	}
}

func TestIsBrewLine(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"brew \"git\"", true},
		{"cask \"firefox\"", true},
		{"tap \"homebrew/cask\"", true},
		{"mas \"Xcode\", id: 497799835", true},
		{"# comment", false},
		{"", false},
		{"notabrew something", false},
	}
	for _, tt := range tests {
		got := isBrewLine(tt.line)
		if got != tt.want {
			t.Errorf("isBrewLine(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestToSet(t *testing.T) {
	lines := []string{"a", "b", "a"}
	s := toSet(lines)
	if !s["a"] || !s["b"] {
		t.Error("expected a and b in set")
	}
	if len(s) != 2 {
		t.Errorf("expected 2 unique keys, got %d", len(s))
	}
}

func TestRun_MissingRepoPath(t *testing.T) {
	err := Run(Options{RepoPath: ""}, nil)
	if err == nil {
		t.Fatal("expected error for empty RepoPath")
	}
}

func TestRelPath(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		src     string
		wantRel string
		wantOK  bool
	}{
		{"normal file", "/repo/dotfiles", "/repo/dotfiles/.zshrc", ".zshrc", true},
		{"nested file", "/repo/dotfiles", "/repo/dotfiles/.config/git/config", ".config/git/config", true},
		{"src equals prefix", "/repo/dotfiles", "/repo/dotfiles", "", false},
		{"src is prefix plus slash", "/repo/dotfiles", "/repo/dotfiles/", "", false},
		{"src shorter than prefix", "/repo/dotfiles", "/repo", "", false},
		{"src not under prefix", "/repo/dotfiles", "/other/file", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, ok := relPath(tt.prefix, tt.src)
			if ok != tt.wantOK || rel != tt.wantRel {
				t.Errorf("relPath(%q, %q) = (%q, %v), want (%q, %v)",
					tt.prefix, tt.src, rel, ok, tt.wantRel, tt.wantOK)
			}
		})
	}
}

func TestLinkMatches(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "repo", "dotfiles", ".zshrc")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("# zshrc\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	home := filepath.Join(dir, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, ".zshrc")

	tests := []struct {
		name string
		link string // value stored in the symlink
		want bool
	}{
		{"absolute symlink to src", src, true},
		{"relative symlink to src", relTo(t, home, src), true},
		{"foreign absolute", filepath.Join(dir, "elsewhere", ".zshrc"), false},
		{"foreign relative", "../somewhere/.zshrc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Remove(target)
			if err := os.Symlink(tt.link, target); err != nil {
				t.Fatal(err)
			}
			if got := linkMatches(target, tt.link, src); got != tt.want {
				t.Errorf("linkMatches(%q, %q, %q) = %v, want %v",
					target, tt.link, src, got, tt.want)
			}
		})
	}
}

func relTo(t *testing.T, base, target string) string {
	t.Helper()
	rel, err := filepath.Rel(base, target)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}
