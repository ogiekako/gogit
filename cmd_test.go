package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ogiekako/gogit/testutil"
)

var gitdir = func() string {
	s, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return s
}()

var prog string

func TestMain(m *testing.M) {
	td, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(td)
	prog = filepath.Join(td, "git")
	b, err := exec.Command("go", "build", "-o", prog).CombinedOutput()
	if err != nil {
		panic(string(b) + ": " + err.Error())
	}
	os.Exit(m.Run())
}

type testD struct {
	t   *testing.T
	dir string
	ctx context.Context
}

func testData(t *testing.T) (_ testD, cancel func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("set up: ", err)
	}
	ctx, cl := context.WithTimeout(context.Background(), time.Second)
	return testD{t, dir, ctx}, func() {
		cl()
		if err := os.RemoveAll(dir); err != nil {
			t.Error("clean up: ", err)
		}
	}
}

func run(td testD, args ...string) string {
	must := func(err error) {
		if err != nil {
			td.t.Error(err)
		}
	}
	must(os.Chdir(td.dir))
	defer func() { must(os.Chdir(gitdir)) }()

	cmd := exec.CommandContext(td.ctx, prog, args...)
	var eb bytes.Buffer
	cmd.Stderr = &eb
	b, err := cmd.Output()
	if err != nil {
		td.t.Fatalf("%v: %v", eb.String(), err)
	}
	if s := eb.String(); s != "" {
		td.t.Log("Log: ", s)
	}
	return string(b)
}

func TestInit(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	if s := run(td, "init"); s != "" {
		t.Errorf(`%q != ""`, s)
	}
	if b, err := ioutil.ReadFile(filepath.Join(td.dir, ".git", "HEAD")); err != nil {
		t.Error(err)
	} else if got, want := string(b), "ref: refs/heads/master\n"; got != want {
		t.Errorf("%q != %q", got, want)
	}
}

func TestCatFile(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	run(td, "init")

	b := testutil.ReadFile(t, "testdata/2262de0c121f22df8e78f5a37d6e114fd322c0b0")
	testutil.WriteFile(t, b, td.dir, ".git", "objects", "22", "62de0c121f22df8e78f5a37d6e114fd322c0b0")

	if s := run(td, "cat-file", "blob", "2262de0c121f22df8e78f5a37d6e114fd322c0b0"); s != "hoge\n" {
		t.Errorf(`%q != "hoge\n"`, s)
	}
}

func TestHashObject(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	run(td, "init")
	testutil.WriteFile(t, []byte("hoge\n"), td.dir, "a")

	if got, want := run(td, "hash-object", "a"), "2262de0c121f22df8e78f5a37d6e114fd322c0b0\n"; got != want {
		t.Errorf("%q != %q", got, want)
	}
	objPath := filepath.Join(td.dir, ".git", "objects", "22", "62de0c121f22df8e78f5a37d6e114fd322c0b0")
	if _, err := os.Stat(objPath); !os.IsNotExist(err) {
		t.Errorf("%s exists", objPath)
	}

	run(td, "hash-object", "-w", "a")
	if _, err := os.Stat(objPath); err != nil {
		t.Errorf("%s not exists: %v", objPath, err)
	}
}
func TestLog(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	testutil.Copy(t, filepath.Join(td.dir, ".git"), "testdata/gitdir")

	got := run(td, "log", "0a380ee19ff3c304bd4c6bd8d0000d4c1070b3d3")
	want := `digraph wyaglog{
c_0a380ee19ff3c304bd4c6bd8d0000d4c1070b3d3 -> c_6aba443f3b8da367cafd04b17c0d33acbdec8475
c_6aba443f3b8da367cafd04b17c0d33acbdec8475 -> c_f6cd3846af74cdaf49efe8874e0ecdf6b8c56327
c_0a380ee19ff3c304bd4c6bd8d0000d4c1070b3d3 -> c_8c93c7625fe3d44432383432565e2fc31090833d
c_8c93c7625fe3d44432383432565e2fc31090833d -> c_f6cd3846af74cdaf49efe8874e0ecdf6b8c56327
}
`
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("(-got +want)\n%s", diff)
	}
}
