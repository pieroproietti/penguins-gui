package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
)

const MaxBioLength = 200

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
// Email is mandatory when enabled, but is strictly private and NEVER exposed in public logs or metadata.
type CommunityData struct {
	Enabled   bool             `yaml:"enabled" json:"enabled"`
	Email     string           `yaml:"email,omitempty" json:"email,omitempty"`
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

// ValidateCommunityData validates community fields: email and source URL are required when enabled,
// bio must not exceed MaxBioLength (200 characters), and public fields must never contain emails.
func ValidateCommunityData(data CommunityData) error {
	if !data.Enabled {
		return nil
	}

	trimmedEmail := strings.TrimSpace(data.Email)
	if trimmedEmail == "" {
		return errors.New("email address is required")
	}
	if !IsValidEmail(trimmedEmail) {
		return errors.New("please enter a valid email address (e.g. user@example.com)")
	}

	trimmedSource := strings.TrimSpace(data.SourceURL)
	if trimmedSource == "" {
		return errors.New("source repository or website URL is required (e.g. https://github.com/user/distro)")
	}
	cleanSource := CleanSourceURL(trimmedSource)
	if !IsValidURL(cleanSource) {
		return errors.New("please enter a valid source repository or website URL")
	}
	if err := ValidateContact(trimmedSource); err != nil {
		return fmt.Errorf("source URL: %w", err)
	}

	if len([]rune(data.Bio)) > MaxBioLength {
		return fmt.Errorf("bio must not exceed %d characters (currently %d)", MaxBioLength, len([]rune(data.Bio)))
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

			// Clean email, source URL, and bio
			comm.Email = strings.TrimSpace(comm.Email)
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
			Email:     strings.TrimSpace(data.Email),
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

	return os.WriteFile(path, out, 0644)
}

// buildCommunityPanel creates the UI container for the Community configuration panel.
func buildCommunityPanel(w fyne.Window, onUpdate func(CommunityData)) (fyne.CanvasObject, func() CommunityData) {
	title := widget.NewLabelWithStyle("Community", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	checkEnabled := widget.NewCheck("Enable Community", nil)

	// Base profile fields
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("e.g. artisan@example.com (required, kept strictly private)")

	emailHelp := widget.NewLabel("🔒 Required: Your email will NEVER be published, exposed, or shared publicly.")
	emailHelp.TextStyle = fyne.TextStyle{Italic: true}
	emailHelp.Wrapping = fyne.TextWrapWord

	nicknameEntry := widget.NewEntry()
	nicknameEntry.SetPlaceHolder("e.g. tux_artisan")

	locationEntry := widget.NewEntry()
	locationEntry.SetPlaceHolder("e.g. Rome, Madrid, Buenos Aires, Beijing...")

	sourceURLEntry := widget.NewEntry()
	sourceURLEntry.SetPlaceHolder("e.g. https://github.com/user/project or https://distro.org")

	sourceURLHelp := widget.NewLabel("🌐 Required: Source code repository, issue tracker, or project website.")
	sourceURLHelp.TextStyle = fyne.TextStyle{Italic: true}
	sourceURLHelp.Wrapping = fyne.TextWrapWord

	// Bio entry with 200 characters maximum length
	bioEntry := widget.NewMultiLineEntry()
	bioEntry.SetPlaceHolder("Short description of your project, distro, or interests (max 200 characters)...")
	bioEntry.Wrapping = fyne.TextWrapWord
	bioEntry.SetMinRowsVisible(3)

	bioCounterLabel := widget.NewLabel(fmt.Sprintf("0/%d characters", MaxBioLength))
	bioCounterLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Social Networks in alphabetical order:
	// 1. Facebook
	// 2. Gitee
	// 3. GitHub
	// 4. Telegram
	// 5. WeChat
	// 6. WhatsApp
	// 7. X (Twitter)

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

	socialHelp := widget.NewLabel("ℹ️ Optional: Select any social links to share publicly. Email addresses cannot be entered here.")
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
			Email:     strings.TrimSpace(emailEntry.Text),
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
			emailEntry.Enable()
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

			statusLabel.SetText("Participation active: enter your required private email, source URL, optional profile details, and save preferences.")
		} else {
			emailEntry.Disable()
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
			statusLabel.SetText("Community participation is disabled.")
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
	emailEntry.OnChanged = onTextChange
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
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d characters", len(runes), MaxBioLength))
		validateInputs()
	}

	entryFacebook.OnChanged = onTextChange
	entryGitee.OnChanged = onTextChange
	entryGitHub.OnChanged = onTextChange
	entryTelegram.OnChanged = onTextChange
	entryWeChat.OnChanged = onTextChange
	entryWhatsApp.OnChanged = onTextChange
	entryX.OnChanged = onTextChange

	saveBtn = widget.NewButtonWithIcon("Save Preferences", theme.DocumentSaveIcon(), func() {
		if checkEnabled.Checked && !validateInputs() {
			dialog.ShowError(errors.New("please correct the validation errors before saving"), w)
			return
		}
		data := currentData()
		if err := SaveCommunityData(data); err != nil {
			dialog.ShowError(fmt.Errorf("unable to save preferences: %w", err), w)
			return
		}
		if data.Enabled {
			statusLabel.SetText("✅ Preferences saved to " + getEggsConfigPath())
		} else {
			statusLabel.SetText("✅ Preferences saved (participation disabled).")
		}
		if onUpdate != nil {
			onUpdate(data)
		}
	})

	clearBtn := widget.NewButtonWithIcon("Disable / Clear Fields", theme.ContentClearIcon(), func() {
		checkEnabled.SetChecked(false)
		emailEntry.SetText("")
		nicknameEntry.SetText("")
		locationEntry.SetText("")
		sourceURLEntry.SetText("")
		bioEntry.SetText("")
		bioCounterLabel.SetText(fmt.Sprintf("0/%d characters", MaxBioLength))
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
	})

	checkEnabled.OnChanged = func(checked bool) {
		updateFieldStates()
		if !checked {
			data := currentData()
			_ = SaveCommunityData(data)
			if onUpdate != nil {
				onUpdate(data)
			}
		}
	}

	// Load existing configuration
	initial := LoadCommunityData()
	if initial.Enabled {
		checkEnabled.SetChecked(true)
		emailEntry.SetText(initial.Email)
		nicknameEntry.SetText(initial.Nickname)
		locationEntry.SetText(initial.Location)
		sourceURLEntry.SetText(initial.SourceURL)
		bioEntry.SetText(initial.Bio)
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d characters", len([]rune(initial.Bio)), MaxBioLength))

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
			"Community",
			"Join the global Penguins Eggs community!\n\n"+
				"Connect with other Linux artisans and remasters.\n\n"+
				"• Email is required for registration and will never be exposed under any circumstances.\n"+
				"• Source repository or website is required.\n"+
				"• Social links are optional and in alphabetical order.",
			w,
		)
	})

	btnRow := container.NewHBox(saveBtn, clearBtn, infoBtn, layout.NewSpacer())

	topBar := container.NewVBox(
		title,
		btnRow,
		statusLabel,
		validationError,
		widget.NewSeparator(),
	)

	// Profile Information & Main Contacts -> Bio (max 200 chars) -> Social Networks (alphabetical)
	formContent := container.NewVBox(
		checkEnabled,
		widget.NewSeparator(),

		// Profile Information
		widget.NewLabelWithStyle("Profile Information:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Email (required - strictly private):"),
		emailEntry,
		emailHelp,
		widget.NewLabel("Nickname / Public handle:"),
		nicknameEntry,
		widget.NewLabel("Location / City:"),
		locationEntry,
		widget.NewLabel("Source / Repository / Website (required):"),
		sourceURLEntry,
		sourceURLHelp,

		// Bio / Description immediately below main profile contacts (before Social Networks)
		widget.NewLabelWithStyle("Bio / Description (optional, max 200 characters):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		bioEntry,
		bioCounterLabel,
		widget.NewSeparator(),

		// Social Networks in alphabetical order
		widget.NewLabelWithStyle("Social Networks (alphabetical order, optional):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		socialHelp,
		socialsForm,
	)

	return container.NewBorder(topBar, nil, nil, nil, container.NewScroll(formContent)), currentData
}
