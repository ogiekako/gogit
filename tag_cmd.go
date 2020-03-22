package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

type tagCmd struct {
	object bool
}

func (*tagCmd) Name() string     { return "tag" }
func (*tagCmd) Synopsis() string { return "git tag" }
func (*tagCmd) Usage() string    { return "git tag [-a] name object" }
func (c *tagCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.object, "a", false, "create tag object instead of lightweight tag")
}
func (c *tagCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		fmt.Fprint(os.Stderr, "Usage: ", c.Usage())
		return subcommands.ExitFailure
	}
	if err := tag(f.Arg(0), f.Arg(1), c.object); err != nil {
		fmt.Fprintln(os.Stderr, "tag: ", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func tag(name, sha string, object bool) error {
	r, err := git.NewRepo("", false)
	if err != nil {
		return err
	}
	return git.Tag(r, name, sha, object)
}
