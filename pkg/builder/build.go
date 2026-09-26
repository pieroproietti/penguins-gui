// Package builder creates native Linux packages without administrative privileges.
package builder

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed assets/*
var assets embed.FS

// Build packages the native binary produced by make build, from the project root.
func Build(root string) (string, error) {
	if os.Geteuid() == 0 {
		return "", fmt.Errorf("run packaging as a normal user, without sudo")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	family, err := detectFamily()
	if err != nil {
		return "", err
	}
	base, revision := getGitVersion(root)
	dist := filepath.Join(root, "dist")
	if err := os.MkdirAll(dist, 0755); err != nil {
		return "", err
	}
	work, err := os.MkdirTemp(dist, ".package-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(work)
	stage := filepath.Join(work, "stage")
	if err := staging(root, stage); err != nil {
		return "", err
	}
	data := RecipeData{BaseVersion: base, Rel: revision}
	switch family {
	case "fedora", "opensuse":
		return packageRPM(root, work, stage, dist, data, family)
	case "arch":
		return packageArch(root, work, stage, dist, data)
	case "debian":
		return packageDebian(root, work, stage, dist, data)
	default:
		return "", fmt.Errorf("unsupported package family: %s", family)
	}
}

func staging(root, stage string) error {
	binary, err := os.ReadFile(filepath.Join(root, "penguins-gui"))
	if err != nil {
		return err
	}
	if err := writeFile(stage, "usr/bin/penguins-gui", binary, 0755); err != nil {
		return err
	}
	for source, dest := range map[string]string{
		"penguins-gui.desktop":  "usr/share/applications/penguins-gui.desktop",
		"penguins-gui.svg":      "usr/share/icons/hicolor/scalable/apps/penguins-gui.svg",
		"49-penguins-gui.rules": "usr/share/polkit-1/rules.d/49-penguins-gui.rules",
	} {
		data, err := assets.ReadFile("assets/" + source)
		if err != nil {
			return err
		}
		if err := writeFile(stage, dest, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(root, name string, data []byte, mode os.FileMode) error {
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, mode)
}

func output(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", cmd.String(), err)
	}
	return strings.TrimSpace(string(out)), nil
}
