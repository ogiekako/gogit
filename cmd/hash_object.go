package cmd

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

func init() {
	subcommands.Register(&hashObjectCmd{}, "")
}

type hashObjectCmd struct {
	typ   string
	write bool
}

func (*hashObjectCmd) Name() string     { return "hash-object" }
func (*hashObjectCmd) Synopsis() string { return "git hash-object" }
func (*hashObjectCmd) Usage() string    { return "git hash-object [-t type] [-w write] path\n" }
func (c *hashObjectCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.typ, "t", "blob", "type")
	f.BoolVar(&c.write, "w", false, "write")
}
func (c *hashObjectCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Fprint(os.Stderr, "Usage: ", c.Usage())
		return subcommands.ExitFailure
	}
	if err := hashObject(f.Arg(0), c.typ, c.write); err != nil {
		fmt.Fprint(os.Stderr, err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func hashObject(path, typ string, write bool) error {
	var r *git.Repo
	if write {
		var err error
		r, err = git.NewRepo("", false)
		if err != nil {
			return err
		}
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	sha, err := git.ObjectHash(b, typ, r)
	if err != nil {
		return err
	}
	fmt.Println(sha)
	return nil
}
