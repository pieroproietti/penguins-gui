/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package system

import (
	_ "embed"
	"fmt"
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
