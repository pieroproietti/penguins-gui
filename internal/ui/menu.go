package ui

import (
	"fmt"
	"math"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *App) setupMenusAndShortcuts() {
	ui.createISOMenuItem = fyne.NewMenuItem("Create ISO", func() {
		ui.setExplanation("Create ISO: Remastering running system into a live ISO…")
		ui.createISO()
	})
	if ui.detectErr != nil {
		ui.createISOMenuItem.Disabled = true
	}

	fileMenu := fyne.NewMenu("File",
		ui.createISOMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", ui.quit),
	)

	ui.cookMenuItem = fyne.NewMenuItem("Cook…", ui.cookAction)
	ui.grubMenuItem = fyne.NewMenuItem("Boot ISO…", ui.grubAction)
	ui.communityMenuItem = fyne.NewMenuItem("Community…", func() {
		if ui.mainTabs != nil {
			ui.mainTabs.SelectIndex(1)
		}
	})

	editMenu := fyne.NewMenu("Edit",
		ui.communityMenuItem,
		ui.cookMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Install Penguins Tools", func() { ui.repoAction("add") }),
		fyne.NewMenuItem("Update /etc/skel", ui.skelAction),
		ui.grubMenuItem,
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Testing & Booting the ISO…", func() {
			showBootGuideDialog(ui.win)
		}),
		fyne.NewMenuItem("Documentation", func() {
			if u, err := url.Parse("https://penguins-eggs.net"); err == nil {
				_ = ui.fyneApp.OpenURL(u)
			}
		}),
		fyne.NewMenuItem("Stats (Census)", func() {
			if u, err := url.Parse("https://penguins-eggs.net/api/census"); err == nil {
				_ = ui.fyneApp.OpenURL(u)
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
				GetGUIVersion(),
				Author,
				ui.eggsVersion,
			)
			about := dialog.NewCustom("About Penguins GUI", "Close", widget.NewLabel(aboutText), ui.win)
			about.SetIcon(ui.appIcon)
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
		item.Checked = (math.Abs(float64(ui.currentScale-p.Scale)) < 0.05)
		presetItems[i] = item
	}

	applyFontScale = func(newScale float32) {
		newScale = clampScale(newScale)
		if math.Abs(float64(newScale-ui.currentScale)) < 0.01 {
			return
		}
		ui.currentScale = newScale
		ui.fyneApp.Preferences().SetFloat(preferenceFontScale, float64(newScale))
		ui.fyneApp.Settings().SetTheme(newScaledTheme(theme.DefaultTheme(), newScale))

		for i, preset := range scalePresets {
			presetItems[i].Checked = (math.Abs(float64(ui.currentScale-preset.Scale)) < 0.05)
		}
		if ui.win.MainMenu() != nil {
			ui.win.MainMenu().Refresh()
		}

		fyne.Do(func() {
			min := ui.win.Content().MinSize()
			curr := ui.win.Canvas().Size()
			newW := curr.Width
			newH := curr.Height
			if min.Width > newW {
				newW = min.Width
			}
			if min.Height > newH {
				newH = min.Height
			}
			if newW != curr.Width || newH != curr.Height {
				ui.win.Resize(fyne.NewSize(newW, newH))
			}
		})
	}

	zoomIn := func() {
		applyFontScale(ui.currentScale + scaleStep)
	}
	zoomOut := func() {
		applyFontScale(ui.currentScale - scaleStep)
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

	ui.win.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyPlus,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		zoomIn()
	})
	ui.win.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyEqual,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		zoomIn()
	})
	ui.win.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyMinus,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		zoomOut()
	})
	ui.win.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.Key0,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(_ fyne.Shortcut) {
		resetZoom()
	})

	ui.win.SetMainMenu(fyne.NewMainMenu(fileMenu, editMenu, viewMenu, helpMenu))
}
