package audit

import (
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
