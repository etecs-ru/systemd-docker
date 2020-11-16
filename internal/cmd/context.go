package cmd

import (
	"os/exec"

	dockerClient "github.com/fsouza/go-dockerclient"
)

type Context struct {
	CGroups      []string
	RunArgs      ParsedArgs
	Logs         bool
	Notify       bool
	Env          bool
	ID           string
	NotifySocket string
	Cmd          *exec.Cmd
	Pid          int
	PidFile      string
	Client       *dockerClient.Client
}
