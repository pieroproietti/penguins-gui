package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"text/template"
)

func fedoraRecipe(data RecipeData) ([]byte, error) {
	// Tags must not inject RPM macros or spec sections.
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.+_]*$`).MatchString(data.BaseVersion) ||
		!regexp.MustCompile(`^[1-9][0-9]*$`).MatchString(data.Rel) {
		return nil, fmt.Errorf("invalid Fedora version: %s-%s", data.BaseVersion, data.Rel)
	}
	source, err := assets.ReadFile("assets/fedora.tmpl")
	if err != nil {
		return nil, err
	}
	recipe, err := template.New("penguins-gui.spec").Parse(string(source))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := recipe.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func packageFedora(root, work, stage, dist string, data RecipeData) (string, error) {
	switch runtime.GOARCH {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("unsupported Fedora architecture: %s", runtime.GOARCH)
	}
	recipe, err := fedoraRecipe(data)
	if err != nil {
		return "", err
	}
	if err := writeFile(work, "SPECS/penguins-gui.spec", recipe, 0644); err != nil {
		return "", err
	}
	cmd := exec.Command("rpmbuild", "-bb", "--define", "_topdir "+work,
		filepath.Join(work, "SPECS/penguins-gui.spec"))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PENGUINS_GUI_STAGE="+stage)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("rpmbuild: %w", err)
	}
	packages, err := filepath.Glob(filepath.Join(work, "RPMS", "*", "*.rpm"))
	if err != nil {
		return "", err
	}
	if len(packages) != 1 {
		return "", fmt.Errorf("expected one Fedora RPM, found %d", len(packages))
	}
	dest := filepath.Join(dist, filepath.Base(packages[0]))
	if err := os.Rename(packages[0], dest); err != nil {
		return "", err
	}
	return dest, nil
}
