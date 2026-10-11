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
