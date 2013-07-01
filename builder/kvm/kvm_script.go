package kvm

import (
	"bufio"
	"strings"
	"io"
	"bytes"
	"fmt"
	"os"
)

type KvmSettings map[string][]string

func (s KvmSettings) Add(name, value string) {
	s[name] = append(s[name], value)
}

func (s KvmSettings) Del(name string) {
	delete(s, name)
}

func (s KvmSettings) Enable(name string) {
	s[name] = []string{}
}

func (s KvmSettings) Disable(name string) {
	delete(s, name)
}

func (s KvmSettings) Set(name, value string) {
	s[name] = []string{value}
}

func (s KvmSettings) Get(name string) []string {
	v, _ := s[name]
	return v
}

func (s KvmSettings) GetOk(name string) ([]string, bool) {
	v, ok := s[name]
	return v, ok
}

func (s KvmSettings) GetSingle(name string) string {
	v, ok := s[name]
	if ok {
		return v[0]
	}
	return ""
}

func (s KvmSettings) GetSingleOk(name string) (string, bool) {
	v, ok := s[name]
	if ok {
		return v[0], true
	}
	return "", false
}

func ParseKvmScript(r io.Reader) (KvmSettings, error) {
	settings := make(KvmSettings)

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\\")
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] != '-' {
			continue
		}
		parts := strings.SplitN(line[1:], " ", 2)
		if len(parts) != 2 {
			settings.Enable(parts[0])
			continue
		}
		settings.Add(parts[0], strings.Trim(parts[1], "\""))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}

func EncodeKvmScript(settings map[string][]string) string {
	buf := &bytes.Buffer{}

	buf.WriteString("#!/bin/sh\n")
	buf.WriteString("\n")
	buf.WriteString("cd \"$(dirname \"$0\")\"\n")
	buf.WriteString("\n")
	buf.WriteString("exec kvm \\\n")

	for key, value := range settings {
		if len(value) == 0 {
			buf.WriteString(fmt.Sprintf("  -%s \\\n", key))
		} else {
			for _, line := range value {
				buf.WriteString(fmt.Sprintf("  -%s \"%s\" \\\n", key, line))
			}
		}
	}

	buf.WriteString("\n")

	return buf.String()
}

// WriteKvmScript takes a path to a Kvm script and contents in the form of a
// map and writes it out.
func WriteKvmScript(path string, data map[string][]string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(EncodeKvmScript(data))
	if err != nil {
		return err
	}

	return nil
}
