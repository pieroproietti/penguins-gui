package main

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"math"

	"github.com/pieroproietti/penguins-gui/internal/tools/eggs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	applicationID = "net.penguins-eggs.gui"
	defaultNest   = "/home/eggs"
	defaultAuthor = "Piero Proietti <piero.proietti@gmail.com>"
)

//go:embed pkg/builder/assets/penguins-gui.svg
var iconSVG []byte

var (
	appIcon           = fyne.NewStaticResource("penguins-gui.svg", iconSVG)
	ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
	guiVersion        = "v26.9.23"
	author            = defaultAuthor
)

func getGUIVersion() string {
	if guiVersion != "" && guiVersion != "devel" {
		return guiVersion
	}
	return "v26.9.23"
}

func main() {
	a := app.NewWithID(applicationID)
	a.SetIcon(appIcon)

	currentScale := float32(a.Preferences().FloatWithFallback(preferenceFontScale, float64(defaultScale)))
	currentScale = clampScale(currentScale)
	if currentScale != defaultScale {
		a.Settings().SetTheme(newScaledTheme(theme.DefaultTheme(), currentScale))
	}

	w := a.NewWindow("Penguins GUI")
	w.SetIcon(appIcon)

	initialWidth := float32(math.Round(float64(900 * currentScale)))
	initialHeight := float32(math.Round(float64(700 * currentScale)))
	if initialWidth < 900 {
		initialWidth = 900
	}
	if initialHeight < 700 {
		initialHeight = 700
	}
	w.Resize(fyne.NewSize(initialWidth, initialHeight))

	eggsCLI := eggs.CLIAdapter{}
	eggsPath, eggsVersion, detectErr := eggsCLI.Detect()

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

	var logMu sync.Mutex
	logText := ""
	appendLog := func(text string) {
		logMu.Lock()
		logText += stripANSI(text)
		logMu.Unlock()
		fyne.Do(func() {
			logMu.Lock()
			current := logText
			logMu.Unlock()
			logGrid.SetText(current)
			logScroll.ScrollToBottom()
		})
	}

	var isBusy bool
	var busyMu sync.Mutex
	var createISOMenuItem *fyne.MenuItem
	var tbCreateISO, tbKill, tbClean, tbGrub, tbDocs *actionButton
	var setButtonsEnabled func(bool)

	const defaultExplanation = "Select or hover over an action to view its description."
	explanationLabel := widget.NewLabel(defaultExplanation)
	explanationLabel.TextStyle = fyne.TextStyle{Italic: true}
	explanationLabel.Wrapping = fyne.TextWrapWord

	setExplanation := func(text string) {
		fyne.Do(func() {
			explanationLabel.SetText(text)
		})
	}
	resetExplanation := func() {
		fyne.Do(func() {
			explanationLabel.SetText(defaultExplanation)
		})
	}

	copyLog := func() {
		logMu.Lock()
		text := logText
		logMu.Unlock()
		if text == "" {
			setExplanation("The log is empty. Run an operation to see its output here.")
			return
		}
		w.Clipboard().SetContent(text)
		setExplanation("Log copied to clipboard.")
	}
	tbCopyLog := newActionButton("Copy", theme.ContentCopyIcon(), copyLog,
		func() { setExplanation("Copy the current operation log to the clipboard.") }, resetExplanation)
	clearLog := func() {
		logMu.Lock()
		logText = ""
		logMu.Unlock()
		logGrid.SetText("")
		setExplanation("Log cleared. Any new output will continue to appear here.")
	}
	tbClearLog := newActionButton("Clear", theme.ContentClearIcon(), clearLog,
		func() { setExplanation("Clear the displayed log without stopping the current operation.") }, resetExplanation)

	quit := func() {
		busyMu.Lock()
		busy := isBusy
		busyMu.Unlock()
		if busy {
			dialog.ShowInformation("Operation in progress", "Please wait until the current operation finishes before closing Penguins GUI.", w)
			return
		}
		a.Quit()
	}
	w.SetCloseIntercept(quit)
	tbQuit := newActionButton("Exit", theme.LogoutIcon(), quit,
		func() { setExplanation("Close Penguins GUI after the current operation has finished.") }, resetExplanation)

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

			if setButtonsEnabled != nil {
				setButtonsEnabled(false)
			}
			openFolder.Hide()
			howToBoot.Hide()

			result.SetText(fmt.Sprintf("%s in progress…", title))
			logMu.Lock()
			logText = ""
			logMu.Unlock()
			logGrid.SetText("")

			go func() {
				err := eggsCLI.RunPrivileged(eggsPath, args, "", "", appendLog)
				fyne.Do(func() {
					busyMu.Lock()
					isBusy = false
					busyMu.Unlock()

					if setButtonsEnabled != nil {
						setButtonsEnabled(true)
					}

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

	needsSudoPassword := func() bool {
		if os.Geteuid() == 0 {
			return false
		}
		return exec.Command("sudo", "-n", "true").Run() != nil
	}

	createISO := func() {
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

		nest := defaultNest
		selectedMode := eggs.RemasterMode(mode.Selected)
		args := eggsCLI.RemasterArgs(selectedMode, nest)

		runRemaster := func(adminPassword, luksPassphrase string) {
			busyMu.Lock()
			if isBusy {
				busyMu.Unlock()
				return
			}
			isBusy = true
			busyMu.Unlock()

			if setButtonsEnabled != nil {
				setButtonsEnabled(false)
			}
			result.SetText("Remaster in progress…")
			openFolder.Hide()
			howToBoot.Hide()
			logMu.Lock()
			logText = ""
			logMu.Unlock()
			logGrid.SetText("")

			startedAt := time.Now()
			go func() {
				err := eggsCLI.RunPrivileged(eggsPath, args, adminPassword, luksPassphrase, appendLog)

				fyne.Do(func() {
					busyMu.Lock()
					isBusy = false
					busyMu.Unlock()

					if setButtonsEnabled != nil {
						setButtonsEnabled(true)
					}

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

		if selectedMode == eggs.ModeClone && needsSudoPassword() {
			pwEntry := widget.NewPasswordEntry()
			pwEntry.PlaceHolder = "Administrator password"
			items := []*widget.FormItem{
				widget.NewFormItem("", widget.NewIcon(theme.LoginIcon())),
				widget.NewFormItem("Password", pwEntry),
			}
			d := dialog.NewForm("Authentication Required", "Authenticate", "Cancel", items, func(ok bool) {
				if !ok {
					return
				}
				if strings.TrimSpace(pwEntry.Text) == "" {
					dialog.ShowError(errors.New("password cannot be empty"), w)
					return
				}
				runRemaster(pwEntry.Text, "")
			}, w)
			d.Resize(fyne.NewSize(380, 160))
			d.Show()
			return
		}

		if selectedMode == eggs.ModeEncrypted {
			luksEntry := widget.NewPasswordEntry()
			luksEntry.PlaceHolder = "Leave empty for default ('evolution')"
			luksConfirm := widget.NewPasswordEntry()
			luksConfirm.PlaceHolder = "Confirm LUKS passphrase"

			items := []*widget.FormItem{
				widget.NewFormItem("", widget.NewIcon(theme.AccountIcon())),
				widget.NewFormItem("LUKS Passphrase", luksEntry),
				widget.NewFormItem("Confirm Passphrase", luksConfirm),
			}

			var adminEntry *widget.Entry
			if needsSudoPassword() {
				adminEntry = widget.NewPasswordEntry()
				adminEntry.PlaceHolder = "Administrator password"
				items = append(items, widget.NewFormItem("Admin Password", adminEntry))
			}

			d := dialog.NewForm("Encrypted Clone Configuration", "Create ISO", "Cancel", items, func(ok bool) {
				if !ok {
					return
				}
				if luksEntry.Text != luksConfirm.Text {
					dialog.ShowError(errors.New("LUKS passphrases do not match"), w)
					return
				}
				adminPw := ""
				if adminEntry != nil {
					if strings.TrimSpace(adminEntry.Text) == "" {
						dialog.ShowError(errors.New("administrator password cannot be empty"), w)
						return
					}
					adminPw = adminEntry.Text
				}
				pass := luksEntry.Text
				if pass == "" {
					pass = "evolution"
				}
				runRemaster(adminPw, pass)
			}, w)
			d.Resize(fyne.NewSize(420, 220))
			d.Show()
			return
		}

		runRemaster("", "")
	}

	killAction := func() {
		runToolCommand(
			"Kill (delete previous ISOs)",
			[]string{"kill"},
			"Delete previous ISOs and clean the build nest (/home/eggs)?",
		)
	}

	cleanAction := func() {
		runToolCommand(
			"Clean system remnants",
			[]string{"tools", "clean"},
			"Clean log rotation, package manager cache, and host system remnants?",
		)
	}

	grubAction := func() {
		runToolCommand(
			"Configure GRUB loopback",
			[]string{"tools", "grub40"},
			"Generate GRUB configuration to boot ANY ISO via loopback?",
		)
	}

	repoAction := func() {
		runToolCommand(
			"Manage repository",
			[]string{"tools", "repo"},
			"Add or remove the official penguins-eggs repository?",
		)
	}

	skelAction := func() {
		runToolCommand(
			"Update /etc/skel",
			[]string{"tools", "skel"},
			"Create /etc/skel based on the current user's configurations?",
		)
	}

	tbCreateISO = newActionButton(
		"Create ISO",
		theme.DocumentCreateIcon(),
		func() {
			setExplanation("Create ISO: Remastering running system into a live ISO…")
			createISO()
		},
		func() {
			setExplanation("Create ISO: Remaster the running system into a live ISO image using the selected mode.")
		},
		resetExplanation,
	)
	tbCreateISO.Importance = widget.HighImportance

	tbKill = newActionButton(
		"Kill",
		theme.DeleteIcon(),
		func() {
			setExplanation("Kill (eggs kill): Delete previous ISOs and clean the build nest (/home/eggs).")
			killAction()
		},
		func() {
			setExplanation("Kill (eggs kill): Delete previous ISOs and clean the build nest (/home/eggs).")
		},
		resetExplanation,
	)

	tbClean = newActionButton(
		"Clean",
		theme.ViewRefreshIcon(),
		func() {
			setExplanation("Clean (eggs tools clean): Clean log rotation, package cache, and host remnants.")
			cleanAction()
		},
		func() {
			setExplanation("Clean (eggs tools clean): Clean log rotation, package manager cache, and host system remnants.")
		},
		resetExplanation,
	)

	tbGrub = newActionButton(
		"GRUB",
		theme.ComputerIcon(),
		func() {
			setExplanation("Configure GRUB loopback (eggs tools grub40): Generate GRUB config to boot ANY ISO via loopback.")
			grubAction()
		},
		func() {
			setExplanation("Configure GRUB loopback (eggs tools grub40): Generate GRUB configuration to boot ANY ISO via loopback.")
		},
		resetExplanation,
	)

	tbDocs = newActionButton(
		"Docs",
		theme.HelpIcon(),
		func() {
			setExplanation("Documentation: Opening official site (https://penguins-eggs.net)…")
			if u, err := url.Parse("https://penguins-eggs.net"); err == nil {
				_ = a.OpenURL(u)
			}
		},
		func() {
			setExplanation("Documentation: Open official Penguins' Eggs documentation website (https://penguins-eggs.net).")
		},
		resetExplanation,
	)

	setButtonsEnabled = func(enabled bool) {
		if enabled && detectErr == nil {
			mode.Enable()
			if tbCreateISO != nil {
				tbCreateISO.Enable()
			}
			if tbKill != nil {
				tbKill.Enable()
			}
			if tbClean != nil {
				tbClean.Enable()
			}
			if tbGrub != nil {
				tbGrub.Enable()
			}
			if createISOMenuItem != nil {
				createISOMenuItem.Disabled = false
				if w.MainMenu() != nil {
					w.MainMenu().Refresh()
				}
			}
		} else {
			mode.Disable()
			if tbCreateISO != nil {
				tbCreateISO.Disable()
			}
			if tbKill != nil {
				tbKill.Disable()
			}
			if tbClean != nil {
				tbClean.Disable()
			}
			if tbGrub != nil {
				tbGrub.Disable()
			}
			if createISOMenuItem != nil {
				createISOMenuItem.Disabled = true
				if w.MainMenu() != nil {
					w.MainMenu().Refresh()
				}
			}
		}
	}

	if detectErr != nil {
		setButtonsEnabled(false)
	}

	createISOMenuItem = fyne.NewMenuItem("Create ISO", func() {
		setExplanation("Create ISO: Remastering running system into a live ISO…")
		createISO()
	})
	if detectErr != nil {
		createISOMenuItem.Disabled = true
	}

	fileMenu := fyne.NewMenu("File",
		createISOMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", quit),
	)
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Copy log", copyLog),
		fyne.NewMenuItem("Clear log", clearLog),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Kill…", func() {
			setExplanation("Kill (eggs kill): Delete previous ISOs and clean the build nest (/home/eggs).")
			killAction()
		}),
	)
	toolsMenu := fyne.NewMenu("Tools",
		fyne.NewMenuItem("Clean system remnants (clean)…", func() {
			setExplanation("Clean (eggs tools clean): Clean log rotation, package manager cache, and host system remnants.")
			cleanAction()
		}),
		fyne.NewMenuItem("Configure GRUB loopback (grub40)…", func() {
			setExplanation("Configure GRUB loopback (eggs tools grub40): Generate GRUB configuration to boot ANY ISO via loopback.")
			grubAction()
		}),
		fyne.NewMenuItem("Manage repository (repo)…", func() {
			setExplanation("Manage repository (eggs tools repo): Add or remove the official penguins-eggs repository.")
			repoAction()
		}),
		fyne.NewMenuItem("Update /etc/skel (skel)…", func() {
			setExplanation("Update /etc/skel (eggs tools skel): Create /etc/skel based on current user's configurations.")
			skelAction()
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
				"Penguins GUI %s\n\n"+
					"Desktop interface for Penguins' Eggs.\n\n"+
					"Author: %s\n\n"+
					"Eggs: %s\n\n"+
					"Homepage: https://penguins-eggs.net",
				getGUIVersion(),
				author,
				eggsVersion,
			)
			about := dialog.NewCustom("About Penguins GUI", "Close", widget.NewLabel(aboutText), w)
			about.SetIcon(appIcon)
			about.Show()
		}),
	)
	var applyFontScale func(float32)

	presetItems := make([]*fyne.MenuItem, len(scalePresets))
	for i, preset := range scalePresets {
		p := preset
		item := fyne.NewMenuItem(p.Label, func() {
			applyFontScale(p.Scale)
		})
		item.Checked = (math.Abs(float64(currentScale-p.Scale)) < 0.05)
		presetItems[i] = item
	}

	applyFontScale = func(newScale float32) {
		newScale = clampScale(newScale)
		if math.Abs(float64(newScale-currentScale)) < 0.01 {
			return
		}
		currentScale = newScale
		a.Preferences().SetFloat(preferenceFontScale, float64(newScale))
		a.Settings().SetTheme(newScaledTheme(theme.DefaultTheme(), newScale))

		for i, preset := range scalePresets {
			presetItems[i].Checked = (math.Abs(float64(currentScale-preset.Scale)) < 0.05)
		}
		if w.MainMenu() != nil {
			w.MainMenu().Refresh()
		}

		fyne.Do(func() {
			min := w.Content().MinSize()
			curr := w.Canvas().Size()
			newW := curr.Width
			newH := curr.Height
			if min.Width > newW {
				newW = min.Width
			}
			if min.Height > newH {
				newH = min.Height
			}
			if newW != curr.Width || newH != curr.Height {
				w.Resize(fyne.NewSize(newW, newH))
			}
		})
	}

	zoomIn := func() {
		applyFontScale(currentScale + scaleStep)
	}
	zoomOut := func() {
		applyFontScale(currentScale - scaleStep)
	}
	resetZoom := func() {
		applyFontScale(defaultScale)
	}

	zoomInItem := fyne.NewMenuItem("Zoom In (Ctrl++)", zoomIn)
	zoomOutItem := fyne.NewMenuItem("Zoom Out (Ctrl+-)", zoomOut)
	resetZoomItem := fyne.NewMenuItem("Reset Zoom (Ctrl+0)", resetZoom)

	viewMenuItems := []*fyne.MenuItem{
		zoomInItem,
		zoomOutItem,
		resetZoomItem,
		fyne.NewMenuItemSeparator(),
	}
	viewMenuItems = append(viewMenuItems, presetItems...)
	viewMenu := fyne.NewMenu("View", viewMenuItems...)

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyPlus,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		zoomIn()
	})
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyEqual,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		zoomIn()
	})
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyMinus,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		zoomOut()
	})
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.Key0,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		resetZoom()
	})

	w.SetMainMenu(fyne.NewMainMenu(fileMenu, editMenu, viewMenu, toolsMenu, helpMenu))

	toolbarActions := container.NewHBox(
		tbCreateISO,
		widget.NewSeparator(),
		tbKill,
		tbClean,
		tbGrub,
		widget.NewSeparator(),
		tbDocs,
	)
	toolbarRow := container.NewBorder(nil, nil, toolbarActions, tbQuit, layout.NewSpacer())

	toolbarBox := container.NewVBox(
		toolbarRow,
		container.NewBorder(nil, nil, widget.NewIcon(theme.InfoIcon()), nil, explanationLabel),
		widget.NewSeparator(),
	)

	controls := container.NewVBox(
		widget.NewLabelWithStyle("Mode", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mode,
		modeHelp,
		widget.NewSeparator(),
		container.NewBorder(nil, nil,
			widget.NewLabelWithStyle("Operation log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(tbClearLog, tbCopyLog), nil),
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		result,
		container.NewHBox(layout.NewSpacer(), howToBoot, openFolder),
	)
	w.SetContent(container.NewPadded(container.NewBorder(container.NewVBox(toolbarBox, controls), footer, nil, nil, logScroll)))
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
	d.SetIcon(theme.HelpIcon())
	d.Resize(fyne.NewSize(620, 480))
	d.Show()
}
