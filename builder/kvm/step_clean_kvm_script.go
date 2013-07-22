package kvm

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"os"
)

// This step cleans up the KVM Script by removing or changing this prior to
// being ready for use.
//
// Uses:
//   kvm_path string
//
// Produces:
//   <nothing>
type stepCleanKvmScript struct{}

func (s stepCleanKvmScript) Run(state map[string]interface{}) multistep.StepAction {
	kvmPath := state["kvm_path"].(string)

	kvmData, err := s.readKvmScript(kvmPath)
	if err != nil {
		state["error"] = fmt.Errorf("Error reading KVM Script: %s", err)
		return multistep.ActionHalt
	}

	delete(kvmData, "S")
	delete(kvmData, "cdrom")

	// Rewrite the KvmScript
	if err := WriteKvmScript(kvmPath, kvmData); err != nil {
		state["error"] = fmt.Errorf("Error writing KvmScript: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCleanKvmScript) Cleanup(map[string]interface{}) {}

func (stepCleanKvmScript) readKvmScript(kvmPath string) (KvmSettings, error) {
	kvmF, err := os.Open(kvmPath)
	if err != nil {
		return nil, err
	}
	defer kvmF.Close()

	return ParseKvmScript(kvmF)
}
