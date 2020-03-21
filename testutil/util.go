package testutil

import (
	"io/ioutil"
	"os"
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
