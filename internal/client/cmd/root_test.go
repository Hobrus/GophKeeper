package cmd

import (
	"bytes"
	"os"
	"runtime"
	"testing"
)

func withTempHome(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	oldHOME, hadHOME := os.LookupEnv("HOME")
	oldUSERPROFILE, hadUSERPROFILE := os.LookupEnv("USERPROFILE")
	os.Setenv("HOME", dir)
	os.Setenv("USERPROFILE", dir)
	if runtime.GOOS == "windows" {
		os.Setenv("HOMEDRIVE", "")
		os.Setenv("HOMEPATH", "")
	}
	return func() {
		if hadHOME {
			os.Setenv("HOME", oldHOME)
		} else {
			os.Unsetenv("HOME")
		}
		if hadUSERPROFILE {
			os.Setenv("USERPROFILE", oldUSERPROFILE)
		} else {
			os.Unsetenv("USERPROFILE")
		}
	}
}

func TestRoot_VersionAndVault(t *testing.T) {
	cleanup := withTempHome(t)
	defer cleanup()

	root := NewRootCmd("1.0.0", "2025-08-13")
	out := new(bytes.Buffer)
	root.SetOut(out)

	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 {
		t.Fatalf("no version output")
	}

	// vault init + status
	out.Reset()
	root.SetArgs([]string{"vault", "init"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	root.SetArgs([]string{"vault", "status"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}
