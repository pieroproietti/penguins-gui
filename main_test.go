package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestRemasterArgs(t *testing.T) {
	tests := []struct {
		name string
		mode remasterMode
		want []string
	}{
		{"standard", modeStandard, []string{"remaster", "--path", "/home/eggs"}},
		{"clone", modeClone, []string{"remaster", "--clone", "--path", "/home/eggs"}},
		{"encrypted", modeEncrypted, []string{"remaster", "--crypted", "--path", "/home/eggs"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := remasterArgs(test.mode, "/home/eggs"); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("remasterArgs() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestNewestISO(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "old.iso")
	newPath := filepath.Join(root, "isodir", "new.iso")
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if err := os.Chtimes(oldPath, started.Add(-time.Hour), started.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new iso"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := newestISO(root, started)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != newPath {
		t.Fatalf("newestISO() path = %q, want %q", got.Path, newPath)
	}
	if got.Size != int64(len("new iso")) {
		t.Fatalf("newestISO() size = %d", got.Size)
	}
}

func TestHumanSize(t *testing.T) {
	if got := humanSize(3 * 1024 * 1024); got != "3.0 MiB" {
		t.Fatalf("humanSize() = %q", got)
	}
}
