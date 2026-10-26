package ui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// actionButton extends widget.Button to support hover detection for status/info descriptions.
type actionButton struct {
	widget.Button
	onHoverIn  func()
	onHoverOut func()
}

func newActionButton(label string, icon fyne.Resource, tapped func(), hoverIn func(), hoverOut func()) *actionButton {
	b := &actionButton{
		onHoverIn:  hoverIn,
		onHoverOut: hoverOut,
	}
	b.Text = label
	b.Icon = icon
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

func (b *actionButton) MouseIn(e *desktop.MouseEvent) {
	b.Button.MouseIn(e)
	if b.onHoverIn != nil {
		b.onHoverIn()
	}
}

func (b *actionButton) MouseMoved(e *desktop.MouseEvent) {
	b.Button.MouseMoved(e)
}

func (b *actionButton) MouseOut() {
	b.Button.MouseOut()
	if b.onHoverOut != nil {
		b.onHoverOut()
	}
}

func (ui *App) buildToolbarRow() *fyne.Container {
	ui.tbCreateISO = newActionButton(
		"Create ISO",
		theme.DocumentCreateIcon(),
		func() {
			ui.setExplanation("Create ISO: Remastering running system into a live ISO…")
			ui.createISO()
		},
		func() {
			ui.setExplanation("Create ISO: Remaster the running system into a live ISO image using the selected mode.")
		},
		ui.resetExplanation,
	)
	ui.tbCreateISO.Importance = widget.HighImportance

	ui.tbKill = newActionButton(
		"Kill",
		theme.DeleteIcon(),
		func() {
			ui.setExplanation("Kill (eggs kill): Delete previous ISOs and clean the build nest (/home/eggs).")
			ui.killAction()
		},
		func() {
			ui.setExplanation("Kill (eggs kill): Delete previous ISOs and clean the build nest (/home/eggs).")
		},
		ui.resetExplanation,
	)

	ui.tbClean = newActionButton(
		"Clean",
		theme.ViewRefreshIcon(),
		func() {
			ui.setExplanation("Clean (eggs tools clean): Clean log rotation, package cache, and host remnants.")
			ui.cleanAction()
		},
		func() {
			ui.setExplanation("Clean (eggs tools clean): Clean log rotation, package manager cache, and host system remnants.")
		},
		ui.resetExplanation,
	)

	ui.tbCook = newActionButton(
		"Cook",
		theme.ColorPaletteIcon(),
		func() {
			ui.setExplanation("Cook (penguins-chef): Select and apply desktop and system recipes.")
			ui.cookAction()
		},
		func() {
			ui.setExplanation("Cook (penguins-chef): Select and apply configuration and desktop recipes.")
		},
		ui.resetExplanation,
	)

	ui.tbGrub = newActionButton(
		"Boot ISO",
		theme.ComputerIcon(),
		func() {
			ui.setExplanation(ui.grubExplanation)
			ui.grubAction()
		},
		func() {
			ui.setExplanation(ui.grubExplanation)
		},
		ui.resetExplanation,
	)

	ui.tbDocs = newActionButton(
		"Docs",
		theme.HelpIcon(),
		func() {
			ui.setExplanation("Documentation: Opening official site (https://penguins-eggs.net)…")
			if u, err := url.Parse("https://penguins-eggs.net"); err == nil {
				_ = ui.fyneApp.OpenURL(u)
			}
		},
		func() {
			ui.setExplanation("Documentation: Open official Penguins' Eggs documentation website (https://penguins-eggs.net).")
		},
		ui.resetExplanation,
	)

	ui.tbStats = newActionButton(
		"Stats",
		theme.InfoIcon(),
		func() {
			ui.setExplanation("Stats: Opening Penguins' Eggs census (https://penguins-eggs.net/api/census)…")
			if u, err := url.Parse("https://penguins-eggs.net/api/census"); err == nil {
				_ = ui.fyneApp.OpenURL(u)
			}
		},
		func() {
			ui.setExplanation("Stats: Open Penguins' Eggs census and statistics (https://penguins-eggs.net/api/census).")
		},
		ui.resetExplanation,
	)

	ui.tbCommunity = newActionButton(
		"Community",
		theme.AccountIcon(),
		func() {
			ui.setExplanation("Community: Configure optional participation before remastering.")
			if ui.mainTabs != nil {
				ui.mainTabs.SelectIndex(1)
			}
		},
		func() {
			ui.setExplanation("Community: Configure your profile and optional participation before remastering.")
		},
		ui.resetExplanation,
	)

	ui.tbQuit = newActionButton(
		"Exit",
		theme.LogoutIcon(),
		ui.quit,
		func() {
			ui.setExplanation("Close Penguins GUI after the current operation has finished.")
		},
		ui.resetExplanation,
	)

	toolbarActions := container.NewHBox(
		ui.tbCreateISO,
		widget.NewSeparator(),
		ui.tbKill,
		ui.tbClean,
		ui.tbCook,
		ui.tbGrub,
		ui.tbCommunity,
		widget.NewSeparator(),
		ui.tbDocs,
		ui.tbStats,
	)
	return container.NewBorder(nil, nil, toolbarActions, ui.tbQuit, layout.NewSpacer())
}
