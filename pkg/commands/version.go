package commands

import (
	"fmt"

	"github.com/acorn-io/acorn-dns/pkg/version"
	"github.com/urfave/cli/v2"
)

func execute(c *cli.Context) error {
	fmt.Printf("%s\n", version.Get())

	return nil
}

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "print version",
		Action: execute,
	}
}
