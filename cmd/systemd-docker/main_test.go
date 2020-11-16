package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"testing"

	dockerClient "github.com/fsouza/go-dockerclient"

	"github.com/etecs-ru/systemd-docker/v2/internal/cmd"
)

func init() {
	Interval = 100
}

func TestParseNoRun(t *testing.T) {
	_, err := parseContext([]string{"a", "b", "-d"})
	if err == nil {
		t.Fatal("parse succeeded")
	}
}

func TestParseCgroupsAll(t *testing.T) {
	c, err := parseContext([]string{"--cgroups", "all", "run"})
	if err != nil {
		t.Fatal("parse failed", err)
	}

	if !c.AllCGroups {
		t.Fatal("all cgroups should be true")
	}
}

func TestParseCgroupList(t *testing.T) {
	c, err := parseContext([]string{"--cgroups", "a", "--cgroups", "b", "run"})
	if err != nil {
		t.Fatal("parse failed", err)
	}

	if c.AllCGroups {
		t.Fatal("all cgroups should be false")
	}

	if c.CGroups[0] != "a" ||
		c.CGroups[1] != "b" {
		t.Fatal("Invalid cgroups value", c.CGroups)
	}
}

func TestParseNotify(t *testing.T) {
	c, err := parseContext([]string{"run"})
	if err != nil {
		t.Fatal("parse failed", err)
	}

	if c.Notify {
		t.Fatal("notify should be false")
	}

	c, err = parseContext([]string{"--notify", "run"})
	if err != nil {
		t.Fatal("parse failed", err)
	}

	if c.Notify {
		t.Fatal("notify should be false because NOTIFY_SOCKET is unset")
	}
}

func TestParseArgs(t *testing.T) {
	c, err := parseContext([]string{"--logs=false", "run", "c", "-rm", "d"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if c.Args[0] != "-d" ||
		c.Args[1] != "c" ||
		c.Args[2] != "d" {
		t.Fatal("Invalid args", c.Args)
	}
}

func TestParseEnv(t *testing.T) {
	c, err := parseContext([]string{"run", "--env", "-d"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if c.Env {
		t.Fatal("env shouldn't be set")
	}

	c, err = parseContext([]string{"--env", "run", "-d"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if !c.Env {
		t.Fatal("env should be set")
	}
}

func TestParseLogs(t *testing.T) {
	c, err := parseContext([]string{"run", "--logs", "false", "-d"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if !c.Logs {
		t.Fatal("logs should be set")
	}

	c, err = parseContext([]string{"--logs=false", "run", "-d"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if c.Logs {
		t.Fatal("logs shouldn't be set")
	}
}

func TestParseName(t *testing.T) {
	c, err := parseContext([]string{"run", "-d", "--logs", "--name=blah"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if c.Name != "blah" {
		t.Fatal("failed to parse name", c.Name)
	}
}

func TestParseName2(t *testing.T) {
	c, err := parseContext([]string{"run", "-d", "--logs", "--name", "blah"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if c.Name != "blah" {
		t.Fatal("failed to parse name", c.Name)
	}
}

func TestParseName3(t *testing.T) {
	c, err := parseContext([]string{"run", "-d", "--logs", "-name", "blah"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if c.Name != "blah" {
		t.Fatal("failed to parse name", c.Name)
	}
}

func TestParseName4(t *testing.T) {
	c, err := parseContext([]string{"run", "-d", "--logs", "-name"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if len(c.Name) != 0 {
		t.Fatal("failed to parse name", c.Name)
	}
}

func TestParseRm(t *testing.T) {
	c, err := parseContext([]string{"run", "-d", "--logs", "-name"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if c.Rm {
		t.Fatal("failed to parse rm", c.Rm)
	}
}

func TestParseRmSet(t *testing.T) {
	c, err := parseContext([]string{"run", "-d", "--logs", "-rm"})
	if err != nil {
		t.Fatal("failed to parse:", err)
	}

	if !c.Rm {
		t.Fatal("failed to parse rm", c.Rm)
	}
}

func TestParseCgroups(t *testing.T) {
	cgroups, err := getCGroupsForPid(os.Getpid())
	if err != nil {
		log.Fatal("Error:", err)
	}

	if val, ok := cgroups["blkio"]; ok {
		p := path.Join(SysFS, "blkio", val)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			log.Fatalf("path does not exist %s : %s", p, err)
		}
	} else {
		log.Fatal("Failed to find blkio cgroup", val)
	}
}

func TestMoveCgroup(t *testing.T) {
	c := &cmd.Context{
		Args: []string{"-d", "busybox", "echo", "hi"},
	}

	err := runContainer(c)

	if err != nil {
		t.Fatal("Exec should not have failed", err)
		return
	}

	if c.Cmd.ProcessState.Pid() <= 0 {
		t.Fatal("Bad pid", c.Cmd.ProcessState.Pid())
	}

	if c.Pid <= 0 {
		t.Fatal("Bad container pid", c.Pid)
	}

	moved, err := moveCGroups(c)
	if !moved || err != nil {
		t.Fatal("Failed to move namespaces ", moved, err)
	}
}

func TestRemoveNoLogs(t *testing.T) {
	c, err := mainWithArgs([]string{"--logs=false", "run", "-rm", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	client, err := getClient(c)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.InspectContainer(c.ID)
	if _, ok := err.(*dockerClient.NoSuchContainer); !ok {
		t.Fatal("Should have failed")
	}
}

func TestRemoveWithLogs(t *testing.T) {
	c, err := mainWithArgs([]string{"--logs", "run", "-rm", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	client, err := getClient(c)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.InspectContainer(c.ID)
	if _, ok := err.(*dockerClient.NoSuchContainer); !ok {
		t.Fatal("Should have failed")
	}
}

func deleteTestContainer(t *testing.T) {
	client, err := getClient(&cmd.Context{})
	if err != nil {
		t.Fatal(err)
	}

	container, err := client.InspectContainer("systemd-docker-test")
	if err == nil {
		log.Println("Deleting existing container", container.ID)
		err = client.RemoveContainer(dockerClient.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestNamedContainerNoRm(t *testing.T) {
	client, err := getClient(&cmd.Context{})
	if err != nil {
		t.Fatal(err)
	}

	deleteTestContainer(t)

	c, err := mainWithArgs([]string{"--logs", "run", "--privileged=true", "--name", "systemd-docker-test", "--privileged=true", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	container, err := client.InspectContainer("systemd-docker-test")
	if err != nil {
		t.Fatal(err)
	}

	if container.State.Running {
		t.Fatal("Should not be running")
	}

	c, err = mainWithArgs([]string{"--logs", "run", "--privileged=true", "--name", "systemd-docker-test", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	container2, err := client.InspectContainer(c.ID)
	if err != nil {
		t.Fatal(err)
	}

	if container2.State.Running {
		t.Fatal("Should not be running")
	}

	if container.ID != container2.ID {
		t.Fatal("Should be the same container", container.ID, container2.ID)
	}

	if !container2.HostConfig.Privileged {
		t.Fatal("Container2 is not privileged")
	}

	deleteTestContainer(t)
}

func TestNamedContainerRmPrevious(t *testing.T) {
	client, err := getClient(&cmd.Context{})
	if err != nil {
		t.Fatal(err)
	}

	deleteTestContainer(t)

	c, err := mainWithArgs([]string{"--logs", "run", "--name", "systemd-docker-test", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	container, err := client.InspectContainer("systemd-docker-test")
	if err != nil {
		t.Fatal(err)
	}

	if container.State.Running {
		t.Fatal("Should not be running")
	}

	c, err = mainWithArgs([]string{"--logs", "run", "--rm", "--name", "systemd-docker-test", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.InspectContainer(c.ID)
	if err == nil {
		t.Fatal("Should not exists")
	}

	if container.ID == c.ID {
		t.Fatal("Should not be the same container", container.ID, c.ID)
	}

	deleteTestContainer(t)
}

func TestNamedContainerAttach(t *testing.T) {
	client, err := getClient(&cmd.Context{})
	if err != nil {
		t.Fatal(err)
	}

	deleteTestContainer(t)

	c, err := mainWithArgs([]string{"--logs=false", "run", "--name", "systemd-docker-test", "busybox", "sleep", "2"})
	if err != nil {
		t.Fatal(err)
	}

	container, err := client.InspectContainer("systemd-docker-test")
	if err != nil {
		t.Fatal(err)
	}

	if !container.State.Running {
		t.Fatal("Should be running")
	}

	c, err = mainWithArgs([]string{"--logs=false", "run", "--name", "systemd-docker-test", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	container2, err := client.InspectContainer(c.ID)
	if err != nil {
		t.Fatal("Should exists", err)
	}

	if !container2.State.Running {
		t.Fatal("Should be running")
	}

	if container.ID != container2.ID {
		t.Fatal("Should not be the same container", container.ID, container2.ID)
	}

	deleteTestContainer(t)
}

func Exist(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

func TestPidFile(t *testing.T) {
	client, err := getClient(&cmd.Context{})
	if err != nil {
		t.Fatal(err)
	}

	pidFileName := "./pid-file"

	os.Remove(pidFileName)

	c, err := mainWithArgs([]string{"--logs=false", "--pid-file", "./pid-file", "run", "--rm", "busybox", "echo", "hi"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.InspectContainer(c.ID)
	if err == nil {
		t.Fatal("Container should not exist")
	}

	bytes, err := ioutil.ReadFile(pidFileName)
	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != strconv.Itoa(c.Pid) {
		t.Fatal("Failed to write pid file")
	}

	os.Remove(pidFileName)
}
