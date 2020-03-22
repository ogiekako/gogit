package main

// https://wyag.thb.lt/ in Golang

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&initCmd{}, "")
	subcommands.Register(&catFileCmd{}, "")
	subcommands.Register(&hashObjectCmd{}, "")
	subcommands.Register(&logCmd{}, "")
	subcommands.Register(&lsTreeCmd{}, "")
	subcommands.Register(&checkoutCmd{}, "")
	subcommands.Register(&showRefCmd{}, "")
	subcommands.Register(&tagCmd{}, "")
	subcommands.Register(&revParseCmd{}, "")
	flag.Parse()
	os.Exit(int(subcommands.Execute(context.Background())))
}
