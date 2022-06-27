package main

import (
	"os"
	"path"

	"github.com/acorn-io/acorn-dns/pkg/commands"
	_ "github.com/acorn-io/acorn-dns/pkg/commands"
	"github.com/acorn-io/acorn-dns/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// log panics forces exit
			if _, ok := r.(*logrus.Entry); ok {
				os.Exit(1)
			}
			panic(r)
		}
	}()

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Let's do DNS"
	app.Version = version.Get().String()
	app.Authors = []*cli.Author{
		{
			Name:  "The Acorn Labs Dev Team",
			Email: "engineering@acorn.io",
		},
	}

	app.Commands = commands.GetCommands()
	app.CommandNotFound = func(context *cli.Context, command string) {
		logrus.Fatalf("Command %s not found.", command)
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
