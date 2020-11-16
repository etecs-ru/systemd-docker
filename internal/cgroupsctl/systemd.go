package cgroupsctl

import (
	"errors"
	"fmt"
	"os"

	"github.com/coreos/go-systemd/v22/daemon"
)

func notify(pid int, forceNotify bool) error {
	if isPidDied(pid) {
		return ErrExitedBeforeNotifySystemD
	}
	// NOTIFY_SOCKET env variable is expected
	if err := sdNotify(false, fmt.Sprintf("MAINPID=%d", pid)); err != nil {
		return err
	}
	if isPidDied(pid) {
		_ = sdNotify(false, fmt.Sprintf("MAINPID=%d", os.Getpid()))
		return ErrExitedBeforeNotifySystemD
	}
	if forceNotify {
		if err := sdNotify(false, "READY=1"); err != nil {
			return err
		}
	}
	return nil
}

func sdNotify(unsetEnvironment bool, state string) error {
	ok, err := daemon.SdNotify(unsetEnvironment, state)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("failed to notify SystemD")
	}
	return nil
}
