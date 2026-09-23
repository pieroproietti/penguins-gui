package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
)

type RecipeData struct {
	BaseVersion string
	Rel         string
	Arch        string
}

func detectFamily() (string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	return familyFromRelease(string(data))
}

func familyFromRelease(release string) (string, error) {
	values := map[string]string{}
	for _, line := range strings.Split(release, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[key] = strings.Trim(value, "\"'")
		}
	}
	for _, id := range strings.Fields(values["ID"] + " " + values["ID_LIKE"]) {
		switch id {
		case "fedora":
			return "fedora", nil
		case "arch", "manjaro":
			return "arch", nil
		case "debian", "ubuntu", "devuan":
			return "debian", nil
		}
	}
	return "", fmt.Errorf("unsupported distribution: %s (ID_LIKE=%s)", values["ID"], values["ID_LIKE"])
}

func packageArch(root, work, stage, dist string, data RecipeData) (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		data.Arch = "x86_64"
	case "arm64":
		data.Arch = "aarch64"
	default:
		return "", fmt.Errorf("unsupported Arch architecture: %s", runtime.GOARCH)
	}
	// Git tags are data, never executable PKGBUILD shell syntax.
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.+_]*$`).MatchString(data.BaseVersion) ||
		!regexp.MustCompile(`^[1-9][0-9]*$`).MatchString(data.Rel) {
		return "", fmt.Errorf("invalid Arch version: %s-%s", data.BaseVersion, data.Rel)
	}
	source, err := assets.ReadFile("assets/arch.tmpl")
	if err != nil {
		return "", err
	}
	recipe, err := template.New("PKGBUILD").Parse(string(source))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := recipe.Execute(&buf, data); err != nil {
		return "", err
	}
	if err := writeFile(stage, "PKGBUILD", buf.Bytes(), 0644); err != nil {
		return "", err
	}
	// The binary is already built. Runtime dependencies need not be installed
	// on the packaging host; pacman enforces them when installing the archive.
	cmd := exec.Command("makepkg", "--nodeps", "--force", "--noconfirm")
	cmd.Dir = stage
	cmd.Env = append(os.Environ(), "PKGDEST="+work, "BUILDDIR="+work, "PKGEXT=.pkg.tar.zst")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("makepkg: %w", err)
	}
	name := fmt.Sprintf("penguins-gui-%s-%s-%s.pkg.tar.zst", data.BaseVersion, data.Rel, data.Arch)
	dest := filepath.Join(dist, name)
	if err := os.Rename(filepath.Join(work, name), dest); err != nil {
		return "", err
	}
	return dest, nil
}
