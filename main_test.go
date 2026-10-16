package main

import (
	"testing"

	"github.com/pieroproietti/penguins-gui/internal/tools/eggs"
)

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

func TestGUIVersionAndAuthor(t *testing.T) {
	if got := getGUIVersion(); got == "" {
		t.Fatalf("expected non-empty GUI version")
	}
	expectedAuthor := "Piero Proietti <piero.proietti@gmail.com>"
	if author != expectedAuthor {
		t.Fatalf("author = %q, want %q", author, expectedAuthor)
	}
}

func TestModeDescription(t *testing.T) {
	tests := []struct {
		mode eggs.RemasterMode
		want string
	}{
		{
			mode: eggs.ModeStandard,
			want: "Create a clean live system without user accounts or personal data.",
		},
		{
			mode: eggs.ModeClone,
			want: "Include users, configurations and data in an unencrypted clone.",
		},
		{
			mode: eggs.ModeEncrypted,
			want: "Create an encrypted clone; passphrase and parameters are requested by Eggs.",
		},
	}

	for _, tt := range tests {
		got := modeDescription(tt.mode)
		if got != tt.want {
			t.Errorf("modeDescription(%q) = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

