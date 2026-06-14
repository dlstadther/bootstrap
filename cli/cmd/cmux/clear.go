package cmux

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	iexec "github.com/dlstadther/bootstrap/cli/internal/exec"

	"github.com/spf13/cobra"

	icmux "github.com/dlstadther/bootstrap/cli/internal/cmux"
)

var clearYes bool

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Close all cmux workspaces except the current one",
	Long: `Close every open cmux workspace except the one you are currently in.

Without --yes, lists the workspaces that will be closed and prompts for
confirmation before proceeding.

Must be run from inside cmux — the active workspace is identified via the
CMUX_WORKSPACE_ID environment variable.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		skipRef := callerWorkspaceRef()

		toClose, err := icmux.ListOtherWorkspaces(skipRef, &iexec.CMux{})
		if err != nil {
			return err
		}
		if len(toClose) == 0 {
			fmt.Println("No other workspaces to close.")
			return nil
		}

		if !clearYes {
			fmt.Printf("The following %d workspace(s) will be closed:\n", len(toClose))
			for _, ws := range toClose {
				fmt.Printf("  %s  %s\n", ws.Ref, ws.Title)
			}
			fmt.Print("Close these workspaces? [y/N] ")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		icmux.ClearWorkspaces(toClose, &iexec.CMux{})
		fmt.Printf("Closed %d workspace(s).\n", len(toClose))
		return nil
	},
}

func init() {
	clearCmd.Flags().BoolVarP(&clearYes, "yes", "y", false, "Skip confirmation prompt")
}
