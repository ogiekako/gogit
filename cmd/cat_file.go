package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

func init() {
	subcommands.Register(&catFileCmd{}, "")
}

type catFileCmd struct{}

func (*catFileCmd) Name() string             { return "cat-file" }
func (*catFileCmd) Synopsis() string         { return "git cat-file" }
func (*catFileCmd) Usage() string            { return "git cat-file type object\n" }
func (*catFileCmd) SetFlags(f *flag.FlagSet) {}
func (c *catFileCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		fmt.Fprint(os.Stderr, "Usage: ", c.Usage())
		return subcommands.ExitFailure
	}
	if err := catFile(f.Arg(0), f.Arg(1)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func catFile(typ, sha string) error {
	r, err := git.NewRepo("", false)
	if err != nil {
		return err
	}
	o, err := git.ReadObject(r, sha)
	if err != nil {
		return err
	}
	fmt.Print(string(o.Encode()))
	return nil
}
