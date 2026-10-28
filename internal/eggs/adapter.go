/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

// Package eggs integrates with the existing Eggs command-line interface.
package eggs

import (
	"os/exec"
	"strings"
)

// CLIAdapter preserves the textual CLI integration, including interactive
// encrypted remasters and filesystem-based ISO discovery. Its zero value is ready
// to use and it has no dependency on the GUI.
type CLIAdapter struct{}

// Detect locates Eggs and returns its textual version, with the existing fallback.
func (CLIAdapter) Detect() (path, version string, err error) {
	path, err = exec.LookPath("eggs")
	if err != nil {
		return "", "", err
	}
	out, versionErr := exec.Command(path, "version").CombinedOutput()
	version = strings.TrimSpace(string(out))
	if version == "" {
		version = "Penguins’ Eggs found: " + path
	}
	if versionErr != nil {
		version += " (version undetermined)"
	}
	return path, version, nil
}
