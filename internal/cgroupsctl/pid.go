package cgroupsctl

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

func isPidDied(pid int) bool {
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return os.IsNotExist(err)
}

func pidFile(pid int, pidFilePath string) error {
	if pidFilePath == "" || pid <= 0 {
		return nil
	}

	err := ioutil.WriteFile(pidFilePath, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return err
	}

	return nil
}

func writePid(pid string, path string) error {
	return ioutil.WriteFile(path, []byte(pid), 0644)
}
