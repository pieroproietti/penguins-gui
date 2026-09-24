package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryFamily(t *testing.T) {
	for _, tc := range []struct{ release, want string }{
		{"ID=debian", "debian"}, {"ID=manjaro\nID_LIKE=arch", "manjaro"},
		{"ID=fedora", "fedora"}, {"ID=almalinux\nID_LIKE=\"rhel centos fedora\"", "el9"},
		{"ID=opensuse-leap", "suse"}, {"ID=alpine", "alpine"},
	} {
		got, err := repositoryFamily(tc.release)
		if err != nil || got != tc.want {
			t.Fatalf("%q: %q, %v; want %q", tc.release, got, err, tc.want)
		}
	}
	if _, _, err := RepositoryCommand("add; injected"); err == nil {
		t.Fatal("invalid action accepted")
	}
}

// Exercise real file edits in a temporary filesystem, replacing only external
// network/key tools. The scripts never see the host's repository files.
func TestRepositoryLifecycle(t *testing.T) {
	for _, family := range []string{"debian", "ubuntu", "arch", "manjaro", "fedora", "el9", "suse", "alpine"} {
		t.Run(family, func(t *testing.T) {
			root, script := repositoryFixture(t)
			run := func(action string) {
				cmd := exec.Command("/bin/sh", "-c", script, "test", family, action)
				cmd.Env = append(os.Environ(), "PATH="+root+"/bin:/usr/bin:/bin")
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s: %v: %s", action, err, out)
				}
			}
			path, marker := "", ""
			switch family {
			case "debian", "ubuntu":
				path, marker = "/etc/apt/sources.list.d/penguins-repos.sources", "URIs: https://penguins-eggs.net/repos/deb"
			case "arch", "manjaro":
				path, marker = "/etc/pacman.conf", "[penguins-eggs]"
			case "fedora", "el9":
				path, marker = "/etc/yum.repos.d/penguins-eggs.repo", "[penguins-eggs]"
			case "suse":
				path, marker = "/etc/zypp/repos.d/penguins-eggs.repo", "[penguins-eggs]"
			case "alpine":
				path, marker = "/etc/apk/repositories", "https://penguins-eggs.net/repos/alpine/"
			}
			run("add")
			run("add")
			data, err := os.ReadFile(root + path)
			if err != nil || strings.Count(string(data), marker) != 1 {
				t.Fatalf("not idempotent: %s, %v", data, err)
			}
			if family == "manjaro" && !strings.Contains(string(data), "/repos/manjaro") {
				t.Fatal("wrong Manjaro channel")
			}
			if family == "el9" && !strings.Contains(string(data), "/rpm/el9") {
				t.Fatal("wrong EL channel")
			}
			run("rm")
			run("rm")
			data, err = os.ReadFile(root + path)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if strings.Contains(string(data), marker) {
				t.Fatalf("repository not removed: %s", data)
			}
			for _, name := range []string{"/etc/pacman.conf", "/etc/apk/repositories", "/etc/apt/sources.list.d/unrelated.list"} {
				data, err := os.ReadFile(root + name)
				if err != nil || !strings.Contains(string(data), "unrelated") {
					t.Fatalf("unrelated repository changed: %s: %v", name, err)
				}
			}
		})
	}
}

func TestRepositoryDownloadFailurePreservesConfiguration(t *testing.T) {
	root, script := repositoryFixture(t)
	key := root + "/usr/share/keyrings/penguins-repos.gpg"
	if err := os.WriteFile(key, []byte("existing key"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(root+"/bin/curl", []byte("#!/bin/sh\nexit 22\n"), 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", "-c", script, "test", "debian", "add")
	cmd.Env = append(os.Environ(), "PATH="+root+"/bin:/usr/bin:/bin")
	if err := cmd.Run(); err == nil {
		t.Fatal("download failure ignored")
	}
	data, err := os.ReadFile(key)
	if err != nil || string(data) != "existing key" {
		t.Fatalf("key damaged: %s, %v", data, err)
	}
	if _, err := os.Stat(root + "/etc/apt/sources.list.d/penguins-repos.sources"); !os.IsNotExist(err) {
		t.Fatal("repository enabled after failed download")
	}
}

func repositoryFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"bin", "etc/apt/sources.list.d", "usr/share/keyrings", "etc/apk/keys"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"etc/pacman.conf":                       "[unrelated]\nServer = https://example.org/arch\n",
		"etc/apk/repositories":                  "https://example.org/unrelated\n",
		"etc/apt/sources.list.d/unrelated.list": "# unrelated\n",
		"bin/curl":                              "#!/bin/sh\nwhile [ $# -gt 0 ]; do if [ \"$1\" = -o ]; then shift; printf 'test key' > \"$1\"; exit; fi; shift; done\nexit 1\n",
		"bin/gpg":                               "#!/bin/sh\nwhile [ $# -gt 0 ]; do if [ \"$1\" = --output ]; then shift; printf 'dearmored key' > \"$1\"; exit; fi; shift; done\nexit 1\n",
		"bin/rpm":                               "#!/bin/sh\nexit 0\n",
		"bin/pacman-key":                        "#!/bin/sh\nexit 0\n",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0755); err != nil {
			t.Fatal(err)
		}
	}
	script := strings.NewReplacer("/etc/", root+"/etc/", "/usr/share/keyrings/", root+"/usr/share/keyrings/").Replace(repositoryScript)
	return root, script
}

func TestEggsInstallWithoutEggs(t *testing.T) {
	for _, family := range []string{"debian", "ubuntu", "arch", "manjaro", "fedora", "el9", "suse", "alpine"} {
		t.Run(family, func(t *testing.T) {
			root := t.TempDir()
			for _, name := range []string{"apt-get", "pacman", "dnf", "zypper", "apk"} {
				if err := os.WriteFile(filepath.Join(root, name), []byte("#!/bin/sh\necho \"$*\"\n"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("/bin/sh", "-c", eggsInstallScript, "test", family)
			cmd.Env = []string{"PATH=" + root}
			out, err := cmd.CombinedOutput()
			if err != nil || !strings.Contains(string(out), "penguins-eggs") {
				t.Fatalf("%v: %s", err, out)
			}
		})
	}
}

func TestRepositorySetupAutomaticallyInstallsEggs(t *testing.T) {
	for _, tc := range []struct {
		name, action                                      string
		downloadFails, installFails, wantInstall, wantErr bool
	}{
		{name: "add installs Eggs", action: "add", wantInstall: true},
		{name: "repository failure stops installation", action: "add", downloadFails: true, wantErr: true},
		{name: "installation failure is reported", action: "add", installFails: true, wantInstall: true, wantErr: true},
		{name: "remove does not install Eggs", action: "rm"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := repositoryFixture(t)
			script := strings.NewReplacer("/etc/", root+"/etc/", "/usr/share/keyrings/", root+"/usr/share/keyrings/").Replace(repositorySetupScript)
			stub := `#!/bin/sh
# The repository and key must exist before refreshing or installing.
test -f "$TEST_ROOT/etc/apt/sources.list.d/penguins-repos.sources" || exit 90
test -f "$TEST_ROOT/usr/share/keyrings/penguins-repos.gpg" || exit 91
echo "$*" >> "$TEST_ROOT/package-calls"
if [ "$1" = install ] && [ "$INSTALL_FAILS" = true ]; then exit 42; fi
`
			if err := os.WriteFile(root+"/bin/apt-get", []byte(stub), 0755); err != nil {
				t.Fatal(err)
			}
			if tc.downloadFails {
				if err := os.WriteFile(root+"/bin/curl", []byte("#!/bin/sh\nexit 22\n"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			failure := "false"
			if tc.installFails {
				failure = "true"
			}
			cmd := exec.Command("/bin/sh", "-c", script, "test", "debian", tc.action)
			cmd.Env = append(os.Environ(), "PATH="+root+"/bin:/usr/bin:/bin", "TEST_ROOT="+root, "INSTALL_FAILS="+failure)
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, output: %s", err, out)
			}
			calls, err := os.ReadFile(root + "/package-calls")
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if got := strings.Contains(string(calls), "install --yes penguins-eggs"); got != tc.wantInstall {
				t.Fatalf("unexpected package commands: %s", calls)
			}
		})
	}
}
