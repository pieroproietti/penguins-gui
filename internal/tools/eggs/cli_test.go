package eggs

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
		mode RemasterMode
		want []string
	}{
		{"standard", ModeStandard, []string{"remaster", "--path", "/home/eggs"}},
		{"clone", ModeClone, []string{"remaster", "--clone", "--path", "/home/eggs"}},
		{"encrypted", ModeEncrypted, []string{"remaster", "--crypted", "--path", "/home/eggs"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := (CLIAdapter{}).RemasterArgs(test.mode, "/home/eggs"); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("(CLIAdapter{}).RemasterArgs() = %#v, want %#v", got, test.want)
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

	got, err := (CLIAdapter{}).NewestISO(root, started)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != newPath {
		t.Fatalf("(CLIAdapter{}).NewestISO() path = %q, want %q", got.Path, newPath)
	}
	if got.Size != int64(len("new iso")) {
		t.Fatalf("(CLIAdapter{}).NewestISO() size = %d", got.Size)
	}
}
