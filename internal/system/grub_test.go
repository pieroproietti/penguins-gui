package system

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testGRUBEntry(name string) string {
	return grubStart + name + " <<<\nmenuentry \"test\" {\n}\n" + grubEnd + name + " <<<\n"
}

func TestReplaceEggsBlocks(t *testing.T) {
	old, next := testGRUBEntry("old.iso"), testGRUBEntry("new.iso")
	prefix := "#!/bin/sh\nexec tail -n +3 $0\nmenuentry 'other OS' {}\n"
	suffix := "# user content without final newline"
	got, err := replaceEggsBlocks(prefix+old+"# between\n"+old+suffix, next)
	if err != nil || got != prefix+next+"# between\n"+suffix {
		t.Fatalf("unexpected replacement: %q, %v", got, err)
	}
	again, err := replaceEggsBlocks(got, next)
	if err != nil || again != got {
		t.Fatalf("not idempotent: %q, %v", again, err)
	}
	got, err = replaceEggsBlocks(prefix, next)
	if err != nil || got != prefix+next {
		t.Fatalf("first entry: %q, %v", got, err)
	}
}

func TestEggsBlocksRejectAmbiguousMarkers(t *testing.T) {
	for _, content := range []string{
		grubStart + "a.iso <<<\n",
		grubEnd + "a.iso <<<\n",
		grubStart + "a.iso <<<\n" + testGRUBEntry("b.iso"),
		grubStart + "a.iso <<<\n" + grubEnd + "b.iso <<<\n",
		grubStart + "a.iso\n",
		"  " + testGRUBEntry("a.iso"),
	} {
		if _, err := replaceEggsBlocks(content, testGRUBEntry("new.iso")); err == nil {
			t.Fatalf("accepted %q", content)
		}
	}
}

func TestConfigureEggsGRUB(t *testing.T) {
	for _, scenario := range []string{"success", "generation failure", "missing block", "wrong ISO", "multiple blocks", "bad existing markers", "concurrent edit"} {
		t.Run(scenario, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "40_custom")
			original := "#!/bin/sh\nexec tail -n +3 $0\n" + testGRUBEntry("old.iso") + "# untouched\n"
			if scenario == "bad existing markers" {
				original += grubStart + "broken.iso <<<\n"
			}
			if err := os.WriteFile(path, []byte(original), 0751); err != nil {
				t.Fatal(err)
			}
			// Reproduce an executable backup left by an older GUI version.
			if err := os.WriteFile(path+".penguins-gui.bak", []byte(original), 0751); err != nil {
				t.Fatal(err)
			}
			expected := original
			called := false
			err := configureEggsGRUB(path, "/home/eggs/new.iso", func() ([]byte, error) {
				called = true
				switch scenario {
				case "generation failure":
					return nil, errors.New("inspection failed")
				case "missing block":
					return []byte("log only"), nil
				case "wrong ISO":
					return []byte(testGRUBEntry("other.iso")), nil
				case "multiple blocks":
					return []byte(testGRUBEntry("new.iso") + testGRUBEntry("other.iso")), nil
				case "concurrent edit":
					expected += "# edited\n"
					if err := os.WriteFile(path, []byte(expected), 0751); err != nil {
						t.Fatal(err)
					}
				}
				return []byte("\x1b[36m[coa]\x1b[0m\n# diagnostics\n" + testGRUBEntry("new.iso") + "\nmore logs\n"), nil
			})
			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if scenario != "success" {
				if err == nil || string(got) != expected {
					t.Fatalf("failure changed file: %q, %v", got, err)
				}
				if scenario == "bad existing markers" && called {
					t.Fatal("generated despite invalid existing markers")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(got), grubStart) != 1 || !strings.Contains(string(got), "new.iso") || !strings.HasSuffix(string(got), "# untouched\n") {
				t.Fatalf("bad output: %q", got)
			}
			backup, err := os.ReadFile(path + ".penguins-gui.bak")
			if err != nil || string(backup) != original {
				t.Fatalf("bad backup: %q, %v", backup, err)
			}
			backupInfo, err := os.Stat(path + ".penguins-gui.bak")
			if err != nil {
				t.Fatal(err)
			}
			if backupInfo.Mode().Perm() != 0640 {
				t.Fatalf("backup must preserve read/write permissions without execute bits: %v", backupInfo.Mode())
			}
			info, _ := os.Stat(path)
			if info.Mode().Perm() != 0751 {
				t.Fatalf("changed permissions: %v", info.Mode())
			}
		})
	}
}
