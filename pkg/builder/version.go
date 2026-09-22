// Version rules adapted from penguins-tailor/pkg/builder/u_get-git-version.go.
// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
package builder

import (
	"os/exec"
	"strings"
)

func getGitVersion(root string) (string, string) {
	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}
	tag, err := git("describe", "--tags", "--abbrev=0")
	base := strings.TrimPrefix(tag, "v")
	if err != nil || base == "" {
		revision, err := git("rev-list", "--count", "HEAD")
		if err != nil || revision == "" {
			revision = "1"
		}
		return "0.1.0", revision
	}
	revision, err := git("rev-list", "--count", "HEAD", "--not", "--tags")
	if err != nil || revision == "" || revision == "0" {
		revision = "1"
	}
	return strings.NewReplacer("-", ".", "_", ".").Replace(base), revision
}
