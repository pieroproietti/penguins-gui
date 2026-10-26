package eggs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ISOArtifact describes an ISO found in the working directory.
type ISOArtifact struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// NewestISO finds the latest ISO directly in root, without scanning subdirectories,
// allowing two seconds of timestamp tolerance.
func (CLIAdapter) NewestISO(root string, notBefore time.Time) (ISOArtifact, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return ISOArtifact{}, err
	}
	var found []ISOArtifact
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".iso") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return ISOArtifact{}, err
		}
		if info.ModTime().Before(notBefore.Add(-2 * time.Second)) {
			continue
		}
		found = append(found, ISOArtifact{Path: filepath.Join(root, entry.Name()), Size: info.Size(), ModTime: info.ModTime()})
	}
	if len(found) == 0 {
		return ISOArtifact{}, errors.New("no new ISO found")
	}
	sort.Slice(found, func(i, j int) bool { return found[i].ModTime.After(found[j].ModTime) })
	return found[0], nil
}

// ListISOs returns the names of regular .iso files directly inside root.
func ListISOs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".iso") {
			continue
		}
		info, err := os.Stat(filepath.Join(root, entry.Name()))
		if err == nil && info.Mode().IsRegular() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// HumanSize formats a byte count using binary IEC units (KiB, MiB, GiB, etc.).
func HumanSize(size int64) string {
	const unit = int64(1024)
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := unit, 0
	for value := size / unit; value >= unit && exp < 5; value /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}
