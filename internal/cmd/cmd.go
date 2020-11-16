package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

func RunCommand() *cli.Command {
	return &cli.Command{
		Name:         "run",
		Usage:        "",
		UsageText:    "",
		Description:  "A wrapper for docker run so that you can sanely run Docker containers under systemd",
		ArgsUsage:    "",
		Category:     "",
		BashComplete: nil,
		Before: func(c *cli.Context) error {
			cmdCtx := &Context{
				RunArgs:      Parse(c.Args().Slice()),
				CGroups:      c.StringSlice(flagNameCGroups),
				Logs:         c.Bool(flagNameLogs),
				Env:          c.Bool(flagNameEnv),
				ID:           "",
				NotifySocket: os.Getenv("NOTIFY_SOCKET"),
				Cmd:          nil,
				Pid:          0,
				PidFile:      "",
				Client:       nil,
			}

			SetupEnvironment(cmdCtx)
			return nil
		},
		After: func(c *cli.Context) error {
			return nil
		},
		Action: func(c *cli.Context) error {
			return nil
		},
		OnUsageError:           nil,
		Subcommands:            nil,
		Flags:                  nil,
		SkipFlagParsing:        true,
		HideHelp:               true,
		HideHelpCommand:        true,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "",
		CustomHelpTemplate:     "",
	}
}
