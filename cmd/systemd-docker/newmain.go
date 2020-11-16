package main

import (
	"context"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/etecs-ru/systemd-docker/v2/internal/cmd"
	"github.com/etecs-ru/systemd-docker/v2/internal/cmd/runcmd"
)

const (
	flagPIDFile     = "pid-file"
	flagNameLogs    = "logs"
	flagNameNotify  = "notify"
	flagNameEnv     = "env"
	flagNameCGroups = "cgroups"
)

var (
	Version    = "latest"
	CompiledAt = time.Now()
)

func parse(args []string) error {
	app := &cli.App{
		Name:     "systemd-docker",
		Version:  Version,
		Compiled: CompiledAt,
		Flags:    Options(),
		Commands: []*cli.Command{
			RunCommand(),
		},
		Usage:     "demonstrate available API",
		UsageText: "contrive - demonstrating the available API",
		ArgsUsage: "[args and such]",
	}
	app.EnableBashCompletion = true
	app.Setup()
	return app.RunContext(context.TODO(), args)
}
