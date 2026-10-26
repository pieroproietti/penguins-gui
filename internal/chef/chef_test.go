package chef

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetChefRoot(t *testing.T) {
	root, err := GetChefRoot()
	if err != nil {
		t.Fatalf("unexpected error getting chef root: %v", err)
	}
	if filepath.Base(root) != ".chef" {
		t.Errorf("expected root to end in .chef, got %s", root)
	}
}

func TestIsChefDirMissingOrEmpty(t *testing.T) {
	// Non-existent directory
	tmpNonExistent := filepath.Join(os.TempDir(), "nonexistent-chef-dir-test-12345")
	if !IsChefDirMissingOrEmpty(tmpNonExistent) {
		t.Errorf("expected non-existent directory to be recognized as missing/empty")
	}

	// Empty directory
	tmpEmpty, err := os.MkdirTemp("", "empty-chef-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpEmpty)

	if !IsChefDirMissingOrEmpty(tmpEmpty) {
		t.Errorf("expected empty directory to be recognized as missing/empty")
	}

	// Directory with only .git
	gitDir := filepath.Join(tmpEmpty, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if !IsChefDirMissingOrEmpty(tmpEmpty) {
		t.Errorf("expected directory with only .git to be recognized as empty")
	}

	// Directory with a yaml file
	yamlFile := filepath.Join(tmpEmpty, "recipe.yaml")
	if err := os.WriteFile(yamlFile, []byte("name: test"), 0o644); err != nil {
		t.Fatal(err)
	}
	if IsChefDirMissingOrEmpty(tmpEmpty) {
		t.Errorf("expected directory with yaml file not to be recognized as missing/empty")
	}
}

func TestFindRecipesAndGroupByCategory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chef-recipes-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	deDir := filepath.Join(tmpDir, "recipes", "de")
	costumesDir := filepath.Join(tmpDir, "recipes", "costumes", "colibri")
	baseDir := filepath.Join(tmpDir, "recipes", "base")

	for _, d := range []string{deDir, costumesDir, baseDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	xfceContent := `version: 1
name: xfce4-desktop
description: Fast XFCE desktop
profiles:
  debian:
    packages: [xfce4]
  archlinux:
    packages: [xfce4]
`
	if err := os.WriteFile(filepath.Join(deDir, "xfce.yaml"), []byte(xfceContent), 0o644); err != nil {
		t.Fatal(err)
	}

	colibriContent := `version: 1
name: colibri
description: Elegant colibri costume
sysroot: sysroot
`
	if err := os.WriteFile(filepath.Join(costumesDir, "colibri.yaml"), []byte(colibriContent), 0o644); err != nil {
		t.Fatal(err)
	}

	baseContent := `version: 1
name: base
description: Minimal base system
`
	if err := os.WriteFile(filepath.Join(baseDir, "base.yml"), []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}

	recipes, err := FindRecipes(filepath.Join(tmpDir, "recipes"))
	if err != nil {
		t.Fatalf("unexpected error finding recipes: %v", err)
	}

	if len(recipes) != 3 {
		t.Fatalf("expected 3 recipes, got %d", len(recipes))
	}

	groups := GroupByCategory(recipes)
	if len(groups) != 3 {
		t.Fatalf("expected 3 category groups, got %d", len(groups))
	}

	// Categories should be sorted: "base", "costumes", "de"
	if groups[0].Name != "base" || groups[1].Name != "costumes" || groups[2].Name != "de" {
		t.Errorf("unexpected category order: %v, %v, %v", groups[0].Name, groups[1].Name, groups[2].Name)
	}

	// Check xfce recipe details
	deGroup := groups[2]
	if len(deGroup.Recipes) != 1 {
		t.Fatalf("expected 1 recipe in de group, got %d", len(deGroup.Recipes))
	}
	r := deGroup.Recipes[0]
	if r.Name != "xfce4-desktop" {
		t.Errorf("expected name 'xfce4-desktop', got %q", r.Name)
	}
	if r.Description != "Fast XFCE desktop" {
		t.Errorf("expected description 'Fast XFCE desktop', got %q", r.Description)
	}
	if len(r.Profiles) != 2 || r.Profiles[0] != "archlinux" || r.Profiles[1] != "debian" {
		t.Errorf("expected profiles [archlinux, debian], got %v", r.Profiles)
	}
}
