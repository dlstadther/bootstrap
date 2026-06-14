// Package toolupgrade checks and upgrades top-level CLI tools the dotfiles
// assume are installed but do not manage via Brewfile or mise.
package toolupgrade

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/dlstadther/bootstrap/cli/internal/writeutil"
)

// State is the upgrade status of a tool.
type State int

const (
	StateUpToDate State = iota
	StateUpdateAvailable
	StateUnknown
	StateNotInstalled
)

// Status is the evaluated state of one tool.
type Status struct {
	Name    string
	Current string
	Latest  string // "" means unknown
	State   State
}

// Tool is one top-level managed binary.
type Tool interface {
	Name() string
	Installed(exec iexec.LookPathExecutor) bool
	CurrentVersion(exec iexec.LookPathExecutor) (string, error)
	LatestVersion(exec iexec.LookPathExecutor) (string, error) // "" + nil means unknown, not an error
	Upgrade(exec iexec.LookPathExecutor) error                 // idempotent: no-op if already current
}

// Evaluate determines the Status of a tool. Pure derivation: any error or empty
// version collapses to StateUnknown rather than failing the run.
func Evaluate(t Tool, exec iexec.LookPathExecutor) Status {
	s := Status{Name: t.Name()}
	if !t.Installed(exec) {
		s.State = StateNotInstalled
		return s
	}
	cur, err := t.CurrentVersion(exec)
	if err != nil || cur == "" {
		s.State = StateUnknown
		return s
	}
	s.Current = cur

	latest, err := t.LatestVersion(exec)
	if err != nil || latest == "" {
		s.State = StateUnknown
		return s
	}
	s.Latest = latest

	if cur == latest {
		s.State = StateUpToDate
	} else {
		s.State = StateUpdateAvailable
	}
	return s
}

// Options configures a Run.
type Options struct {
	Check bool      // print the table and exit; no prompts, no upgrades
	Out   io.Writer // defaults to os.Stdout if nil
}

// Decider receives every upgrade candidate and returns the set of tool names the
// user approved. It MUST collect all answers before returning; Run applies
// upgrades only after it returns, so no upgrade can run mid-prompt.
type Decider func(candidates []Status) (approved map[string]bool, err error)

// Run evaluates all tools, prints a status table, and (unless Check) prompts via
// decider and applies the approved upgrades.
func Run(opts Options, exec iexec.LookPathExecutor, tools []Tool, decider Decider) error {
	dst := opts.Out
	if dst == nil {
		dst = os.Stdout
	}
	out := writeutil.New(dst)

	statuses := make([]Status, 0, len(tools))
	byName := make(map[string]Tool, len(tools))
	for _, t := range tools {
		statuses = append(statuses, Evaluate(t, exec))
		byName[t.Name()] = t
	}
	if err := renderTable(out, statuses); err != nil {
		return err
	}

	if opts.Check {
		return out.Err
	}

	var candidates []Status
	for _, s := range statuses {
		if s.State == StateUpdateAvailable || s.State == StateUnknown {
			candidates = append(candidates, s)
		}
	}
	if len(candidates) == 0 {
		fmt.Fprintln(out, "\nAll tools up to date.")
		return out.Err
	}

	approved, err := decider(candidates)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	var upgraded, skipped, failed int
	for _, c := range candidates {
		if !approved[c.Name] {
			skipped++
			continue
		}
		fmt.Fprintf(out, "  → %s … ", c.Name)
		if err := byName[c.Name].Upgrade(exec); err != nil {
			fmt.Fprintf(out, "FAILED: %v\n", err)
			failed++
			continue
		}
		fmt.Fprintln(out, "done")
		upgraded++
	}
	fmt.Fprintf(out, "\nSummary: %d upgraded, %d skipped, %d failed.\n", upgraded, skipped, failed)
	return out.Err
}

func renderTable(out io.Writer, statuses []Status) error {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TOOL\tCURRENT\tLATEST\tSTATUS")
	for _, s := range statuses {
		cur := s.Current
		if cur == "" {
			cur = "—"
		}
		latest := s.Latest
		if latest == "" {
			latest = "—"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Name, cur, latest, stateLabel(s.State))
	}
	return tw.Flush()
}

func stateLabel(st State) string {
	switch st {
	case StateUpToDate:
		return "up to date"
	case StateUpdateAvailable:
		return "update available"
	case StateNotInstalled:
		return "not installed"
	default:
		return "unknown"
	}
}

// StdinDecider returns a Decider that asks y/N for each candidate, reading from in
// and prompting to out. It reads all answers before returning. Empty/EOF ⇒ No.
func StdinDecider(in io.Reader, out io.Writer) Decider {
	return func(candidates []Status) (map[string]bool, error) {
		approved := make(map[string]bool, len(candidates))
		reader := bufio.NewReader(in)
		for _, c := range candidates {
			latest := c.Latest
			if latest == "" {
				latest = "?"
			}
			fmt.Fprintf(out, "Upgrade %s (%s → %s)? [y/N] ", c.Name, c.Current, latest)
			line, _ := reader.ReadString('\n')
			ans := strings.ToLower(strings.TrimSpace(line))
			approved[c.Name] = ans == "y" || ans == "yes"
		}
		return approved, nil
	}
}
