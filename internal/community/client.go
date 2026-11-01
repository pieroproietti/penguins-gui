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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrUnauthorizedEmail = errors.New("this email is already registered. Please paste your valid egg-pat token to associate this machine")
	ErrTokenRequired     = errors.New("your egg-pat token is required to deactivate and remove this machine from the server")
	ErrMissingEmail      = errors.New("please enter a valid email address")
)

func doCommunityRequest(method, pathSuffix, token string, payload []byte) (*http.Response, error) {
	endpoint := strings.TrimSuffix(Endpoint(), "/")
	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	buildReq := func(baseURL string) (*http.Request, error) {
		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(payload)
		}
		req, err := http.NewRequest(method, baseURL+"/"+strings.TrimPrefix(pathSuffix, "/"), bodyReader)
		if err != nil {
			return nil, err
		}
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("X-Eggs-Artisan-Token", token)
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return req, nil
	}

	req, err := buildReq(endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil && endpoint == DefaultCommunityEndpoint {
		if reqFallback, fErr := buildReq(FallbackCommunityEndpoint); fErr == nil {
			if fResp, fDoErr := client.Do(reqFallback); fDoErr == nil {
				return fResp, nil
			}
		}
	}
	return resp, err
}

// SyncCommunityProfile communicates with /api/community:
//
// Case A (First run / TOFU, empty token):
//   Sends PUT with author_id and uuid without auth headers.
//   If author_id is newly created, server returns artisan_token which is returned to caller.
//
// Case B (Existing author, token provided):
//   Sends PUT with author_id, uuid, and auth headers (X-Eggs-Artisan-Token / Authorization).
//   If 401 Unauthorized, returns ErrUnauthorizedEmail.
//
// Case C (Deactivation):
//   Sends DELETE with author_id and uuid. Requires token (ErrTokenRequired if empty).
func SyncCommunityProfile(data CommunityData) (string, error) {
	uuid := GetMachineUUID()
	if uuid == "" {
		return "", errors.New("unable to determine machine UUID")
	}

	authorID := strings.TrimSpace(data.AuthorID)
	email := strings.TrimSpace(strings.ToLower(data.Email))
	if authorID == "" {
		if email != "" {
			authorID = ComputeAuthorID(email)
		}
	}

	if authorID == "" {
		return "", ErrMissingEmail
	}

	tok := strings.TrimSpace(data.ArtisanToken)

	if !data.Enabled {
		if tok == "" {
			return "", ErrTokenRequired
		}
		delPayload, _ := json.Marshal(map[string]interface{}{
			"author_id": authorID,
			"uuid":      uuid,
			"enabled":   false,
		})
		resp, err := doCommunityRequest(http.MethodDelete, authorID, tok, delPayload)
		if err != nil {
			return "", fmt.Errorf("network error during removal: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return "", errors.New("invalid egg-pat token. Unable to remove machine")
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return "", nil
	}

	payloadMap := map[string]interface{}{
		"author_id": authorID,
		"uuid":      uuid,
		"email":     email,
		"enabled":   true,
	}

	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return "", err
	}

	resp, err := doCommunityRequest(http.MethodPut, authorID, tok, payloadBytes)
	if err != nil {
		return "", fmt.Errorf("network error during update: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorizedEmail
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var respData map[string]interface{}
	newTok := ""
	if err := json.Unmarshal(bodyBytes, &respData); err == nil {
		if t, ok := respData["artisan_token"].(string); ok && strings.TrimSpace(t) != "" {
			newTok = strings.TrimSpace(t)
		} else if t, ok := respData["token"].(string); ok && strings.TrimSpace(t) != "" {
			newTok = strings.TrimSpace(t)
		} else if t, ok := respData["egg_pat"].(string); ok && strings.TrimSpace(t) != "" {
			newTok = strings.TrimSpace(t)
		}
	}

	return newTok, nil
}

// FetchGravatarImage downloads the raw avatar image bytes from Gravatar.
func FetchGravatarImage(email string, size int) ([]byte, error) {
	u := GravatarURL(email, size)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "penguins-gui/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gravatar returned status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
