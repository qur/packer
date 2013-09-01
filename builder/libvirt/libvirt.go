package libvirt

import (
	"bytes"
	"log"
	"strings"
	"fmt"
	"os/exec"
)

var (
	virshCmd = ""
	qemuImgCmd = ""
)

func virsh(args ...string) (string, string, error) {
	if virshCmd == "" {
		cmd, err := exec.LookPath("virsh")
		if err != nil {
			return "", "", err
		}
		virshCmd = cmd
	}

	cmd := exec.Command(virshCmd, args...)

	return runAndLog(cmd)
}

func isRunning(name string) (bool, error) {
	output, _, err := virsh("domstate", name)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(output) == "running", nil
}

func qemuImg(args ...string) (string, string, error) {
	if virshCmd == "" {
		cmd, err := exec.LookPath("qemu-img")
		if err != nil {
			return "", "", err
		}
		qemuImgCmd = cmd
	}

	cmd := exec.Command(qemuImgCmd, args...)

	return runAndLog(cmd)
}

func runAndLog(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing: %s %v", cmd.Path, cmd.Args[1:])
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("VMware error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	// Replace these for Windows, we only want to deal with Unix
	// style line endings.
	returnStdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	returnStderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return returnStdout, returnStderr, err
}
