package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func packageDebian(root, work, stage, dist string, data RecipeData) (string, error) {
	version := data.BaseVersion + "-" + data.Rel
	if _, err := output(root, "dpkg", "--validate-version", version); err != nil {
		return "", err
	}
	arch, err := output(root, "dpkg", "--print-architecture")
	if err != nil {
		return "", err
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
	depends := strings.TrimPrefix(deps, prefix) + ", pkexec, sudo, xdg-utils, curl, gnupg"
	control := fmt.Sprintf("Package: penguins-gui\nVersion: %s\nSection: utils\nPriority: optional\nArchitecture: %s\nMaintainer: Piero Proietti <piero.proietti@gmail.com>\nDepends: %s\nSuggests: penguins-eggs\nHomepage: https://github.com/pieroproietti/penguins-gui\nDescription: Desktop interface for Penguins' Eggs\n Create live and cloned system images through the Eggs command-line interface.\n", version, arch, depends)
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
