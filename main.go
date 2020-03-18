package main

import "os"

const DIR = ".git"

func main() {
	os.Open(DIR)
}

type blob []byte
type tree map[]

// object
// - blob : only file contents.
// - tree  : a directory contents (filename -> object hash)
// - commit: commit object (tree hash, parent hash + commit info)
