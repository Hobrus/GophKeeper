package main

import (
	"bytes"
	"testing"

	icmd "gophkeeper/internal/client/cmd"
)

func TestVersionCommand(t *testing.T) {
	root := icmd.NewRootCmd("1.2.3", "2025-08-13")
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if out == "" || out[:10] == "gophkeeper" && len(out) < 10 {
		t.Fatalf("unexpected version output: %q", out)
	}
}
