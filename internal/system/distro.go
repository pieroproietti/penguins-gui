// Package system provides host package installation, bootloader detection, and GRUB operations for the GUI.
package system

import (
	"fmt"
	"os"
	"strings"
)

func distroFamily(release string) (string, error) {
	values := map[string]string{}
	for _, line := range strings.Split(release, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[key] = strings.Trim(value, "\"'")
		}
	}
	for _, id := range strings.Fields(values["ID"] + " " + values["ID_LIKE"]) {
		switch id {
		case "ubuntu":
			return "ubuntu", nil
		case "debian", "devuan":
			return "debian", nil
		case "arch", "manjaro":
			return "arch", nil
		case "fedora", "rhel", "centos":
			return "fedora", nil
		case "opensuse", "opensuse-leap", "opensuse-tumbleweed", "suse":
			return "suse", nil
		case "alpine":
			return "alpine", nil
		}
	}
	return "", fmt.Errorf("Package setup is not supported on distribution %q", values["ID"])
}

func repositoryFamily(release string) (string, error) {
	family, err := distroFamily(release)
	if err != nil {
		return "", err
	}
	// Preserve the distinct repository channels used by Eggs for these hosts.
	var id string
	for _, line := range strings.Split(release, "\n") {
		if value, ok := strings.CutPrefix(line, "ID="); ok {
			id = strings.Trim(value, "\"'")
		}
	}
	if id == "manjaro" {
		return "manjaro", nil
	}
	if family == "fedora" && id != "fedora" {
		return "el9", nil
	}
	return family, nil
}

func bootstrapCommand(script, action string) (string, []string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", nil, err
	}
	family, err := repositoryFamily(string(data))
	if err != nil {
		return "", nil, err
	}
	return "/bin/sh", []string{"-c", script, "penguins-gui-setup", family, action}, nil
}
