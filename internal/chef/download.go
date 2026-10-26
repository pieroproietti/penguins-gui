package chef

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DownloadChef downloads or updates the chef recipes repository in chefRoot.
func DownloadChef(chefRoot string, logFn func(string)) error {
	chefPath, err := exec.LookPath("chef")
	if err == nil {
		if logFn != nil {
			logFn("$ chef cook --pull\n")
		}
		cmd := exec.Command(chefPath, "cook", "--pull")
		out, err := cmd.CombinedOutput()
		if logFn != nil && len(out) > 0 {
			logFn(string(out) + "\n")
		}
		if err == nil {
			fixChefOwnership(chefRoot)
			return nil
		}
		if logFn != nil {
			logFn(fmt.Sprintf("chef cook --pull exited with error (%v), attempting git clone fallback...\n", err))
		}
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("neither chef nor git command found in PATH")
	}

	if _, err := os.Stat(chefRoot); err == nil {
		_ = os.RemoveAll(chefRoot)
	}

	if logFn != nil {
		logFn(fmt.Sprintf("$ git clone https://github.com/pieroproietti/penguins-chef %s\n", chefRoot))
	}
	cmd := exec.Command(gitPath, "clone", "https://github.com/pieroproietti/penguins-chef", chefRoot)
	out, err := cmd.CombinedOutput()
	if logFn != nil && len(out) > 0 {
		logFn(string(out) + "\n")
	}
	if err != nil {
		return fmt.Errorf("git clone failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	fixChefOwnership(chefRoot)
	return nil
}
