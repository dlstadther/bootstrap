package claude

import (
	"fmt"
	"os"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/spf13/cobra"

	"github.com/dlstadther/bootstrap/cli/internal/pluginupgrade"
)

var pluginCheckOnly bool

var pluginUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check and optionally upgrade Claude Code plugins",
	Long: `upgrade lists all installed, enabled plugins, prompts yes/no for each
one up front, then applies only the approved updates. The claude CLI cannot
report the latest available plugin version, so every installed plugin is
offered; "claude plugins update" always pulls the latest.

Use --check to print the status table without prompting or upgrading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		executor := &iexec.Real{}
		tools, err := pluginupgrade.Discover(executor)
		if err != nil {
			return err
		}
		if len(tools) == 0 {
			fmt.Fprintln(os.Stdout, "No enabled plugins found.")
			return nil
		}
		return pluginupgrade.Run(
			pluginupgrade.Options{Check: pluginCheckOnly, Out: os.Stdout},
			executor,
			tools,
			pluginupgrade.StdinDecider(os.Stdin, os.Stdout),
		)
	},
}

func init() {
	pluginUpgradeCmd.Flags().BoolVar(&pluginCheckOnly, "check", false, "print status and exit without prompting or upgrading")
}
