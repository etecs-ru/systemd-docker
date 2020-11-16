package cgroupsctl

// import "github.com/containers/libpod/pkg/cgroups"

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/etecs-ru/systemd-docker/v2/internal/contains"
	"github.com/etecs-ru/systemd-docker/v2/internal/zaplog"
)

var (
	ErrExitedBeforeNotifySystemD = errors.New("container exited before an attempt to notify SystemD has been made")
)

const (
	pathSysFS      = "/sys/fs/cgroup"
	pathCGroupProc = "/proc/%d/cgroup"
	pathProcs      = "cgroup.procs"
)

func GetCGroupsForPid(pid int) (map[string]string, error) {
	p := fmt.Sprintf(pathCGroupProc, pid)
	return parseCgroupFile(p)
}

// parseCgroupFile parses /proc/PID/cgroup file and return string
func parseCgroupFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseCGroupsFromReader(f)
}

func parseCGroupsFromReader(r io.Reader) (map[string]string, error) {
	var (
		s = bufio.NewScanner(r)
	)
	cGroups := make(map[string]string)
	for s.Scan() {
		var (
			text  = s.Text()
			parts = strings.SplitN(text, ":", 3)
		)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid cgroup entry: %q", text)
		}
		cGroups[parts[1]] = parts[2]
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return cGroups, nil
}

func getCgroupPids(cgroupName string, cgroupPath string) ([]string, error) {
	var ret []string

	file, err := os.Open(constructCgroupPath(cgroupName, cgroupPath))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ret = append(ret, strings.TrimSpace(scanner.Text()))
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func constructCgroupPath(cgroupName string, cgroupPath string) string {
	return path.Join(pathSysFS, strings.TrimPrefix(cgroupName, "name="), cgroupPath, pathProcs)
}

func MoveCGroups(ctx context.Context, pid int, cGroups []string) (bool, error) {
	moved := false
	currentCGroups, err := GetCGroupsForPid(os.Getpid())
	if err != nil {
		return false, err
	}

	containerCGroups, err := GetCGroupsForPid(pid)
	if err != nil {
		return false, err
	}

	var ns []string

	if len(cGroups) == 0 || contains.String(cGroups, "all") {
		ns = make([]string, 0, len(containerCGroups))
		for value := range containerCGroups {
			ns = append(ns, value)
		}
	} else {
		ns = cGroups
	}

	for _, nsName := range ns {
		currentPath, ok := currentCGroups[nsName]
		if !ok {
			continue
		}

		containerPath, ok := containerCGroups[nsName]
		if !ok {
			continue
		}

		if currentPath == containerPath || containerPath == "/" {
			continue
		}

		pids, err := getCgroupPids(nsName, containerPath)
		if err != nil {
			return false, err
		}

		for _, pid := range pids {
			pidInt, err := strconv.Atoi(pid)
			if err != nil {
				continue
			}

			if isPidDied(pidInt) {
				continue
			}

			currentFullPath := constructCgroupPath(nsName, currentPath)
			zaplog.Extract(ctx).Info("Moving pid",
				zap.String("pid", pid), zap.String("target_path", currentFullPath))
			err = writePid(pid, currentFullPath)
			if err != nil {
				return false, err
			}

			moved = true
		}
	}

	return moved, nil
}
