package main

import (
	"github.com/mitchellh/packer/builder/libvirt"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(libvirt.Builder))
}
