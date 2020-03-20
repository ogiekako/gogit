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
	flag.Parse()
	os.Exit(int(subcommands.Execute(context.Background())))
}
