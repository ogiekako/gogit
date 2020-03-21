package git

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

func init() {
	ini.PrettyFormat = false
	ini.PrettyEqual = true
}

// Object represents a git object.
type Object struct {
	// Format is blob, commit, tag or tree.
	Format string
	repo   *Repo

	// Encode encodes itself to bytes.
	Encode func() []byte
	// Decode decodes object into format specific representation.
	Decode func() interface{}
}

func ReadObject(repo *Repo, sha string) (*Object, error) {
	path := repo.path("objects", sha[0:2], sha[2:])

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
	size, err := strconv.Atoi(string(b[x+1 : y]))
	if err != nil {
		return nil, err
	}
	if size != len(b)-y-1 {
		return nil, fmt.Errorf("Malformed object %s: bad length", sha)
	}
	switch kind {
	case "commit":
		return newCommit(repo, b[y+1:])
	case "tree":
		return newTree(repo, b[y+1:])
	// case "tag":
	// 	return NewTag(repo, b[y+1:])
	case "blob":
		return newBlob(repo, b[y+1:]), nil
	default:
		return nil, fmt.Errorf("Unknown type %s for object %s", kind, sha)
	}
}

// ObjectHash computes object hash from the data, if repo is not nil stores the object into repo.
func ObjectHash(data []byte, fmt string, repo *Repo) (string, error) {
	var o *Object
	switch fmt {
	case "blob":
		o = newBlob(repo, data)
	default:
	}
	return o.HashData(repo != nil)
}

// HashData computes sha1 hash of the object.
// Stores object if write is true.
// If always succeeds if write is false.
func (o *Object) HashData(write bool) (string, error) {
	data := o.Encode()
	hash := sha1.New()
	result := fmt.Sprintf("%s %d\x00%s", o.Format, len(data), data)
	fmt.Fprint(hash, result)
	sha := hex.EncodeToString(hash.Sum(nil))

	if write {
		p := o.repo.path("objects", sha[0:2], sha[2:])
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return "", err
		}
		f, err := os.Create(p)
		if err != nil {
			return "", err
		}
		w := zlib.NewWriter(f)
		defer w.Close()
		if _, err := w.Write([]byte(result)); err != nil {
			return "", err
		}
	}
	return sha, nil
}

// newBlob creates a blob object from the object file data.
func newBlob(repo *Repo, data []byte) *Object {
	return &Object{
		Format: "blob",
		repo:   repo,
		Encode: func() []byte {
			return data
		},
		Decode: func() interface{} {
			return data
		},
	}
}

// newCommit creates a commit object from the object file data.
func newCommit(repo *Repo, data []byte) (*Object, error) {
	kvlm, err := decodeKVLM(string(data))
	if err != nil {
		return nil, err
	}
	return &Object{
		Format: "commit",
		repo:   repo,
		Encode: func() []byte {
			return data
		},
		Decode: func() interface{} {
			return kvlm
		},
	}, nil
}

// WriteLog writes log to w in graphvis format.
func WriteLog(w io.Writer, repo *Repo, sha string) error {
	if repo == nil {
		return errors.New("repo is nil")
	}
	if _, err := fmt.Fprintln(w, "digraph wyaglog{"); err != nil {
		return err
	}
	if err := writeLog(w, repo, sha); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "}"); err != nil {
		return err
	}
	return nil
}

func writeLog(w io.Writer, repo *Repo, sha string) error {
	o, err := ReadObject(repo, sha)
	if err != nil {
		return err
	}
	if o.Format != "commit" {
		return fmt.Errorf("format %s != commit", o.Format)
	}
	m := o.Decode().(map[string][]string)
	for _, p := range m["parent"] {
		fmt.Fprintf(w, "c_%s -> c_%s\n", sha, p)
		if err := writeLog(w, repo, p); err != nil {
			return err
		}
	}
	return nil
}

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

type Repo struct {
	worktree, gitDir string
	conf             *ini.File
}

func NewRepo(path string, create bool) (*Repo, error) {
	r := &Repo{
		worktree: path,
		gitDir:   filepath.Join(path, ".git"),
	}

	if create {
		var err error
		mkdir := func(elem ...string) {
			if err != nil {
				return
			}
			err = os.MkdirAll(r.path(elem...), 0755)
		}
		mkdir()
		mkdir("objects")
		mkdir("refs", "tags")
		mkdir("refs", "heads")
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(r.path("description"), []byte("Unnamed repository; edit this file 'description' to name the repository.\n"), 0644); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(r.path("HEAD"), []byte("ref: refs/heads/master\n"), 0644); err != nil {
			return nil, err
		}
		if err := defaultConfig().SaveTo(r.path("config")); err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(r.gitDir); err != nil {
		return nil, err
	}
	cf := filepath.Join(r.gitDir, "config")
	var err error
	r.conf, err = ini.Load(cf)
	if err != nil {
		return nil, err
	}
	if vers := r.conf.Section("core").Key("repositoryformatversion").MustInt(); vers != 0 {
		return nil, fmt.Errorf("Unsupported repositoryformatversion %d", vers)
	}

	return r, nil
}

func (r *Repo) path(elem ...string) string {
	return filepath.Join(append([]string{r.gitDir}, elem...)...)
}

func defaultConfig() *ini.File {
	f := ini.Empty()
	s, err := f.NewSection("core")
	if err != nil { // never happen
		panic(err)
	}
	s.Key("repositoryformatversion").SetValue("0")
	s.Key("filemode").SetValue("false")
	s.Key("bare").SetValue("false")
	return f
}

func newTree(repo *Repo, data []byte) (*Object, error) {
	t, err := parseTree(data)
	if err != nil {
		return nil, err
	}
	return &Object{
		Format: "tree",
		repo:   repo,
		Encode: func() []byte {
			return data
		},
		Decode: func() interface{} {
			return t
		},
	}, nil
}

type Tree []*TreeLeaf

type TreeLeaf struct {
	Mode, Path []byte
	SHA        string
}

func parseTree(raw []byte) (Tree, error) {
	pos := 0
	var res []*TreeLeaf
	for pos < len(raw) {
		nPos, l, err := parseTreeEntry(raw, pos)
		if err != nil {
			return nil, err
		}
		pos = nPos
		res = append(res, l)
	}
	return Tree(res), nil
}

func parseTreeEntry(raw []byte, start int) (int, *TreeLeaf, error) {
	x := bytes.IndexByte(raw[start:], ' ') + start
	if x-start != 5 && x-start != 6 {
		return 0, nil, fmt.Errorf("illegal format")
	}
	mode := raw[start:x]
	y := bytes.IndexByte(raw[x:], '\x00') + x
	path := raw[x+1 : y]

	v := big.NewInt(0)
	sha := v.SetBytes(raw[y+1 : y+21]).Text(16)
	return y + 21, &TreeLeaf{mode, path, sha}, nil
}
