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

	"github.com/pieroproietti/penguins-gui/internal/tools/chef"
	"github.com/pieroproietti/penguins-gui/internal/tools/eggs"
	"github.com/pieroproietti/penguins-gui/internal/tools/system"

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
	applicationID      = "net.penguins-eggs.gui"
	defaultNest        = "/home/eggs"
	defaultAuthor      = "Piero Proietti <piero.proietti@gmail.com>"
	missingEggsMessage = "The eggs command was not found in PATH. Use Edit → Install penguins-egg CLI to configure the repository and install it. ISO creation will become available after installation."
)

//go:embed pkg/builder/assets/penguins-gui.svg
var iconSVG []byte

var (
	appIcon           = fyne.NewStaticResource("penguins-gui.svg", iconSVG)
	ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
	guiVersion        = "v26.10.21"
	author            = defaultAuthor
)

func getGUIVersion() string {
	if guiVersion != "" && guiVersion != "devel" {
		return guiVersion
	}
	return "v26.10.21"
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == system.GRUBHelperArg {
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "Expected Eggs executable and ISO path")
			os.Exit(1)
		}
		if err := system.ConfigureEggsGRUB(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("ISO boot entry saved; previous marked Eggs entries replaced.")
		return
	}
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
	grubBoot := system.DetectGRUBBoot()
	grubExplanation := "Boot ISO: Add the ISO to this computer's startup menu so you can select it the next time you restart."
	if !grubBoot.Detected {
		grubExplanation = grubBoot.Reason
	}

	mode := widget.NewRadioGroup([]string{
		string(eggs.ModeStandard),
		string(eggs.ModeClone),
		string(eggs.ModeEncrypted),
	}, nil)
	mode.Horizontal = true
	mode.Required = true
	mode.SetSelected(string(eggs.ModeStandard))

	modeHelp := widget.NewLabel(modeDescription(eggs.ModeStandard))
	modeHelp.Truncation = fyne.TextTruncateEllipsis
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

	const maxLogLines = 1000
	const uiThrottleInterval = 80 * time.Millisecond

	var logMu sync.Mutex
	var pendingBuf strings.Builder
	var logLines []string
	var incompleteLine string
	var logText string
	var uiUpdatePending bool
	var updateTimer *time.Timer

	applyPendingLogLocked := func(flushIncomplete bool) string {
		if pendingBuf.Len() > 0 || (flushIncomplete && incompleteLine != "") {
			chunk := pendingBuf.String()
			pendingBuf.Reset()

			combined := incompleteLine + chunk
			parts := strings.Split(combined, "\n")
			if flushIncomplete {
				incompleteLine = ""
				for i, p := range parts {
					if i == len(parts)-1 && p == "" {
						continue
					}
					logLines = append(logLines, p)
				}
			} else {
				incompleteLine = parts[len(parts)-1]
				if len(parts) > 1 {
					logLines = append(logLines, parts[:len(parts)-1]...)
				}
			}

			if len(logLines) > maxLogLines {
				trimmed := make([]string, maxLogLines)
				copy(trimmed, logLines[len(logLines)-maxLogLines:])
				logLines = trimmed
			}

			if incompleteLine != "" {
				if len(logLines) > 0 {
					logText = strings.Join(logLines, "\n") + "\n" + incompleteLine
				} else {
					logText = incompleteLine
				}
			} else {
				logText = strings.Join(logLines, "\n")
			}
		}
		return logText
	}

	renderLog := func() {
		logMu.Lock()
		if updateTimer != nil {
			updateTimer.Stop()
			updateTimer = nil
		}
		uiUpdatePending = false
		current := applyPendingLogLocked(false)
		logMu.Unlock()

		logGrid.SetText(current)
		logScroll.ScrollToBottom()
	}

	appendLog := func(text string) {
		clean := stripANSI(text)
		if clean == "" {
			return
		}
		logMu.Lock()
		pendingBuf.WriteString(clean)
		if !uiUpdatePending {
			uiUpdatePending = true
			updateTimer = time.AfterFunc(uiThrottleInterval, func() {
				fyne.Do(renderLog)
			})
		}
		logMu.Unlock()
	}

	flushLog := func() {
		done := make(chan struct{})
		fyne.Do(func() {
			logMu.Lock()
			if updateTimer != nil {
				updateTimer.Stop()
				updateTimer = nil
			}
			uiUpdatePending = false
			current := applyPendingLogLocked(true)
			logMu.Unlock()

			logGrid.SetText(current)
			logScroll.ScrollToBottom()
			close(done)
		})
		<-done
	}

	resetLog := func() {
		logMu.Lock()
		if updateTimer != nil {
			updateTimer.Stop()
			updateTimer = nil
		}
		uiUpdatePending = false
		pendingBuf.Reset()
		logLines = nil
		incompleteLine = ""
		logText = ""
		logMu.Unlock()
		fyne.Do(func() {
			logGrid.SetText("")
		})
	}

	var isBusy bool
	var busyMu sync.Mutex
	var createISOMenuItem *fyne.MenuItem
	var cookMenuItem *fyne.MenuItem
	var grubMenuItem *fyne.MenuItem
	var tbCreateISO, tbKill, tbClean, tbCook, tbGrub, tbDocs *actionButton
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
		_ = applyPendingLogLocked(false)
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
		resetLog()
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

	runCommand := func(title string, executable string, args []string, confirmPrompt string, successMessage string) {
		busyMu.Lock()
		if isBusy {
			busyMu.Unlock()
			dialog.ShowInformation("Busy", "Another operation is currently in progress. Please wait until it completes.", w)
			return
		}
		busyMu.Unlock()

		execute := func() {
			busyMu.Lock()
			if isBusy {
				busyMu.Unlock()
				dialog.ShowInformation("Busy", "Another operation is currently in progress.", w)
				return
			}
			isBusy = true
			busyMu.Unlock()

			if setButtonsEnabled != nil {
				setButtonsEnabled(false)
			}
			openFolder.Hide()
			howToBoot.Hide()

			result.SetText(fmt.Sprintf("%s in progress…", title))
			resetLog()

			go func() {
				err := eggsCLI.RunPrivileged(executable, args, "", "", appendLog)
				detectedPath, detectedVersion, detectedErr := eggsCLI.Detect()
				flushLog()
				fyne.Do(func() {
					eggsPath, eggsVersion, detectErr = detectedPath, detectedVersion, detectedErr
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
						message := fmt.Sprintf("%s completed successfully.", title)
						if successMessage != "" {
							message = successMessage
						}
						result.SetText(message)
						dialog.ShowInformation(title, message, w)
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

	runToolCommand := func(title string, args []string, prompt string, successMessage string) {
		if detectErr != nil {
			dialog.ShowInformation("penguins-eggs CLI not found", missingEggsMessage, w)
			return
		}
		runCommand(title, eggsPath, args, prompt, successMessage)
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
			resetLog()

			startedAt := time.Now()
			go func() {
				err := eggsCLI.RunPrivileged(eggsPath, args, adminPassword, luksPassphrase, appendLog)
				flushLog()

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
			"",
		)
	}

	cleanAction := func() {
		runToolCommand(
			"Clean system remnants",
			[]string{"tools", "clean"},
			"Clean log rotation, package manager cache, and host system remnants?",
			"",
		)
	}

	cookAction := func() {
		busyMu.Lock()
		if isBusy {
			busyMu.Unlock()
			dialog.ShowInformation("Busy", "Another operation is currently in progress. Please wait until it completes.", w)
			return
		}
		busyMu.Unlock()

		chefPath, err := exec.LookPath("chef")
		if err != nil {
			dialog.ShowInformation("penguins-chef CLI not found", "The chef command was not found in PATH.\nPlease install penguins-chef to use the Cook feature.", w)
			return
		}

		chefRoot, err := chef.GetChefRoot()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		openPicker := func() {
			showCookDialog(w, chefPath, chefRoot, runCommand, appendLog)
		}

		if chef.IsChefDirMissingOrEmpty(chefRoot) {
			prog := widget.NewProgressBarInfinite()
			msg := widget.NewLabel("The chef recipes directory (~/.chef) is missing or empty.\n\nDownloading recipes repository…")
			msg.Wrapping = fyne.TextWrapWord
			content := container.NewVBox(msg, prog)
			d := dialog.NewCustomWithoutButtons("Downloading Chef Recipes", container.NewPadded(content), w)
			d.SetIcon(theme.DownloadIcon())
			d.Resize(fyne.NewSize(450, 160))
			d.Show()

			go func() {
				err := chef.DownloadChef(chefRoot, appendLog)
				flushLog()
				fyne.Do(func() {
					d.Hide()
					if err != nil {
						dialog.ShowError(fmt.Errorf("failed to download recipes: %w", err), w)
						return
					}
					openPicker()
				})
			}()
			return
		}

		openPicker()
	}

	grubAction := func() {
		if !grubBoot.Detected {
			dialog.ShowInformation("Boot ISO", grubBoot.Reason, w)
			return
		}
		entries, err := os.ReadDir(defaultNest)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		var names []string
		for _, entry := range entries {
			if !strings.EqualFold(filepath.Ext(entry.Name()), ".iso") {
				continue
			}
			info, err := os.Stat(filepath.Join(defaultNest, entry.Name()))
			if err == nil && info.Mode().IsRegular() {
				names = append(names, entry.Name())
			}
		}
		if len(names) == 0 {
			dialog.ShowInformation("Boot ISO", "No ISO files found in "+defaultNest, w)
			return
		}
		selection := widget.NewSelect(names, nil)
		selection.SetSelected(names[0])
		picker := dialog.NewForm("Select ISO from "+defaultNest, "Continue", "Cancel",
			[]*widget.FormItem{widget.NewFormItem("ISO", selection)}, func(ok bool) {
				if !ok || selection.Selected == "" {
					return
				}
				isoPath := filepath.Join(defaultNest, selection.Selected)
				executable, err := os.Executable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				runCommand(
					"Boot ISO",
					executable,
					[]string{system.GRUBHelperArg, eggsPath, isoPath},
					fmt.Sprintf("Use this ISO in this computer's startup menu, replacing any previous Penguins' Eggs entries?\n\n%s\n\nAfter this operation, you must update GRUB before restarting to make the ISO available in the startup menu.", isoPath),
					"ISO boot entry saved. Update GRUB before restarting.\n\nOn Debian/Ubuntu, run in a terminal:\nsudo update-grub\n\nOn other distributions, use the GRUB update command provided by your distribution. Then restart and select the ISO from the startup menu.",
				)
			}, w)
		picker.Resize(fyne.NewSize(720, 180))
		picker.Show()
	}

	repoAction := func(action string) {
		title, prompt := "Install penguins-eggs CLI", "Add the official repository and signing keys, then automatically install Penguins’ Eggs?\n\nOn Arch/Manjaro this also performs the system upgrade required when refreshing package databases."
		if action == "rm" {
			title, prompt = "Remove native repository", "Remove the official Penguins’ Eggs repository? Shared RPM/Pacman signing keys will be retained."
		}
		executable, args, err := system.RepositoryCommand(action)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		runCommand(title, executable, args, prompt, "")
	}

	chefAction := func() {
		executable, args, err := system.ChefCommand()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		runCommand("Install penguins-chef", executable, args,
			"Install penguins-chef?\n\nPackages will be downloaded from your configured repositories. On Arch/Manjaro this also performs the system upgrade required when refreshing package databases. If penguins-chef is unavailable, add the native Penguins’ Eggs repository first.", "")
	}

	calamaresAction := func() {
		executable, args, err := system.CalamaresCommand()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		runCommand("Install Calamares", executable, args,
			"Install Calamares and the Qt/QML packages required by its slideshow?\n\nPackages will be downloaded from your configured repositories. If Calamares is unavailable, add the native Penguins’ Eggs repository first.", "")
	}

	skelAction := func() {
		runToolCommand(
			"Update /etc/skel",
			[]string{"tools", "skel"},
			"Create /etc/skel based on the current user's configurations?",
			"",
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

	tbCook = newActionButton(
		"Cook",
		theme.ColorPaletteIcon(),
		func() {
			setExplanation("Cook (penguins-chef): Select and apply desktop and system recipes.")
			cookAction()
		},
		func() {
			setExplanation("Cook (penguins-chef): Select and apply configuration and desktop recipes.")
		},
		resetExplanation,
	)

	tbGrub = newActionButton(
		"Boot ISO",
		theme.ComputerIcon(),
		func() {
			setExplanation(grubExplanation)
			grubAction()
		},
		func() {
			setExplanation(grubExplanation)
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
			if createISOMenuItem != nil {
				createISOMenuItem.Disabled = true
				if w.MainMenu() != nil {
					w.MainMenu().Refresh()
				}
			}
		}
		if tbCook != nil {
			if enabled {
				tbCook.Enable()
			} else {
				tbCook.Disable()
			}
		}
		if cookMenuItem != nil {
			cookMenuItem.Disabled = !enabled
		}
		grubEnabled := enabled && detectErr == nil && grubBoot.Detected
		if grubEnabled {
			tbGrub.Enable()
		} else {
			tbGrub.Disable()
		}
		if grubMenuItem != nil {
			grubMenuItem.Disabled = !grubEnabled
		}
		if w.MainMenu() != nil {
			w.MainMenu().Refresh()
		}
	}

	if detectErr != nil {
		setButtonsEnabled(false)
		result.SetText(missingEggsMessage)
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
	cookMenuItem = fyne.NewMenuItem("Cook…", cookAction)
	grubMenuItem = fyne.NewMenuItem("Boot ISO…", grubAction)
	setButtonsEnabled(true)
	editMenu := fyne.NewMenu("Edit",
		cookMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Install penguins-egg CLI", func() { repoAction("add") }),
		fyne.NewMenuItem("Install penguins-chef", chefAction),
		fyne.NewMenuItem("Install calamares", calamaresAction),
		fyne.NewMenuItem("Update /etc/skel", skelAction),
		grubMenuItem,
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

	w.SetMainMenu(fyne.NewMainMenu(fileMenu, editMenu, viewMenu, helpMenu))

	toolbarActions := container.NewHBox(
		tbCreateISO,
		widget.NewSeparator(),
		tbKill,
		tbClean,
		tbCook,
		tbGrub,
		widget.NewSeparator(),
		tbDocs,
	)
	toolbarRow := container.NewBorder(nil, nil, toolbarActions, tbQuit, layout.NewSpacer())

	statusBar := container.NewBorder(nil, nil, widget.NewIcon(theme.InfoIcon()), nil, explanationLabel)

	modeRow := container.NewBorder(
		nil, nil,
		container.NewHBox(
			mode,
			widget.NewSeparator(),
		),
		nil,
		modeHelp,
	)

	logHeader := container.NewBorder(nil, nil,
		widget.NewLabelWithStyle("Operation log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(tbClearLog, tbCopyLog), nil)

	topControls := container.NewVBox(
		toolbarRow,
		statusBar,
		widget.NewSeparator(),
		modeRow,
		widget.NewSeparator(),
		logHeader,
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		result,
		container.NewHBox(layout.NewSpacer(), howToBoot, openFolder),
	)
	w.SetContent(container.NewPadded(container.NewBorder(topControls, footer, nil, nil, logScroll)))
	if detectErr != nil {
		message := widget.NewLabel(missingEggsMessage)
		message.Wrapping = fyne.TextWrapWord
		warning := dialog.NewCustom("penguins-eggs CLI not found", "OK", message, w)
		warning.SetIcon(theme.WarningIcon())
		warning.Resize(fyne.NewSize(480, 180))
		warning.Show()
	}
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
