// Package chef integrates with the penguins-chef command-line interface and its recipes repository.
package chef

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// GetChefRoot returns ~/.chef for the calling user.
// If running under sudo, it resolves the home directory of the invoking user.
func GetChefRoot() (string, error) {
	var homeDir string

	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		if u, err := user.Lookup(sudoUser); err == nil {
			homeDir = u.HomeDir
		}
	}

	if homeDir == "" {
		if out, err := exec.Command("logname").Output(); err == nil {
			login := strings.TrimSpace(string(out))
			if login != "" && login != "root" {
				if u, err := user.Lookup(login); err == nil {
					homeDir = u.HomeDir
				}
			}
		}
	}

	if homeDir == "" {
		if u := firstHumanUser(); u != nil {
			homeDir = u.HomeDir
		}
	}

	if homeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to determine home directory: %w", err)
		}
		homeDir = home
	}

	return filepath.Join(homeDir, ".chef"), nil
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
	if os.Geteuid() != 0 {
		return
	}
	username := os.Getenv("SUDO_USER")
	if username == "" {
		if out, err := exec.Command("logname").Output(); err == nil {
			login := strings.TrimSpace(string(out))
			if login != "" && login != "root" {
				username = login
			}
		}
	}
	if username == "" {
		if u := firstHumanUser(); u != nil {
			username = u.Username
		}
	}
	if username != "" {
		_ = exec.Command("chown", "-R", username+":"+username, root).Run()
	}
}
