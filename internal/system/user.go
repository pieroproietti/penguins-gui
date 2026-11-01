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
	"bufio"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// UserIdentity represents a resolved user identity and home directory.
type UserIdentity struct {
	Username string
	HomeDir  string
	UID      int
	GID      int
}

// ResolveTargetUser finds the invoking or primary human user.
// When running under sudo or with root privileges, it resolves to the real user (SUDO_USER, etc.).
func ResolveTargetUser() UserIdentity {
	// 1. SUDO_USER: explicitly set when invoked via sudo
	if sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER")); sudoUser != "" && sudoUser != "root" {
		if id, ok := lookupUserIdentity(sudoUser); ok {
			return id
		}
		// Fallback for minimal environments without /etc/passwd entries
		home := filepath.Join("/home", sudoUser)
		uid, gid := -1, -1
		if envUID, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SUDO_UID"))); err == nil && envUID > 0 {
			uid = envUID
		}
		if envGID, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SUDO_GID"))); err == nil && envGID >= 0 {
			gid = envGID
		}
		return UserIdentity{
			Username: sudoUser,
			HomeDir:  home,
			UID:      uid,
			GID:      gid,
		}
	}

	// 2. PKEXEC_UID: set when invoked via pkexec
	if pkUIDStr := strings.TrimSpace(os.Getenv("PKEXEC_UID")); pkUIDStr != "" && pkUIDStr != "0" {
		if u, err := user.LookupId(pkUIDStr); err == nil {
			if id, ok := identityFromUser(u); ok {
				return id
			}
		}
	}

	// 3. If running as root, try kernel login name (logname)
	if os.Geteuid() == 0 {
		if out, err := exec.Command("logname").Output(); err == nil {
			login := strings.TrimSpace(string(out))
			if login != "" && login != "root" {
				if id, ok := lookupUserIdentity(login); ok {
					return id
				}
			}
		}

		// 4. If running as root, scan /etc/passwd for first human user (UID 1000..59999)
		if u := firstHumanUser(); u != nil {
			if id, ok := identityFromUser(u); ok {
				return id
			}
		}
	}

	// 5. Current process user
	if curr, err := user.Current(); err == nil {
		if id, ok := identityFromUser(curr); ok {
			return id
		}
	}

	// 6. Fallback from os.UserHomeDir()
	home, _ := os.UserHomeDir()
	return UserIdentity{
		Username: os.Getenv("USER"),
		HomeDir:  home,
		UID:      os.Getuid(),
		GID:      os.Getgid(),
	}
}

func lookupUserIdentity(username string) (UserIdentity, bool) {
	u, err := user.Lookup(username)
	if err != nil {
		return UserIdentity{}, false
	}
	return identityFromUser(u)
}

func identityFromUser(u *user.User) (UserIdentity, bool) {
	if u == nil {
		return UserIdentity{}, false
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	return UserIdentity{
		Username: u.Username,
		HomeDir:  u.HomeDir,
		UID:      uid,
		GID:      gid,
	}, true
}

// firstHumanUser scans /etc/passwd for the first non-system user (UID 1000..59999).
func firstHumanUser() *user.User {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 7 {
			continue
		}
		uid, err := strconv.Atoi(fields[2])
		if err != nil || uid < 1000 || uid >= 60000 {
			continue
		}
		shell := fields[6]
		if strings.HasSuffix(shell, "nologin") || strings.HasSuffix(shell, "/false") {
			continue
		}
		if u, err := user.Lookup(fields[0]); err == nil {
			return u
		}
	}
	return nil
}

// GetUserHomeDir returns the home directory of the invoking user.
// If running under sudo or elevated, returns SUDO_USER's home directory.
func GetUserHomeDir() string {
	id := ResolveTargetUser()
	if id.HomeDir != "" {
		return id.HomeDir
	}
	home, _ := os.UserHomeDir()
	return home
}

// GetUserConfigDir returns the configuration directory (~/.config) for the invoking user.
// If running under sudo (or elevated as root), it resolves to the real user's ~/.config.
func GetUserConfigDir() string {
	// If SUDO_USER is set
	if sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER")); sudoUser != "" && sudoUser != "root" {
		id := ResolveTargetUser()
		if id.HomeDir != "" {
			if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" && strings.HasPrefix(xdg, id.HomeDir) {
				return xdg
			}
			return filepath.Join(id.HomeDir, ".config")
		}
	}

	// If running as root without SUDO_USER
	if os.Geteuid() == 0 {
		id := ResolveTargetUser()
		if id.Username != "" && id.Username != "root" && id.HomeDir != "" {
			if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" && !strings.HasPrefix(xdg, "/root") {
				return xdg
			}
			return filepath.Join(id.HomeDir, ".config")
		}
	}

	// Standard non-privileged user
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return xdg
	}
	if cfgDir, err := os.UserConfigDir(); err == nil && cfgDir != "" {
		return cfgDir
	}
	home := GetUserHomeDir()
	if home != "" {
		return filepath.Join(home, ".config")
	}
	return "/root/.config"
}

// FixPathOwnership recursively changes ownership of targetPath to the invoking user
// if running with root privileges (os.Geteuid() == 0).
func FixPathOwnership(targetPath string) error {
	if os.Geteuid() != 0 {
		return nil
	}
	id := ResolveTargetUser()
	if id.Username == "" || id.Username == "root" {
		return nil
	}

	// Ensure ~/.config is also user-owned if targetPath is inside ~/.config
	if id.HomeDir != "" {
		userConfig := filepath.Join(id.HomeDir, ".config")
		if strings.HasPrefix(targetPath, userConfig) {
			if id.UID >= 0 && id.GID >= 0 {
				_ = os.Chown(userConfig, id.UID, id.GID)
			} else {
				_ = exec.Command("chown", id.Username+":"+id.Username, userConfig).Run()
			}
		}
	}

	if id.UID >= 0 && id.GID >= 0 {
		return filepath.Walk(targetPath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			_ = os.Chown(p, id.UID, id.GID)
			return nil
		})
	}

	return exec.Command("chown", "-R", id.Username+":"+id.Username, targetPath).Run()
}

// SetupUserEnvironment configures environment variables so that desktop libraries
// (like Fyne) and configuration lookups target the real user (SUDO_USER) when running elevated.
func SetupUserEnvironment() {
	if os.Geteuid() == 0 || strings.TrimSpace(os.Getenv("SUDO_USER")) != "" {
		cfgDir := GetUserConfigDir()
		if cfgDir != "" && !strings.HasPrefix(cfgDir, "/root") {
			_ = os.Setenv("XDG_CONFIG_HOME", cfgDir)
			if os.Geteuid() == 0 {
				if err := os.MkdirAll(cfgDir, 0755); err == nil {
					_ = FixPathOwnership(cfgDir)
				}
			}
		}
	}
}
