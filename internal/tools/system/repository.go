package system

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

//go:embed repository.sh
var repositoryScript string

//go:embed eggs.sh
var eggsInstallScript string

// repositorySetupScript keeps configuration and installation in one privileged
// operation. set -e stops before installation if repository configuration fails.
var repositorySetupScript = repositoryScript + "\nif [ \"$action\" = add ]; then\n" + eggsInstallScript + "\nfi\n"

// RepositoryCommand configures native repositories without requiring Eggs.
// Adding a repository also installs Eggs; removing one leaves packages installed.
func RepositoryCommand(action string) (string, []string, error) {
	if action != "add" && action != "rm" {
		return "", nil, fmt.Errorf("invalid repository action: %q", action)
	}
	return bootstrapCommand(repositorySetupScript, action)
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
