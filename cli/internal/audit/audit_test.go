package audit

import (
	"os"
	"path/filepath"
	"testing"
)

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
