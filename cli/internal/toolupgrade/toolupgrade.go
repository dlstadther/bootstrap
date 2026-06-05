// Package toolupgrade checks and upgrades top-level CLI tools the dotfiles
// assume are installed but do not manage via Brewfile or mise.
package toolupgrade

// Executor runs a command and returns combined output. LookPath reports whether
// a binary is resolvable on PATH. Both are seams for testing.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
	LookPath(name string) (string, error)
}

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
	Installed(exec Executor) bool
	CurrentVersion(exec Executor) (string, error)
	LatestVersion(exec Executor) (string, error) // "" + nil means unknown, not an error
	Upgrade(exec Executor) error                 // idempotent: no-op if already current
}

// Evaluate determines the Status of a tool. Pure derivation: any error or empty
// version collapses to StateUnknown rather than failing the run.
func Evaluate(t Tool, exec Executor) Status {
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
