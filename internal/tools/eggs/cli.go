// Package eggs integrates with the existing Eggs command-line interface.
package eggs

import (
	"bufio"
	"errors"
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
	ModeStandard  RemasterMode = "Live standard"
	ModeClone     RemasterMode = "Clone del sistema"
	ModeEncrypted RemasterMode = "Clone cifrato"
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
		version = "Penguins’ Eggs rilevato: " + path
	}
	if versionErr != nil {
		version += " (versione non determinata)"
	}
	return path, version, nil
}

// RemasterArgs builds the current CLI arguments without shell interpretation.
func (CLIAdapter) RemasterArgs(mode RemasterMode, nest string) []string {
	args := []string{"remaster"}
	switch mode {
	case ModeClone:
		args = append(args, "--clone")
	case ModeEncrypted:
		args = append(args, "--crypted")
	}
	return append(args, "--path", nest)
}

// RunPrivileged executes Eggs as root or through pkexec and streams both outputs.
func (CLIAdapter) RunPrivileged(eggsPath string, args []string, appendLog func(string)) error {
	command, commandArgs, err := privilegedCommand(eggsPath, args)
	if err != nil {
		return err
	}

	appendLog("$ " + strings.Join(append([]string{command}, commandArgs...), " ") + "\n\n")
	cmd := exec.Command(command, commandArgs...)
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

func privilegedCommand(eggsPath string, args []string) (string, []string, error) {
	if os.Geteuid() == 0 {
		return eggsPath, args, nil
	}
	pkexecPath, err := exec.LookPath("pkexec")
	if err != nil {
		return "", nil, errors.New("pkexec non trovato: installa polkit oppure avvia temporaneamente la GUI come root")
	}
	return pkexecPath, append([]string{eggsPath}, args...), nil
}

// RunInteractiveTerminal launches the encrypted wizard in the first supported terminal.
func (CLIAdapter) RunInteractiveTerminal(eggsPath string, args []string) error {
	command, commandArgs, err := privilegedCommand(eggsPath, args)
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
	return errors.New("nessun terminale grafico supportato trovato per il wizard cifrato")
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
		return ISOArtifact{}, errors.New("nessuna nuova ISO trovata")
	}
	sort.Slice(found, func(i, j int) bool { return found[i].ModTime.After(found[j].ModTime) })
	return found[0], nil
}
