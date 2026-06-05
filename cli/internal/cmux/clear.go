package cmux

import (
	"encoding/json"
	"fmt"
)

// WorkspaceInfo holds the ref and title of a cmux workspace.
type WorkspaceInfo struct {
	Ref   string
	Title string
}

// ListOtherWorkspaces returns all open workspaces except skipRef.
// Returns an error if cmux is not running.
func ListOtherWorkspaces(skipRef string, exec Executor) ([]WorkspaceInfo, error) {
	if out, err := exec.Run("cmux", "ping"); err != nil {
		detail := out
		if detail == "" {
			detail = err.Error()
		}
		return nil, fmt.Errorf("cmux is not running (%s): ensure cmux is installed and running", detail)
	}

	out, err := exec.Run("cmux", "workspace", "list", "--json")
	if err != nil || out == "" {
		return nil, nil
	}
	var result workspaceListJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil, nil
	}

	var infos []WorkspaceInfo
	for _, ws := range result.Workspaces {
		if skipRef != "" && ws.Ref == skipRef {
			continue
		}
		infos = append(infos, WorkspaceInfo{Ref: ws.Ref, Title: ws.Title})
	}
	return infos, nil
}

// ClearWorkspaces closes each workspace in the provided list.
func ClearWorkspaces(toClose []WorkspaceInfo, exec Executor) {
	for _, ws := range toClose {
		exec.Run("cmux", "workspace", "close", ws.Ref) //nolint:errcheck
	}
}
