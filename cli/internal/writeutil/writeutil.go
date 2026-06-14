// Package writeutil provides small helpers for working with io.Writer.
package writeutil

import "io"

// ErrWriter wraps an io.Writer and remembers the first write error so callers
// can perform a series of writes and check the result once at the end, instead
// of checking after every Fprint call. Once an error occurs, all subsequent
// writes are skipped and the stored error is returned.
type ErrWriter struct {
	w   io.Writer
	Err error
}

// New returns an ErrWriter wrapping w.
func New(w io.Writer) *ErrWriter { return &ErrWriter{w: w} }

// Write implements io.Writer, short-circuiting after the first error.
func (e *ErrWriter) Write(p []byte) (int, error) {
	if e.Err != nil {
		return 0, e.Err
	}
	var n int
	n, e.Err = e.w.Write(p)
	return n, e.Err
}
