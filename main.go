package main

// https://wyag.thb.lt/ in Golang

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	_ "github.com/ogiekako/gogit/cmd" // register commands
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	// TODO: implement ls-files command, parsing index file.
	flag.Parse()
	os.Exit(int(subcommands.Execute(context.Background())))
}
