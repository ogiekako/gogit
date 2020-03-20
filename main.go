package main

// https://wyag.thb.lt/ in Golang

import (
	"os"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Init()
	switch os.Args[1] {
	case "init":
		doInit(os.Args[2:])
	}
}
