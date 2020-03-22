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
	subcommands.Register(&logCmd{}, "")
}

type logCmd struct {
	typ   string
	write bool
}

func (*logCmd) Name() string             { return "log" }
func (*logCmd) Synopsis() string         { return "git log" }
func (*logCmd) Usage() string            { return "git log [object]\n" }
func (*logCmd) SetFlags(f *flag.FlagSet) {}
func (*logCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	sha := f.Arg(0)
	if sha == "" {
		sha = "HEAD"
	}
	if err := doLog(sha); err != nil {
		fmt.Fprint(os.Stderr, err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func doLog(sha string) error {
	r, err := git.NewRepo("", false)
	if err != nil {
		return err
	}
	return git.WriteLog(os.Stdout, r, sha)
}
