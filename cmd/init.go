package cmd

import (
	"context"
	"flag"
	"log"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

func init() {
	subcommands.Register(&initCmd{}, "")
}

type initCmd struct{}

func (*initCmd) Name() string             { return "init" }
func (*initCmd) Synopsis() string         { return "git init" }
func (*initCmd) Usage() string            { return "git init [path]" }
func (*initCmd) SetFlags(f *flag.FlagSet) {}
func (*initCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if _, err := git.NewRepo(f.Arg(0), true); err != nil {
		log.Fatal("init:  ", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
