package main

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/pieroproietti/penguins-gui/internal/tools/chef"
)

func TestCookDialogRecipeDiscoveryAndSelection(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	tmpDir, err := os.MkdirTemp("", "cook-dialog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	recipesDir := filepath.Join(tmpDir, "recipes", "de")
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

	// Test calling showCookDialog
	showCookDialog(w, "/usr/bin/chef", tmpDir, fakeRunCommand, fakeAppendLog)

	// Verify that fakeRunCommand is not called before user interaction
	if executedTitle != "" || executedExec != "" {
		t.Errorf("expected no execution before user action")
	}
	_ = executedArgs
	_ = logged
}
