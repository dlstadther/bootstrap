package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// syncExec returns a fixed result for every call and counts invocations.
type syncExec struct {
	err   error
	calls int
}

func (f *syncExec) Run(_ string, _ ...string) (string, error) {
	f.calls++
	return "", f.err
}

func writeSettings(t *testing.T, plugins string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	body := `{"enabledPlugins": {` + plugins + `}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunSync_AllSucceed(t *testing.T) {
	settings := writeSettings(t, `"p@m": true`)
	ex := &syncExec{}
	var errOut bytes.Buffer

	if err := runSync(&errOut, settings, ex, false); err != nil {
		t.Fatalf("runSync() error: %v", err)
	}
	if errOut.Len() != 0 {
		t.Errorf("expected no error output, got %q", errOut.String())
	}
}

func TestRunSync_AggregatesErrors(t *testing.T) {
	settings := writeSettings(t, `"p@m": true`)
	ex := &syncExec{err: errors.New("command failed")}
	var errOut bytes.Buffer

	err := runSync(&errOut, settings, ex, true)
	if err == nil {
		t.Fatal("expected error when steps fail")
	}
	for _, want := range []string{"mise error", "brew error", "plugins error"} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("stderr missing %q, got:\n%s", want, errOut.String())
		}
	}
}
