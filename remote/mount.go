package remote

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

func mount(source, target, fstype, options string) error {
	mounted, err := checkMount(target)
	if err != nil {
		return err
	}
	if mounted {
		return nil
	}
	timeout := 30 * time.Second

	info, err := os.Stat(target)
	if err != nil {
		err = os.MkdirAll(target, 0777)
		if err != nil {
			return err
		}
	} else {
		if !info.IsDir() {
			return errors.New("target should be a directory and is a file")
		}
	}

	doneCmd := make(chan struct{})
	// sudo mount -t ...
	mountCmd := "sudo"
	args := []string{"mount"}
	if len(fstype) > 0 {
		args = append(args, "-t", fstype)
	}
	if len(options) > 0 {
		args = append(args, "-o", options)
	}
	args = append(args, source, target)
	cmd := exec.Command(mountCmd, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	// killer after the timeout
	go func() {
		select {
		case <-time.After(timeout):
			cmd.Process.Kill()
		case <-doneCmd:
			return
		}
	}()

	errBuf, _ := ioutil.ReadAll(stderr)
	err = cmd.Wait()
	close(doneCmd)
	if err != nil {
		errBufStr := ""
		if len(errBuf) > 0 {
			errBufStr = ": " + string(errBuf)
		}
		err = fmt.Errorf("command '%s %s' failed: %s%s", mountCmd, strings.Join(args, " "), err, errBufStr)
		return err
	}
	return nil
}

func checkMount(target string) (bool, error) {
	timeout := 10 * time.Second
	mountCmd, err := exec.LookPath("mount")
	if err != nil {
		return false, errors.New("mount not installed")
	}
	doneCmd := make(chan struct{})
	args := []string{}
	cmd := exec.Command(mountCmd, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return false, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}

	err = cmd.Start()
	if err != nil {
		return false, err
	}
	// killer after the timeout
	go func() {
		select {
		case <-time.After(timeout):
			cmd.Process.Kill()
		case <-doneCmd:
			return
		}
	}()

	s := bufio.NewScanner(stdout)
	found := false
	for s.Scan() {
		if strings.Contains(s.Text(), target) {
			found = true
		}
	}
	errBuf, _ := ioutil.ReadAll(stderr)
	err = cmd.Wait()
	close(doneCmd)
	if err != nil {
		errBufStr := ""
		if len(errBuf) > 0 {
			errBufStr = ": " + string(errBuf)
		}
		err = fmt.Errorf("command 'mount %s' failed: %s%s", strings.Join(args, " "), err, errBufStr)
		return false, err
	}
	return found, nil
}
