package eggs

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	newPath := filepath.Join(root, "new.iso")
	nestedPath := filepath.Join(root, "isodir", "nested.iso")
	if err := os.MkdirAll(filepath.Dir(nestedPath), 0o755); err != nil {
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
	if err := os.WriteFile(nestedPath, []byte("nested iso"), 0o644); err != nil {
		t.Fatal(err)
	}
	newer := started.Add(time.Minute)
	if err := os.Chtimes(nestedPath, newer, newer); err != nil {
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

func TestScanLinesOrCR(t *testing.T) {
	input := "Line 1\nLine 2\r\nProgress 10%\rProgress 20%\rProgress 100%\nFinal line"
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(scanLinesOrCR)

	var got []string
	for scanner.Scan() {
		got = append(got, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	want := []string{
		"Line 1",
		"Line 2",
		"Progress 10%",
		"Progress 20%",
		"Progress 100%",
		"Final line",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scanLinesOrCR got %#v, want %#v", got, want)
	}
}
