package commands

import (
	"fmt"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func GlobalFlags() []cli.Flag {
	globalFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "log-level",
			Usage:   "Log Level",
			Aliases: []string{"l"},
			EnvVars: []string{"LOGLEVEL"},
			Value:   "info",
		},
		&cli.BoolFlag{
			Name:  "log-caller",
			Usage: "log the caller (aka line number and file)",
		},
	}

	return globalFlags
}

func Before(c *cli.Context) error {
	formatter := &logrus.JSONFormatter{}

	if c.Bool("log-caller") {
		logrus.SetReportCaller(true)

		formatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		}
	}

	logrus.SetFormatter(formatter)

	switch c.String("log-level") {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	}

	return nil
}
