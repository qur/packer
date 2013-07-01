package main

import (
	"github.com/mitchellh/packer/builder/kvm"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(kvm.Builder))
}
