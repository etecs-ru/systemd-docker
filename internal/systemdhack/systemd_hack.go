package systemdhack

import "github.com/etecs-ru/systemd-docker/v2/internal/container"

func Do() {
	c, err := parseContext(args)
	if err != nil {
		return c, err
	}

	err = container.RunContainer(c)
	if err != nil {
		return c, err
	}

	_, err = moveCGroups(c)
	if err != nil {
		return c, err
	}

	err = notify(c)
	if err != nil {
		return c, err
	}

	err = pidFile(c)
	if err != nil {
		return c, err
	}

	go pipeLogs(c)

	err = keepAlive(c)
	if err != nil {
		return c, err
	}

	err = container.RemoveContainer(c)
	if err != nil {
		return c, err
	}

	return c, nil
}
