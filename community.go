package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
)

const (
	MaxBioLength              = 200
	DefaultCommunityEndpoint  = "https://api.penguins-eggs.net/api/community"
	FallbackCommunityEndpoint = "https://penguins-eggs.net/api/community"
)

var (
	emailValidationPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	emailForbiddenPattern  = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
)

// CommunitySocials holds individual social platform handles in alphabetical order.
type CommunitySocials struct {
	Facebook string `yaml:"facebook,omitempty" json:"facebook,omitempty"`
	Gitee    string `yaml:"gitee,omitempty" json:"gitee,omitempty"`
	GitHub   string `yaml:"github,omitempty" json:"github,omitempty"`
	Telegram string `yaml:"telegram,omitempty" json:"telegram,omitempty"`
	WeChat   string `yaml:"wechat,omitempty" json:"wechat,omitempty"`
	WhatsApp string `yaml:"whatsapp,omitempty" json:"whatsapp,omitempty"`
	X        string `yaml:"x,omitempty" json:"x,omitempty"`
}

// CommunityData represents user preferences for community participation.
// Email is strictly forbidden to prevent spam.
type CommunityData struct {
	Enabled   bool             `yaml:"enabled" json:"enabled"`
	Nickname  string           `yaml:"nickname,omitempty" json:"nickname,omitempty"`
	Location  string           `yaml:"location,omitempty" json:"location,omitempty"`
	SourceURL string           `yaml:"source_url,omitempty" json:"source_url,omitempty"` // repository or project website
	Bio       string           `yaml:"bio,omitempty" json:"bio,omitempty"`               // max 200 characters
	Contact   string           `yaml:"contact,omitempty" json:"contact,omitempty"`       // primary public contact handle/URL (never email!)
	Socials   CommunitySocials `yaml:"socials,omitempty" json:"socials,omitempty"`
}

// IsValidEmail verifies if an email address conforms to standard syntax.
func IsValidEmail(email string) bool {
	return emailValidationPattern.MatchString(strings.TrimSpace(email))
}

// CleanSourceURL normalizes a source repository or project website URL.
func CleanSourceURL(input string) string {
	s := strings.TrimSpace(input)
	if s == "" {
		return ""
	}
	s = strings.TrimPrefix(s, "https//")
	s = strings.TrimPrefix(s, "http//")
	if strings.Contains(s, "://") {
		return s
	}
	return "https://" + s
}

// IsValidURL checks whether a URL is syntactically valid and has an http/https scheme and a valid host.
func IsValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return strings.Contains(u.Host, ".") || u.Host == "localhost"
}

// CleanPlatformHandle cleans and extracts the base handle from an input for a specific platform.
// It handles raw handles, @handles, full URLs, and malformed prefixes (e.g. https//).
func CleanPlatformHandle(platform, input string) string {
	s := strings.TrimSpace(input)
	if s == "" {
		return ""
	}

	// Fix common typos like https// or http// without colon
	s = strings.TrimPrefix(s, "https//")
	s = strings.TrimPrefix(s, "http//")
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")

	switch strings.ToLower(platform) {
	case "facebook":
		s = strings.TrimPrefix(s, "facebook.com/")
		s = strings.TrimPrefix(s, "fb.com/")
		s = strings.Trim(s, "/")
		return s
	case "gitee":
		s = strings.TrimPrefix(s, "gitee.com/")
		s = strings.TrimPrefix(s, "@")
		s = strings.Trim(s, "/")
		return s
	case "github":
		s = strings.TrimPrefix(s, "github.com/")
		s = strings.TrimPrefix(s, "@")
		s = strings.Trim(s, "/")
		return s
	case "telegram":
		s = strings.TrimPrefix(s, "t.me/")
		s = strings.TrimPrefix(s, "telegram.me/")
		s = strings.Trim(s, "/")
		if s != "" && !strings.HasPrefix(s, "@") {
			s = "@" + s
		}
		return s
	case "wechat":
		return s
	case "whatsapp":
		s = strings.TrimPrefix(s, "wa.me/")
		s = strings.Trim(s, "/")
		s = strings.ReplaceAll(s, " ", "")
		return s
	case "x", "twitter":
		s = strings.TrimPrefix(s, "x.com/")
		s = strings.TrimPrefix(s, "twitter.com/")
		s = strings.Trim(s, "/")
		if s != "" && !strings.HasPrefix(s, "@") {
			s = "@" + s
		}
		return s
	default:
		return s
	}
}

// FormatContact returns a clean, fully-qualified URL or handle for contact display and storage.
func FormatContact(platform, val string) string {
	handle := CleanPlatformHandle(platform, val)
	if handle == "" {
		return ""
	}

	switch strings.ToLower(platform) {
	case "facebook":
		return "https://facebook.com/" + handle
	case "gitee":
		return "https://gitee.com/" + handle
	case "github":
		return "https://github.com/" + handle
	case "telegram":
		return handle
	case "wechat":
		return handle
	case "whatsapp":
		if strings.HasPrefix(handle, "+") {
			return "https://wa.me/" + strings.TrimPrefix(handle, "+")
		}
		return "https://wa.me/" + handle
	case "x", "twitter":
		return handle
	default:
		return handle
	}
}

// ComputePrimaryContact determines the public contact handle/URL from socials in alphabetical order.
// No social network is given preference over others. Email is NEVER used as public contact.
func ComputePrimaryContact(socials CommunitySocials) string {
	if socials.Facebook != "" {
		return FormatContact("facebook", socials.Facebook)
	}
	if socials.Gitee != "" {
		return FormatContact("gitee", socials.Gitee)
	}
	if socials.GitHub != "" {
		return FormatContact("github", socials.GitHub)
	}
	if socials.Telegram != "" {
		return FormatContact("telegram", socials.Telegram)
	}
	if socials.WeChat != "" {
		return FormatContact("wechat", socials.WeChat)
	}
	if socials.WhatsApp != "" {
		return FormatContact("whatsapp", socials.WhatsApp)
	}
	if socials.X != "" {
		return FormatContact("x", socials.X)
	}
	return ""
}

// ValidateContact ensures public fields (socials, nickname, location, source URL) never contain email addresses.
func ValidateContact(contact string) error {
	trimmed := strings.TrimSpace(contact)
	if trimmed == "" {
		return nil
	}
	// Fediverse handles like @user@instance.social are valid handles
	if strings.HasPrefix(trimmed, "@") && strings.Count(trimmed, "@") >= 2 {
		return nil
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "mailto:") || emailForbiddenPattern.MatchString(trimmed) {
		return errors.New("to protect your privacy, email addresses are not permitted in public fields")
	}
	return nil
}

// ValidateCommunityData validates community fields:
// bio must not exceed MaxBioLength (200 characters),
// source URL must be valid if provided,
// and all public fields must never contain email addresses.
func ValidateCommunityData(data CommunityData) error {
	if !data.Enabled {
		return nil
	}

	trimmedSource := strings.TrimSpace(data.SourceURL)
	if trimmedSource != "" {
		cleanSource := CleanSourceURL(trimmedSource)
		if !IsValidURL(cleanSource) {
			return errors.New("please enter a valid source repository or website URL")
		}
		if err := ValidateContact(trimmedSource); err != nil {
			return fmt.Errorf("source URL: %w", err)
		}
	}

	if len([]rune(data.Bio)) > MaxBioLength {
		return fmt.Errorf("bio must not exceed %d characters (currently %d)", MaxBioLength, len([]rune(data.Bio)))
	}
	if err := ValidateContact(data.Bio); err != nil {
		return fmt.Errorf("bio: %w", err)
	}

	if err := ValidateContact(data.Nickname); err != nil {
		return fmt.Errorf("nickname: %w", err)
	}
	if err := ValidateContact(data.Location); err != nil {
		return fmt.Errorf("location: %w", err)
	}

	socials := []struct {
		name string
		val  string
	}{
		{"Facebook", data.Socials.Facebook},
		{"Gitee", data.Socials.Gitee},
		{"GitHub", data.Socials.GitHub},
		{"Telegram", data.Socials.Telegram},
		{"WeChat", data.Socials.WeChat},
		{"WhatsApp", data.Socials.WhatsApp},
		{"X", data.Socials.X},
		{"Contact", data.Contact},
	}
	for _, s := range socials {
		if err := ValidateContact(s.val); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}

	return nil
}

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

// LoadCommunityData loads community preferences from community.json (with fallback to legacy paths).
func LoadCommunityData() CommunityData {
	userPath := getEggsConfigPath()
	candidates := []string{
		userPath,
		"/etc/penguins-eggs.d/community.json",
		// Legacy fallbacks for migration
		filepath.Join(filepath.Dir(userPath), "custom.yaml"),
		filepath.Join(filepath.Dir(userPath), "custom.json"),
		"/etc/penguins-eggs.d/custom.yaml",
		"/etc/penguins-eggs.d/custom.yml",
		"/etc/penguins-eggs.d/custom.json",
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

	// Also persist community settings into custom.yaml in user configuration directory (without email)
	saveToCustomYAML(filepath.Join(dir, "custom.yaml"), toSave)

	return nil
}

func saveToCustomYAML(customYamlPath string, data CommunityData) {
	var root map[string]interface{}
	if existingBytes, err := os.ReadFile(customYamlPath); err == nil {
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
		_ = os.WriteFile(customYamlPath, yamlBytes, 0644)
	}
}

func getCommunityEndpoint() string {
	if ep := strings.TrimSpace(os.Getenv("EGGS_COMMUNITY_ENDPOINT")); ep != "" {
		return ep
	}
	if ep := strings.TrimSpace(os.Getenv("CENSUS_COMMUNITY_URL")); ep != "" {
		return ep
	}
	return DefaultCommunityEndpoint
}

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

// SyncCommunityProfile communicates directly with /api/community:
// If data.Enabled is true, it sends a PUT request to https://api.penguins-eggs.net/api/community/{uuid}.
// If data.Enabled is false, it sends a DELETE request to https://api.penguins-eggs.net/api/community/{uuid}.
// Fail-safe: uses a short 7s timeout and falls back to FallbackCommunityEndpoint if the primary host fails.
func SyncCommunityProfile(data CommunityData) error {
	uuid := GetMachineUUID()
	if uuid == "" {
		return errors.New("unable to determine machine UUID")
	}

	endpoint := getCommunityEndpoint()
	endpoint = strings.TrimSuffix(endpoint, "/")

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	if !data.Enabled {
		delURL := fmt.Sprintf("%s/%s", endpoint, uuid)
		req, err := http.NewRequest(http.MethodDelete, delURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil && endpoint == DefaultCommunityEndpoint {
			// Try fallback host
			fallbackURL := fmt.Sprintf("%s/%s", FallbackCommunityEndpoint, uuid)
			if reqFallback, fErr := http.NewRequest(http.MethodDelete, fallbackURL, nil); fErr == nil {
				if fResp, fDoErr := client.Do(reqFallback); fDoErr == nil {
					resp = fResp
					err = nil
				}
			}
		}
		if err != nil {
			return fmt.Errorf("network error removing profile: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return nil
	}

	socials := make(map[string]string)
	if s := CleanPlatformHandle("facebook", data.Socials.Facebook); s != "" {
		socials["facebook"] = s
	}
	if s := CleanPlatformHandle("gitee", data.Socials.Gitee); s != "" {
		socials["gitee"] = s
	}
	if s := CleanPlatformHandle("github", data.Socials.GitHub); s != "" {
		socials["github"] = s
	}
	if s := CleanPlatformHandle("telegram", data.Socials.Telegram); s != "" {
		socials["telegram"] = s
	}
	if s := CleanPlatformHandle("wechat", data.Socials.WeChat); s != "" {
		socials["wechat"] = s
	}
	if s := CleanPlatformHandle("whatsapp", data.Socials.WhatsApp); s != "" {
		socials["whatsapp"] = s
	}
	if s := CleanPlatformHandle("x", data.Socials.X); s != "" {
		socials["x"] = s
	}

	payloadMap := map[string]interface{}{
		"uuid":       uuid,
		"nickname":   strings.TrimSpace(data.Nickname),
		"location":   strings.TrimSpace(data.Location),
		"source_url": CleanSourceURL(data.SourceURL),
		"bio":        strings.TrimSpace(data.Bio),
		"contact":    data.Contact,
		"socials":    socials,
		"enabled":    true,
	}

	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	putURL := fmt.Sprintf("%s/%s", endpoint, uuid)
	req, err := http.NewRequest(http.MethodPut, putURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil && endpoint == DefaultCommunityEndpoint {
		// Try fallback host
		fallbackURL := fmt.Sprintf("%s/%s", FallbackCommunityEndpoint, uuid)
		if reqFallback, fErr := http.NewRequest(http.MethodPut, fallbackURL, bytes.NewReader(payloadBytes)); fErr == nil {
			reqFallback.Header.Set("Content-Type", "application/json")
			if fResp, fDoErr := client.Do(reqFallback); fDoErr == nil {
				resp = fResp
				err = nil
			}
		}
	}
	if err != nil {
		return fmt.Errorf("network error updating profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// buildCommunityPanel creates the UI container for the Community configuration panel.
func buildCommunityPanel(w fyne.Window, onUpdate func(CommunityData)) (fyne.CanvasObject, func() CommunityData) {
	checkEnabled := widget.NewCheck("Join the Penguins Eggs Community", nil)

	// 2. Nickname / Public handle
	nicknameEntry := widget.NewEntry()
	nicknameEntry.SetPlaceHolder("e.g. tux_artisan or @nickname")

	// 3. Location / City
	locationEntry := widget.NewEntry()
	locationEntry.SetPlaceHolder("e.g. Rome, IT")

	// 4. Repository / Website URL (source_url)
	sourceURLEntry := widget.NewEntry()
	sourceURLEntry.SetPlaceHolder("e.g. https://github.com/user/distro or https://mydistro.org")

	// 5. Bio / Description (before socials, max 200 characters)
	bioEntry := widget.NewMultiLineEntry()
	bioEntry.SetPlaceHolder("Short description of your project, distro, or interests (max 200 characters)...")
	bioEntry.Wrapping = fyne.TextWrapWord
	bioEntry.SetMinRowsVisible(3)

	bioCounterLabel := widget.NewLabel(fmt.Sprintf("0/%d characters", MaxBioLength))
	bioCounterLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 6. Social Networks / Contact (alphabetical order)
	checkFacebook := widget.NewCheck("Facebook", nil)
	entryFacebook := widget.NewEntry()
	entryFacebook.SetPlaceHolder("e.g. username or facebook.com/username")

	checkGitee := widget.NewCheck("Gitee", nil)
	entryGitee := widget.NewEntry()
	entryGitee.SetPlaceHolder("e.g. username or gitee.com/username")

	checkGitHub := widget.NewCheck("GitHub", nil)
	entryGitHub := widget.NewEntry()
	entryGitHub.SetPlaceHolder("e.g. username or github.com/username")

	checkTelegram := widget.NewCheck("Telegram", nil)
	entryTelegram := widget.NewEntry()
	entryTelegram.SetPlaceHolder("e.g. @username or t.me/username")

	checkWeChat := widget.NewCheck("WeChat", nil)
	entryWeChat := widget.NewEntry()
	entryWeChat.SetPlaceHolder("e.g. WeChat ID")

	checkWhatsApp := widget.NewCheck("WhatsApp", nil)
	entryWhatsApp := widget.NewEntry()
	entryWhatsApp.SetPlaceHolder("e.g. +39 333 1234567 or wa.me/number")

	checkX := widget.NewCheck("X (Twitter)", nil)
	entryX := widget.NewEntry()
	entryX.SetPlaceHolder("e.g. @username or x.com/username")

	socialHelp := widget.NewLabel("ℹ️ Optional: enter your public social channels. Email addresses are strictly forbidden; Fediverse handles (@user@instance.social) are allowed.")
	socialHelp.TextStyle = fyne.TextStyle{Italic: true}
	socialHelp.Wrapping = fyne.TextWrapWord

	validationError := widget.NewLabel("")
	validationError.Wrapping = fyne.TextWrapWord
	validationError.TextStyle = fyne.TextStyle{Bold: true}

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	var saveBtn *widget.Button

	currentData := func() CommunityData {
		socials := CommunitySocials{}
		if checkFacebook.Checked {
			socials.Facebook = CleanPlatformHandle("facebook", entryFacebook.Text)
		}
		if checkGitee.Checked {
			socials.Gitee = CleanPlatformHandle("gitee", entryGitee.Text)
		}
		if checkGitHub.Checked {
			socials.GitHub = CleanPlatformHandle("github", entryGitHub.Text)
		}
		if checkTelegram.Checked {
			socials.Telegram = CleanPlatformHandle("telegram", entryTelegram.Text)
		}
		if checkWeChat.Checked {
			socials.WeChat = CleanPlatformHandle("wechat", entryWeChat.Text)
		}
		if checkWhatsApp.Checked {
			socials.WhatsApp = CleanPlatformHandle("whatsapp", entryWhatsApp.Text)
		}
		if checkX.Checked {
			socials.X = CleanPlatformHandle("x", entryX.Text)
		}

		primary := ComputePrimaryContact(socials)

		return CommunityData{
			Enabled:   checkEnabled.Checked,
			Nickname:  strings.TrimSpace(nicknameEntry.Text),
			Location:  strings.TrimSpace(locationEntry.Text),
			SourceURL: CleanSourceURL(sourceURLEntry.Text),
			Bio:       strings.TrimSpace(bioEntry.Text),
			Contact:   primary,
			Socials:   socials,
		}
	}

	validateInputs := func() bool {
		if !checkEnabled.Checked {
			validationError.SetText("")
			if saveBtn != nil {
				saveBtn.Enable()
			}
			return true
		}
		data := currentData()
		if err := ValidateCommunityData(data); err != nil {
			validationError.SetText("⚠️ " + err.Error())
			if saveBtn != nil {
				saveBtn.Disable()
			}
			return false
		}
		validationError.SetText("")
		if saveBtn != nil {
			saveBtn.Enable()
		}
		return true
	}

	updateFieldStates := func() {
		enabled := checkEnabled.Checked
		if enabled {
			nicknameEntry.Enable()
			locationEntry.Enable()
			sourceURLEntry.Enable()
			bioEntry.Enable()
			checkFacebook.Enable()
			checkGitee.Enable()
			checkGitHub.Enable()
			checkTelegram.Enable()
			checkWeChat.Enable()
			checkWhatsApp.Enable()
			checkX.Enable()

			if checkFacebook.Checked {
				entryFacebook.Enable()
			} else {
				entryFacebook.Disable()
			}
			if checkGitee.Checked {
				entryGitee.Enable()
			} else {
				entryGitee.Disable()
			}
			if checkGitHub.Checked {
				entryGitHub.Enable()
			} else {
				entryGitHub.Disable()
			}
			if checkTelegram.Checked {
				entryTelegram.Enable()
			} else {
				entryTelegram.Disable()
			}
			if checkWeChat.Checked {
				entryWeChat.Enable()
			} else {
				entryWeChat.Disable()
			}
			if checkWhatsApp.Checked {
				entryWhatsApp.Enable()
			} else {
				entryWhatsApp.Disable()
			}
			if checkX.Checked {
				entryX.Enable()
			} else {
				entryX.Disable()
			}

			statusLabel.SetText("Partecipazione attiva: compila i campi e premi Salva per sincronizzare il profilo con la community.")
		} else {
			nicknameEntry.Disable()
			locationEntry.Disable()
			sourceURLEntry.Disable()
			bioEntry.Disable()
			checkFacebook.Disable()
			checkGitee.Disable()
			checkGitHub.Disable()
			checkTelegram.Disable()
			checkWeChat.Disable()
			checkWhatsApp.Disable()
			checkX.Disable()
			entryFacebook.Disable()
			entryGitee.Disable()
			entryGitHub.Disable()
			entryTelegram.Disable()
			entryWeChat.Disable()
			entryWhatsApp.Disable()
			entryX.Disable()
			validationError.SetText("")
			statusLabel.SetText("Partecipazione alla community disattivata.")
		}
		validateInputs()
	}

	// Wire checkbox toggles
	checkFacebook.OnChanged = func(checked bool) {
		if checked {
			entryFacebook.Enable()
		} else {
			entryFacebook.Disable()
		}
		validateInputs()
	}
	checkGitee.OnChanged = func(checked bool) {
		if checked {
			entryGitee.Enable()
		} else {
			entryGitee.Disable()
		}
		validateInputs()
	}
	checkGitHub.OnChanged = func(checked bool) {
		if checked {
			entryGitHub.Enable()
		} else {
			entryGitHub.Disable()
		}
		validateInputs()
	}
	checkTelegram.OnChanged = func(checked bool) {
		if checked {
			entryTelegram.Enable()
		} else {
			entryTelegram.Disable()
		}
		validateInputs()
	}
	checkWeChat.OnChanged = func(checked bool) {
		if checked {
			entryWeChat.Enable()
		} else {
			entryWeChat.Disable()
		}
		validateInputs()
	}
	checkWhatsApp.OnChanged = func(checked bool) {
		if checked {
			entryWhatsApp.Enable()
		} else {
			entryWhatsApp.Disable()
		}
		validateInputs()
	}
	checkX.OnChanged = func(checked bool) {
		if checked {
			entryX.Enable()
		} else {
			entryX.Disable()
		}
		validateInputs()
	}

	// Validate on entry text changes
	onTextChange := func(string) {
		validateInputs()
	}
	nicknameEntry.OnChanged = onTextChange
	locationEntry.OnChanged = onTextChange
	sourceURLEntry.OnChanged = onTextChange

	// Bio length constraint and visual counter
	bioEntry.OnChanged = func(s string) {
		runes := []rune(s)
		if len(runes) > MaxBioLength {
			bioEntry.SetText(string(runes[:MaxBioLength]))
			runes = []rune(bioEntry.Text)
		}
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d caratteri", len(runes), MaxBioLength))
		validateInputs()
	}

	entryFacebook.OnChanged = onTextChange
	entryGitee.OnChanged = onTextChange
	entryGitHub.OnChanged = onTextChange
	entryTelegram.OnChanged = onTextChange
	entryWeChat.OnChanged = onTextChange
	entryWhatsApp.OnChanged = onTextChange
	entryX.OnChanged = onTextChange

	saveBtn = widget.NewButtonWithIcon("Salva / Aggiorna profilo Community", theme.DocumentSaveIcon(), func() {
		if checkEnabled.Checked && !validateInputs() {
			dialog.ShowError(errors.New("correggi gli errori di validazione prima di salvare"), w)
			return
		}
		data := currentData()
		if err := SaveCommunityData(data); err != nil {
			dialog.ShowError(fmt.Errorf("impossibile salvare le preferenze locali: %w", err), w)
			return
		}

		statusLabel.SetText("Sincronizzazione in corso con il server community...")
		if onUpdate != nil {
			onUpdate(data)
		}

		go func() {
			err := SyncCommunityProfile(data)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText("⚠️ Salvato localmente, errore server: " + err.Error())
				} else {
					if data.Enabled {
						statusLabel.SetText("✅ Profilo salvato e sincronizzato con successo su " + getCommunityEndpoint())
					} else {
						statusLabel.SetText("✅ Profilo disattivato e rimosso dal server.")
					}
				}
			})
		}()
	})

	clearBtn := widget.NewButtonWithIcon("Disabilita / Cancella Campi", theme.ContentClearIcon(), func() {
		checkEnabled.SetChecked(false)
		nicknameEntry.SetText("")
		locationEntry.SetText("")
		sourceURLEntry.SetText("")
		bioEntry.SetText("")
		bioCounterLabel.SetText(fmt.Sprintf("0/%d caratteri", MaxBioLength))
		entryFacebook.SetText("")
		entryGitee.SetText("")
		entryGitHub.SetText("")
		entryTelegram.SetText("")
		entryWeChat.SetText("")
		entryWhatsApp.SetText("")
		entryX.SetText("")
		checkFacebook.SetChecked(false)
		checkGitee.SetChecked(false)
		checkGitHub.SetChecked(false)
		checkTelegram.SetChecked(false)
		checkWeChat.SetChecked(false)
		checkWhatsApp.SetChecked(false)
		checkX.SetChecked(false)
		updateFieldStates()
		data := CommunityData{Enabled: false}
		_ = SaveCommunityData(data)
		if onUpdate != nil {
			onUpdate(data)
		}

		statusLabel.SetText("Rimozione profilo dal server in corso...")
		go func() {
			err := SyncCommunityProfile(data)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText("⚠️ Preferenze azzerate localmente, errore server: " + err.Error())
				} else {
					statusLabel.SetText("✅ Profilo disabilitato e rimosso dal server.")
				}
			})
		}()
	})

	checkEnabled.OnChanged = func(checked bool) {
		updateFieldStates()
		if !checked {
			data := currentData()
			_ = SaveCommunityData(data)
			if onUpdate != nil {
				onUpdate(data)
			}
			go func() {
				_ = SyncCommunityProfile(data)
			}()
		}
	}

	// Load existing configuration
	initial := LoadCommunityData()
	if initial.Enabled {
		checkEnabled.SetChecked(true)
		nicknameEntry.SetText(initial.Nickname)
		locationEntry.SetText(initial.Location)
		sourceURLEntry.SetText(initial.SourceURL)
		bioEntry.SetText(initial.Bio)
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d caratteri", len([]rune(initial.Bio)), MaxBioLength))

		if initial.Socials.Facebook != "" {
			checkFacebook.SetChecked(true)
			entryFacebook.SetText(initial.Socials.Facebook)
		}
		if initial.Socials.Gitee != "" {
			checkGitee.SetChecked(true)
			entryGitee.SetText(initial.Socials.Gitee)
		}
		if initial.Socials.GitHub != "" {
			checkGitHub.SetChecked(true)
			entryGitHub.SetText(initial.Socials.GitHub)
		}
		if initial.Socials.Telegram != "" {
			checkTelegram.SetChecked(true)
			entryTelegram.SetText(initial.Socials.Telegram)
		}
		if initial.Socials.WeChat != "" {
			checkWeChat.SetChecked(true)
			entryWeChat.SetText(initial.Socials.WeChat)
		}
		if initial.Socials.WhatsApp != "" {
			checkWhatsApp.SetChecked(true)
			entryWhatsApp.SetText(initial.Socials.WhatsApp)
		}
		if initial.Socials.X != "" {
			checkX.SetChecked(true)
			entryX.SetText(initial.Socials.X)
		}
		updateFieldStates()
	} else {
		checkEnabled.SetChecked(false)
		updateFieldStates()
	}

	// Socials Form in strictly alphabetical order:
	// Facebook, Gitee, GitHub, Telegram, WeChat, WhatsApp, X
	socialsForm := container.New(layout.NewFormLayout(),
		checkFacebook, entryFacebook,
		checkGitee, entryGitee,
		checkGitHub, entryGitHub,
		checkTelegram, entryTelegram,
		checkWeChat, entryWeChat,
		checkWhatsApp, entryWhatsApp,
		checkX, entryX,
	)

	infoBtn := widget.NewButtonWithIcon("Info", theme.InfoIcon(), func() {
		dialog.ShowInformation(
			"Penguins Community",
			"Partecipa alla community globale di Penguins Eggs!\n\n"+
				"Connettiti con altri sviluppatori e artigiani di distribuzioni Linux.\n\n"+
				"• Nessun indirizzo email viene richiesto o salvato (anti-spam).\n"+
				"• Il profilo è pubblicato direttamente su "+getCommunityEndpoint()+"\n"+
				"• La bio ha un limite massimo di 200 caratteri.\n"+
				"• I link social sono opzionali e supportano handle Fediverse.",
			w,
		)
	})

	btnRow := container.NewHBox(saveBtn, clearBtn, infoBtn, layout.NewSpacer())

	topBar := container.NewVBox(
		btnRow,
		statusLabel,
		validationError,
		widget.NewSeparator(),
	)

	// Nuovo ordine del layout:
	// 1. Spunta principale: Join the Penguins Eggs Community
	// 2. Nickname / Public handle
	// 3. Location / City (e.g. Roma, IT)
	// 4. Repository / Website URL (source_url)
	// 5. Bio / Description (before socials, max 200 characters)
	// 6. Social Networks / Contact
	formContent := container.NewVBox(
		checkEnabled,
		widget.NewSeparator(),

		// 2. Nickname / Public handle
		widget.NewLabelWithStyle("Nickname / Public handle:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nicknameEntry,

		// 3. Location / City
		widget.NewLabelWithStyle("Location / City (e.g. Roma, IT):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		locationEntry,

		// 4. Repository / Website URL (source_url)
		widget.NewLabelWithStyle("Repository / Website URL (source_url):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sourceURLEntry,

		// 5. Bio / Description (spostata PRIMA dei social network)
		widget.NewLabelWithStyle("Bio / Description (max 200 characters):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		bioEntry,
		bioCounterLabel,
		widget.NewSeparator(),

		// 6. Social Networks / Contact
		widget.NewLabelWithStyle("Social Networks / Contact (optional):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		socialHelp,
		socialsForm,
	)

	return container.NewBorder(topBar, nil, nil, nil, container.NewScroll(formContent)), currentData
}
