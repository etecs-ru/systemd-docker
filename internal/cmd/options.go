package cmd

import "github.com/urfave/cli/v2"

const (
	flagPIDFile     = "pid-file"
	flagNameLogs    = "logs"
	flagNameNotify  = "notify"
	flagNameEnv     = "env"
	flagNameCGroups = "cgroups"
)

func Options() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flagPIDFile,
			Aliases: []string{"p"},
			Value:   "",
			Usage:   "Pipe file",
		},
		&cli.BoolFlag{
			Name:    flagNameLogs,
			Aliases: []string{"l"},
			Value:   true,
			Usage:   "Pipe logs",
		},
		&cli.BoolFlag{
			Name:    flagNameNotify,
			Aliases: []string{"n"},
			Usage:   "Setup SystemD notify for container",
		},
		&cli.BoolFlag{
			Name:    flagNameEnv,
			Aliases: []string{"e"},
			Usage:   "Inherit environment variables",
		},
		&cli.StringSliceFlag{
			Name:    flagNameCGroups,
			Aliases: []string{"c"},
			Usage:   "cgroups to take ownership of or 'all' for all cgroups available",
		},
	}
}
