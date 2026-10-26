package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDistroFamily(t *testing.T) {
	for _, tc := range []struct{ release, want string }{
		{"ID=debian", "debian"}, {"ID=devuan", "debian"},
		{"ID=linuxmint\nID_LIKE=\"ubuntu debian\"", "ubuntu"},
		{"ID=endeavouros\nID_LIKE=arch", "arch"},
		{"ID=almalinux\nID_LIKE=\"rhel centos fedora\"", "fedora"},
		{"ID=opensuse-tumbleweed", "suse"}, {"ID=alpine", "alpine"},
	} {
		got, err := distroFamily(tc.release)
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %q, %v; want %q", tc.release, got, err, tc.want)
		}
	}
	if _, err := distroFamily("ID=unknown"); err == nil {
		t.Fatal("unsupported distribution accepted")
	}
}

func TestCalamaresInstaller(t *testing.T) {
	for _, tc := range []struct {
		name, family, qt, fail, want, absent string
		wantErr                              bool
	}{
		{"Debian Qt5", "debian", "5", "", "qml-module-qtquick-controls", "qml6-module", false},
		{"Ubuntu Qt6", "ubuntu", "6", "", "language-selector-common", "qml-module-qtquick2", false},
		{"Arch", "arch", "6", "", "qt6-declarative", "-Sy", false},
		{"Fedora", "fedora", "6", "", "qt6-qtdeclarative", "apt-get", false},
		{"SUSE", "suse", "6", "", "qt6-declarative-imports", "apt-get", false},
		{"Alpine", "alpine", "6", "", "calamares-mod-unpackfs", "apt-get", false},
		{"update failure", "debian", "6", "update", "apt-get update", "install --yes", true},
		{"install failure", "debian", "6", "calamares", "install --yes calamares", "qml6-module", true},
		{"slideshow failure", "debian", "6", "qml6-module", "qml6-module-qtquick", "installed successfully", true},
		{"unknown Qt", "debian", "unknown", "", "calamares", "qml6-module", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			log := filepath.Join(dir, "calls")
			stub := `#!/bin/sh
name=${0##*/}
echo "$name $*" >> "$CALL_LOG"
if [ "$name" = ldd ]; then echo "libQt${QT_VERSION}Core.so"; exit 0; fi
if [ -n "$FAIL_MATCH" ]; then
 case "$*" in *"$FAIL_MATCH"*) exit 42 ;; esac
fi
`
			for _, name := range []string{"apt-get", "pacman", "dnf", "zypper", "apk", "calamares", "ldd"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(stub), 0755); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("/bin/sh", "-c", calamaresScript, "test-installer", tc.family)
			cmd.Env = []string{"PATH=" + dir, "CALL_LOG=" + log, "QT_VERSION=" + tc.qt, "FAIL_MATCH=" + tc.fail}
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, output: %s", err, out)
			}
			calls, readErr := os.ReadFile(log)
			if readErr != nil {
				t.Fatal(readErr)
			}
			all := string(calls) + string(out)
			if !strings.Contains(all, tc.want) || strings.Contains(all, tc.absent) {
				t.Fatalf("unexpected execution: %s", all)
			}
			if tc.family == "ubuntu" && !strings.Contains(all, "qml6-module-qtquick-controls") {
				t.Fatal("missing Qt6 controls")
			}
		})
	}
}
