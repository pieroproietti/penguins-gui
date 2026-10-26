package system

import (
	"encoding/binary"
	"os"
	"testing"
	"unicode/utf16"
)

func TestDetectGRUBBoot(t *testing.T) {
	params := func(loader byte) []byte {
		data := make([]byte, 4096)
		data[0x210] = loader
		return data
	}
	loaderInfo := func(name string) []byte {
		data := make([]byte, 4)
		for _, unit := range utf16.Encode([]rune(name + "\x00")) {
			data = binary.LittleEndian.AppendUint16(data, unit)
		}
		return data
	}
	for _, tt := range []struct {
		name   string
		params []byte
		info   []byte
		want   bool
	}{
		{"GRUB BIOS", params(0x70), nil, true},
		{"GRUB version nibble", params(0x72), nil, true},
		{"GRUB EFI", nil, loaderInfo("GRUB 2.12"), true},
		{"systemd boot", nil, loaderInfo("systemd-boot 257"), false},
		{"Syslinux", params(0x31), nil, false},
		{"unknown loader", params(0xff), nil, false},
		{"no evidence", nil, nil, false},
		{"truncated kernel data", []byte{0x70}, nil, false},
		{"truncated EFI data", nil, []byte{7, 0, 0, 0, 'G'}, false},
		{"unterminated EFI data", nil, loaderInfo("GRUB")[:12], false},
		{"GRUB substring is not identification", nil, loaderInfo("not-GRUB 1.0"), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			read := func(path string) ([]byte, error) {
				var data []byte
				switch path {
				case "/sys/kernel/boot_params/data":
					data = tt.params
				case loaderInfoPath:
					data = tt.info
				default:
					t.Fatalf("unexpected read: %s", path)
				}
				if data == nil {
					return nil, os.ErrPermission
				}
				return data, nil
			}
			got := detectGRUBBoot(read)
			if got.Detected != tt.want {
				t.Fatalf("Detected = %v, want %v (%s)", got.Detected, tt.want, got.Reason)
			}
			if !got.Detected && got.Reason == "" {
				t.Fatal("unavailable action needs an explanation")
			}
		})
	}
}
