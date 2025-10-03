package vault

import (
	"os"
	"runtime"
	"testing"
)

func TestGenerateSaveLoad(t *testing.T) {
	// isolate from user home by overriding Path via temp HOME
	dir := t.TempDir()
	// Override home dir env for both Unix and Windows
	oldHOME, hadHOME := os.LookupEnv("HOME")
	oldUSERPROFILE, hadUSERPROFILE := os.LookupEnv("USERPROFILE")
	os.Setenv("HOME", dir)
	os.Setenv("USERPROFILE", dir)
	if runtime.GOOS == "windows" {
		// ensure HOMEDRIVE+HOMEPATH do not interfere
		os.Setenv("HOMEDRIVE", "")
		os.Setenv("HOMEPATH", "")
	}
	t.Cleanup(func() {
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
	})

	if Exists() {
		t.Fatalf("key should not exist")
	}
	key, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != KeyLength {
		t.Fatalf("len: %d", len(key))
	}
	if !Exists() {
		t.Fatalf("key must exist after generate")
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != KeyLength {
		t.Fatalf("loaded len %d", len(loaded))
	}
}
