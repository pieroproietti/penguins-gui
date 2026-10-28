package ui

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/pieroproietti/penguins-gui/internal/community"
	"github.com/pieroproietti/penguins-gui/internal/chef"
	"github.com/pieroproietti/penguins-gui/internal/eggs"
	"github.com/pieroproietti/penguins-gui/internal/system"
)

const (
	ApplicationID      = "net.penguins-eggs.gui"
	DefaultNest        = "/home/eggs"
	DefaultAuthor      = "Piero Proietti <piero.proietti@gmail.com>"
	missingEggsMessage = "The eggs command was not found in PATH. Use Edit → Install penguins-egg CLI to configure the repository and install it. ISO creation will become available after installation."
	defaultExplanation = "Select or hover over an action to view its description."
)

var (
	guiVersion = "v26.10.28"
	Author     = DefaultAuthor
)

// GetGUIVersion returns the current application version string.
func GetGUIVersion() string {
	if guiVersion != "" && guiVersion != "devel" {
		return guiVersion
	}
	return "v26.10.28"
}

// App coordinates the Fyne window, widgets, state, and external tool adapters.
type App struct {
	fyneApp      fyne.App
	win          fyne.Window
	appIcon      fyne.Resource
	currentScale float32

	eggsCLI         eggs.CLIAdapter
	eggsPath        string
	eggsVersion     string
	detectErr       error
	grubBoot        system.GRUBBootStatus
	grubExplanation string

	busyMu sync.Mutex
	isBusy bool

	logView          *LogView
	explanationLabel *widget.Label
	mode             *widget.RadioGroup
	result           *widget.Label
	openFolder       *widget.Button
	howToBoot        *widget.Button
	communityBtn     *widget.Button
	mainTabs         *container.AppTabs
	getCommunityData func() community.CommunityData

	createISOMenuItem *fyne.MenuItem
	cookMenuItem      *fyne.MenuItem
	grubMenuItem      *fyne.MenuItem
	communityMenuItem *fyne.MenuItem

	tbCreateISO *actionButton
	tbKill      *actionButton
	tbClean     *actionButton
	tbCook      *actionButton
	tbGrub      *actionButton
	tbDocs      *actionButton
	tbStats     *actionButton
	tbQuit      *actionButton
}

// Run initializes and starts the Penguins GUI desktop application.
func Run(iconSVG []byte) {
	a := app.NewWithID(ApplicationID)
	appIcon := fyne.NewStaticResource("penguins-gui.svg", iconSVG)
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

	explanationLabel := widget.NewLabel(defaultExplanation)
	explanationLabel.TextStyle = fyne.TextStyle{Italic: true}
	explanationLabel.Wrapping = fyne.TextWrapWord

	ui := &App{
		fyneApp:          a,
		win:              w,
		appIcon:          appIcon,
		currentScale:     currentScale,
		eggsCLI:          eggsCLI,
		eggsPath:         eggsPath,
		eggsVersion:      eggsVersion,
		detectErr:        detectErr,
		grubBoot:         grubBoot,
		grubExplanation:  grubExplanation,
		logView:          newLogView(),
		explanationLabel: explanationLabel,
	}

	w.SetCloseIntercept(ui.quit)

	remasterContent := ui.buildRemasterTab()
	communityPanel, getCommunityData := buildCommunityPanel(w, ui.updateCommunityStatus)
	ui.getCommunityData = getCommunityData
	communityTabContent := container.NewPadded(communityPanel)

	ui.updateCommunityStatus(ui.getCommunityData())

	ui.mainTabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Remaster", theme.DocumentCreateIcon(), remasterContent),
		container.NewTabItemWithIcon("Community", theme.AccountIcon(), communityTabContent),
	)
	mainContent := ui.buildMainContent(ui.mainTabs)

	ui.setupMenusAndShortcuts()
	ui.setButtonsEnabled(true)

	if ui.detectErr != nil {
		ui.setButtonsEnabled(false)
		ui.result.SetText(missingEggsMessage)
	}

	w.SetContent(mainContent)
	if ui.detectErr != nil {
		message := widget.NewLabel(missingEggsMessage)
		message.Wrapping = fyne.TextWrapWord
		warning := dialog.NewCustom("penguins-eggs CLI not found", "OK", message, w)
		warning.SetIcon(theme.WarningIcon())
		warning.Resize(fyne.NewSize(480, 180))
		warning.Show()
	}
	w.ShowAndRun()
}

func (ui *App) setExplanation(text string) {
	fyne.Do(func() {
		ui.explanationLabel.SetText(text)
	})
}

func (ui *App) resetExplanation() {
	fyne.Do(func() {
		ui.explanationLabel.SetText(defaultExplanation)
	})
}

func (ui *App) quit() {
	ui.busyMu.Lock()
	busy := ui.isBusy
	ui.busyMu.Unlock()
	if busy {
		dialog.ShowInformation("Operation in progress", "Please wait until the current operation finishes before closing Penguins GUI.", ui.win)
		return
	}
	ui.fyneApp.Quit()
}

func (ui *App) setButtonsEnabled(enabled bool) {
	if enabled && ui.detectErr == nil {
		ui.mode.Enable()
		if ui.tbCreateISO != nil {
			ui.tbCreateISO.Enable()
		}
		if ui.tbKill != nil {
			ui.tbKill.Enable()
		}
		if ui.tbClean != nil {
			ui.tbClean.Enable()
		}
		if ui.createISOMenuItem != nil {
			ui.createISOMenuItem.Disabled = false
			if ui.win.MainMenu() != nil {
				ui.win.MainMenu().Refresh()
			}
		}
	} else {
		ui.mode.Disable()
		if ui.tbCreateISO != nil {
			ui.tbCreateISO.Disable()
		}
		if ui.tbKill != nil {
			ui.tbKill.Disable()
		}
		if ui.tbClean != nil {
			ui.tbClean.Disable()
		}
		if ui.createISOMenuItem != nil {
			ui.createISOMenuItem.Disabled = true
			if ui.win.MainMenu() != nil {
				ui.win.MainMenu().Refresh()
			}
		}
	}
	if ui.tbCook != nil {
		if enabled {
			ui.tbCook.Enable()
		} else {
			ui.tbCook.Disable()
		}
	}
	if ui.cookMenuItem != nil {
		ui.cookMenuItem.Disabled = !enabled
	}
	if ui.communityMenuItem != nil {
		ui.communityMenuItem.Disabled = !enabled
	}
	grubEnabled := enabled && ui.detectErr == nil && ui.grubBoot.Detected
	if grubEnabled {
		ui.tbGrub.Enable()
	} else {
		ui.tbGrub.Disable()
	}
	if ui.grubMenuItem != nil {
		ui.grubMenuItem.Disabled = !grubEnabled
	}
	if ui.win.MainMenu() != nil {
		ui.win.MainMenu().Refresh()
	}
}

func (ui *App) runCommand(title string, executable string, args []string, confirmPrompt string, successMessage string) {
	ui.busyMu.Lock()
	if ui.isBusy {
		ui.busyMu.Unlock()
		dialog.ShowInformation("Busy", "Another operation is currently in progress. Please wait until it completes.", ui.win)
		return
	}
	ui.busyMu.Unlock()

	execute := func() {
		ui.busyMu.Lock()
		if ui.isBusy {
			ui.busyMu.Unlock()
			dialog.ShowInformation("Busy", "Another operation is currently in progress.", ui.win)
			return
		}
		ui.isBusy = true
		ui.busyMu.Unlock()

		ui.setButtonsEnabled(false)
		ui.openFolder.Hide()
		ui.howToBoot.Hide()

		ui.result.SetText(fmt.Sprintf("%s in progress…", title))
		ui.logView.Reset()

		go func() {
			err := ui.eggsCLI.RunPrivileged(executable, args, "", "", ui.logView.Append)
			detectedPath, detectedVersion, detectedErr := ui.eggsCLI.Detect()
			ui.logView.Flush()
			fyne.Do(func() {
				ui.eggsPath, ui.eggsVersion, ui.detectErr = detectedPath, detectedVersion, detectedErr
				ui.busyMu.Lock()
				ui.isBusy = false
				ui.busyMu.Unlock()

				ui.setButtonsEnabled(true)

				if err != nil {
					ui.result.SetText(fmt.Sprintf("%s failed: %v", title, err))
					dialog.ShowError(err, ui.win)
				} else {
					message := fmt.Sprintf("%s completed successfully.", title)
					if successMessage != "" {
						message = successMessage
					}
					ui.result.SetText(message)
					dialog.ShowInformation(title, message, ui.win)
				}
			})
		}()
	}

	if confirmPrompt != "" {
		dialog.ShowConfirm(title, confirmPrompt, func(ok bool) {
			if ok {
				execute()
			}
		}, ui.win)
	} else {
		execute()
	}
}

func (ui *App) runToolCommand(title string, args []string, prompt string, successMessage string) {
	if ui.detectErr != nil {
		dialog.ShowInformation("penguins-eggs CLI not found", missingEggsMessage, ui.win)
		return
	}
	ui.runCommand(title, ui.eggsPath, args, prompt, successMessage)
}

func (ui *App) createISO() {
	if ui.detectErr != nil {
		dialog.ShowError(errors.New("Penguins’ Eggs was not found in PATH"), ui.win)
		return
	}

	ui.busyMu.Lock()
	if ui.isBusy {
		ui.busyMu.Unlock()
		dialog.ShowInformation("Busy", "Another operation is currently in progress. Please wait until it completes.", ui.win)
		return
	}
	ui.busyMu.Unlock()

	nest := DefaultNest
	selectedMode := eggs.RemasterMode(ui.mode.Selected)
	args := ui.eggsCLI.RemasterArgs(selectedMode, nest)

	runRemaster := func(adminPassword, luksPassphrase string) {
		ui.busyMu.Lock()
		if ui.isBusy {
			ui.busyMu.Unlock()
			return
		}
		ui.isBusy = true
		ui.busyMu.Unlock()

		ui.setButtonsEnabled(false)
		ui.result.SetText("Remaster in progress…")
		ui.openFolder.Hide()
		ui.howToBoot.Hide()
		ui.logView.Reset()

		if ui.getCommunityData != nil {
			comm := ui.getCommunityData()
			if comm.Enabled {
				nick := comm.Nickname
				if nick == "" {
					nick = "Anonymous"
				}
				var details []string
				if comm.Contact != "" {
					details = append(details, comm.Contact)
				}
				if comm.SourceURL != "" {
					details = append(details, "source: "+comm.SourceURL)
				}
				if len(details) > 0 {
					ui.logView.Append(fmt.Sprintf("[community] Community participation active (@%s, %s - %s)\n\n", nick, comm.Location, strings.Join(details, " - ")))
				} else {
					ui.logView.Append(fmt.Sprintf("[community] Community participation active (@%s, %s)\n\n", nick, comm.Location))
				}
			} else {
				ui.logView.Append("[community] Community participation: disabled\n\n")
			}
		}

		startedAt := time.Now()
		go func() {
			err := ui.eggsCLI.RunPrivileged(ui.eggsPath, args, adminPassword, luksPassphrase, ui.logView.Append)
			ui.logView.Flush()

			fyne.Do(func() {
				ui.busyMu.Lock()
				ui.isBusy = false
				ui.busyMu.Unlock()

				ui.setButtonsEnabled(true)

				if err != nil {
					ui.result.SetText(fmt.Sprintf("Remaster failed: %v", err))
					dialog.ShowError(err, ui.win)
				} else {
					artifact, findErr := ui.eggsCLI.NewestISO(nest, startedAt)
					if findErr != nil {
						ui.result.SetText("Remaster completed, but no new ISO was found in the selected directory.")
					} else {
						ui.result.SetText(fmt.Sprintf("ISO created: %s (%s)", artifact.Path, eggs.HumanSize(artifact.Size)))
						ui.openFolder.OnTapped = func() {
							if err := openDirectory(filepath.Dir(artifact.Path)); err != nil {
								dialog.ShowError(err, ui.win)
							}
						}
						ui.openFolder.Show()
						ui.howToBoot.Show()
					}

					if ui.getCommunityData != nil {
						comm := ui.getCommunityData()
						if comm.Enabled {
							go func() {
								if syncErr := community.SyncCommunityProfile(comm); syncErr != nil {
									ui.logView.Append(fmt.Sprintf("[community] Warning: failed to sync profile to community server: %v\n", syncErr))
								} else {
									ui.logView.Append(fmt.Sprintf("[community] Profile successfully published to %s\n", community.Endpoint()))
								}
							}()
						}
					}

					dialog.ShowInformation("Remaster completed", "Penguins’ Eggs finished successfully.\n\n"+ui.result.Text, ui.win)
				}
			})
		}()
	}

	if selectedMode == eggs.ModeClone && needsSudoPassword() {
		showCloneAuthDialog(ui.win, func(adminPw string) {
			runRemaster(adminPw, "")
		})
		return
	}

	if selectedMode == eggs.ModeEncrypted {
		showEncryptedCloneDialog(ui.win, needsSudoPassword(), runRemaster)
		return
	}

	runRemaster("", "")
}

func (ui *App) killAction() {
	ui.runToolCommand(
		"Kill (delete previous ISOs)",
		[]string{"kill"},
		"Delete previous ISOs and clean the build nest (/home/eggs)?",
		"",
	)
}

func (ui *App) cleanAction() {
	ui.runToolCommand(
		"Clean system remnants",
		[]string{"tools", "clean"},
		"Clean log rotation, package manager cache, and host system remnants?",
		"",
	)
}

func (ui *App) cookAction() {
	ui.busyMu.Lock()
	if ui.isBusy {
		ui.busyMu.Unlock()
		dialog.ShowInformation("Busy", "Another operation is currently in progress. Please wait until it completes.", ui.win)
		return
	}
	ui.busyMu.Unlock()

	chefPath, err := exec.LookPath("chef")
	if err != nil {
		dialog.ShowInformation("penguins-chef CLI not found", "The chef command was not found in PATH.\nPlease install penguins-chef to use the Cook feature.", ui.win)
		return
	}

	chefRoot, err := chef.GetChefRoot()
	if err != nil {
		dialog.ShowError(err, ui.win)
		return
	}

	openPicker := func() {
		showCookDialog(ui.win, chefPath, chefRoot, ui.runCommand, ui.logView.Append)
	}

	if chef.IsChefDirMissingOrEmpty(chefRoot) {
		prog := widget.NewProgressBarInfinite()
		msg := widget.NewLabel("The chef recipes directory (~/.chef) is missing or empty.\n\nDownloading recipes repository…")
		msg.Wrapping = fyne.TextWrapWord
		content := container.NewVBox(msg, prog)
		d := dialog.NewCustomWithoutButtons("Downloading Chef Recipes", container.NewPadded(content), ui.win)
		d.SetIcon(theme.DownloadIcon())
		d.Resize(fyne.NewSize(450, 160))
		d.Show()

		go func() {
			err := chef.DownloadChef(chefRoot, ui.logView.Append)
			ui.logView.Flush()
			fyne.Do(func() {
				d.Hide()
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to download recipes: %w", err), ui.win)
					return
				}
				openPicker()
			})
		}()
		return
	}

	openPicker()
}

func (ui *App) grubAction() {
	if !ui.grubBoot.Detected {
		dialog.ShowInformation("Boot ISO", ui.grubBoot.Reason, ui.win)
		return
	}
	names, err := eggs.ListISOs(DefaultNest)
	if err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	if len(names) == 0 {
		dialog.ShowInformation("Boot ISO", "No ISO files found in "+DefaultNest, ui.win)
		return
	}
	showGRUBISOPickerDialog(ui.win, DefaultNest, names, func(selectedName string) {
		isoPath := filepath.Join(DefaultNest, selectedName)
		executable, err := os.Executable()
		if err != nil {
			dialog.ShowError(err, ui.win)
			return
		}
		ui.runCommand(
			"Boot ISO",
			executable,
			[]string{system.GRUBHelperArg, ui.eggsPath, isoPath},
			fmt.Sprintf("Use this ISO in this computer's startup menu, replacing any previous Penguins' Eggs entries?\n\n%s\n\nAfter this operation, you must update GRUB before restarting to make the ISO available in the startup menu.", isoPath),
			"ISO boot entry saved. Update GRUB before restarting.\n\nOn Debian/Ubuntu, run in a terminal:\nsudo update-grub\n\nOn other distributions, use the GRUB update command provided by your distribution. Then restart and select the ISO from the startup menu.",
		)
	})
}

func (ui *App) repoAction(action string) {
	title, prompt := "Install penguins-eggs CLI", "Add the official repository and signing keys, then automatically install Penguins’ Eggs?\n\nOn Arch/Manjaro this also performs the system upgrade required when refreshing package databases."
	if action == "rm" {
		title, prompt = "Remove native repository", "Remove the official Penguins’ Eggs repository? Shared RPM/Pacman signing keys will be retained."
	}
	executable, args, err := system.RepositoryCommand(action)
	if err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	ui.runCommand(title, executable, args, prompt, "")
}

func (ui *App) chefAction() {
	executable, args, err := system.ChefCommand()
	if err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	ui.runCommand("Install penguins-chef", executable, args,
		"Install penguins-chef?\n\nPackages will be downloaded from your configured repositories. On Arch/Manjaro this also performs the system upgrade required when refreshing package databases. If penguins-chef is unavailable, add the native Penguins’ Eggs repository first.", "")
}

func (ui *App) calamaresAction() {
	executable, args, err := system.CalamaresCommand()
	if err != nil {
		dialog.ShowError(err, ui.win)
		return
	}
	ui.runCommand("Install Calamares", executable, args,
		"Install Calamares and the Qt/QML packages required by its slideshow?\n\nPackages will be downloaded from your configured repositories. If Calamares is unavailable, add the native Penguins’ Eggs repository first.", "")
}

func (ui *App) skelAction() {
	ui.runToolCommand(
		"Update /etc/skel",
		[]string{"tools", "skel"},
		"Create /etc/skel based on the current user's configurations?",
		"",
	)
}
