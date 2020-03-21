package git

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
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
	// format is blob, commit, tag or tree.
	format string
	repo   *Repo

	// Encode encodes itself to bytes.
	Encode func() []byte
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
	// case "commit":
	// 	return NewCommit(repo, b[y+1:])
	// case "tree":
	// 	return NewTree(repo, b[y+1:])
	// case "tag":
	// 	return NewTag(repo, b[y+1:])
	case "blob":
		return NewBlob(repo, b[y+1:]), nil
	default:
		return nil, fmt.Errorf("Unknown type %s for object %s", kind, sha)
	}
}

// ObjectHash computes object hash from the data, if repo is not nil stores the object into repo.
func ObjectHash(data []byte, fmt string, repo *Repo) (string, error) {
	var o *Object
	switch fmt {
	case "blob":
		o = NewBlob(repo, data)
	default:
	}
	return o.HashData()
}

// HashData computes sha1 hash of the object.
// Stores object if o.repo is non-nil.
func (o *Object) HashData() (string, error) {
	data := o.Encode()
	hash := sha1.New()
	result := fmt.Sprintf("%s %d\x00%s", o.format, len(data), data)
	fmt.Fprint(hash, result)
	sha := hex.EncodeToString(hash.Sum(nil))

	if o.repo != nil {
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

func NewBlob(repo *Repo, data []byte) *Object {
	return &Object{
		format: "blob",
		repo:   repo,
		Encode: func() []byte {
			return data
		},
	}
}

type Commit struct {
}

func NewCommit(repo string, data []byte) *Commit {
	return nil
}

type Blob struct {
	data []byte
}

func (b *Blob) Encode() []byte {
	return nil
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
