package cmd

import (
	"fmt"
	"os"
	"strings"
)

type ParsedArgs struct {
	ContainerName string
	IsRmCommand   bool
	Args          []string
}

// Parse translates arguments passed to `systemd-docker run` to `docker run` arguments
func Parse(args []string) ParsedArgs {
	var result ParsedArgs
	for i, arg := range args {
		switch arg {
		case "-rm", "--rm":
			result.IsRmCommand = true
			continue
		case "-d", "-detach", "--detach":
			result.Args = append([]string{"-d"}, result.Args...)
		default:
			if strings.HasPrefix(arg, "-name") || strings.HasPrefix(arg, "--name") {
				if strings.Contains(arg, "=") {
					result.ContainerName = strings.SplitN(arg, "=", 2)[1]
				} else if len(args) > i+1 {
					result.ContainerName = args[i+1]
				}
			}
		}
		result.Args = append(result.Args, arg)
	}
	return result
}

func SetupEnvironment(c *Context) {
	var newArgs []string
	if c.Notify && len(c.NotifySocket) > 0 {
		newArgs = append(newArgs, "-e", fmt.Sprintf("NOTIFY_SOCKET=%s", c.NotifySocket))
		newArgs = append(newArgs, "-v", fmt.Sprintf("%s:%s", c.NotifySocket, c.NotifySocket))
	} else {
		c.Notify = false
	}

	if c.Env {
		for _, val := range os.Environ() {
			if !strings.HasPrefix(val, "HOME=") && !strings.HasPrefix(val, "PATH=") {
				newArgs = append(newArgs, "-e", val)
			}
		}
	}

	if len(newArgs) > 0 {
		c.RunArgs.Args = append(newArgs, c.RunArgs.Args...)
	}
}
