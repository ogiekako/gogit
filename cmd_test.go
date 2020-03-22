package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	testutil.Copy(t, filepath.Join(td.dir, "a.tag"), "testdata/a.tag")
	if got, want := run(td, "hash-object", "-t", "tag", "a.tag"), "6521f7bf9c42c397be87988657092931e32ca56f\n"; got != want {
		t.Errorf("got %s; want %s", got, want)
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
func TestLsTree(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	testutil.Copy(t, filepath.Join(td.dir, ".git"), "testdata/gitdir")

	got := run(td, "ls-tree", "32179141295ce1e34105ed7619916254cbe00e6a")
	want := `100644 blob 2262de0c121f22df8e78f5a37d6e114fd322c0b0	a
100644 blob e69de29bb2d1d6434b8b29ae775ad8c2e48c5391	b
100644 blob e69de29bb2d1d6434b8b29ae775ad8c2e48c5391	c
`
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("(-got +want)\n%s", diff)
	}
}
func TestCheckout(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	testutil.Copy(t, filepath.Join(td.dir, ".git"), "testdata/gitdir2")

	for _, sha := range []string{"7a7dd58919381869a1e39be3d0c7f45978a3a04f", "2823188337a27d8b30fa3b1876d1e46ef8f4ba57"} {
		dir := "d_" + sha
		run(td, "checkout", sha, dir)

		if got := testutil.ReadFile(t, td.dir, dir, "a"); string(got) != "hoge\n" {
			t.Errorf("%s != hoge", string(got))
		}
		if got := testutil.ReadFile(t, td.dir, dir, "d", "a"); string(got) != "" {
			t.Errorf(`%q != ""`, got)
		}
	}
}

func TestShowRef(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	testutil.Copy(t, filepath.Join(td.dir, ".git"), "testdata/gitdir2")

	got := run(td, "show-ref")
	want := `6aba443f3b8da367cafd04b17c0d33acbdec8475 refs/heads/c
8c93c7625fe3d44432383432565e2fc31090833d refs/heads/hoge
7a7dd58919381869a1e39be3d0c7f45978a3a04f refs/heads/master
`
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("(-got +want)\n%s", diff)
	}
}

func TestTag(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	testutil.Copy(t, filepath.Join(td.dir, ".git"), "testdata/gitdir2")

	const head = "7a7dd58919381869a1e39be3d0c7f45978a3a04f"

	run(td, "tag", "hoge", head)

	got := run(td, "show-ref")
	want := `6aba443f3b8da367cafd04b17c0d33acbdec8475 refs/heads/c
8c93c7625fe3d44432383432565e2fc31090833d refs/heads/hoge
7a7dd58919381869a1e39be3d0c7f45978a3a04f refs/heads/master
7a7dd58919381869a1e39be3d0c7f45978a3a04f refs/tags/hoge
`
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("(-got +want)\n%s", diff)
	}

	run(td, "tag", "-a", "piyo", head)
	s := run(td, "show-ref")
	var sha string
	for _, l := range strings.Split(s, "\n") {
		if strings.HasSuffix(l, "piyo") {
			sha = strings.Split(l, " ")[0]
		}
	}
	if sha == "" {
		t.Fatalf("tag piyo not found\n%s", s)
	}
	got = run(td, "cat-file", "tag", sha)
	want = `tag piyo
tagger dummy name <dummy@example.com>
object 7a7dd58919381869a1e39be3d0c7f45978a3a04f
type commit

Dummy commit message.
`
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("(-got +want)\n%s", diff)
	}
}

func TestRevParse(t *testing.T) {
	td, cancel := testData(t)
	defer cancel()

	testutil.Copy(t, filepath.Join(td.dir, ".git"), "testdata/gitdir3")

	for _, tc := range []struct {
		name, want string
	}{
		{"HEAD", "7a7dd58919381869a1e39be3d0c7f45978a3a04f\n"},
		{"HEAD^{tree}", "2823188337a27d8b30fa3b1876d1e46ef8f4ba57\n"},
		{"hevy^{commit}", "7a7dd58919381869a1e39be3d0c7f45978a3a04f\n"},
		{"hevy", "cae02c8b5610cb970fa2f5c16b1a9d53b38221f4\n"},
	} {
		if got := run(td, "rev-parse", tc.name); got != tc.want {
			t.Errorf("%s: got %s; want %s", tc.name, got, tc.want)
		}
	}
}
