package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dlstadther/bootstrap/cli/cmd/brew"
	"github.com/dlstadther/bootstrap/cli/cmd/claude"
	"github.com/dlstadther/bootstrap/cli/cmd/cmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tool"
)

var rootCmd = &cobra.Command{
	Use:   "bs",
	Short: "Bootstrap CLI — dotfile and machine management",
	Long: `bs automates personal developer-machine setup: it symlinks dotfiles from the
bootstrap repo into your home directory and manages Homebrew packages, mise tools,
Claude plugins, and tmux/cmux agent workspaces.

It is a single-user tool, intended for the owner of the bootstrap repo to keep
their macOS and Linux machines in sync with version-controlled configuration.

Use 'bs <command> --help' for details on any subcommand.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(brew.Cmd)
	rootCmd.AddCommand(claude.Cmd)
	rootCmd.AddCommand(cmux.Cmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(tmux.Cmd)
	rootCmd.AddCommand(tool.Cmd)
}
