package audit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ibrew "github.com/dlstadther/bootstrap/cli/internal/brew"
)

// Executor runs a command and returns combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
}

// Options configures the audit run.
type Options struct {
	All      bool   // show OK symlinks in addition to problems
	RepoPath string // absolute path to the bootstrap repo
}

// Run audits dotfile symlinks and brew package drift.
func Run(opts Options, exec Executor) error {
	if opts.RepoPath == "" {
		return fmt.Errorf("repo path is not set; run 'make install' or set $BOOTSTRAP_REPO")
	}

	machine, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("hostname: %w", err)
	}
	// Use short hostname (up to first '.')
	if idx := strings.Index(machine, "."); idx >= 0 {
		machine = machine[:idx]
	}

	dotfilesDir := filepath.Join(opts.RepoPath, "dotfiles")
	hostsDir := filepath.Join(opts.RepoPath, "hosts")
	hostDir := filepath.Join(hostsDir, machine)

	printSection("Shared dotfiles (" + dotfilesDir + ")")
	if err := checkDir(dotfilesDir, dotfilesDir, opts.All); err != nil {
		return err
	}

	if _, statErr := os.Stat(hostDir); os.IsNotExist(statErr) {
		printSection("Host-specific dotfiles")
		fmt.Printf("  no host directory for %q (hosts/%s not found)\n", machine, machine)
	} else {
		printSection("Host-specific dotfiles (hosts/" + machine + ")")
		if err := checkDir(hostDir, hostDir, opts.All); err != nil {
			return err
		}
	}

	printSection("Brew package drift")
	return checkBrew(opts.RepoPath, exec)
}

func printSection(title string) {
	fmt.Printf("\n=== %s ===\n", title)
}

func checkDir(dir, prefix string, showAll bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var entries []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == ".gitkeep" {
			return nil
		}
		entries = append(entries, path)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(entries)

	if len(entries) == 0 {
		fmt.Println("  (no files)")
		return nil
	}

	for _, src := range entries {
		rel, ok := relPath(prefix, src)
		if !ok {
			continue
		}
		target := filepath.Join(home, rel)

		info, statErr := os.Lstat(target)
		if os.IsNotExist(statErr) {
			fmt.Printf("  MISSING   %s\n", rel)
			continue
		}
		if statErr != nil {
			return statErr
		}

		if info.Mode()&os.ModeSymlink != 0 {
			link, readErr := os.Readlink(target)
			if readErr != nil {
				return readErr
			}
			if linkMatches(target, link, src) {
				if showAll {
					fmt.Printf("  OK        %s\n", rel)
				}
			} else {
				fmt.Printf("  FOREIGN   %s -> %s\n", rel, link)
			}
		} else {
			fmt.Printf("  REAL FILE %s\n", rel)
		}
	}
	return nil
}

// relPath returns src relative to prefix. ok is false when src is not a proper
// descendant of prefix (guards the slice against root/short paths from WalkDir).
func relPath(prefix, src string) (rel string, ok bool) {
	if !strings.HasPrefix(src, prefix) || len(src) <= len(prefix)+1 {
		return "", false
	}
	return src[len(prefix)+1:], true
}

// linkMatches reports whether a symlink at target with the given Readlink value
// points at src. link may be relative (resolved against target's directory) or
// absolute; both are cleaned and, as a fallback, fully resolved before compare.
func linkMatches(target, link, src string) bool {
	resolved := link
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(target), resolved)
	}
	resolved = filepath.Clean(resolved)
	if resolved == filepath.Clean(src) {
		return true
	}
	rl, err1 := filepath.EvalSymlinks(resolved)
	rs, err2 := filepath.EvalSymlinks(src)
	return err1 == nil && err2 == nil && rl == rs
}

func checkBrew(repoPath string, exec Executor) error {
	brewfileSrc, err := ibrew.BrewfilePath(repoPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(brewfileSrc); os.IsNotExist(err) {
		fmt.Println("  no Brewfile found in repo")
		return nil
	}

	tmp, err := os.CreateTemp("", ".Brewfile.current.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	_ = tmp.Close()
	defer func() { _ = os.Remove(tmpName) }()

	_, err = exec.Run("brew", "bundle", "dump", "--force", "--file="+tmpName)
	if err != nil {
		return fmt.Errorf("brew bundle dump: %w", err)
	}

	repoLines, err := readBrewLines(brewfileSrc)
	if err != nil {
		return err
	}
	machineLines, err := readBrewLines(tmpName)
	if err != nil {
		return err
	}

	repoSet := toSet(repoLines)
	machineSet := toSet(machineLines)

	fmt.Println("\n  In repo but NOT installed on machine:")
	printed := false
	for _, l := range repoLines {
		if !machineSet[l] {
			fmt.Printf("    %s\n", l)
			printed = true
		}
	}
	if !printed {
		fmt.Println("    (none)")
	}

	fmt.Println("\n  Installed on machine but NOT in repo:")
	printed = false
	for _, l := range machineLines {
		if !repoSet[l] {
			fmt.Printf("    %s\n", l)
			printed = true
		}
	}
	if !printed {
		fmt.Println("    (none)")
	}

	return nil
}

func readBrewLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if isBrewLine(line) {
			lines = append(lines, line)
		}
	}
	sort.Strings(lines)
	return lines, nil
}

func isBrewLine(line string) bool {
	for _, prefix := range []string{"brew ", "cask ", "tap ", "mas ", "vscode ", "npm ", "go ", "uv "} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func toSet(lines []string) map[string]bool {
	s := make(map[string]bool, len(lines))
	for _, l := range lines {
		s[l] = true
	}
	return s
}
