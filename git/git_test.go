package git

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

func TestEncodeTree(t *testing.T) {
	tr := Tree([]*TreeLeaf{
		&TreeLeaf{"100644", "a", "2262de0c121f22df8e78f5a37d6e114fd322c0b0"},
		&TreeLeaf{"40000", "hoge", "496d6428b9cf92981dc9495211e6e1120fb6f2ba"},
	})
	raw := encodeTree(tr)
	got, err := parseTree(raw)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(tr, got); diff != "" {
		t.Errorf("(-got +want)\n%s", diff)
	}
}
