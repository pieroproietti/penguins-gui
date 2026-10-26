package system

import (
	_ "embed"
	"os"
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
