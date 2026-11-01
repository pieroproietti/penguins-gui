/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/pieroproietti/penguins-gui/internal/eggs"
)

func (ui *App) buildRemasterTab() *fyne.Container {
	ui.mode = widget.NewRadioGroup([]string{
		string(eggs.ModeStandard),
		string(eggs.ModeClone),
		string(eggs.ModeEncrypted),
	}, nil)
	ui.mode.Horizontal = true
	ui.mode.Required = true
	ui.mode.SetSelected(string(eggs.ModeStandard))

	modeHelp := widget.NewLabel(eggs.ModeDescription(eggs.ModeStandard))
	modeHelp.Truncation = fyne.TextTruncateEllipsis
	ui.mode.OnChanged = func(selected string) {
		modeHelp.SetText(eggs.ModeDescription(eggs.RemasterMode(selected)))
	}

	ui.result = widget.NewLabel("")
	ui.result.Wrapping = fyne.TextWrapBreak

	ui.openFolder = widget.NewButton("Open ISO folder", func() {})
	ui.openFolder.Hide()

	ui.howToBoot = widget.NewButton("How to boot ISO…", func() {
		showBootGuideDialog(ui.win)
	})
	ui.howToBoot.Hide()

	copyLog := func() {
		text := ui.logView.Text()
		if text == "" {
			ui.setExplanation("The log is empty. Run an operation to see its output here.")
			return
		}
		ui.win.Clipboard().SetContent(text)
		ui.setExplanation("Log copied to clipboard.")
	}
	tbCopyLog := newActionButton("Copy", theme.ContentCopyIcon(), copyLog,
		func() { ui.setExplanation("Copy the current operation log to the clipboard.") }, ui.resetExplanation)

	clearLog := func() {
		ui.logView.Reset()
		ui.setExplanation("Log cleared. Any new output will continue to appear here.")
	}
	tbClearLog := newActionButton("Clear", theme.ContentClearIcon(), clearLog,
		func() { ui.setExplanation("Clear the displayed log without stopping the current operation.") }, ui.resetExplanation)

	toolbarRow := ui.buildToolbarRow()
	statusBar := container.NewBorder(nil, nil, widget.NewIcon(theme.InfoIcon()), nil, ui.explanationLabel)

	modeRow := container.NewBorder(
		nil, nil,
		ui.mode,
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
		ui.result,
		container.NewHBox(layout.NewSpacer(), ui.howToBoot, ui.openFolder),
	)

	return container.NewPadded(container.NewBorder(topControls, footer, nil, nil, ui.logView.Scroll))
}
