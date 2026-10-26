package community

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const CommunitySecretFileName = "penguins-community.secret"

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

// GetCommunitySecretPath returns the path to ~/.ssh/penguins-community.secret.
func GetCommunitySecretPath() string {
	if custom := strings.TrimSpace(os.Getenv("EGGS_COMMUNITY_SECRET_PATH")); custom != "" {
		return custom
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		// In isolated test environments using XDG_CONFIG_HOME, place .ssh sibling to config dir
		return filepath.Join(filepath.Dir(xdg), ".ssh", CommunitySecretFileName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/root", ".ssh", CommunitySecretFileName)
	}
	return filepath.Join(home, ".ssh", CommunitySecretFileName)
}

func loadCommunitySecret() string {
	data, err := os.ReadFile(GetCommunitySecretPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveCommunitySecret(token string) error {
	secretPath := GetCommunitySecretPath()
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		if err := os.Remove(secretPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	sshDir := filepath.Dir(secretPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}
	return os.WriteFile(secretPath, []byte(trimmed+"\n"), 0600)
}

// LoadCommunityData loads community preferences from community.json or community.yaml,
// and loads the private Community Token from ~/.ssh/penguins-community.secret.
func LoadCommunityData() CommunityData {
	secretToken := loadCommunitySecret()

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
			// Migrate legacy artisan_token from config file into ~/.ssh/penguins-community.secret
			var legacyTok string
			if commMap != nil {
				if rawTok, ok := commMap["artisan_token"].(string); ok {
					legacyTok = strings.TrimSpace(rawTok)
				}
			}
			if secretToken != "" {
				comm.ArtisanToken = secretToken
			} else if legacyTok != "" {
				comm.ArtisanToken = legacyTok
				_ = saveCommunitySecret(legacyTok)
			}

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

			// If legacy token was present in the public config file, rewrite config without it
			if legacyTok != "" {
				_ = SaveCommunityData(comm)
			}

			return comm
		}
	}

	return CommunityData{
		ArtisanToken: secretToken,
		Enabled:      false,
	}
}

// SaveCommunityData persists public community preferences into community.json and community.yaml,
// while storing the secret Community Token exclusively in ~/.ssh/penguins-community.secret (0600).
func SaveCommunityData(data CommunityData) error {
	path := getEggsConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := saveCommunitySecret(data.ArtisanToken); err != nil {
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
			ID:        strings.TrimSpace(data.ID),
			Enabled:   true,
			Nickname:  strings.TrimSpace(data.Nickname),
			Location:  strings.TrimSpace(data.Location),
			SourceURL: cleanSource,
			Bio:       cleanBio,
			Contact:   ComputePrimaryContact(socials),
			Socials:   socials,
		}
	} else {
		toSave = CommunityData{
			ID:      strings.TrimSpace(data.ID),
			Enabled: false,
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

	// Also persist community settings into community.yaml in user configuration directory (without email or token)
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
