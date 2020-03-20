package main

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// Object represents a git object.
type Object interface {
	// Encode encodes itself to bytes
	Encode() []byte
}

func NewObject(repo, sha string) (Object, error) {
	path := filepath.Join(repo, "objects", sha[0:2], sha[2:])

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	rc, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	x := bytes.IndexByte(b, ' ')
	kind := string(b[0:x])

	y := bytes.IndexByte(b[x:], '\x00') + x
	size, err := strconv.Atoi(string(b[x:y]))
	if err != nil {
		return nil, err
	}
	if size != len(b)-y-1 {
		return nil, fmt.Errorf("Malformed object %s: bad length", sha)
	}
	switch kind {
	case "commit":
		return NewCommit(repo, b[y+1:])
	case "tree":
		return NewTree(repo, b[y+1:])
	case "tag":
		return NewTag(repo, b[y+1:])
	case "blob":
		return NewBlob(repo, b[y+1:])
	default:
		return nil, fmt.Errorf("Unknown type %s for object %s", kind, sha)
	}
}

type Commit struct {
}

func NewCommit(repo string, data []byte) *Commit {

}

type Blob struct {
	data []byte
}

func (b *Blob) Encode() []byte {

}

type entry struct {
	filename string
}

type Tree struct {
	es []entry
}

type blob []byte
type tree []entry

// type commit ()

// object
// - blob: only file contents.
// - tree: a directory contents: list of (filemode, filename, object hash)
// - commit: commit object (tree hash, parent hash + commit info)

func findRepo(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if s, err := os.Stat(filepath.Join(path, ".git")); err == nil && s.IsDir() {
		return path, nil
	}
	if path == "/" {
		return "", errors.New("no git dir")
	}
	return findRepo(filepath.Join(path, ".."))
}
