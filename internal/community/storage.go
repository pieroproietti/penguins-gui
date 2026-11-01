/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package community

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pieroproietti/penguins-gui/internal/system"
	"gopkg.in/yaml.v3"
)

// getEggsConfigDir returns the user configuration directory for penguins-eggs.
// When running with sudo or root privileges, it resolves to SUDO_USER's config directory.
func getEggsConfigDir() (string, error) {
	cfgDir := system.GetUserConfigDir()
	return filepath.Join(cfgDir, "penguins-eggs"), nil
}

// getEggsConfigPath returns the primary configuration file path for community settings.
func getEggsConfigPath() string {
	dir, err := getEggsConfigDir()
	if err != nil {
		return "/etc/penguins-eggs.d/community.json"
	}
	return filepath.Join(dir, "community.json")
}

// LoadCommunityData loads community preferences from community.json or community.yaml.
// Note: ArtisanToken (egg-pat) is managed strictly in memory and is NEVER read from disk.
func LoadCommunityData() CommunityData {
	userPath := getEggsConfigPath()
	candidates := []string{
		userPath,
		filepath.Join(filepath.Dir(userPath), "community.yaml"),
		"/etc/penguins-eggs.d/community.json",
		"/etc/penguins-eggs.d/community.yaml",
		"/etc/penguins-eggs.d/community.yml",
	}

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var root map[string]interface{}
		if err := json.Unmarshal(data, &root); err != nil {
			if err := yaml.Unmarshal(data, &root); err != nil {
				continue
			}
		}

		var commMap map[string]interface{}
		var commBytes []byte
		if commRaw, ok := root["community"]; ok && commRaw != nil {
			commBytes, err = json.Marshal(commRaw)
			if err != nil {
				continue
			}
			_ = json.Unmarshal(commBytes, &commMap)
		} else if _, ok := root["enabled"]; ok {
			commBytes = data
			commMap = root
		} else {
			continue
		}

		var comm CommunityData
		if err := json.Unmarshal(commBytes, &comm); err == nil {
			// Backward compatibility: if author_id is missing but id is present
			if comm.AuthorID == "" {
				if id, ok := commMap["id"].(string); ok && id != "" {
					comm.AuthorID = strings.TrimSpace(id)
				}
			}
			if comm.AuthorID == "" && comm.Email != "" {
				comm.AuthorID = ComputeAuthorID(comm.Email)
			}
			comm.ArtisanToken = "" // strictly in-memory
			return comm
		}
	}

	return CommunityData{
		Enabled: false,
	}
}

// SaveCommunityData persists community preferences into community.json and community.yaml.
// The private egg-pat token is NEVER saved to disk.
func SaveCommunityData(data CommunityData) error {
	path := getEggsConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	authorID := strings.TrimSpace(data.AuthorID)
	email := strings.TrimSpace(strings.ToLower(data.Email))
	if authorID == "" && email != "" {
		authorID = ComputeAuthorID(email)
	}

	toSave := CommunityData{
		AuthorID: authorID,
		Email:    email,
		Enabled:  data.Enabled,
	}

	out, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return err
	}

	saveToCommunityYAML(filepath.Join(dir, "community.yaml"), toSave)
	_ = system.FixPathOwnership(dir)
	return nil
}

func saveToCommunityYAML(communityYamlPath string, data CommunityData) {
	var root map[string]interface{}
	if existingBytes, err := os.ReadFile(communityYamlPath); err == nil {
		_ = yaml.Unmarshal(existingBytes, &root)
	}
	if root == nil {
		root = make(map[string]interface{})
	}

	root["community"] = data
	if yamlBytes, err := yaml.Marshal(root); err == nil {
		_ = os.WriteFile(communityYamlPath, yamlBytes, 0644)
	}
}
