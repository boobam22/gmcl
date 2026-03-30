package main

import (
	"log"

	"github.com/boobam22/gmcl/cli"
	"github.com/boobam22/gmcl/cmd"
)

func main() {
	if err := execute(); err != nil {
		log.SetFlags(0)
		log.Fatalln(err)
	}
}

func execute() error {
	gmcl, err := cli.NewGmcl()
	if err != nil {
		return err
	}
	defer gmcl.Close()

	root := cmd.NewRootCmd(gmcl)
	return root.Execute()
}
