package builder

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFedoraRecipeVersionValidation(t *testing.T) {
	for _, tc := range []struct {
		base, revision string
		valid          bool
	}{
		{"26.9.23", "1", true},
		{"1.2.3.rc.1", "12", true},
		{"", "1", false},
		{"1.0%{lua:print(1)}", "1", false},
		{"1.0\nRequires: injected", "1", false},
		{"1.0", "1%{?dist}", false},
		{"1.0", "0", false},
		{"1.0", "-1", false},
	} {
		_, err := rpmRecipe(RecipeData{BaseVersion: tc.base, Rel: tc.revision}, "fedora")
		if (err == nil) != tc.valid {
			t.Errorf("fedoraRecipe(%q, %q): %v; want valid=%v", tc.base, tc.revision, err, tc.valid)
		}
	}
}

func TestOpenSUSERecipe(t *testing.T) {
	data := RecipeData{BaseVersion: "1.2.3", Rel: "4"}
	recipe, err := rpmRecipe(data, "opensuse")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Release: 4.opensuse", "Suggests: penguins-eggs", "Requires: polkit", "Requires: sudo", "Requires: curl", "Requires: xdg-utils"} {
		if !strings.Contains(string(recipe), want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, bad := range []RecipeData{{BaseVersion: "1%{bad}", Rel: "1"}, {BaseVersion: "1", Rel: "0"}} {
		if _, err := rpmRecipe(bad, "opensuse"); err == nil {
			t.Errorf("accepted invalid version: %+v", bad)
		}
	}
	if _, err := rpmRecipe(data, "unknown"); err == nil {
		t.Fatal("accepted unknown RPM family")
	}
}

// Exercise the real RPM toolchain with a small ELF fixture, without needing
// the GUI's graphics build dependencies or installing anything on the host.
func TestOpenSUSERPM(t *testing.T) {
	if _, err := exec.LookPath("rpmbuild"); err != nil {
		t.Skip("rpmbuild unavailable")
	}
	if _, err := exec.LookPath("rpm"); err != nil {
		t.Skip("rpm unavailable")
	}
	root := t.TempDir()
	binary, err := os.ReadFile("/usr/bin/true")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "penguins-gui"), binary, 0755); err != nil {
		t.Fatal(err)
	}
	stage, work, dist := filepath.Join(root, "stage"), filepath.Join(root, "work"), filepath.Join(root, "dist")
	if err := staging(root, stage); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dist, 0755); err != nil {
		t.Fatal(err)
	}
	artifact, err := packageRPM(root, work, stage, dist, RecipeData{BaseVersion: "1.2.3", Rel: "4"}, "opensuse")
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := output(root, "rpm", "-qp", "--queryformat", "%{NAME} %{VERSION} %{RELEASE}", artifact)
	if err != nil {
		t.Fatal(err)
	}
	if metadata != "penguins-gui 1.2.3 4.opensuse" {
		t.Fatalf("unexpected metadata: %s", metadata)
	}
	files, err := output(root, "rpm", "-qlp", artifact)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"/usr/bin/penguins-gui", "/usr/share/applications/penguins-gui.desktop", "/usr/share/icons/hicolor/scalable/apps/penguins-gui.svg", "/usr/share/polkit-1/rules.d/49-penguins-gui.rules"} {
		if !strings.Contains(files, want) {
			t.Errorf("RPM missing %s", want)
		}
	}
	requires, err := output(root, "rpm", "-qp", "--requires", artifact)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(requires, "penguins-eggs") {
		t.Fatal("Eggs should remain optional")
	}
	if !strings.Contains(requires, "libc.so") {
		t.Fatal("ELF library dependencies missing")
	}
}
