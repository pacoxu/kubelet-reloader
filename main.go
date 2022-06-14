/*
Copyright 2022 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/fsnotify.v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	kubeletBinPath           = "/usr/bin/kubelet"
	defaultKubeletNewBinPath = "/usr/bin/kubelet-new"
)

var (
	watchedFiles = []string{
		// kubeletBinPath,
		getKubeletNewBinPath(),
	}
)

func getKubeletNewBinPath() string {
	kubeletNewBinPath := os.Getenv("NEW_KUBELET_PATH")
	if kubeletNewBinPath == "" {
		return defaultKubeletNewBinPath
	}
	return kubeletNewBinPath
}

type cmd struct {
	command string
	args    []string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

func newCmd(command string, args ...string) *cmd {
	return &cmd{
		command: command,
		args:    args,
	}
}

func (c *cmd) Run() error {
	return c.runInnnerCommand()
}

func (c *cmd) RunWithEcho() error {
	c.stdout = os.Stderr
	c.stderr = os.Stdout
	return c.runInnnerCommand()
}

func (c *cmd) RunAndCapture() (lines []string, err error) {
	var buff bytes.Buffer
	c.stdout = &buff
	c.stderr = &buff
	err = c.runInnnerCommand()

	scanner := bufio.NewScanner(&buff)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())

	}
	return lines, err
}

func (c *cmd) Stdin(in io.Reader) *cmd {
	c.stdin = in
	return c
}

func (c *cmd) runInnnerCommand() error {
	cmd := exec.Command(c.command, c.args...)

	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}
	if c.stdout != nil {
		cmd.Stdout = c.stdout
	}
	if c.stderr != nil {
		cmd.Stderr = c.stderr
	}

	return cmd.Run()
}

func getKubeletVersion(filepath string) (string, error) {
	cmd := newCmd(filepath, "--version")

	lines, err := cmd.RunAndCapture()
	if err != nil {
		return "", errors.WithStack(errors.WithMessage(err, strings.Join(lines, "\n")))
	}
	return lines[0], nil
}

func isChanged() bool {
	kubeletVersion, err := getKubeletVersion(kubeletBinPath)
	if err != nil {
		fmt.Printf("Kubelet binary get version Error : %v \n", err)
		return false
	}
	fmt.Printf("Kubelet binary version: %s \n", kubeletVersion)
	kubeletNewVersion, err := getKubeletVersion(getKubeletNewBinPath())
	if err != nil {
		fmt.Printf("Kubelet new binary get version Error : %v \n", err)
		return false
	}
	fmt.Printf("Kubelet new binary version: %s \n", kubeletNewVersion)
	return kubeletVersion != kubeletNewVersion
}

func replaceKubelet() error {
	fmt.Printf("Replace kubelet with kubelet-new \n")

	cmd := exec.Command("/usr/bin/cp", "-f", getKubeletNewBinPath(), kubeletBinPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Cp replace Error : %v \n", err)
	}
	return err
}

func stopKubelet() error {
	_, err := exec.Command("systemctl", "stop", "kubelet").Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: systemctl stop kubelet")
	}
	return err
}

func restartKubelet() error {
	_, err := exec.Command("systemctl", "restart", "kubelet").Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: systemctl restart kubelet")
	}
	return err
}

func main() {
	fmt.Println("kubelet reloader start")
	defer fmt.Println("kubelet reloader ended")

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Kubelet binary watcher new Error : %v \n", err)
		os.Exit(1)
	}

	for _, watchFile := range watchedFiles {
		if err = fswatcher.Add(watchFile); err != nil {
			fmt.Printf("Kubelet binary watch %s Error : %v \n", watchFile, err)
			os.Exit(1)
		}
	}

	newKubeletBin := make(chan bool)
	// caluate md5
	go func() {
		for ; ; time.Sleep(5 * time.Minute) {
			if isChanged() {
				newKubeletBin <- true
			}
		}
	}()

	for {
		select {
		case flag := <-newKubeletBin:
			if flag {
				wait.Poll(1*time.Second, 60*time.Second, func() (bool, error) {
					return kubeletReplaceAndRestart()
				})
			}
		case event := <-fswatcher.Events:
			if fsnotify.Create == event.Op || fsnotify.Write == event.Op || fsnotify.Remove == event.Op || fsnotify.Rename == event.Op {
				if isChanged() {
					wait.Poll(1*time.Second, 60*time.Second, func() (bool, error) {
						return kubeletReplaceAndRestart()
					})
				}
			}
		case err := <-fswatcher.Errors:
			fmt.Printf("Kubelet binary watcher error : %v \n", err)
			os.Exit(1)
		}
	}
}

func kubeletReplaceAndRestart() (bool, error) {
	waitErr := stopKubelet()
	if waitErr != nil {
		return false, waitErr
	}
	waitErr = replaceKubelet()
	if waitErr != nil {
		return false, waitErr
	}
	waitErr = restartKubelet()
	if waitErr != nil {
		return false, waitErr
	}
	return true, nil
}
