package main

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
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

//go:embed pkg/builder/assets/penguins-gui.svg
var iconSVG []byte

var (
	appIcon           = fyne.NewStaticResource("penguins-gui.svg", iconSVG)
	ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
)

func main() {
	a := app.NewWithID(applicationID)
	a.SetIcon(appIcon)
	w := a.NewWindow("Penguins GUI")
	w.SetIcon(appIcon)
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
			return errors.New("select a working directory")
		}
		return nil
	}

	browse := widget.NewButton("Browse…", func() {
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
	logGrid.SetText("Waiting.\n")
	logScroll := container.NewScroll(logGrid)
	logScroll.SetMinSize(fyne.NewSize(760, 260))

	result := widget.NewLabel("")
	result.Wrapping = fyne.TextWrapBreak
	openFolder := widget.NewButton("Open ISO folder", func() {})
	openFolder.Hide()

	howToBoot := widget.NewButton("How to boot ISO…", func() {
		showBootGuideDialog(w)
	})
	howToBoot.Hide()

	start := widget.NewButton("Create ISO", nil)
	if detectErr != nil {
		start.Disable()
		status.SetText("Penguins’ Eggs was not found in PATH. Install eggs and restart the application.")
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

	var isBusy bool
	var busyMu sync.Mutex

	runToolCommand := func(title string, args []string, confirmPrompt string) {
		if detectErr != nil {
			dialog.ShowError(errors.New("Penguins’ Eggs was not found in PATH"), w)
			return
		}

		busyMu.Lock()
		if isBusy {
			busyMu.Unlock()
			dialog.ShowInformation("Busy", "Another operation is currently in progress. Please wait until it completes.", w)
			return
		}
		busyMu.Unlock()

		execute := func() {
			busyMu.Lock()
			isBusy = true
			busyMu.Unlock()

			start.Disable()
			browse.Disable()
			mode.Disable()
			pathEntry.Disable()
			openFolder.Hide()
			howToBoot.Hide()

			result.SetText(fmt.Sprintf("%s in progress…", title))
			logMu.Lock()
			logText = ""
			logMu.Unlock()
			logGrid.SetText("")

			go func() {
				err := eggsCLI.RunPrivileged(eggsPath, args, appendLog)
				fyne.Do(func() {
					busyMu.Lock()
					isBusy = false
					busyMu.Unlock()

					start.Enable()
					browse.Enable()
					mode.Enable()
					pathEntry.Enable()

					if err != nil {
						result.SetText(fmt.Sprintf("%s failed: %v", title, err))
						dialog.ShowError(err, w)
					} else {
						result.SetText(fmt.Sprintf("%s completed successfully.", title))
						dialog.ShowInformation(title, fmt.Sprintf("%s finished successfully.", title), w)
					}
				})
			}()
		}

		if confirmPrompt != "" {
			dialog.ShowConfirm(title, confirmPrompt, func(ok bool) {
				if ok {
					execute()
				}
			}, w)
		} else {
			execute()
		}
	}

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Quit", func() {
			a.Quit()
		}),
	)
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Delete previous ISOs (eggs kill)…", func() {
			runToolCommand(
				"Delete previous ISOs",
				[]string{"kill"},
				"Are you sure you want to delete previous ISOs and clean the build nest?",
			)
		}),
	)
	toolsMenu := fyne.NewMenu("Tools",
		fyne.NewMenuItem("Clean system remnants (clean)…", func() {
			runToolCommand(
				"Clean system remnants",
				[]string{"tools", "clean"},
				"Clean log rotation, package manager cache, and host system remnants?",
			)
		}),
		fyne.NewMenuItem("Configure GRUB loopback (grub40)…", func() {
			runToolCommand(
				"Configure GRUB loopback",
				[]string{"tools", "grub40"},
				"Generate GRUB configuration to boot ANY ISO via loopback?",
			)
		}),
		fyne.NewMenuItem("Manage repository (repo)…", func() {
			runToolCommand(
				"Manage repository",
				[]string{"tools", "repo"},
				"Add or remove the official penguins-eggs repository?",
			)
		}),
		fyne.NewMenuItem("Update /etc/skel (skel)…", func() {
			runToolCommand(
				"Update /etc/skel",
				[]string{"tools", "skel"},
				"Create /etc/skel based on the current user's configurations?",
			)
		}),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Testing & Booting the ISO…", func() {
			showBootGuideDialog(w)
		}),
		fyne.NewMenuItem("Documentation", func() {
			if u, err := url.Parse("https://penguins-eggs.net"); err == nil {
				_ = a.OpenURL(u)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About", func() {
			aboutText := fmt.Sprintf(
				"Penguins GUI\n\n"+
					"Desktop interface for Penguins' Eggs.\n\n"+
					"Eggs: %s\n\n"+
					"Homepage: https://penguins-eggs.net",
				eggsVersion,
			)
			dialog.ShowInformation("About Penguins GUI", aboutText, w)
		}),
	)
	w.SetMainMenu(fyne.NewMainMenu(fileMenu, editMenu, toolsMenu, helpMenu))

	start.OnTapped = func() {
		if err := pathEntry.Validate(); err != nil {
			dialog.ShowError(err, w)
			return
		}

		busyMu.Lock()
		if isBusy {
			busyMu.Unlock()
			return
		}
		isBusy = true
		busyMu.Unlock()

		nest := filepath.Clean(pathEntry.Text)
		selectedMode := eggs.RemasterMode(mode.Selected)
		args := eggsCLI.RemasterArgs(selectedMode, nest)

		start.Disable()
		browse.Disable()
		mode.Disable()
		pathEntry.Disable()
		result.SetText("Remaster in progress…")
		openFolder.Hide()
		howToBoot.Hide()
		logMu.Lock()
		logText = ""
		logMu.Unlock()
		logGrid.SetText("")

		startedAt := time.Now()
		go func() {
			var err error
			if selectedMode == eggs.ModeEncrypted {
				appendLog("Encrypted mode uses the interactive Eggs wizard.\n")
				appendLog("Passphrase and encryption parameters will be requested in a separate terminal.\n\n")
				err = eggsCLI.RunInteractiveTerminal(eggsPath, args)
			} else {
				err = eggsCLI.RunPrivileged(eggsPath, args, appendLog)
			}

			fyne.Do(func() {
				busyMu.Lock()
				isBusy = false
				busyMu.Unlock()

				start.Enable()
				browse.Enable()
				mode.Enable()
				pathEntry.Enable()

				if err != nil {
					result.SetText(fmt.Sprintf("Remaster failed: %v", err))
					dialog.ShowError(err, w)
				} else {
					artifact, findErr := eggsCLI.NewestISO(nest, startedAt)
					if findErr != nil {
						result.SetText("Remaster completed, but no new ISO was found in the selected directory.")
					} else {
						result.SetText(fmt.Sprintf("ISO created: %s (%s)", artifact.Path, humanSize(artifact.Size)))
						openFolder.OnTapped = func() {
							if err := openDirectory(filepath.Dir(artifact.Path)); err != nil {
								dialog.ShowError(err, w)
							}
						}
						openFolder.Show()
						howToBoot.Show()
					}
					dialog.ShowInformation("Remaster completed", "Penguins’ Eggs finished successfully.\n\n"+result.Text, w)
				}
			})
		}()
	}

	header := container.NewVBox(
		widget.NewLabelWithStyle("Create an ISO of the running system", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		status,
		widget.NewSeparator(),
	)

	form := container.NewVBox(
		widget.NewLabelWithStyle("Mode", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mode,
		modeHelp,
		widget.NewLabelWithStyle("Working directory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, nil, browse, pathEntry),
		container.NewHBox(layout.NewSpacer(), start),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		result,
		container.NewHBox(layout.NewSpacer(), howToBoot, openFolder),
	)
	w.SetContent(container.NewBorder(container.NewVBox(header, form), footer, nil, nil, logScroll))
	w.ShowAndRun()
}

func stripANSI(text string) string {
	return ansiEscapePattern.ReplaceAllString(text, "")
}

func modeDescription(mode eggs.RemasterMode) string {
	switch mode {
	case eggs.ModeClone:
		return "Include users, configurations and data in an unencrypted clone."
	case eggs.ModeEncrypted:
		return "Create an encrypted clone; passphrase and parameters are requested by Eggs."
	default:
		return "Create a clean live system without user accounts or personal data."
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
		return errors.New("xdg-open not found")
	}
	return exec.Command(launcher, path).Start()
}

func showBootGuideDialog(parent fyne.Window) {
	markdown := `### How to Boot and Test Your ISO

#### 1. Virtual Machines (Quickest & Safest)
Test your ISO immediately without rebooting your physical computer:
* **GNOME Boxes** or **virt-manager**: create a new virtual machine and select the ISO file.
* **VirtualBox**: create an OS machine and mount the ISO in the virtual optical drive.
* **QEMU CLI**:
` + "```bash" + `
qemu-system-x86_64 -enable-kvm -m 4G -cdrom /path/to/egg.iso
` + "```" + `

#### 2. Ventoy USB (Recommended for Hardware)
The most convenient method for physical hardware:
* Install **Ventoy** on your USB flash drive once (https://www.ventoy.net).
* Simply copy and paste the ` + "`.iso`" + ` file onto the Ventoy USB partition.
* No need to reformat when generating new ISOs!

#### 3. Direct USB Flashing (Dedicated Live USB)
Write the ISO directly to a dedicated USB drive:
* **Graphical tools**: Balena Etcher, Popsicle, Raspberry Pi Imager, or your desktop's image writer.
* **Terminal (dd)**:
` + "```bash" + `
sudo dd if=/path/to/egg.iso of=/dev/sdX bs=4M status=progress oflag=sync
` + "```" + `
*(Replace /dev/sdX with your actual USB drive; take extra care not to overwrite system drives!)*
`
	rich := widget.NewRichTextFromMarkdown(markdown)
	rich.Wrapping = fyne.TextWrapWord
	scroll := container.NewScroll(rich)
	scroll.SetMinSize(fyne.NewSize(580, 420))

	d := dialog.NewCustom("Testing & Booting the ISO", "Close", scroll, parent)
	d.Resize(fyne.NewSize(620, 480))
	d.Show()
}
