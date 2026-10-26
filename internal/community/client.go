package community

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SyncCommunityProfile communicates directly with /api/community:
// If data.Enabled is true, it sends a PUT request to https://api.penguins-eggs.net/api/community/{id-or-uuid}.
// If data.Enabled is false, it sends a DELETE request to https://api.penguins-eggs.net/api/community/{id-or-uuid}.
// Fail-safe: uses a short 7s timeout and falls back to FallbackCommunityEndpoint if the primary host fails.
// Injects X-Eggs-Artisan-Token when available and captures auto-issued tokens on first creation (TOFU).
func SyncCommunityProfile(data CommunityData) error {
	uuid := GetMachineUUID()
	if uuid == "" {
		return errors.New("unable to determine machine UUID")
	}

	endpoint := Endpoint()
	endpoint = strings.TrimSuffix(endpoint, "/")

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	tok := strings.TrimSpace(data.ArtisanToken)
	if tok == "" {
		if loaded := LoadCommunityData(); loaded.ArtisanToken != "" {
			tok = strings.TrimSpace(loaded.ArtisanToken)
			data.ArtisanToken = tok
		}
	}

	setAuthHeaders := func(req *http.Request) {
		if tok != "" {
			req.Header.Set("X-Eggs-Artisan-Token", tok)
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	}

	targetID := strings.TrimSpace(data.ID)
	if targetID == "" {
		targetID = uuid
	}

	if !data.Enabled {
		delURL := fmt.Sprintf("%s/%s", endpoint, targetID)
		req, err := http.NewRequest(http.MethodDelete, delURL, nil)
		if err != nil {
			return err
		}
		setAuthHeaders(req)
		resp, err := client.Do(req)
		if err != nil && endpoint == DefaultCommunityEndpoint {
			// Try fallback host
			fallbackURL := fmt.Sprintf("%s/%s", FallbackCommunityEndpoint, targetID)
			if reqFallback, fErr := http.NewRequest(http.MethodDelete, fallbackURL, nil); fErr == nil {
				setAuthHeaders(reqFallback)
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
	if data.ID != "" {
		payloadMap["id"] = strings.TrimSpace(data.ID)
	}

	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	putURL := fmt.Sprintf("%s/%s", endpoint, targetID)
	req, err := http.NewRequest(http.MethodPut, putURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	setAuthHeaders(req)

	resp, err := client.Do(req)
	if err != nil && endpoint == DefaultCommunityEndpoint {
		// Try fallback host
		fallbackURL := fmt.Sprintf("%s/%s", FallbackCommunityEndpoint, targetID)
		if reqFallback, fErr := http.NewRequest(http.MethodPut, fallbackURL, bytes.NewReader(payloadBytes)); fErr == nil {
			reqFallback.Header.Set("Content-Type", "application/json")
			setAuthHeaders(reqFallback)
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

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	// Trust-On-First-Use: Extract artisan_token or id if returned
	var respData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &respData); err == nil {
		shouldSave := false
		if returnedTok, ok := respData["artisan_token"].(string); ok && strings.TrimSpace(returnedTok) != "" {
			trimmedTok := strings.TrimSpace(returnedTok)
			if data.ArtisanToken != trimmedTok {
				data.ArtisanToken = trimmedTok
				shouldSave = true
			}
		}
		if returnedID, ok := respData["id"].(string); ok && strings.TrimSpace(returnedID) != "" {
			trimmedID := strings.TrimSpace(returnedID)
			if data.ID != trimmedID {
				data.ID = trimmedID
				shouldSave = true
			}
		}
		if shouldSave {
			_ = SaveCommunityData(data)
		}
	}

	return nil
}

// FetchProfileWithToken retrieves the community profile for the given token from /api/community/me
// and persists it into the local community.json / community.yaml configurations.
func FetchProfileWithToken(token string) error {
	_, err := FetchProfileWithTokenData(token)
	return err
}

// FetchProfileWithTokenData fetches the profile with token and returns the populated CommunityData.
func FetchProfileWithTokenData(token string) (*CommunityData, error) {
	trimmedToken := strings.TrimSpace(token)
	if trimmedToken == "" {
		return nil, errors.New("artisan token cannot be empty")
	}

	endpoint := Endpoint()
	endpoint = strings.TrimSuffix(endpoint, "/")

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	meURL := fmt.Sprintf("%s/me", endpoint)
	req, err := http.NewRequest(http.MethodGet, meURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Eggs-Artisan-Token", trimmedToken)
	req.Header.Set("Authorization", "Bearer "+trimmedToken)

	resp, err := client.Do(req)
	if err != nil && endpoint == DefaultCommunityEndpoint {
		fallbackURL := fmt.Sprintf("%s/me", FallbackCommunityEndpoint)
		if reqFallback, fErr := http.NewRequest(http.MethodGet, fallbackURL, nil); fErr == nil {
			reqFallback.Header.Set("X-Eggs-Artisan-Token", trimmedToken)
			reqFallback.Header.Set("Authorization", "Bearer "+trimmedToken)
			if fResp, fDoErr := client.Do(reqFallback); fDoErr == nil {
				resp = fResp
				err = nil
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("network error fetching profile: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read server response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var p struct {
		ID        string            `json:"id"`
		UUID      string            `json:"uuid"`
		Nickname  string            `json:"nickname"`
		Location  string            `json:"location"`
		Bio       string            `json:"bio"`
		Contact   string            `json:"contact"`
		Socials   map[string]string `json:"socials"`
		SourceURL string            `json:"source_url"`
		Enabled   *bool             `json:"enabled"`
	}
	if err := json.Unmarshal(bodyBytes, &p); err != nil {
		return nil, fmt.Errorf("failed to parse profile JSON: %w", err)
	}

	enabled := true
	if p.Enabled != nil {
		enabled = *p.Enabled
	}

	socials := CommunitySocials{}
	if p.Socials != nil {
		socials.Facebook = CleanPlatformHandle("facebook", p.Socials["facebook"])
		socials.Gitee = CleanPlatformHandle("gitee", p.Socials["gitee"])
		socials.GitHub = CleanPlatformHandle("github", p.Socials["github"])
		socials.Telegram = CleanPlatformHandle("telegram", p.Socials["telegram"])
		socials.WeChat = CleanPlatformHandle("wechat", p.Socials["wechat"])
		socials.WhatsApp = CleanPlatformHandle("whatsapp", p.Socials["whatsapp"])
		socials.X = CleanPlatformHandle("x", p.Socials["x"])
		if socials.X == "" {
			socials.X = CleanPlatformHandle("x", p.Socials["twitter"])
		}
	}

	contact := p.Contact
	if contact == "" {
		contact = ComputePrimaryContact(socials)
	}

	id := strings.TrimSpace(p.ID)
	if id == "" {
		id = strings.TrimSpace(p.UUID)
	}

	data := CommunityData{
		ID:           id,
		ArtisanToken: trimmedToken,
		Enabled:      enabled,
		Nickname:     strings.TrimSpace(p.Nickname),
		Location:     strings.TrimSpace(p.Location),
		SourceURL:    CleanSourceURL(p.SourceURL),
		Bio:          strings.TrimSpace(p.Bio),
		Contact:      contact,
		Socials:      socials,
	}

	if err := SaveCommunityData(data); err != nil {
		return nil, fmt.Errorf("failed to save imported community data: %w", err)
	}

	return &data, nil
}
