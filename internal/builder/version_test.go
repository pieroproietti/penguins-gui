package builder

import (
	"os/exec"
	"testing"
)

func TestGitVersion(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	check := func(base, revision string) {
		t.Helper()
		gotBase, gotRevision := getGitVersion(dir)
		if gotBase != base || gotRevision != revision {
			t.Fatalf("got %s-%s; want %s-%s", gotBase, gotRevision, base, revision)
		}
	}
	check("0.1.0", "1")
	git("init")
	git("config", "user.name", "Builder test")
	git("config", "user.email", "builder@example.invalid")
	commit := func() { git("-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "fixture") }
	commit()
	check("0.1.0", "1")
	commit()
	check("0.1.0", "2")
	git("tag", "v1.2.3-rc_1")
	check("1.2.3.rc.1", "1")
	commit()
	check("1.2.3.rc.1", "1")
	commit()
	check("1.2.3.rc.1", "2")
}
