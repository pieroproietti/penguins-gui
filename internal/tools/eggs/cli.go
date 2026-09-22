// Package eggs integrates with the existing Eggs command-line interface.
package eggs

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// CLIAdapter preserves the textual CLI integration, including interactive
// encrypted remasters and filesystem-based ISO discovery. Its zero value is ready
// to use and it has no dependency on the GUI.
type CLIAdapter struct{}

// RemasterMode identifies one of the existing remaster choices.
type RemasterMode string

// Remaster modes retain the labels used by the current GUI.
const (
	ModeStandard  RemasterMode = "Standard live"
	ModeClone     RemasterMode = "System clone"
	ModeEncrypted RemasterMode = "Encrypted clone"
)

// ISOArtifact describes an ISO found in the working directory.
type ISOArtifact struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// Detect locates Eggs and returns its textual version, with the existing fallback.
func (CLIAdapter) Detect() (path, version string, err error) {
	path, err = exec.LookPath("eggs")
	if err != nil {
		return "", "", err
	}
	out, versionErr := exec.Command(path, "version").CombinedOutput()
	version = strings.TrimSpace(string(out))
	if version == "" {
		version = "Penguins’ Eggs found: " + path
	}
	if versionErr != nil {
		version += " (version undetermined)"
	}
	return path, version, nil
}

// RemasterArgs produces the arguments for a remaster operation.
func (CLIAdapter) RemasterArgs(mode RemasterMode, workingDirectory string) []string {
	args := []string{"remaster"}
	switch mode {
	case ModeClone:
		args = append(args, "--clone")
	case ModeEncrypted:
		args = append(args, "--crypted")
	}
	if workingDirectory != "" {
		args = append(args, "--path", workingDirectory)
	}
	return args
}

// RunPrivileged executes Eggs with root privileges and streams both outputs.
// If adminPassword is provided, it uses sudo -S to elevate without prompting on the terminal.
// If luksPassphrase is provided, it passes EGGS_LUKS_PASSPHRASE in the environment for non-interactive encryption.
func (CLIAdapter) RunPrivileged(eggsPath string, args []string, adminPassword, luksPassphrase string, appendLog func(string)) error {
	command, commandArgs, err := privilegedCommand(eggsPath, args, adminPassword, luksPassphrase)
	if err != nil {
		return err
	}

	displayArgs := append([]string{command}, commandArgs...)
	appendLog("$ " + strings.Join(displayArgs, " ") + "\n\n")

	cmd := exec.Command(command, commandArgs...)
	if luksPassphrase != "" {
		cmd.Env = append(os.Environ(), "EGGS_LUKS_PASSPHRASE="+luksPassphrase)
	}

	var stdinWriter io.WriteCloser
	if adminPassword != "" {
		var pipeErr error
		stdinWriter, pipeErr = cmd.StdinPipe()
		if pipeErr != nil {
			return pipeErr
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	if stdinWriter != nil {
		_, _ = stdinWriter.Write([]byte(adminPassword + "\n"))
		_ = stdinWriter.Close()
	}

	var readers sync.WaitGroup
	readers.Add(2)
	stream := func(scanner *bufio.Scanner) {
		defer readers.Done()
		for scanner.Scan() {
			appendLog(scanner.Text() + "\n")
		}
	}
	go stream(bufio.NewScanner(stdout))
	go stream(bufio.NewScanner(stderr))

	waitErr := cmd.Wait()
	readers.Wait()
	return waitErr
}

func privilegedCommand(eggsPath string, args []string, adminPassword, luksPassphrase string) (string, []string, error) {
	if os.Geteuid() == 0 {
		return eggsPath, args, nil
	}

	if adminPassword != "" {
		sudoPath, err := exec.LookPath("sudo")
		if err != nil {
			return "", nil, errors.New("sudo not found: install sudo or run GUI as root")
		}
		cmdArgs := []string{"-S", "-p", ""}
		if luksPassphrase != "" {
			cmdArgs = append(cmdArgs, "env", "EGGS_LUKS_PASSPHRASE="+luksPassphrase)
		}
		cmdArgs = append(cmdArgs, eggsPath)
		cmdArgs = append(cmdArgs, args...)
		return sudoPath, cmdArgs, nil
	}

	// If running passwordless sudo works, use sudo
	if sudoPath, err := exec.LookPath("sudo"); err == nil {
		if exec.Command("sudo", "-n", "true").Run() == nil {
			cmdArgs := []string{}
			if luksPassphrase != "" {
				cmdArgs = append(cmdArgs, "env", "EGGS_LUKS_PASSPHRASE="+luksPassphrase)
			}
			cmdArgs = append(cmdArgs, eggsPath)
			cmdArgs = append(cmdArgs, args...)
			return sudoPath, cmdArgs, nil
		}
	}

	// Fallback to pkexec
	pkexecPath, err := exec.LookPath("pkexec")
	if err != nil {
		return "", nil, errors.New("neither sudo nor pkexec found: install sudo/polkit or run GUI as root")
	}
	return pkexecPath, append([]string{eggsPath}, args...), nil
}

// RunInteractiveTerminal launches the encrypted wizard in the first supported terminal.
func (CLIAdapter) RunInteractiveTerminal(eggsPath string, args []string) error {
	command, commandArgs, err := privilegedCommand(eggsPath, args, "", "")
	if err != nil {
		return err
	}

	type terminalCandidate struct {
		name   string
		prefix []string
	}
	candidates := []terminalCandidate{
		{name: "x-terminal-emulator", prefix: []string{"-e"}},
		{name: "gnome-terminal", prefix: []string{"--"}},
		{name: "konsole", prefix: []string{"-e"}},
		{name: "xfce4-terminal", prefix: []string{"-x"}},
		{name: "xterm", prefix: []string{"-e"}},
	}
	for _, candidate := range candidates {
		terminalPath, lookErr := exec.LookPath(candidate.name)
		if lookErr != nil {
			continue
		}
		terminalArgs := append([]string{}, candidate.prefix...)
		terminalArgs = append(terminalArgs, command)
		terminalArgs = append(terminalArgs, commandArgs...)
		return exec.Command(terminalPath, terminalArgs...).Run()
	}
	return errors.New("no supported graphical terminal found for encrypted wizard")
}

// NewestISO finds the latest ISO directly in root, without scanning subdirectories,
// allowing two seconds of timestamp tolerance.
func (CLIAdapter) NewestISO(root string, notBefore time.Time) (ISOArtifact, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return ISOArtifact{}, err
	}
	var found []ISOArtifact
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".iso") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return ISOArtifact{}, err
		}
		if info.ModTime().Before(notBefore.Add(-2 * time.Second)) {
			continue
		}
		found = append(found, ISOArtifact{Path: filepath.Join(root, entry.Name()), Size: info.Size(), ModTime: info.ModTime()})
	}
	if len(found) == 0 {
		return ISOArtifact{}, errors.New("no new ISO found")
	}
	sort.Slice(found, func(i, j int) bool { return found[i].ModTime.After(found[j].ModTime) })
	return found[0], nil
}
