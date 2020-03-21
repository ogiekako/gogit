package main

import (
	"bytes"
	"testing"
	"github.com/google/go-cmp/cmp"
)

func TestDefaultConfig(t *testing.T) {
	var b bytes.Buffer
	defaultConfig().WriteToIndent(&b, "\t")
	got := b.String()
	want := `[core]
	repositoryformatversion = 0
	filemode = false
	bare = false

`
	if diff := cmp.Diff(got, want); diff != "" {
        t.Errorf("(-got +want)\n%s", diff)
	}
}