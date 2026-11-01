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
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestResolveTargetUser_Current(t *testing.T) {
	t.Setenv("SUDO_USER", "")
	t.Setenv("PKEXEC_UID", "")

	id := ResolveTargetUser()
	if id.HomeDir == "" {
		t.Errorf("expected non-empty home directory")
	}
}

func TestResolveTargetUser_SudoUser(t *testing.T) {
	curr, err := user.Current()
	if err != nil {
		t.Skip("unable to get current user")
	}

	// Pretend current user is SUDO_USER
	t.Setenv("SUDO_USER", curr.Username)
	t.Setenv("SUDO_UID", curr.Uid)
	t.Setenv("SUDO_GID", curr.Gid)

	id := ResolveTargetUser()
	if id.Username != curr.Username {
		t.Errorf("expected username %s, got %s", curr.Username, id.Username)
	}
	if id.HomeDir != curr.HomeDir {
		t.Errorf("expected home %s, got %s", curr.HomeDir, id.HomeDir)
	}

	cfgDir := GetUserConfigDir()
	expectedCfgDir := filepath.Join(curr.HomeDir, ".config")
	if cfgDir != expectedCfgDir {
		t.Errorf("expected config dir %s, got %s", expectedCfgDir, cfgDir)
	}
}

func TestResolveTargetUser_SudoUserFallback(t *testing.T) {
	t.Setenv("SUDO_USER", "nonexistentuser999")
	t.Setenv("SUDO_UID", "1234")
	t.Setenv("SUDO_GID", "1234")

	id := ResolveTargetUser()
	if id.Username != "nonexistentuser999" {
		t.Errorf("expected username nonexistentuser999, got %s", id.Username)
	}
	if id.HomeDir != "/home/nonexistentuser999" {
		t.Errorf("expected fallback home /home/nonexistentuser999, got %s", id.HomeDir)
	}
	if id.UID != 1234 || id.GID != 1234 {
		t.Errorf("expected UID/GID 1234, got %d/%d", id.UID, id.GID)
	}

	cfgDir := GetUserConfigDir()
	expectedCfgDir := "/home/nonexistentuser999/.config"
	if cfgDir != expectedCfgDir {
		t.Errorf("expected config dir %s, got %s", expectedCfgDir, cfgDir)
	}
}

func TestSetupUserEnvironment(t *testing.T) {
	curr, err := user.Current()
	if err != nil {
		t.Skip("unable to get current user")
	}

	t.Setenv("SUDO_USER", curr.Username)
	SetupUserEnvironment()

	expectedCfgDir := filepath.Join(curr.HomeDir, ".config")
	if os.Getenv("XDG_CONFIG_HOME") != expectedCfgDir {
		t.Errorf("expected XDG_CONFIG_HOME to be %s, got %s", expectedCfgDir, os.Getenv("XDG_CONFIG_HOME"))
	}
}

func TestFixPathOwnership_NonRoot(t *testing.T) {
	// Should safely return nil when not root
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := FixPathOwnership(testFile); err != nil {
		t.Errorf("expected nil error on non-root, got %v", err)
	}
}
