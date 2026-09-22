package main

import "testing"

func TestHumanSize(t *testing.T) {
	if got := humanSize(3 * 1024 * 1024); got != "3.0 MiB" {
		t.Fatalf("humanSize() = %q", got)
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[36m[parser]\x1b[0m Profile loaded successfully."
	want := "[parser] Profile loaded successfully."
	if got := stripANSI(input); got != want {
		t.Fatalf("stripANSI() = %q, want %q", got, want)
	}
}
