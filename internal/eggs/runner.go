/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package eggs

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

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
	stream := func(r io.Reader) {
		defer readers.Done()
		scanner := bufio.NewScanner(r)
		scanner.Split(scanLinesOrCR)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			appendLog(scanner.Text() + "\n")
		}
	}
	go stream(stdout)
	go stream(stderr)

	readers.Wait()
	waitErr := cmd.Wait()
	return waitErr
}

// scanLinesOrCR is a bufio.SplitFunc that splits on either '\n' or '\r'.
// It correctly handles CRLF ("\r\n") as a single line break.
func scanLinesOrCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	for i, b := range data {
		if b == '\r' {
			if i+1 < len(data) {
				if data[i+1] == '\n' {
					return i + 2, data[:i], nil
				}
				return i + 1, data[:i], nil
			}
			if atEOF {
				return i + 1, data[:i], nil
			}
			// Request more data to see if '\n' follows.
			return 0, nil, nil
		}
		if b == '\n' {
			return i + 1, data[:i], nil
		}
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
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
