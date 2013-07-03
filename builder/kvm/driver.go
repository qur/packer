package kvm

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string) error

	// Checks if the KVM file at the given path is running.
	IsRunning(*kvmControl) (bool, error)

	// Start starts a VM specified by the path to the KVM given.
	Start(string) (*kvmControl, error)

	// Stop stops a VM specified by the path to the KVM given.
	Stop(*kvmControl) error

	// Verify checks to make sure that this driver should function
	// properly. This should check that all the files it will use
	// appear to exist and so on. If everything is okay, this doesn't
	// return an error. Otherwise, this returns an error.
	Verify() error
}

// KvmDriver is a Driver that runs qemu directly.
type KvmDriver struct {
}

func (d *KvmDriver) CreateDisk(output string, size string) error {
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", output, size)
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *KvmDriver) IsRunning(control *kvmControl) (bool, error) {
	//TODO: ...
	r, err := control.Execute("query-status")
	if _, ok := err.(*notRunning); ok {
		return false, nil
	} else if err != nil {
		return false, err
	}

	running := r["running"].(bool)

	return running, nil
}

func (d *KvmDriver) Start(kvmPath string) (*kvmControl, error) {
	//TODO: ...
	cmd := exec.Command(kvmPath)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	control, err := NewKvmControl(kvmPath)
	if err != nil {
		return nil, err
	}

	_, err = control.Execute("cont")
	if err != nil {
		return nil, err
	}

	return control, nil
}

func (d *KvmDriver) Stop(control *kvmControl) error {
	//TODO: ...
	_, err := control.Execute("quit")
	if err != nil {
		return err
	}

	return nil
}

func (d *KvmDriver) Verify() error {
	//TODO: ...

	return nil
}

func (d *KvmDriver) runAndLog(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing: %s %v", cmd.Path, cmd.Args[1:])
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	log.Printf("stdout: %s", strings.TrimSpace(stdout.String()))
	log.Printf("stderr: %s", strings.TrimSpace(stderr.String()))

	return stdout.String(), stderr.String(), err
}
