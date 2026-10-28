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
	"math"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/pieroproietti/penguins-gui/internal/chef"
)

func TestStripANSI(t *testing.T) {
	input := "\x1b[36m[parser]\x1b[0m Profile loaded successfully."
	want := "[parser] Profile loaded successfully."
	if got := stripANSI(input); got != want {
		t.Fatalf("stripANSI() = %q, want %q", got, want)
	}
}

func TestGUIVersionAndAuthor(t *testing.T) {
	if got := GetGUIVersion(); got == "" {
		t.Fatalf("expected non-empty GUI version")
	}
	expectedAuthor := "Piero Proietti <piero.proietti@gmail.com>"
	if Author != expectedAuthor {
		t.Fatalf("Author = %q, want %q", Author, expectedAuthor)
	}
}

func TestScaledThemeSize(t *testing.T) {
	base := theme.DefaultTheme()
	defaultTextSize := base.Size(theme.SizeNameText)

	scales := []float32{0.85, 1.0, 1.15, 1.3, 1.5, 2.0}
	for _, s := range scales {
		th := newScaledTheme(base, s)
		got := th.Size(theme.SizeNameText)
		want := float32(math.Round(float64(defaultTextSize*s)*100) / 100)
		diff := float64(got - want)
		if math.Abs(diff) > 0.05 {
			t.Errorf("scale %v: got %v, want %v", s, got, want)
		}
	}
}

func TestScaledThemeUnwrap(t *testing.T) {
	base := theme.DefaultTheme()
	th1 := newScaledTheme(base, 1.2)
	th2 := newScaledTheme(th1, 1.5)

	st, ok := th2.(*scaledTheme)
	if !ok {
		t.Fatalf("expected *scaledTheme, got %T", th2)
	}
	if _, isInnerScaled := st.base.(*scaledTheme); isInnerScaled {
		t.Errorf("base theme should be unwrapped, but got nested *scaledTheme")
	}
}

func TestScaledThemeClamping(t *testing.T) {
	thLow := newScaledTheme(nil, 0.1)
	if st, ok := thLow.(*scaledTheme); !ok || st.scale != minScale {
		t.Errorf("expected scale clamped to minScale %v, got %v", minScale, st.scale)
	}

	thHigh := newScaledTheme(nil, 5.0)
	if st, ok := thHigh.(*scaledTheme); !ok || st.scale != maxScale {
		t.Errorf("expected scale clamped to maxScale %v, got %v", maxScale, st.scale)
	}
}

func TestScaledThemeColorAndFont(t *testing.T) {
	test.NewApp()
	base := theme.DefaultTheme()
	th := newScaledTheme(base, 1.25)

	if th.Font(fyne.TextStyle{}) == nil {
		t.Errorf("expected non-nil font")
	}

	darkBg := th.Color(theme.ColorNameBackground, theme.VariantDark)
	lightBg := th.Color(theme.ColorNameBackground, theme.VariantLight)
	if darkBg == nil || lightBg == nil {
		t.Errorf("expected valid background colors")
	}
	if darkBg == lightBg {
		t.Errorf("expected dark and light variant backgrounds to differ")
	}
}

func TestPreferenceFontScale(t *testing.T) {
	app := test.NewApp()
	pref := app.Preferences()

	pref.SetFloat(preferenceFontScale, 1.30)
	got := float32(pref.FloatWithFallback(preferenceFontScale, 1.0))
	if math.Abs(float64(got-1.30)) > 0.01 {
		t.Fatalf("expected 1.30, got %v", got)
	}
}

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
	btnDocs := newActionButton("Docs", theme.HelpIcon(), func() {}, nil, nil)
	btnStats := newActionButton("Stats", theme.InfoIcon(), func() {}, nil, nil)

	if btnClean.Text != "Clean" {
		t.Errorf("expected Clean button text, got %q", btnClean.Text)
	}
	if btnCook.Text != "Cook" {
		t.Errorf("expected Cook button text, got %q", btnCook.Text)
	}
	if btnBootISO.Text != "Boot ISO" {
		t.Errorf("expected Boot ISO button text, got %q", btnBootISO.Text)
	}
	if btnDocs.Text != "Docs" {
		t.Errorf("expected Docs button text, got %q", btnDocs.Text)
	}
	if btnStats.Text != "Stats" {
		t.Errorf("expected Stats button text, got %q", btnStats.Text)
	}
}

func TestBuildCommunityPanel(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))

	w := test.NewWindow(nil)
	defer w.Close()

	panel, getData := buildCommunityPanel(w, nil)
	if panel == nil {
		t.Fatalf("expected non-nil panel")
	}
	if getData == nil {
		t.Fatalf("expected non-nil getData function")
	}

	data := getData()
	if data.Enabled {
		t.Errorf("expected initially disabled data")
	}
}

func TestCookDialogRecipeDiscoveryAndSelection(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	tmpDir, err := os.MkdirTemp("", "cook-dialog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	recipesDir := filepath.Join(tmpDir, "recipes", "desktop")
	if err := os.MkdirAll(recipesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	recipeFile := filepath.Join(recipesDir, "plasma.yaml")
	recipeContent := `version: 1
name: plasma-desktop
description: KDE Plasma Desktop Environment
profiles:
  debian:
    packages: [plasma-desktop]
`
	if err := os.WriteFile(recipeFile, []byte(recipeContent), 0o644); err != nil {
		t.Fatal(err)
	}

	recipes, err := chef.FindRecipes(filepath.Join(tmpDir, "recipes"))
	if err != nil {
		t.Fatalf("unexpected error finding recipes: %v", err)
	}
	if len(recipes) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(recipes))
	}
	if recipes[0].Name != "plasma-desktop" {
		t.Errorf("expected plasma-desktop, got %q", recipes[0].Name)
	}

	var executedTitle, executedExec string
	var executedArgs []string
	fakeRunCommand := func(title string, executable string, args []string, prompt string, success string) {
		executedTitle = title
		executedExec = executable
		executedArgs = args
	}

	var logged []string
	fakeAppendLog := func(s string) {
		logged = append(logged, s)
	}

	showCookDialog(w, "/usr/bin/chef", tmpDir, fakeRunCommand, fakeAppendLog)

	if executedTitle != "" || executedExec != "" {
		t.Errorf("expected no execution before user action")
	}
	_ = executedArgs
	_ = logged
}
