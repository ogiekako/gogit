package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

type revParseCmd struct{}

func (*revParseCmd) Name() string     { return "rev-parse" }
func (*revParseCmd) Synopsis() string { return "git rev-parse name[^{type}]" }
func (*revParseCmd) Usage() string {
	return `git rev-parse name[^{type}]
  Example:
	- git rev-parse HEAD
	- git rev-parse HEAD^{tree}
`
}
func (*revParseCmd) SetFlags(f *flag.FlagSet) {}
func (c *revParseCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 1 {
		fmt.Fprint(os.Stderr, "Usage: ", c.Usage())
		return subcommands.ExitFailure
	}
	if err := revParse(f.Arg(0)); err != nil {
		fmt.Fprintln(os.Stderr, "rev-parse: ", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

var revParseRE = regexp.MustCompile(`^(.*?)(?:\^\{(.+)\})?$`)

func revParse(query string) error {
	m := revParseRE.FindStringSubmatch(query)
	if len(m) == 0 {
		return fmt.Errorf("invalid argument %s", query)
	}
	r, err := git.NewRepo("", false)
	if err != nil {
		return err
	}
	o, err := git.FindObject(r, m[1], m[2])
	if err != nil {
		return err
	}
	if o == nil {
		return fmt.Errorf("found no object")
	}
	sha, err := o.HashData(false)
	if err != nil {
		return err
	}
	fmt.Println(sha)
	return nil
}
