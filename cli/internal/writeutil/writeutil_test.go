package writeutil_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/dlstadther/bootstrap/cli/internal/writeutil"
)

func TestErrWriter_PassThrough(t *testing.T) {
	var buf bytes.Buffer
	ew := writeutil.New(&buf)
	fmt.Fprint(ew, "hello")
	if ew.Err != nil {
		t.Fatalf("unexpected err: %v", ew.Err)
	}
	if buf.String() != "hello" {
		t.Fatalf("got %q", buf.String())
	}
}

type failWriter struct {
	err   error
	calls int
}

func (f *failWriter) Write(p []byte) (int, error) {
	f.calls++
	return 0, f.err
}

func TestErrWriter_StoresAndShortCircuits(t *testing.T) {
	fw := &failWriter{err: errors.New("disk full")}
	ew := writeutil.New(fw)

	fmt.Fprint(ew, "first")
	if ew.Err == nil {
		t.Fatal("expected stored error after failing write")
	}

	fmt.Fprint(ew, "second")
	if fw.calls != 1 {
		t.Fatalf("subsequent writes should be skipped; got %d calls", fw.calls)
	}
}
