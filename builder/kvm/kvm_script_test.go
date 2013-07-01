package kvm

import "testing"

import (
	"reflect"
	"strings"
)

func TestParseKvmScript(t *testing.T) {
	contents := `#!/bin/sh

qemu-kvm \
  -drive "file=foo.img,if=virtio" \
  -name "foo,process=foo.vm" \

`

	results, err := ParseKvmScript(strings.NewReader(contents))
	if err != nil {
		t.Fatalf("error parsing kvm script: %s", err)
	}
	if len(results) != 2 {
		t.Fatalf("not correct number of results: %d", len(results))
	}

	if !reflect.DeepEqual(results["drive"], []string{"file=foo.img,if=virtio"}) {
		t.Errorf("invalid drive: %s", results["drive"])
	}

	if !reflect.DeepEqual(results["name"], []string{"foo,process=foo.vm"}) {
		t.Errorf("invalid name: %s", results["name"])
	}
}

func TestEncodeKvmScript(t *testing.T) {
	contents := KvmSettings{
		"drive": []string{"file=foo.img,if=virtio"},
		"name":  []string{"foo,process=foo.vm"},
	}

	expected := `#!/bin/sh

exec kvm \
  -drive "file=foo.img,if=virtio" \
  -name "foo,process=foo.vm" \

`

	result := EncodeKvmScript(contents)
	if result != expected {
		t.Errorf("invalid results: %s", result)
	}
}
