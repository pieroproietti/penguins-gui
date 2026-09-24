// Package system provides host package installation operations for the GUI.
package system

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed calamares.sh
var calamaresScript string

// CalamaresCommand returns a fixed installer script and a validated distro family.
// The caller must execute it through the normal administrative authorization flow.
func CalamaresCommand() (string, []string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", nil, err
	}
	family, err := distroFamily(string(data))
	if err != nil {
		return "", nil, err
	}
	return "/bin/sh", []string{"-c", calamaresScript, "install-calamares", family}, nil
}

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
