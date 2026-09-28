package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// BuildAppImage bundles the locally built GUI using linuxdeploy and its AppImage plugin.
func BuildAppImage(root string) (string, error) {
	if os.Geteuid() == 0 {
		return "", fmt.Errorf("run packaging as a normal user, without sudo")
	}
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		return "", fmt.Errorf("AppImage packaging currently supports Linux amd64 only")
	}
	tool := os.Getenv("LINUXDEPLOY")
	if tool == "" {
		tool = "linuxdeploy"
	}
	tool, err := exec.LookPath(tool)
	if err != nil {
		return "", fmt.Errorf("install linuxdeploy or set LINUXDEPLOY to its AppImage path: %w", err)
	}
	tool, err = filepath.Abs(tool)
	if err != nil {
		return "", err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return "", err
	}
	dist := filepath.Join(root, "dist")
	if err := os.MkdirAll(dist, 0755); err != nil {
		return "", err
	}
	work, err := os.MkdirTemp(dist, ".appimage-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(work)
	stage := filepath.Join(work, "AppDir")
	binary, err := os.ReadFile(filepath.Join(root, "penguins-gui"))
	if err != nil {
		return "", err
	}
	if err := writeFile(stage, "usr/bin/penguins-gui", binary, 0755); err != nil {
		return "", err
	}
	// Only desktop assets: polkit rules in an AppImage cannot be installed on the host.
	for source, dest := range map[string]string{
		"penguins-gui.desktop": "usr/share/applications/penguins-gui.desktop",
		"penguins-gui.svg":     "usr/share/icons/hicolor/scalable/apps/penguins-gui.svg",
	} {
		data, err := assets.ReadFile("assets/" + source)
		if err != nil {
			return "", err
		}
		if err := writeFile(stage, dest, data, 0644); err != nil {
			return "", err
		}
	}
	base, revision := getGitVersion(root)
	version := base + "-" + revision
	name := "penguins-gui-" + version + "-x86_64.AppImage"
	artifact := filepath.Join(work, name)
	cmd := exec.Command(tool, "--appdir", stage, "--desktop-file", filepath.Join(stage, "usr/share/applications/penguins-gui.desktop"), "--icon-file", filepath.Join(stage, "usr/share/icons/hicolor/scalable/apps/penguins-gui.svg"), "--output", "appimage")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "APPIMAGE_EXTRACT_AND_RUN=1", "ARCH=x86_64", "LDAI_OUTPUT="+artifact, "LINUXDEPLOY_OUTPUT_VERSION="+version)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("linuxdeploy: %w", err)
	}
	info, err := os.Stat(artifact)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return "", fmt.Errorf("linuxdeploy produced an invalid AppImage")
	}
	if err := os.Chmod(artifact, 0755); err != nil {
		return "", err
	}
	target := filepath.Join(dist, name)
	if err := os.Rename(artifact, target); err != nil {
		return "", err
	}
	return target, nil
}
