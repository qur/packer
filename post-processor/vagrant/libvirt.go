package vagrant

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"strings"
)

type LibVirtProvider struct{}

func (p *LibVirtProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Copy the disk image into the temporary directory (as box.img)
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".img") {
			ui.Message(fmt.Sprintf("Copying from artifact: %s", path))
			dstPath := filepath.Join(dir, "box.img")
			if err = CopySparseContents(dstPath, path); err != nil {
				return
			}
		}
	}

	format := artifact.State("diskType").(string)
	size := artifact.State("diskSize").(uint) / 1024 // In MB, want GB
	domainType := artifact.State("domainType").(string)

	// Convert domain type to libvirt driver
	var driver string
	switch domainType {
	case "kvm", "qemu":
		driver = "qemu"
	default:
		return "", nil, fmt.Errorf("Unknown libvirt domain type: %s", domainType)
	}

	// Create the metadata
	metadata = map[string]interface{}{
		"provider":     "libvirt",
		"format":       format,
		"virtual_size": size,
	}

	vagrantfile = fmt.Sprintf(libvirtVagrantfile, driver)
	return
}

var libvirtVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.provider :libvirt do |libvirt|
    libvirt.driver = "%s"
  end
end
`
