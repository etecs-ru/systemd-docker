package container

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"

	"github.com/etecs-ru/systemd-docker/v2/internal/cmd"
	"github.com/etecs-ru/systemd-docker/v2/internal/zaplog"
)

const (
	apiVersion = "1.40"
)

const (
	ExecutorDocker = "docker"
	ExecutorPodman = "podman"
)

type Identifier struct {
	ID  string
	PID int
}

type Container struct {
	log    *zap.Logger
	client *client.Client
	opts   Options
}

type Options struct {
	Executor              string
	APIVersion            string
	APIVersionNegotiation bool
}

func DefaultOptions() Options {
	return Options{
		Executor:              ExecutorDocker,
		APIVersion:            apiVersion,
		APIVersionNegotiation: false,
	}
}

func New(opts Options) (*Container, error) {
	dockerOpts := []client.Opt{client.FromEnv}
	if opts.APIVersion != "" {
		dockerOpts = append(dockerOpts, client.WithVersion(opts.APIVersion))
	}
	if opts.APIVersionNegotiation {
		dockerOpts = append(dockerOpts, client.WithAPIVersionNegotiation())
	}
	cli, err := client.NewClientWithOpts(dockerOpts...)
	if err != nil {
		return nil, err
	}
	return &Container{
		client: cli,
		opts:   opts,
	}, nil
}

func (c *Container) getPID(ctx context.Context, id string) (int, error) {
	resp, err := c.client.ContainerInspect(ctx, id)
	if err != nil {
		return 0, err
	}
	pid := resp.State.Pid

	if pid <= 0 {
		return 0, fmt.Errorf("pid is %d for container %s", pid, id)
	}
	return pid, nil
}

func (c *Container) launchContainer(ctx context.Context, args []string) (string, error) {
	cmdLauncher := exec.CommandContext(ctx, c.opts.Executor, append([]string{"run"}, args...)...)

	errorPipe, err := cmdLauncher.StderrPipe()
	if err != nil {
		return "", err
	}

	outputPipe, err := cmdLauncher.StdoutPipe()
	if err != nil {
		return "", err
	}

	err = cmdLauncher.Start()
	if err != nil {
		return "", err
	}

	go func() {
		if _, err := io.Copy(os.Stderr, errorPipe); err != nil {
			fmt.Println("error", err)

		}
	}()

	bytes, err := ioutil.ReadAll(outputPipe)
	if err != nil {
		return "", err
	}

	id := strings.TrimSpace(string(bytes))

	err = cmdLauncher.Wait()
	if err != nil {
		return "", err
	}

	if !cmdLauncher.ProcessState.Success() {
		return "", err
	}
	return id, nil
}

func (c *Container) Run(ctx context.Context, args cmd.ParsedArgs) error {
	_, err := c.RunC(ctx, args)
	return err
}

func (c *Container) RunC(ctx context.Context, args cmd.ParsedArgs) (int, error) {
	idTuple, err := c.lookupNamedContainer(ctx, args.ContainerName, args.IsRmCommand)
	if err != nil {
		return 0, err
	}
	if idTuple.ID == "" {
		id, err := c.launchContainer(ctx, args.Args)
		if err != nil {
			return 0, err
		}
		idTuple.ID = id
	}
	pid, err := c.getPID(ctx, idTuple.ID)
	if err != nil {
		return 0, err
	}
	if pid == 0 {
		return 0, errors.New("failed to launch container, pid is 0")
	}
	return pid, nil
}

func (c *Container) lookupNamedContainer(ctx context.Context, name string, hasRemoveCmd bool) (Identifier, error) {
	if name == "" {
		return Identifier{}, nil
	}
	resp, err := c.client.ContainerInspect(ctx, name)
	if err != nil {
		if client.IsErrNotFound(err) {
			return Identifier{}, nil
		}
		return Identifier{}, err
	}
	switch {
	case resp.State.Running:
		return Identifier{
			ID:  resp.ID,
			PID: resp.State.Pid,
		}, nil
	case hasRemoveCmd:
		return Identifier{}, c.client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
			RemoveVolumes: false,
			RemoveLinks:   false,
			Force:         true,
		})
	}
	if err := c.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return Identifier{}, err
	}
	inspectOpts, err := c.client.ContainerInspect(ctx, name)
	if err != nil {
		return Identifier{}, err
	}
	return Identifier{
		ID:  inspectOpts.ID,
		PID: inspectOpts.State.Pid,
	}, nil
}

// if c.Logs || c.RunArgs.IsRmCommand {
func (c *Container) KeepAlive(ctx context.Context, id string) error {
	resultC, errC := c.client.ContainerWait(ctx, id, dockerContainer.WaitConditionNotRunning)
	select {
	case result := <-resultC:
		zaplog.Extract(ctx).Warn("container exited with non-zero status",
			zap.String("container_id", id), zap.Int64("status_code", result.StatusCode))
		return nil
	case err := <-errC:
		return err
	}
}

func (c *Container) PipeLogs(ctx context.Context, containerID string) error {
	logs, err := c.client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      "",
		Until:      "",
		Timestamps: true,
		Follow:     false,
		Tail:       "",
		Details:    false,
	})
	if err != nil {
		return err
	}
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	return err
}

//	if !c.RunArgs.IsRmCommand {
//		return nil
//	}
func (c *Container) RemoveContainer(ctx context.Context, id string) error {
	return c.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	})
}

func (c *Container) Close() error {
	return c.client.Close()
}
