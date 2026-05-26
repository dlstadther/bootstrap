package version

// Injected at build time via -ldflags.
var (
	CommitHash = "dev"
	BuildTime  = "unknown"
	RepoPath   = ""
)
