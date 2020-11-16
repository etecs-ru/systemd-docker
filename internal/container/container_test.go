package container_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/etecs-ru/systemd-docker/v2/internal/cmd"
	"github.com/etecs-ru/systemd-docker/v2/internal/container"
)

var app *container.Container

func TestMain(m *testing.M) {
	c, err := container.New(container.DefaultOptions())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := c.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}()
	app = c
	m.Run()
}

func TestBadExec(t *testing.T) {
	args :=
		cmd.ParsedArgs{
			ContainerName: "",
			IsRmCommand:   false,
			Args:          []string{"-bad"},
		}

	err := app.Run(context.Background(), args)
	if err == nil {
		t.Fatal("Exec should have failed")
		return
	}

	if e, ok := err.(*exec.ExitError); ok {
		if status, ok := e.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() != 125 {
				t.Fatal("Expect 125 exit code got ", status.ExitStatus())
			}
		}
	} else {
		t.Fatal("Expect exec.ExitError", err)
	}
}

func TestGoodExec(t *testing.T) {
	args :=
		cmd.ParsedArgs{
			ContainerName: "",
			IsRmCommand:   false,
			Args:          []string{"-d", "busybox", "sleep", "60"},
		}

	pid, err := app.RunC(context.Background(), args)
	if err != nil {
		t.Fatal("Exec should not have failed", err)
		return
	}
	if pid <= 0 {
		t.Fatal("Bad pid", pid)
	}
}
