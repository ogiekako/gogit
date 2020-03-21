package testutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func ReadFile(t *testing.T, elem ...string) []byte {
	b, err := ioutil.ReadFile(filepath.Join(elem...))
	if err != nil {
		t.Error(err)
	}
	return b
}

// WriteFile writes file creating directories if needed.
func WriteFile(t *testing.T, b []byte, elem ...string) {
	p := filepath.Join(elem...)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Error(err)
	}
	if err := ioutil.WriteFile(p, b, 0644); err != nil {
		t.Error(err)
	}
}

// Copy copies src to dst, recursively if src is a directory.
func Copy(t *testing.T, dst, src string) {
	if b, err := exec.Command("cp", "-r", src, dst).CombinedOutput(); err != nil {
		t.Fatalf("%s: %v", string(b), err)
	}
}
