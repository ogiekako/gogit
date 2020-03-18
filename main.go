package main

import "os"

const DIR = ".git"

func main() {
	os.Open(DIR)
}

// object
// - blob : only file contents.
// - tree  : a directory contents (filename -> object hash)
// - commit: commit object (tree hash + commit info)
