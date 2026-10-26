package system

import (
	"encoding/binary"
	"os"
	"strings"
	"unicode/utf16"
)

const loaderInfoPath = "/sys/firmware/efi/efivars/LoaderInfo-4a67b082-0a4c-41cf-b6c7-440b29bb8c4f"

// GRUBBootStatus identifies the loader for the current boot, not installed packages.
type GRUBBootStatus struct {
	Detected bool
	Reason   string
}

// DetectGRUBBoot reads kernel and EFI boot information without requiring privileges.
// Missing or unreadable evidence leaves the action unavailable.
func DetectGRUBBoot() GRUBBootStatus {
	return detectGRUBBoot(os.ReadFile)
}

func detectGRUBBoot(readFile func(string) ([]byte, error)) GRUBBootStatus {
	// The x86 boot protocol assigns loader ID 7 to GRUB. The low nibble
	// contains its version. This also works for legacy BIOS boots.
	// https://www.kernel.org/doc/html/latest/arch/x86/boot.html
	if data, err := readFile("/sys/kernel/boot_params/data"); err == nil && len(data) > 0x210 {
		if data[0x210]>>4 == 7 {
			return GRUBBootStatus{Detected: true}
		}
	}

	// EFI loaders can identify themselves via the Boot Loader Interface.
	// efivarfs prefixes the UTF-16LE payload with four attribute bytes.
	// https://systemd.io/BOOT_LOADER_INTERFACE/
	if data, err := readFile(loaderInfoPath); err == nil {
		if name := decodeLoaderInfo(data); name != "" {
			fields := strings.Fields(name)
			if len(fields) > 0 && strings.EqualFold(fields[0], "GRUB") {
				return GRUBBootStatus{Detected: true}
			}
			return GRUBBootStatus{Reason: "Boot without USB requires GRUB. The reported boot loader is " + name + "."}
		}
	}
	return GRUBBootStatus{Reason: "Boot without USB is unavailable because startup through GRUB could not be confirmed on this system."}
}

func decodeLoaderInfo(data []byte) string {
	if len(data) < 6 || (len(data)-4)%2 != 0 {
		return ""
	}
	var units []uint16
	for i := 4; i < len(data); i += 2 {
		unit := binary.LittleEndian.Uint16(data[i:])
		if unit == 0 {
			return strings.TrimSpace(string(utf16.Decode(units)))
		}
		units = append(units, unit)
	}
	return ""
}
