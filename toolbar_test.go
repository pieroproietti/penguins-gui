package main

import (
	"testing"

	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestActionButtonHoverAndTap(t *testing.T) {
	test.NewApp()

	var hoveredIn, hoveredOut, tapped bool
	btn := newActionButton("Test", theme.ConfirmIcon(), func() {
		tapped = true
	}, func() {
		hoveredIn = true
	}, func() {
		hoveredOut = true
	})

	btn.MouseIn(&desktop.MouseEvent{})
	if !hoveredIn {
		t.Errorf("expected hoveredIn to be true")
	}

	btn.MouseOut()
	if !hoveredOut {
		t.Errorf("expected hoveredOut to be true")
	}

	test.Tap(btn)
	if !tapped {
		t.Errorf("expected tapped to be true")
	}
}

func TestCookAndBootISOButtons(t *testing.T) {
	test.NewApp()

	btnClean := newActionButton("Clean", theme.ViewRefreshIcon(), func() {}, nil, nil)
	btnCook := newActionButton("Cook", theme.ColorPaletteIcon(), func() {}, nil, nil)
	btnBootISO := newActionButton("Boot ISO", theme.ComputerIcon(), func() {}, nil, nil)

	if btnClean.Text != "Clean" {
		t.Errorf("expected Clean button text, got %q", btnClean.Text)
	}
	if btnCook.Text != "Cook" {
		t.Errorf("expected Cook button text, got %q", btnCook.Text)
	}
	if btnBootISO.Text != "Boot ISO" {
		t.Errorf("expected Boot ISO button text, got %q", btnBootISO.Text)
	}
}
