package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pieroproietti/penguins-gui/internal/tools/eggs"

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

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func main() {
	a := app.NewWithID(applicationID)
	w := a.NewWindow("Penguins’ Eggs")
	w.Resize(fyne.NewSize(820, 650))

	eggsCLI := eggs.CLIAdapter{}
	eggsPath, eggsVersion, detectErr := eggsCLI.Detect()
	status := widget.NewLabel(eggsVersion)
	status.Wrapping = fyne.TextWrapWord

	mode := widget.NewRadioGroup([]string{
		string(eggs.ModeStandard),
		string(eggs.ModeClone),
		string(eggs.ModeEncrypted),
	}, nil)
	mode.Required = true
	mode.SetSelected(string(eggs.ModeStandard))

	modeHelp := widget.NewLabel(modeDescription(eggs.ModeStandard))
	modeHelp.Wrapping = fyne.TextWrapWord
	mode.OnChanged = func(selected string) {
		modeHelp.SetText(modeDescription(eggs.RemasterMode(selected)))
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
		selectedMode := eggs.RemasterMode(mode.Selected)
		args := eggsCLI.RemasterArgs(selectedMode, nest)

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
			if selectedMode == eggs.ModeEncrypted {
				appendLog("La modalità cifrata usa il wizard interattivo di Eggs.\n")
				appendLog("Passphrase e parametri crittografici verranno richiesti in un terminale separato.\n\n")
				err = eggsCLI.RunInteractiveTerminal(eggsPath, args)
			} else {
				err = eggsCLI.RunPrivileged(eggsPath, args, appendLog)
			}
			if err != nil {
				fyne.Do(func() {
					result.SetText(fmt.Sprintf("Remaster terminato con errore: %v", err))
					dialog.ShowError(err, w)
				})
			} else {
				artifact, findErr := eggsCLI.NewestISO(nest, startedAt)
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
					dialog.ShowInformation("Remaster completato", "Penguins’ Eggs ha terminato senza errori.\n\n"+result.Text, w)
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

func modeDescription(mode eggs.RemasterMode) string {
	switch mode {
	case eggs.ModeClone:
		return "Include utenti, configurazioni e dati nel clone non cifrato."
	case eggs.ModeEncrypted:
		return "Crea il clone cifrato previsto da Penguins’ Eggs; la passphrase viene richiesta da Eggs."
	default:
		return "Crea una live ripulita, senza includere utenti e dati personali."
	}
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
