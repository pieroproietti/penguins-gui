package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	applicationID = "net.penguins-eggs.gui"
	defaultNest   = "/home/eggs"
)

type remasterMode string

const (
	modeStandard  remasterMode = "Live standard"
	modeClone     remasterMode = "Clone del sistema"
	modeEncrypted remasterMode = "Clone cifrato"
)

type isoArtifact struct {
	Path    string
	Size    int64
	ModTime time.Time
}

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func main() {
	a := app.NewWithID(applicationID)
	w := a.NewWindow("Penguins’ Eggs")
	w.Resize(fyne.NewSize(820, 650))

	eggsPath, eggsVersion, detectErr := detectEggs()
	status := widget.NewLabel(eggsVersion)
	status.Wrapping = fyne.TextWrapWord

	mode := widget.NewRadioGroup([]string{
		string(modeStandard),
		string(modeClone),
		string(modeEncrypted),
	}, nil)
	mode.Required = true
	mode.SetSelected(string(modeStandard))

	modeHelp := widget.NewLabel(modeDescription(modeStandard))
	modeHelp.Wrapping = fyne.TextWrapWord
	mode.OnChanged = func(selected string) {
		modeHelp.SetText(modeDescription(remasterMode(selected)))
	}

	pathEntry := widget.NewEntry()
	pathEntry.SetText(defaultNest)
	pathEntry.Validator = func(value string) error {
		if strings.TrimSpace(value) == "" {
			return errors.New("seleziona una cartella di lavoro")
		}
		return nil
	}

	browse := widget.NewButton("Scegli…", func() {
		d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
			dialog.ShowError(err, w)
			return
			}
			if uri != nil {
				pathEntry.SetText(uri.Path())
			}
		}, w)
		d.Show()
	})

	logGrid := widget.NewTextGrid()
	logGrid.SetText("In attesa.\n")
	logScroll := container.NewScroll(logGrid)
	logScroll.SetMinSize(fyne.NewSize(760, 260))

	result := widget.NewLabel("")
	result.Wrapping = fyne.TextWrapBreak
	openFolder := widget.NewButton("Apri cartella ISO", func() {})
	openFolder.Hide()

	start := widget.NewButton("Crea la ISO", nil)
	if detectErr != nil {
		start.Disable()
		status.SetText("Penguins’ Eggs non trovato nel PATH. Installa eggs e riavvia la GUI.")
	}

	var logMu sync.Mutex
	logText := ""
	appendLog := func(text string) {
		logMu.Lock()
		logText += stripANSI(text)
		current := logText
		logMu.Unlock()
		fyne.Do(func() {
			logGrid.SetText(current)
			logScroll.ScrollToBottom()
		})
	}

	start.OnTapped = func() {
		if err := pathEntry.Validate(); err != nil {
			dialog.ShowError(err, w)
			return
		}

		nest := filepath.Clean(pathEntry.Text)
		selectedMode := remasterMode(mode.Selected)
		args := remasterArgs(selectedMode, nest)

		start.Disable()
		browse.Disable()
		mode.Disable()
		pathEntry.Disable()
		result.SetText("Remaster in esecuzione…")
		openFolder.Hide()
		logMu.Lock()
		logText = ""
		logMu.Unlock()
		logGrid.SetText("")

		startedAt := time.Now()
		go func() {
			var err error
			if selectedMode == modeEncrypted {
				appendLog("La modalità cifrata usa il wizard interattivo di Eggs.\n")
				appendLog("Passphrase e parametri crittografici verranno richiesti in un terminale separato.\n\n")
				err = runInteractiveTerminal(eggsPath, args)
			} else {
				err = runPrivileged(eggsPath, args, appendLog)
			}
			if err != nil {
				fyne.Do(func() {
					result.SetText(fmt.Sprintf("Remaster terminato con errore: %v", err))
					dialog.ShowError(err, w)
				})
			} else {
				artifact, findErr := newestISO(nest, startedAt)
				fyne.Do(func() {
					if findErr != nil {
						result.SetText("Remaster completato, ma non ho trovato una nuova ISO nella cartella selezionata.")
					} else {
						result.SetText(fmt.Sprintf("ISO creata: %s (%s)", artifact.Path, humanSize(artifact.Size)))
						openFolder.OnTapped = func() {
							if err := openDirectory(filepath.Dir(artifact.Path)); err != nil {
								dialog.ShowError(err, w)
							}
						}
						openFolder.Show()
					}
				})
			}

			fyne.Do(func() {
				start.Enable()
				browse.Enable()
				mode.Enable()
				pathEntry.Enable()
			})
		}()
	}

	header := container.NewVBox(
		widget.NewLabelWithStyle("Crea una ISO del sistema in esecuzione", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		status,
		widget.NewSeparator(),
	)

	form := container.NewVBox(
		widget.NewLabelWithStyle("Modalità", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mode,
		modeHelp,
		widget.NewLabelWithStyle("Cartella di lavoro", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, nil, browse, pathEntry),
		container.NewHBox(layout.NewSpacer(), start),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	footer := container.NewVBox(widget.NewSeparator(), result, container.NewHBox(layout.NewSpacer(), openFolder))
	w.SetContent(container.NewBorder(container.NewVBox(header, form), footer, nil, nil, logScroll))
	w.ShowAndRun()
}

func stripANSI(text string) string {
	return ansiEscapePattern.ReplaceAllString(text, "")
}

func detectEggs() (path, version string, err error) {
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

func modeDescription(mode remasterMode) string {
	switch mode {
	case modeClone:
		return "Include utenti, configurazioni e dati nel clone non cifrato."
	case modeEncrypted:
		return "Crea il clone cifrato previsto da Penguins’ Eggs; la passphrase viene richiesta da Eggs."
	default:
		return "Crea una live ripulita, senza includere utenti e dati personali."
	}
}

func remasterArgs(mode remasterMode, nest string) []string {
	args := []string{"remaster"}
	switch mode {
	case modeClone:
		args = append(args, "--clone")
	case modeEncrypted:
		args = append(args, "--crypted")
	}
	return append(args, "--path", nest)
}

func runPrivileged(eggsPath string, args []string, appendLog func(string)) error {
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

func runInteractiveTerminal(eggsPath string, args []string) error {
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

func newestISO(root string, notBefore time.Time) (isoArtifact, error) {
	var found []isoArtifact
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".iso") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.ModTime().Before(notBefore.Add(-2 * time.Second)) {
			return nil
		}
		found = append(found, isoArtifact{Path: path, Size: info.Size(), ModTime: info.ModTime()})
		return nil
	})
	if err != nil {
		return isoArtifact{}, err
	}
	if len(found) == 0 {
		return isoArtifact{}, errors.New("nessuna nuova ISO trovata")
	}
	sort.Slice(found, func(i, j int) bool { return found[i].ModTime.After(found[j].ModTime) })
	return found[0], nil
}

func humanSize(size int64) string {
	const unit = int64(1024)
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := unit, 0
	for value := size / unit; value >= unit && exp < 5; value /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func openDirectory(path string) error {
	launcher, err := exec.LookPath("xdg-open")
	if err != nil {
		return errors.New("xdg-open non trovato")
	}
	return exec.Command(launcher, path).Start()
}
