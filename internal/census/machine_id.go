package census

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func deriveUUIDFromMachineID(machineID string) string {
	normalized := strings.ToLower(strings.TrimSpace(machineID))
	sum := sha256.Sum256([]byte("penguins-eggs-census:" + normalized))
	var b [16]byte
	copy(b[:], sum[:16])
	b[6] = (b[6] & 0x0f) | 0x50 // Version 5
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func readAndDeriveMachineID(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "uninitialized" || strings.Trim(raw, "0") == "" {
		return "", false
	}
	return deriveUUIDFromMachineID(raw), true
}

// GetMachineUUID returns the anonymized UUID derived from machine-id or stored census.json.
func GetMachineUUID() string {
	if custom := strings.TrimSpace(os.Getenv("EGGS_MACHINE_ID_PATH")); custom != "" {
		if id, ok := readAndDeriveMachineID(custom); ok {
			return id
		}
	}

	// 1. Check census.json (etc or user config) which contains the exact eggs uuid
	for _, p := range []string{
		"/etc/penguins-eggs/census.json",
		"/etc/penguins-eggs.d/census.json",
	} {
		if data, err := os.ReadFile(p); err == nil {
			var m map[string]interface{}
			if err := json.Unmarshal(data, &m); err == nil {
				if u, ok := m["uuid"].(string); ok && len(u) >= 8 {
					return strings.TrimSpace(u)
				}
			}
		}
	}

	if userPath := getEggsConfigPath(); userPath != "" {
		dir := filepath.Dir(userPath)
		for _, name := range []string{"census.json", ".census_id"} {
			p := filepath.Join(dir, name)
			if data, err := os.ReadFile(p); err == nil {
				var m map[string]interface{}
				if err := json.Unmarshal(data, &m); err == nil {
					if u, ok := m["uuid"].(string); ok && len(u) >= 8 {
						return strings.TrimSpace(u)
					}
				}
				s := strings.TrimSpace(string(data))
				if len(s) >= 8 && len(s) <= 64 && !strings.Contains(s, "{") {
					return s
				}
			}
		}
	}

	// 2. Check system machine-id files (used by eggs)
	for _, p := range []string{
		"/etc/penguins-eggs.d/machine-id",
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	} {
		if id, ok := readAndDeriveMachineID(p); ok {
			return id
		}
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-host"
	}
	return deriveUUIDFromMachineID(hostname)
}
