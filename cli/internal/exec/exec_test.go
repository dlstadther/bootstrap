package exec

import (
	"slices"
	"testing"
)

func TestStripCmuxContext(t *testing.T) {
	in := []string{
		"PATH=/usr/bin",
		"CMUX_WORKSPACE_ID=ws1",
		"HOME=/home/me",
		"CMUX_TAB_ID=tab1",
		"CMUX_SURFACE_ID=surf1",
		"TERM=xterm",
	}
	got := stripCmuxContext(in)
	want := []string{"PATH=/usr/bin", "HOME=/home/me", "TERM=xterm"}
	if !slices.Equal(got, want) {
		t.Fatalf("stripCmuxContext() = %v, want %v", got, want)
	}
}

func TestStripCmuxContext_NoCmuxVars(t *testing.T) {
	in := []string{"PATH=/usr/bin", "HOME=/home/me"}
	got := stripCmuxContext(in)
	if !slices.Equal(got, in) {
		t.Fatalf("stripCmuxContext() = %v, want unchanged %v", got, in)
	}
}
