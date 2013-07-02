package kvm

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"math/rand"
	"net"
	"os"
)

// This step configures the VM to enable the VNC server.
//
// Uses:
//   config *config
//   ui     packer.Ui
//   kvm_path string
//
// Produces:
//   vnc_port uint - The port that VNC is configured to listen on.
type stepConfigureVNC struct{}

func (stepConfigureVNC) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)
	kvmPath := state["kvm_path"].(string)

	f, err := os.Open(kvmPath)
	if err != nil {
		err := fmt.Errorf("Error reading KVM data: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer f.Close()

	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	log.Printf("Looking for available port between %d and %d", config.VNCPortMin, config.VNCPortMax)
	var vncPort uint
	portRange := int(config.VNCPortMax - config.VNCPortMin)
	for {
		vncPort = uint(rand.Intn(portRange)) + config.VNCPortMin
		log.Printf("Trying port: %d", vncPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", vncPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	log.Printf("Found available VNC port: %d", vncPort)

	kvmData, err := ParseKvmScript(f)
	if err != nil {
		err := fmt.Errorf("Error reading KVM data: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	kvmData.Set("vnc", fmt.Sprintf(":%d,share=ignore", vncPort - 5900))
	kvmData.Add("usbdevice", "tablet")

	if err := WriteKvmScript(kvmPath, kvmData); err != nil {
		err := fmt.Errorf("Error writing KVM data: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["vnc_port"] = vncPort

	return multistep.ActionContinue
}

func (stepConfigureVNC) Cleanup(map[string]interface{}) {
}
