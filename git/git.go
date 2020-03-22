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
	"regexp"
	"strconv"
	"strings"

	"github.com/ogiekako/gogit/kvlm"
	"gopkg.in/ini.v1"
)

func init() {
	ini.PrettyFormat = false
	ini.PrettyEqual = true
}

// Object represents a git object.
type Object struct {
	// Type is blob, commit, tag or tree.
	Type string
	repo *Repo

	// Blob is blob data.
	Blob []byte
	// KVLM is commit or tag data.
	KVLM *kvlm.KVLM
	// Tree is tree data.
	Tree Tree

	// Encode encodes object to bytes.
	Encode func() []byte
	// Decode decodes data and updates an object field and returns the object itself.
	// commit, tag -> updates KVLM
	// blob -> updates Blob
	// tree -> updates Tree
	Decode func(data []byte) (*Object, error)
}

type nameResolutionError struct {
	name  string
	cands []string
}

func (e *nameResolutionError) Error() string {
	if len(e.cands) == 0 {
		return fmt.Sprintf("no such reference %s", e.name)
	} else if len(e.cands) > 1 {
		return fmt.Sprintf("ambiguous reference %s: candidates are %v", e.name, e.cands)
	} else {
		panic(fmt.Sprintf("internal error: %q %q", e.name, e.cands[0]))
	}
}

// FindObject finds the matching object.
// If typ is not the zero value chases until an object with the type if found.
func FindObject(repo *Repo, name, typ string) (*Object, error) {
	ss, err := findSHA(repo, name)
	if err != nil {
		return nil, err
	}
	if len(ss) != 1 {
		return nil, &nameResolutionError{name, ss}
	}
	sha := ss[0]
	for {
		o, err := ReadObject(repo, sha)
		if err != nil {
			return nil, err
		}
		if typ == "" || o.Type == typ {
			return o, nil
		}
		switch o.Type {
		case "tag":
			sha = o.KVLM.Get("object")[0]
		case "commit":
			sha = o.KVLM.Get("tree")[0]
		default:
			return nil, fmt.Errorf("found no object of type %s for %s", typ, name)
		}
	}
}

var shortHashRE = regexp.MustCompile("^[0-9A-Fa-f]{4,16}$")

func findSHA(repo *Repo, name string) ([]string, error) {
	if name == "HEAD" {
		sha, err := resolveRef(repo, repo.path("HEAD"))
		if err != nil {
			return nil, err
		}
		return []string{sha}, nil
	}
	if shortHashRE.MatchString(name) {
		fs, err := filepath.Glob(fmt.Sprintf("%s/%s*", name[0:2], name[2:]))
		if err != nil {
			return nil, err
		}
		var res []string
		for _, f := range fs {
			d := filepath.Dir(f)
			res = append(res, d[len(d)-2:]+f[len(f)-18:])
		}
		return res, nil
	}
	var res []string
	s, err := resolveRef(repo, repo.path("refs", "heads", name))
	if err == nil {
		res = append(res, s)
	}
	s, err = resolveRef(repo, repo.path("refs", "tags", name))
	if err == nil {
		res = append(res, s)
	}
	return res, nil
}

// ReadObject reads object for the hash.
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
	typ := string(b[0:x])

	y := bytes.IndexByte(b[x:], '\x00') + x
	size, err := strconv.Atoi(string(b[x+1 : y]))
	if err != nil {
		return nil, err
	}
	if size != len(b)-y-1 {
		return nil, fmt.Errorf("Malformed object %s: bad length", sha)
	}
	switch typ {
	case "commit":
		return newCommit(repo).Decode(b[y+1:])
	case "tree":
		return newTree(repo).Decode(b[y+1:])
	case "tag":
		return newTag(repo).Decode(b[y+1:])
	case "blob":
		return newBlob(repo).Decode(b[y+1:])
	default:
		return nil, fmt.Errorf("Unknown type %s for object %s", typ, sha)
	}
}

// ObjectHash computes object hash from the data, if repo is not nil stores the object into repo.
func ObjectHash(data []byte, typ string, repo *Repo) (string, error) {
	var o *Object
	var err error
	switch typ {
	case "commit":
		o, err = newCommit(repo).Decode(data)
	case "tree":
		o, err = newTree(repo).Decode(data)
	case "tag":
		o, err = newTag(repo).Decode(data)
	case "blob":
		o, err = newBlob(repo).Decode(data)
	default:
		return "", fmt.Errorf("ObjectHash: unsupported type %s", typ)
	}
	if err != nil {
		return "", err
	}
	return o.HashData(repo != nil)
}

// HashData computes sha1 hash of the object.
// Stores object if write is true.
// If always succeeds if write is false.
func (o *Object) HashData(write bool) (string, error) {
	data := o.Encode()
	hash := sha1.New()
	result := fmt.Sprintf("%s %d\x00%s", o.Type, len(data), data)
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
func newBlob(repo *Repo) *Object {
	o := &Object{
		Type: "blob",
		repo: repo,
	}
	o.Encode = func() []byte {
		return o.Blob
	}
	o.Decode = func(data []byte) (*Object, error) {
		o.Blob = data
		return o, nil
	}
	return o
}

// newTag creates an empty tag object.
func newTag(repo *Repo) *Object {
	o := &Object{
		Type: "tag",
		repo: repo,
		KVLM: kvlm.New(),
	}
	o.Encode = func() []byte {
		return []byte(kvlm.Encode(o.KVLM))
	}
	o.Decode = func(data []byte) (*Object, error) {
		kvlm, err := kvlm.Decode(string(data))
		o.KVLM = kvlm
		return o, err
	}
	return o
}

// newCommit creates an empty commit object.
func newCommit(repo *Repo) *Object {
	o := &Object{
		Type: "commit",
		repo: repo,
	}
	o.Encode = func() []byte {
		return []byte(kvlm.Encode(o.KVLM))
	}
	o.Decode = func(data []byte) (*Object, error) {
		kvlm, err := kvlm.Decode(string(data))
		o.KVLM = kvlm
		return o, err
	}
	return o
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
	if o.Type != "commit" {
		return fmt.Errorf("type %s != commit", o.Type)
	}
	m := o.KVLM
	for _, p := range m.Get("parent") {
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

// Repo represents a git repository.
type Repo struct {
	worktree, gitDir string
	conf             *ini.File
}

// NewRepo reads or creates a repository.
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

// newTree craetes an empty tree object.
func newTree(repo *Repo) *Object {
	o := &Object{
		Type: "tree",
		repo: repo,
	}
	o.Encode = func() []byte {
		return encodeTree(o.Tree)
	}
	o.Decode = func(data []byte) (*Object, error) {
		t, err := parseTree(data)
		if err != nil {
			return nil, err
		}
		o.Tree = t
		return o, nil
	}
	return o
}

// Tree is a representation of a tree object.
type Tree []*TreeLeaf

// TreeLeaf is an entry of Tree
type TreeLeaf struct {
	Mode, Path, SHA string
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
	mode := string(raw[start:x])
	y := bytes.IndexByte(raw[x:], '\x00') + x
	path := string(raw[x+1 : y])

	v := big.NewInt(0)
	hex := v.SetBytes(raw[y+1 : y+21]).Text(16)
	var sha string
	for len(sha)+len(hex) < 40 {
		sha += "0"
	}
	sha += hex
	return y + 21, &TreeLeaf{mode, path, sha}, nil
}

func encodeTree(t Tree) []byte {
	var b bytes.Buffer
	for _, l := range t {
		v := big.NewInt(0)
		v.SetString(l.SHA, 16)
		bs := v.Bytes()
		var sha []byte
		for len(sha)+len(bs) < 20 {
			sha = append(sha, 0)
		}
		sha = append(sha, bs...)
		if len(sha) != 20 {
			panic(string(l.SHA))
		}
		fmt.Fprintf(&b, "%s %s\x00%s", l.Mode, l.Path, string(sha))
	}
	return b.Bytes()
}

func resolveRef(repo *Repo, path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	s := string(b)
	if strings.HasPrefix(s, "ref: ") {
		p := s[len("ref: ") : len(s)-1]
		return resolveRef(repo, repo.path(p))
	}
	return s[:len(s)-1], nil
}

// Refs returns mapping from files in refs directory to its hash.
func Refs(r *Repo) (map[string]string, error) {
	m := make(map[string]string)
	return m, filepath.Walk(r.path("refs"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		sha, err := resolveRef(r, path)
		if err != nil {
			return err
		}
		m[path] = sha
		return nil
	})
}

func createRef(r *Repo, refPath, sha string) error {
	return writeFile([]byte(sha+"\n"), r.path("refs", refPath))
}

// Tag creates a new tag object.
func Tag(r *Repo, name, sha string, object bool) error {
	refPath := filepath.Join("tags", name)
	if !object {
		return createRef(r, refPath, sha)
	}
	to, err := ReadObject(r, sha)
	if err != nil {
		return err
	}

	o := newTag(r)
	o.KVLM.Append("tag", name)
	o.KVLM.Append("tagger", "dummy name <dummy@example.com>")
	o.KVLM.Append("object", sha)
	o.KVLM.Append("type", to.Type)
	o.KVLM.Append("", "Dummy commit message.\n")

	tagSHA, err := o.HashData(true)
	if err != nil {
		return err
	}
	return createRef(r, refPath, tagSHA)
}
