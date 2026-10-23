package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestCleanSourceURLAndIsValidURL(t *testing.T) {
	cases := []struct {
		input       string
		cleaned     string
		shouldValid bool
	}{
		{"https://github.com/pieroproietti/penguins-eggs", "https://github.com/pieroproietti/penguins-eggs", true},
		{"http://my-distro.org", "http://my-distro.org", true},
		{"https//github.com/pieroproietti/penguins-gui", "https://github.com/pieroproietti/penguins-gui", true},
		{"github.com/distro/linux", "https://github.com/distro/linux", true},
		{"sourceforge.net/projects/eggs", "https://sourceforge.net/projects/eggs", true},
		{"", "", false},
		{"   ", "", false},
		{"not-a-domain", "https://not-a-domain", false},
		{"ftp://invalid-scheme.org", "ftp://invalid-scheme.org", false},
	}

	for _, tc := range cases {
		cleaned := CleanSourceURL(tc.input)
		if cleaned != tc.cleaned {
			t.Errorf("CleanSourceURL(%q) = %q, expected %q", tc.input, cleaned, tc.cleaned)
		}
		valid := IsValidURL(cleaned)
		if valid != tc.shouldValid {
			t.Errorf("IsValidURL(%q) = %v, expected %v", cleaned, valid, tc.shouldValid)
		}
	}
}

func TestBioCharacterLimit(t *testing.T) {
	// 200 runes should be valid
	validBio := strings.Repeat("A", 200)
	// 201 runes should fail
	invalidBio := strings.Repeat("A", 201)

	validData := CommunityData{
		Enabled:   true,
		SourceURL: "https://github.com/distro/linux",
		Bio:       validBio,
	}
	if err := ValidateCommunityData(validData); err != nil {
		t.Errorf("expected 200-char bio to be valid, got: %v", err)
	}

	invalidData := validData
	invalidData.Bio = invalidBio
	if err := ValidateCommunityData(invalidData); err == nil {
		t.Errorf("expected 201-char bio to fail validation")
	}
}

func TestCleanPlatformHandle(t *testing.T) {
	cases := []struct {
		platform string
		input    string
		expected string
	}{
		{"facebook", "pierofb", "pierofb"},
		{"facebook", "https://facebook.com/pierofb", "pierofb"},
		{"facebook", "https//facebook.com/pierofb", "pierofb"},
		{"gitee", "piero", "piero"},
		{"gitee", "https://gitee.com/piero", "piero"},
		{"gitee", "https//gitee.com/piero", "piero"},
		{"github", "pieroproietti", "pieroproietti"},
		{"github", "@pieroproietti", "pieroproietti"},
		{"github", "github.com/pieroproietti", "pieroproietti"},
		{"github", "https://github.com/pieroproietti", "pieroproietti"},
		{"github", "https//github.com/pieroproietti", "pieroproietti"},
		{"github", "http://github.com/pieroproietti/", "pieroproietti"},
		{"github", "www.github.com/pieroproietti", "pieroproietti"},
		{"telegram", "ziopele", "@ziopele"},
		{"telegram", "@ziopele", "@ziopele"},
		{"telegram", "t.me/ziopele", "@ziopele"},
		{"telegram", "https://t.me/ziopele", "@ziopele"},
		{"wechat", "wxid_123456", "wxid_123456"},
		{"whatsapp", "+39333123456", "+39333123456"},
		{"whatsapp", "+39 333 123456", "+39333123456"},
		{"whatsapp", "https://wa.me/39333123456", "39333123456"},
		{"x", "pierox", "@pierox"},
		{"x", "@pierox", "@pierox"},
		{"x", "x.com/pierox", "@pierox"},
		{"x", "https://twitter.com/pierox", "@pierox"},
	}

	for _, tc := range cases {
		got := CleanPlatformHandle(tc.platform, tc.input)
		if got != tc.expected {
			t.Errorf("CleanPlatformHandle(%q, %q) = %q, expected %q", tc.platform, tc.input, got, tc.expected)
		}
	}
}

func TestFormatContactAndAlphabeticalComputePrimaryContact(t *testing.T) {
	// Must NEVER produce "github.com/https//github.com/..."
	typoInput := "https//github.com/pieroproietti"
	if got := FormatContact("github", typoInput); got != "https://github.com/pieroproietti" {
		t.Errorf("expected 'https://github.com/pieroproietti', got %q", got)
	}

	if got := FormatContact("facebook", "pierofb"); got != "https://facebook.com/pierofb" {
		t.Errorf("expected 'https://facebook.com/pierofb', got %q", got)
	}
	if got := FormatContact("gitee", "piero"); got != "https://gitee.com/piero" {
		t.Errorf("expected 'https://gitee.com/piero', got %q", got)
	}
	if got := FormatContact("github", "pieroproietti"); got != "https://github.com/pieroproietti" {
		t.Errorf("expected 'https://github.com/pieroproietti', got %q", got)
	}
	if got := FormatContact("telegram", "piero"); got != "@piero" {
		t.Errorf("expected '@piero', got %q", got)
	}
	if got := FormatContact("wechat", "wxid_test"); got != "wxid_test" {
		t.Errorf("expected 'wxid_test', got %q", got)
	}
	if got := FormatContact("whatsapp", "+393331234567"); got != "https://wa.me/393331234567" {
		t.Errorf("expected 'https://wa.me/393331234567', got %q", got)
	}
	if got := FormatContact("x", "pierox"); got != "@pierox" {
		t.Errorf("expected '@pierox', got %q", got)
	}

	// Alphabetical order: Facebook comes before GitHub and Telegram
	socialsFB := CommunitySocials{
		Facebook: "pierofb",
		GitHub:   "pieroproietti",
		Telegram: "@ziopele",
	}
	primaryFB := ComputePrimaryContact(socialsFB)
	if primaryFB != "https://facebook.com/pierofb" {
		t.Errorf("expected alphabetical first to be Facebook, got %q", primaryFB)
	}

	// Alphabetical order: Gitee comes before GitHub
	socialsGitee := CommunitySocials{
		Gitee:    "pierocn",
		GitHub:   "pieroproietti",
		Telegram: "@ziopele",
	}
	primaryGitee := ComputePrimaryContact(socialsGitee)
	if primaryGitee != "https://gitee.com/pierocn" {
		t.Errorf("expected alphabetical first to be Gitee, got %q", primaryGitee)
	}

	// Alphabetical order: GitHub comes before Telegram
	socialsGitHub := CommunitySocials{
		GitHub:   "pieroproietti",
		Telegram: "@ziopele",
		X:        "@pierox",
	}
	primaryGitHub := ComputePrimaryContact(socialsGitHub)
	if primaryGitHub != "https://github.com/pieroproietti" {
		t.Errorf("expected alphabetical first to be GitHub, got %q", primaryGitHub)
	}

	// Alphabetical order: Telegram comes before WeChat, WhatsApp, X
	socialsTelegram := CommunitySocials{
		Telegram: "@ziopele",
		WeChat:   "wxid_test",
		WhatsApp: "+393331234567",
		X:        "@pierox",
	}
	primaryTelegram := ComputePrimaryContact(socialsTelegram)
	if primaryTelegram != "@ziopele" {
		t.Errorf("expected alphabetical first to be Telegram, got %q", primaryTelegram)
	}
}

func TestValidateCommunityData(t *testing.T) {
	// Disabled participation does not require email or source URL
	disabled := CommunityData{Enabled: false}
	if err := ValidateCommunityData(disabled); err != nil {
		t.Errorf("expected disabled to be valid without email or source, got %v", err)
	}

	// Enabled with invalid SourceURL fails
	enabledInvalidSource := CommunityData{
		Enabled:   true,
		SourceURL: "http//invalid-url",
	}
	if err := ValidateCommunityData(enabledInvalidSource); err == nil {
		t.Errorf("expected error when source_url is invalid")
	}

	// Enabled with valid source URL succeeds without email
	validData := CommunityData{
		Enabled:   true,
		Nickname:  "tux_artisan",
		Location:  "Rome, IT",
		SourceURL: "https://github.com/distro/linux",
		Bio:       "Linux distro artisan",
		Socials: CommunitySocials{
			GitHub: "pieroproietti",
		},
	}
	if err := ValidateCommunityData(validData); err != nil {
		t.Errorf("unexpected error on valid community data: %v", err)
	}

	// Email forbidden in source URL
	emailInSource := validData
	emailInSource.SourceURL = "mailto:admin@distro.org"
	if err := ValidateCommunityData(emailInSource); err == nil {
		t.Errorf("expected error when email is entered in source URL")
	}

	// Email forbidden in public social fields
	emailInGitHub := validData
	emailInGitHub.Socials.GitHub = "user@example.com"
	if err := ValidateCommunityData(emailInGitHub); err == nil {
		t.Errorf("expected error when email is entered in social field")
	}

	// Email forbidden in nickname
	emailInNick := validData
	emailInNick.Nickname = "user@example.com"
	if err := ValidateCommunityData(emailInNick); err == nil {
		t.Errorf("expected error when email is entered in nickname")
	}

	// Email forbidden in bio
	emailInBio := validData
	emailInBio.Bio = "Contact me at hello@domain.org"
	if err := ValidateCommunityData(emailInBio); err == nil {
		t.Errorf("expected error when email is entered in bio")
	}

	// Fediverse handle allowed
	fediData := validData
	fediData.Socials.X = "@artisan@fosstodon.org"
	if err := ValidateCommunityData(fediData); err != nil {
		t.Errorf("expected Fediverse handle to be allowed, got: %v", err)
	}
}

func TestSaveAndLoadCommunityData(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	initial := LoadCommunityData()
	if initial.Enabled {
		t.Errorf("expected initial Enabled to be false")
	}

	data := CommunityData{
		Enabled:   true,
		Nickname:  "tux_artisan",
		Location:  "Rome, IT",
		SourceURL: "https//github.com/pieroproietti/penguins-gui", // typo cleaned
		Socials: CommunitySocials{
			Facebook: "pierofb",
			Gitee:    "pierocn",
			GitHub:   "https//github.com/pieroproietti", // typo cleaned
			Telegram: "@ziopele",
			WeChat:   "wx_artisan",
			WhatsApp: "+39 333 1234567",
			X:        "@pierox",
		},
		Bio: "Penguins Eggs builder and Linux artisan",
	}

	if err := SaveCommunityData(data); err != nil {
		t.Fatalf("SaveCommunityData failed: %v", err)
	}

	cfgFile := filepath.Join(tmpDir, "penguins-eggs", "community.json")
	if _, err := os.Stat(cfgFile); err != nil {
		t.Fatalf("community.json was not created: %v", err)
	}

	yamlFile := filepath.Join(tmpDir, "penguins-eggs", "custom.yaml")
	yamlBytes, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Fatalf("custom.yaml was not created: %v", err)
	}
	if strings.Contains(string(yamlBytes), "email") {
		t.Errorf("custom.yaml should NOT contain email: %s", string(yamlBytes))
	}

	loaded := LoadCommunityData()
	if !loaded.Enabled {
		t.Fatalf("expected loaded.Enabled == true")
	}
	if loaded.Nickname != "tux_artisan" {
		t.Errorf("expected nickname 'tux_artisan', got %q", loaded.Nickname)
	}
	if loaded.SourceURL != "https://github.com/pieroproietti/penguins-gui" {
		t.Errorf("expected cleaned source_url 'https://github.com/pieroproietti/penguins-gui', got %q", loaded.SourceURL)
	}
	if loaded.Bio != "Penguins Eggs builder and Linux artisan" {
		t.Errorf("expected bio to match, got %q", loaded.Bio)
	}
	if loaded.Socials.GitHub != "pieroproietti" {
		t.Errorf("expected cleaned github 'pieroproietti', got %q", loaded.Socials.GitHub)
	}
	if loaded.Socials.WhatsApp != "+393331234567" {
		t.Errorf("expected cleaned whatsapp '+393331234567', got %q", loaded.Socials.WhatsApp)
	}

	// Primary contact should be Facebook (alphabetically first)
	if loaded.Contact != "https://facebook.com/pierofb" {
		t.Errorf("expected primary contact 'https://facebook.com/pierofb', got %q", loaded.Contact)
	}
}

func TestEmailNeverExposedInContact(t *testing.T) {
	socials := CommunitySocials{
		Telegram: "@ziopele",
	}
	contact := ComputePrimaryContact(socials)
	if strings.Contains(contact, "@") && strings.Contains(contact, ".") && !strings.HasPrefix(contact, "@") {
		t.Errorf("contact looks like an exposed email: %s", contact)
	}

	// Even if someone only provided empty socials, contact must remain empty
	socialsEmpty := CommunitySocials{}
	emptyContact := ComputePrimaryContact(socialsEmpty)
	if emptyContact != "" {
		t.Errorf("contact should be empty when no socials are provided, got %q", emptyContact)
	}
}

func TestBackwardCompatibilityWithMalformedCustomYaml(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfgDir := filepath.Join(tmpDir, "penguins-eggs")
	_ = os.MkdirAll(cfgDir, 0755)

	malformedYAML := []byte(`community:
  bio: Linux Artisan
  contact: github.com/https//github.com/pieroproietti
  enabled: true
  email: artisan@example.com
  location: Rome, IT
  nickname: artisan
  source_url: https//github.com/pieroproietti/penguins-eggs
  socials:
    github: https//github.com/pieroproietti
    telegram: '@pieroproietti'
`)
	_ = os.WriteFile(filepath.Join(cfgDir, "custom.yaml"), malformedYAML, 0644)

	loaded := LoadCommunityData()
	if !loaded.Enabled {
		t.Fatalf("expected loaded.Enabled == true")
	}
	if loaded.SourceURL != "https://github.com/pieroproietti/penguins-eggs" {
		t.Errorf("expected cleaned source_url 'https://github.com/pieroproietti/penguins-eggs', got %q", loaded.SourceURL)
	}
	if loaded.Socials.GitHub != "pieroproietti" {
		t.Errorf("expected cleaned github 'pieroproietti', got %q", loaded.Socials.GitHub)
	}
	if loaded.Contact != "https://github.com/pieroproietti" {
		t.Errorf("expected corrected contact 'https://github.com/pieroproietti', got %q", loaded.Contact)
	}

	// Resaving must produce clean JSON without email
	if err := SaveCommunityData(loaded); err != nil {
		t.Fatalf("SaveCommunityData failed: %v", err)
	}

	savedBytes, err := os.ReadFile(filepath.Join(cfgDir, "community.json"))
	if err != nil {
		t.Fatalf("expected community.json to exist: %v", err)
	}
	savedStr := string(savedBytes)
	if strings.Contains(savedStr, "github.com/https") {
		t.Errorf("saved JSON still contains invalid string: %s", savedStr)
	}
	if strings.Contains(savedStr, `"email"`) {
		t.Errorf("saved JSON should NOT contain email: %s", savedStr)
	}
	if !strings.Contains(savedStr, `"source_url": "https://github.com/pieroproietti/penguins-eggs"`) {
		t.Errorf("saved JSON missing source_url: %s", savedStr)
	}
}

func TestBuildCommunityPanel(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	w := test.NewWindow(nil)
	defer w.Close()

	panel, getData := buildCommunityPanel(w, nil)
	if panel == nil {
		t.Fatalf("expected non-nil panel")
	}
	if getData == nil {
		t.Fatalf("expected non-nil getData function")
	}

	data := getData()
	if data.Enabled {
		t.Errorf("expected initially disabled data")
	}
}

func TestSyncCommunityProfile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	var lastMethod string
	var lastPath string
	var lastBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		lastBody = body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	t.Setenv("EGGS_COMMUNITY_ENDPOINT", srv.URL+"/api/community")

	// 1. Sync enabled profile -> PUT /api/community/{uuid}
	data := CommunityData{
		Enabled:   true,
		Nickname:  "artisan_sync",
		Location:  "Rome, IT",
		SourceURL: "https://github.com/artisan/distro",
		Bio:       "Distro creator",
		Socials: CommunitySocials{
			GitHub: "artisan",
		},
	}
	if err := SyncCommunityProfile(data); err != nil {
		t.Fatalf("SyncCommunityProfile enabled failed: %v", err)
	}

	if lastMethod != http.MethodPut {
		t.Errorf("expected PUT method, got %s", lastMethod)
	}
	if !strings.HasPrefix(lastPath, "/api/community/") {
		t.Errorf("expected path starting with /api/community/, got %s", lastPath)
	}
	var sentPayload map[string]interface{}
	if err := json.Unmarshal(lastBody, &sentPayload); err != nil {
		t.Fatalf("failed to parse sent JSON: %v", err)
	}
	if sentPayload["nickname"] != "artisan_sync" || sentPayload["location"] != "Rome, IT" {
		t.Errorf("unexpected payload content: %+v", sentPayload)
	}
	if _, hasEmail := sentPayload["email"]; hasEmail {
		t.Errorf("sent payload must NEVER include email field")
	}

	// 2. Sync disabled profile -> DELETE /api/community/{uuid}
	dataDisabled := CommunityData{
		Enabled: false,
	}
	if err := SyncCommunityProfile(dataDisabled); err != nil {
		t.Fatalf("SyncCommunityProfile disabled failed: %v", err)
	}

	if lastMethod != http.MethodDelete {
		t.Errorf("expected DELETE method, got %s", lastMethod)
	}
	if !strings.HasPrefix(lastPath, "/api/community/") {
		t.Errorf("expected path starting with /api/community/, got %s", lastPath)
	}
}

func TestGetMachineUUID(t *testing.T) {
	uuid := GetMachineUUID()
	if uuid == "" {
		t.Fatalf("GetMachineUUID returned empty string")
	}
	if len(uuid) < 8 {
		t.Errorf("GetMachineUUID returned unexpectedly short string: %s", uuid)
	}
}
