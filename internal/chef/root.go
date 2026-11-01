/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

// Package chef integrates with the penguins-chef command-line interface and its recipes repository.
package chef

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pieroproietti/penguins-gui/internal/system"
)

// GetChefRoot returns ~/.chef for the calling user.
// If running under sudo, it resolves the home directory of the invoking user.
func GetChefRoot() (string, error) {
	homeDir := system.GetUserHomeDir()
	if homeDir == "" {
		return "", fmt.Errorf("unable to determine home directory")
	}
	return filepath.Join(homeDir, ".chef"), nil
}

// IsChefDirMissingOrEmpty checks if dir does not exist or has no files/recipes.
func IsChefDirMissingOrEmpty(dir string) bool {
	stat, err := os.Stat(dir)
	if err != nil || !stat.IsDir() {
		return true
	}

	// Check if recipes subdirectory exists and has entries
	recipesDir := filepath.Join(dir, "recipes")
	if rStat, err := os.Stat(recipesDir); err == nil && rStat.IsDir() {
		rEntries, err := os.ReadDir(recipesDir)
		if err == nil && len(rEntries) > 0 {
			return false
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return true
	}

	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			return false
		}
		if e.IsDir() && name != ".git" && name != ".github" {
			return false
		}
	}
	return true
}

func fixChefOwnership(root string) {
	_ = system.FixPathOwnership(root)
}
