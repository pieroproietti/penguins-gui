// Package builder creates Debian packages without administrative privileges.
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
	base, revision := getGitVersion(root)
	version := base + "-" + revision
	if _, err := output(root, "dpkg", "--validate-version", version); err != nil {
		return "", err
	}
	arch, err := output(root, "dpkg", "--print-architecture")
	if err != nil {
		return "", err
	}
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
	if err := os.MkdirAll(filepath.Join(stage, "DEBIAN"), 0755); err != nil {
		return "", err
	}
	binary, err := os.ReadFile(filepath.Join(root, "penguins-gui"))
	if err != nil {
		return "", err
	}
	if err := writeFile(stage, "usr/bin/penguins-gui", binary, 0755); err != nil {
		return "", err
	}
	for source, dest := range map[string]string{
		"penguins-gui.desktop":  "usr/share/applications/penguins-gui.desktop",
		"penguins-gui.svg":      "usr/share/icons/hicolor/scalable/apps/penguins-gui.svg",
		"49-penguins-gui.rules": "usr/share/polkit-1/rules.d/49-penguins-gui.rules",
	} {
		data, err := assets.ReadFile("assets/" + source)
		if err != nil {
			return "", err
		}
		if err := writeFile(stage, dest, data, 0644); err != nil {
			return "", err
		}
	}
	// dpkg-shlibdeps needs a source control file; it is not shipped in the package.
	sourceControl := "Source: penguins-gui\nSection: utils\nPriority: optional\nMaintainer: Piero Proietti <piero.proietti@gmail.com>\n\nPackage: penguins-gui\nArchitecture: any\nDescription: Desktop interface for Penguins' Eggs\n"
	if err := writeFile(work, "debian/control", []byte(sourceControl), 0644); err != nil {
		return "", err
	}
	deps, err := output(work, "dpkg-shlibdeps", "-O", filepath.Join(stage, "usr/bin/penguins-gui"))
	if err != nil {
		return "", err
	}
	const prefix = "shlibs:Depends="
	if !strings.HasPrefix(deps, prefix) {
		return "", fmt.Errorf("unexpected dpkg-shlibdeps output: %q", deps)
	}
	depends := strings.TrimPrefix(deps, prefix) + ", penguins-eggs, pkexec, xdg-utils"
	control := fmt.Sprintf("Package: penguins-gui\nVersion: %s\nSection: utils\nPriority: optional\nArchitecture: %s\nMaintainer: Piero Proietti <piero.proietti@gmail.com>\nDepends: %s\nHomepage: https://github.com/pieroproietti/penguins-gui\nDescription: Desktop interface for Penguins' Eggs\n Create live and cloned system images through the Eggs command-line interface.\n", version, arch, depends)
	if err := writeFile(stage, "DEBIAN/control", []byte(control), 0644); err != nil {
		return "", err
	}
	name := fmt.Sprintf("penguins-gui_%s_%s.deb", version, arch)
	temporary := filepath.Join(work, name)
	if _, err := output(root, "dpkg-deb", "--root-owner-group", "--build", stage, temporary); err != nil {
		return "", err
	}
	dest := filepath.Join(dist, name)
	if err := os.Rename(temporary, dest); err != nil {
		return "", err
	}
	return dest, nil
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
