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
