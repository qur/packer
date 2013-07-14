package kvm

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"path/filepath"
	"text/template"
)

type kvmTemplateData struct {
	Name     string
	DiskName string
	MemSize  uint
	ISOPath  string
}

// This step creates the KVM file for the VM.
//
// Uses:
//   config *config
//   iso_path string
//   ui     packer.Ui
//
// Produces:
//   kvm_path string - The path to the KVM file.
type stepCreateKvmScript struct{}

func (stepCreateKvmScript) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	isoPath := state["iso_path"].(string)
	ui := state["ui"].(packer.Ui)

	ui.Say("Building and writing KVM file")

	tplData := &kvmTemplateData{
		config.VMName,
		config.DiskName,
		config.MemSize,
		isoPath,
	}

	var buf bytes.Buffer
	t := template.Must(template.New("kvm").Parse(DefaultKVMTemplate))
	t.Execute(&buf, tplData)

	kvmData, err := ParseKvmScript(&buf)
	if err != nil {
		err := fmt.Errorf("Error parsing KVM file: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if config.KVMData != nil {
		log.Println("Setting custom KVM data...")
		for k, v := range config.KVMData {
			log.Printf("Setting KVM: '%s' = '%s'", k, v)
			kvmData.Add(k, v)
		}
	}

	kvmPath := filepath.Join(config.OutputDir, config.VMName+".sh")
	if err := WriteKvmScript(kvmPath, kvmData); err != nil {
		err := fmt.Errorf("Error creating KVM file: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["kvm_path"] = kvmPath

	return multistep.ActionContinue
}

func (stepCreateKvmScript) Cleanup(map[string]interface{}) {
}

// This is the default KVM template used if no other template is given.
// This is hardcoded here. If you wish to use a custom template please
// do so by specifying in the builder configuration.
const DefaultKVMTemplate = `#!/bin/sh

exec kvm \
  -S \
  -name "{{ .Name }}" \
  -net "nic,model=virtio" \
  -net "user,hostfwd=tcp::2222-:22" \
  -daemonize \
  -monitor "unix:monitor,server,nowait" \
  -qmp "unix:control,server,nowait" \
  -drive "file={{ .DiskName }}.img,if=virtio" \
  -cdrom "{{ .ISOPath }}"
  -m "{{ .MemSize }}" \

`
