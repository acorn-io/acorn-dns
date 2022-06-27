package commands

import (
	"context"

	"github.com/acorn-io/acorn-dns/pkg/apiserver"
	"github.com/acorn-io/acorn-dns/pkg/backend"
	"github.com/acorn-io/acorn-dns/pkg/db"
	"github.com/acorn-io/acorn-dns/pkg/version"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

type apiServerCommand struct{}

func (s *apiServerCommand) Execute(c *cli.Context) error {
	ctx := signals.SetupSignalHandler(context.Background())

	log := logrus.WithField("command", "api-server")

	log.Infof("version: %v", version.Get())

	database, err := db.New(ctx, c.String("sql-dialect"), c.String("sql-dsn"), &gorm.Config{
		Logger: db.NewLogger(c.String("log-level")),
	})
	if err != nil {
		return err
	}

	back, err := backend.NewBackend(database)
	if err != nil {
		return err
	}

	apiServer := apiserver.NewAPIServer(ctx, log, c.Int("port"))

	if err := apiServer.Start(back); err != nil {
		return err
	}

	return nil
}

func serverCommand() *cli.Command {
	cmd := apiServerCommand{}

	flags := []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Usage:   "Port for the HTTP Server Port",
			EnvVars: []string{"ACORN_PORT", "PORT"},
			Value:   4315,
		},
		&cli.StringFlag{
			Name:    "sql-dialect",
			Usage:   "The type of sql to use, sqlite or mysql",
			EnvVars: []string{"ACORN_SQL_DIALECT", "SQL_DIALECT"},
			Value:   "sqlite",
		},
		&cli.StringFlag{
			Name:    "sql-dsn",
			Usage:   "The DSN to use to connect to",
			EnvVars: []string{"ACORN_SQL_DSN", "SQL_DSN"},
			Value:   "file:acorn.sqlite?_pragma=foreign_keys(1)",
		},
	}

	return &cli.Command{
		Name:   "api-server",
		Usage:  "acorn api server",
		Action: cmd.Execute,
		Flags:  append(flags, GlobalFlags()...),
		Before: Before,
	}
}
