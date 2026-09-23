package builder

import "testing"

func TestFamilyFromRelease(t *testing.T) {
	for _, tc := range []struct{ release, want string }{
		{"ID=arch", "arch"},
		{"ID=manjaro\nID_LIKE=arch", "arch"},
		{"ID=ubuntu\nID_LIKE=debian", "debian"},
		{"ID=devuan", "debian"},
		{"ID=custom\nID_LIKE=\"other arch\"", "arch"},
		{"ID=debian\nID_LIKE=arch", "debian"},
		{"ID=fedora", ""},
		{"", ""},
	} {
		got, err := familyFromRelease(tc.release)
		if got != tc.want || (err != nil) != (tc.want == "") {
			t.Errorf("familyFromRelease(%q) = %q, %v; want %q", tc.release, got, err, tc.want)
		}
	}
}
