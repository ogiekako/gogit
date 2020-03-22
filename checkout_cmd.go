package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/subcommands"
	"github.com/ogiekako/gogit/git"
)

type checkoutCmd struct{}

func (*checkoutCmd) Name() string     { return "checkout" }
func (*checkoutCmd) Synopsis() string { return "git checkout commit path" }
func (*checkoutCmd) Usage() string {
	return `git checkout commit path
  commit - commit or tree to checkout.
  path - The EMPTY directory to checkout on.`
}
func (*checkoutCmd) SetFlags(f *flag.FlagSet) {}
func (c *checkoutCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() < 2 {
		fmt.Fprint(os.Stderr, "Usage: ", c.Usage())
		return subcommands.ExitFailure
	}
	if err := checkout(f.Arg(0), f.Arg(1)); err != nil {
		fmt.Fprintln(os.Stderr, "checkout: ", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func checkout(sha, path string) error {
	if s, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !s.IsDir() {
		return fmt.Errorf("%s is not directory", path)
	} else if m, err := filepath.Glob(path + "/**"); err != nil {
		return err
	} else if len(m) > 0 {
		return fmt.Errorf("%s is not empty", path)
	}

	r, err := git.NewRepo("", false)
	if err != nil {
		return err
	}
	o, err := git.ReadObject(r, sha)
	if err != nil {
		return err
	}
	if o.Format == "commit" {
		m := o.KVLM
		o, err = git.ReadObject(r, m["tree"][0])
		if err != nil {
			return err
		}
	}
	return checkoutTree(r, o, path)
}

func checkoutTree(repo *git.Repo, tree *git.Object, path string) error {
	if tree.Format != "tree" {
		return fmt.Errorf("format %s != tree", tree.Format)
	}
	for _, c := range tree.Tree {
		o, err := git.ReadObject(repo, c.SHA)
		if err != nil {
			return err
		}
		dest := filepath.Join(path, c.Path)
		switch o.Format {
		case "blob":
			if err := ioutil.WriteFile(dest, o.Blob, 0644); err != nil {
				return err
			}
		case "tree":
			if err := os.Mkdir(dest, 0755); err != nil {
				return err
			}
			if err := checkoutTree(repo, o, dest); err != nil {
				return err
			}
		}
	}
	return nil
}
