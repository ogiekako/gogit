package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

type lsTreeCmd struct{}

func (*lsTreeCmd) Name() string             { return "ls-tree" }
func (*lsTreeCmd) Synopsis() string         { return "git ls-tree" }
func (*lsTreeCmd) Usage() string            { return "git ls-tree object" }
func (*lsTreeCmd) SetFlags(f *flag.FlagSet) {}
func (c *lsTreeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Fprint(os.Stderr, "Usage: ", c.Usage())
		return subcommands.ExitFailure
	}
	if err := lsTree(f.Arg(0)); err != nil {
		fmt.Fprintln(os.Stderr, "ls-tree: ", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func lsTree(sha string) error {
	r, err := git.NewRepo("", false)
	if err != nil {
		return err
	}
	o, err := git.ReadObject(r, sha)
	if err != nil {
		return err
	}
	if o.Format != "tree" {
		return fmt.Errorf("format %s != tree", o.Format)
	}
	tree := o.Tree
	for _, c := range tree {
		pad := ""
		if len(c.Mode) == 5 {
			pad = "0"
		}
		o, err := git.ReadObject(r, c.SHA)
		if err != nil {
			return err
		}
		fmt.Printf("%s%s %s %s\t%s\n", pad, string(c.Mode), o.Format, c.SHA, string(c.Path))
	}
	return nil
}
