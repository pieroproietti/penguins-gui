package system

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

const GRUBHelperArg = "--configure-eggs-grub"

const grubStart = "# >>> penguins-eggs start: "
const grubEnd = "# >>> penguins-eggs end: "

type grubBlock struct {
	start, end int
	name       string
}

// eggsBlocks accepts only complete, non-nested pairs with matching ISO names.
// Offsets preserve every byte outside the managed blocks.
func eggsBlocks(content string) ([]grubBlock, error) {
	var blocks []grubBlock
	var active *grubBlock
	offset := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		marker := strings.TrimSuffix(line, "\n")
		start, end := strings.HasPrefix(marker, grubStart), strings.HasPrefix(marker, grubEnd)
		if start || end {
			prefix := grubStart
			if end {
				prefix = grubEnd
			}
			name := strings.TrimPrefix(marker, prefix)
			if !strings.HasSuffix(name, " <<<") || len(name) <= 4 {
				return nil, fmt.Errorf("invalid Penguins' Eggs GRUB marker")
			}
			name = strings.TrimSuffix(name, " <<<")
			if start {
				if active != nil {
					return nil, fmt.Errorf("nested Penguins' Eggs GRUB markers")
				}
				active = &grubBlock{start: offset, name: name}
			} else {
				if active == nil || active.name != name {
					return nil, fmt.Errorf("unmatched Penguins' Eggs GRUB marker")
				}
				active.end = offset + len(line)
				blocks = append(blocks, *active)
				active = nil
			}
		} else if strings.HasPrefix(strings.TrimSpace(marker), "# >>> penguins-eggs") {
			return nil, fmt.Errorf("unrecognized Penguins' Eggs GRUB marker")
		}
		offset += len(line)
	}
	if active != nil {
		return nil, fmt.Errorf("incomplete Penguins' Eggs GRUB block")
	}
	return blocks, nil
}

func replaceEggsBlocks(content, entry string) (string, error) {
	blocks, err := eggsBlocks(content)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	offset := 0
	for i, block := range blocks {
		result.WriteString(content[offset:block.start])
		if i == 0 {
			result.WriteString(entry)
		}
		offset = block.end
	}
	result.WriteString(content[offset:])
	if len(blocks) == 0 {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			result.WriteByte('\n')
		}
		result.WriteString(entry)
	}
	return result.String(), nil
}

// ConfigureEggsGRUB generates first, then replaces only marked Eggs entries.
// It neither removes ISO files nor regenerates the system's GRUB menu.
func ConfigureEggsGRUB(eggsPath, isoPath string) error {
	return configureEggsGRUB("/etc/grub.d/40_custom", isoPath, func() ([]byte, error) {
		cmd := exec.Command(eggsPath, "tools", "grub40", isoPath)
		cmd.Stderr = os.Stderr
		return cmd.Output()
	})
}

func configureEggsGRUB(path, isoPath string, generate func() ([]byte, error)) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s must be a regular file", path)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if _, err := eggsBlocks(string(original)); err != nil {
		return err
	}
	output, err := generate()
	if err != nil {
		return fmt.Errorf("generate ISO boot entry: %w", err)
	}
	// Eggs may color its terminal output; only the marked block is installed.
	generated := regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`).ReplaceAllString(string(output), "")
	blocks, err := eggsBlocks(generated)
	if err != nil {
		return err
	}
	if len(blocks) != 1 || blocks[0].name != filepath.Base(isoPath) {
		return fmt.Errorf("Eggs did not return exactly one marked entry for the selected ISO")
	}
	entry := strings.TrimSuffix(generated[blocks[0].start:blocks[0].end], "\n") + "\n"
	updated, err := replaceEggsBlocks(string(original), entry)
	if err != nil {
		return err
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	currentInfo, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !os.SameFile(info, currentInfo) || !bytes.Equal(current, original) {
		return fmt.Errorf("%s changed during ISO inspection; please retry", path)
	}
	// grub-mkconfig executes executable files in grub.d, including .bak files.
	// Strip execute bits even when replacing a backup from an older GUI version.
	if err := writeGRUBFile(path+".penguins-gui.bak", original, info, info.Mode().Perm()&^0111); err != nil {
		return err
	}
	return writeGRUBFile(path, []byte(updated), info, info.Mode().Perm())
}

func writeGRUBFile(path string, data []byte, info os.FileInfo, mode os.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".penguins-grub-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if err := f.Chown(int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}
	}
	if err := f.Chmod(mode); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
