package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

type showRefCmd struct{}

func (*showRefCmd) Name() string             { return "show-ref" }
func (*showRefCmd) Synopsis() string         { return "git show-ref" }
func (*showRefCmd) Usage() string            { return "git show-ref object" }
func (*showRefCmd) SetFlags(f *flag.FlagSet) {}
func (c *showRefCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := showRef(); err != nil {
		fmt.Fprintln(os.Stderr, "show-ref: ", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func showRef() error {
	r, err := git.NewRepo("", false)
	if err != nil {
		return err
	}
	m, err := git.Refs(r)
	if err != nil {
		return err
	}
	var ps []string
	for p := range m {
		ps = append(ps, p)
	}
	sort.Strings(ps)
	for _, path := range ps {
		fmt.Printf("%s %s\n", m[path], path[len(".git/"):])
	}
	return nil
}
