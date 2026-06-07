package cmd

import (
	"github.com/dlstadther/bootstrap/cli/cmd/brew"
	"github.com/dlstadther/bootstrap/cli/cmd/claude"
	"github.com/dlstadther/bootstrap/cli/cmd/cmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tmux"
	"github.com/dlstadther/bootstrap/cli/cmd/tool"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bs",
	Short: "Bootstrap CLI — dotfile and machine management",
	Long: `bs is a unified CLI for managing your bootstrap dotfiles and machine configuration.

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
	rootCmd.AddCommand(tmux.Cmd)
	rootCmd.AddCommand(tool.Cmd)
}
