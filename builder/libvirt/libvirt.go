package libvirt

import (
	"bufio"
	"bytes"
	"log"
	"strings"
	"fmt"
	"os"
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

func getMac(vmName string) (string, error) {
	xml, _, err := virsh("dumpxml", vmName)
	if err != nil {
		return "", err
	}

	var mac string

	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(xml)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, "mac address") {
			continue
		}
		parts := strings.Split(line, "'")
		mac = parts[1]
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if mac == "" {
		return "", fmt.Errorf("MAC not found for VM '%s'", mac)
	}

	return mac, nil
}

func lookupIp(vmName, mac string) (string, error) {
	leasesPath := "/var/lib/libvirt/dnsmasq/" + vmName + ".leases"
	f, err := os.Open(leasesPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var ip string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if strings.ToLower(parts[1]) == strings.ToLower(mac) {
			ip = parts[2]
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if ip == "" {
		return "", fmt.Errorf("Unknown MAC '%s'", mac)
	}

	return ip, nil
}
