package community

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// getEggsConfigDir returns the user configuration directory for penguins-eggs.
func getEggsConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "penguins-eggs"), nil
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", "penguins-eggs"), nil
	}
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

		var commBytes []byte
		if commRaw, ok := root["community"]; ok && commRaw != nil {
			commBytes, err = json.Marshal(commRaw)
			if err != nil {
				continue
			}
		} else if _, ok := root["enabled"]; ok {
			commBytes = data
		} else {
			continue
		}

		var comm CommunityData
		if err := json.Unmarshal(commBytes, &comm); err == nil {
			// Backward compatibility: map legacy contact/bio fields if socials are missing
			if comm.Socials.Telegram == "" && (strings.HasPrefix(comm.Contact, "@") || strings.Contains(comm.Contact, "t.me/")) {
				comm.Socials.Telegram = comm.Contact
			}
			if comm.Socials.GitHub == "" {
				if strings.Contains(comm.Contact, "github.com") {
					comm.Socials.GitHub = comm.Contact
				} else if strings.Contains(comm.Bio, "github.com") {
					comm.Socials.GitHub = comm.Bio
				}
			}
			if comm.Socials.Gitee == "" && strings.Contains(comm.Contact, "gitee.com") {
				comm.Socials.Gitee = comm.Contact
			}

			// Clean all handles
			if comm.Socials.Facebook != "" {
				comm.Socials.Facebook = CleanPlatformHandle("facebook", comm.Socials.Facebook)
			}
			if comm.Socials.Gitee != "" {
				comm.Socials.Gitee = CleanPlatformHandle("gitee", comm.Socials.Gitee)
			}
			if comm.Socials.GitHub != "" {
				comm.Socials.GitHub = CleanPlatformHandle("github", comm.Socials.GitHub)
			}
			if comm.Socials.Telegram != "" {
				comm.Socials.Telegram = CleanPlatformHandle("telegram", comm.Socials.Telegram)
			}
			if comm.Socials.WeChat != "" {
				comm.Socials.WeChat = CleanPlatformHandle("wechat", comm.Socials.WeChat)
			}
			if comm.Socials.WhatsApp != "" {
				comm.Socials.WhatsApp = CleanPlatformHandle("whatsapp", comm.Socials.WhatsApp)
			}
			if comm.Socials.X != "" {
				comm.Socials.X = CleanPlatformHandle("x", comm.Socials.X)
			}

			// Clean source URL and bio
			comm.SourceURL = strings.TrimSpace(comm.SourceURL)
			if comm.SourceURL != "" {
				comm.SourceURL = CleanSourceURL(comm.SourceURL)
			}
			comm.Bio = strings.TrimSpace(comm.Bio)
			if len([]rune(comm.Bio)) > MaxBioLength {
				comm.Bio = string([]rune(comm.Bio)[:MaxBioLength])
			}

			// Compute canonical primary contact without platform favoritism
			comm.Contact = ComputePrimaryContact(comm.Socials)

			return comm
		}
	}

	return CommunityData{
		Enabled: false,
	}
}

// SaveCommunityData persists community preferences into community.json.
// If enabled is false, community settings are marked as disabled.
func SaveCommunityData(data CommunityData) error {
	path := getEggsConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var toSave CommunityData
	if data.Enabled {
		socials := CommunitySocials{}
		if s := CleanPlatformHandle("facebook", data.Socials.Facebook); s != "" {
			socials.Facebook = s
		}
		if s := CleanPlatformHandle("gitee", data.Socials.Gitee); s != "" {
			socials.Gitee = s
		}
		if s := CleanPlatformHandle("github", data.Socials.GitHub); s != "" {
			socials.GitHub = s
		}
		if s := CleanPlatformHandle("telegram", data.Socials.Telegram); s != "" {
			socials.Telegram = s
		}
		if s := CleanPlatformHandle("wechat", data.Socials.WeChat); s != "" {
			socials.WeChat = s
		}
		if s := CleanPlatformHandle("whatsapp", data.Socials.WhatsApp); s != "" {
			socials.WhatsApp = s
		}
		if s := CleanPlatformHandle("x", data.Socials.X); s != "" {
			socials.X = s
		}

		cleanSource := ""
		if data.SourceURL != "" {
			cleanSource = CleanSourceURL(data.SourceURL)
		}

		cleanBio := strings.TrimSpace(data.Bio)
		if len([]rune(cleanBio)) > MaxBioLength {
			cleanBio = string([]rune(cleanBio)[:MaxBioLength])
		}

		toSave = CommunityData{
			ID:           strings.TrimSpace(data.ID),
			ArtisanToken: strings.TrimSpace(data.ArtisanToken),
			Enabled:      true,
			Nickname:     strings.TrimSpace(data.Nickname),
			Location:     strings.TrimSpace(data.Location),
			SourceURL:    cleanSource,
			Bio:          cleanBio,
			Contact:      ComputePrimaryContact(socials),
			Socials:      socials,
		}
	} else {
		toSave = CommunityData{
			ID:           strings.TrimSpace(data.ID),
			ArtisanToken: strings.TrimSpace(data.ArtisanToken),
			Enabled:      false,
		}
	}

	out, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	if err := os.WriteFile(path, out, 0644); err != nil {
		return err
	}

	// Also persist community settings into community.yaml in user configuration directory (without email)
	saveToCommunityYAML(filepath.Join(dir, "community.yaml"), toSave)

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

	if data.Enabled {
		root["community"] = data
	} else {
		root["community"] = map[string]interface{}{
			"enabled": false,
		}
	}

	if yamlBytes, err := yaml.Marshal(root); err == nil {
		_ = os.WriteFile(communityYamlPath, yamlBytes, 0644)
	}
}
