package system

import (
	_ "embed"
	"os"
)

//go:embed chef.sh
var chefInstallScript string

// ChefCommand returns a fixed installer script for penguins-chef and a validated distro family.
// The caller must execute it through the normal administrative authorization flow.
func ChefCommand() (string, []string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", nil, err
	}
	family, err := repositoryFamily(string(data))
	if err != nil {
		return "", nil, err
	}
	return "/bin/sh", []string{"-c", chefInstallScript, "install-chef", family}, nil
}
