/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package system

import (
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
