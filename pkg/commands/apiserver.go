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

	database, err := db.New(ctx,
		c.String("sql-dialect"),
		c.String("sql-dsn"),
		&gorm.Config{Logger: db.NewLogger(c.String("log-level"))})
	if err != nil {
		return err
	}

	back, err := backend.NewBackend(
		c.String("route53-zone-id"),
		c.Int64("route53-record-ttl-seconds"),
		c.Int64("purge-interval-seconds"),
		c.Int64("domain-max-age-seconds"),
		c.Int64("record-max-age-seconds"),
		database)
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
			Usage:   "HTTP Server Port",
			EnvVars: []string{"ACORN_DNS_PORT"},
			Value:   4315,
		},
		&cli.StringFlag{
			Name:     "route53-zone-id",
			Usage:    "AWS Route53 Zone ID where records will be created",
			EnvVars:  []string{"ACORN_ROUTE53_ZONE_ID"},
			Required: true,
		},
		&cli.Int64Flag{
			Name:    "route53-record-ttl-seconds",
			Usage:   "AWS Route53 record TTL",
			EnvVars: []string{"ACORN_ROUTE53_RECORD_TTL_SECONDS"},
			Value:   300,
		},
		&cli.Int64Flag{
			Name:    "purge-interval-seconds",
			Usage:   "How often to run the domain and record purge daemon. Default 86,400 (1 day)",
			EnvVars: []string{"ACORN_PURGE_INTERVAL_SECONDS"},
			Value:   86400,
		},
		&cli.Int64Flag{
			Name:    "domain-max-age-seconds",
			Usage:   "Max age a domain can be without being renewed before it's deleted. Default 2,592,000 (30 days)",
			EnvVars: []string{"ACORN_DOMAIN_MAX_AGE_SECONDS"},
			Value:   2592000,
		},
		&cli.Int64Flag{
			Name:    "record-max-age-seconds",
			Usage:   "Max age a domain can be without being renewed before it's deleted. Default 172,800 (2 days)",
			EnvVars: []string{"ACORN_RECORD_MAX_AGE_SECONDS"},
			Value:   172800,
		},
		&cli.StringFlag{
			Name:    "sql-dialect",
			Usage:   "The type of sql to use, sqlite or mysql",
			EnvVars: []string{"ACORN_SQL_DIALECT"},
			Value:   "sqlite",
		},
		&cli.StringFlag{
			Name:    "sql-dsn",
			Usage:   "The DSN to use to connect to",
			EnvVars: []string{"ACORN_SQL_DSN"},
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
