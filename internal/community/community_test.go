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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputeAuthorID(t *testing.T) {
	testEmail := "  Artisan.Penguin@Example.COM  "
	expectedClean := "artisan.penguin@example.com"
	expectedSum := sha256.Sum256([]byte(expectedClean))
	expectedHex := hex.EncodeToString(expectedSum[:])

	authorID := ComputeAuthorID(testEmail)
	if authorID != expectedHex {
		t.Errorf("ComputeAuthorID(%q) = %q, expected %q", testEmail, authorID, expectedHex)
	}

	if emptyID := ComputeAuthorID("   "); emptyID != "" {
		t.Errorf("expected empty string for blank email, got %q", emptyID)
	}
}

func TestGravatarURL(t *testing.T) {
	email := "user@example.org"
	expectedHash := ComputeAuthorID(email)
	url := GravatarURL(email, 96)

	if !strings.Contains(url, "https://www.gravatar.com/avatar/"+expectedHash) {
		t.Errorf("expected Gravatar URL containing hash %s, got %s", expectedHash, url)
	}
	if !strings.Contains(url, "s=96") {
		t.Errorf("expected size s=96 in URL, got %s", url)
	}
	if !strings.Contains(url, "d=identicon") {
		t.Errorf("expected fallback identicon in URL, got %s", url)
	}
}

func TestIsValidEmail(t *testing.T) {
	valid := []string{
		"artisan@penguins-eggs.net",
		"tux.artisan@sub.example.com",
		"user+tag@domain.co.uk",
	}
	for _, em := range valid {
		if !IsValidEmail(em) {
			t.Errorf("expected %q to be valid email", em)
		}
	}

	invalid := []string{
		"",
		"   ",
		"not-an-email",
		"user@domain",
		"@domain.com",
		"user@",
	}
	for _, em := range invalid {
		if IsValidEmail(em) {
			t.Errorf("expected %q to be invalid email", em)
		}
	}
}

func TestStorageNoTokenOnDisk(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))

	data := CommunityData{
		AuthorID:     ComputeAuthorID("artisan@example.com"),
		Email:        "artisan@example.com",
		Enabled:      true,
		ArtisanToken: "egg_pat_super_secret_in_memory_only",
	}

	if err := SaveCommunityData(data); err != nil {
		t.Fatalf("SaveCommunityData failed: %v", err)
	}

	cfgDir := filepath.Join(tmpDir, "config", "penguins-eggs")
	jsonPath := filepath.Join(cfgDir, "community.json")
	yamlPath := filepath.Join(cfgDir, "community.yaml")

	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("community.json was not created: %v", err)
	}
	if strings.Contains(string(jsonBytes), "egg_pat_") || strings.Contains(string(jsonBytes), "ArtisanToken") {
		t.Errorf("community.json must NEVER contain the secret token: %s", string(jsonBytes))
	}

	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("community.yaml was not created: %v", err)
	}
	if strings.Contains(string(yamlBytes), "egg_pat_") || strings.Contains(string(yamlBytes), "artisan_token") {
		t.Errorf("community.yaml must NEVER contain the secret token: %s", string(yamlBytes))
	}

	// Verify no secret file was written to disk in ~/.ssh
	secretPath := filepath.Join(tmpDir, ".ssh", "penguins-community.secret")
	if _, err := os.Stat(secretPath); !os.IsNotExist(err) {
		t.Errorf("secret file must NOT exist on disk: %s", secretPath)
	}

	// Verify LoadCommunityData returns token as empty
	loaded := LoadCommunityData()
	if loaded.ArtisanToken != "" {
		t.Errorf("expected loaded.ArtisanToken to be empty, got %q", loaded.ArtisanToken)
	}
	if loaded.Email != "artisan@example.com" {
		t.Errorf("expected loaded.Email to be 'artisan@example.com', got %q", loaded.Email)
	}
	if loaded.AuthorID != data.AuthorID {
		t.Errorf("expected loaded.AuthorID to match %q, got %q", data.AuthorID, loaded.AuthorID)
	}
	if !loaded.Enabled {
		t.Errorf("expected loaded.Enabled == true")
	}
}

func TestStorageSudoUser(t *testing.T) {
	curr, err := user.Current()
	if err != nil {
		t.Skip("unable to get current user")
	}

	t.Setenv("SUDO_USER", curr.Username)
	t.Setenv("SUDO_UID", curr.Uid)
	t.Setenv("SUDO_GID", curr.Gid)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(curr.HomeDir, ".config-test-penguins"))

	cfgDir, err := getEggsConfigDir()
	if err != nil {
		t.Fatalf("getEggsConfigDir failed: %v", err)
	}
	expected := filepath.Join(curr.HomeDir, ".config-test-penguins", "penguins-eggs")
	if cfgDir != expected {
		t.Errorf("expected %s, got %s", expected, cfgDir)
	}
}

func TestStorageSudoUserRootXDG(t *testing.T) {
	curr, err := user.Current()
	if err != nil {
		t.Skip("unable to get current user")
	}

	t.Setenv("SUDO_USER", curr.Username)
	t.Setenv("SUDO_UID", curr.Uid)
	t.Setenv("SUDO_GID", curr.Gid)
	t.Setenv("XDG_CONFIG_HOME", "/root/.config")

	cfgDir, err := getEggsConfigDir()
	if err != nil {
		t.Fatalf("getEggsConfigDir failed: %v", err)
	}
	expected := filepath.Join(curr.HomeDir, ".config", "penguins-eggs")
	if cfgDir != expected {
		t.Errorf("expected %s ignoring /root/.config, got %s", expected, cfgDir)
	}
}

func TestSyncCommunityProfile_CaseA_TOFU(t *testing.T) {
	var lastMethod, lastPath string
	var lastHeaders http.Header
	var lastBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		lastHeaders = r.Header.Clone()
		lastBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"ok", "author_id":"12345", "artisan_token":"egg_pat_tofu_generated_123"}`))
	}))
	defer srv.Close()

	t.Setenv("EGGS_COMMUNITY_ENDPOINT", srv.URL+"/api/community")

	email := "new.artisan@example.com"
	authorID := ComputeAuthorID(email)
	data := CommunityData{
		AuthorID:     authorID,
		Email:        email,
		Enabled:      true,
		ArtisanToken: "", // Case A: empty token
	}

	newTok, err := SyncCommunityProfile(data)
	if err != nil {
		t.Fatalf("SyncCommunityProfile Case A failed: %v", err)
	}

	if newTok != "egg_pat_tofu_generated_123" {
		t.Errorf("expected generated token 'egg_pat_tofu_generated_123', got %q", newTok)
	}

	if lastMethod != http.MethodPut {
		t.Errorf("expected PUT method, got %s", lastMethod)
	}
	if !strings.HasSuffix(lastPath, "/"+authorID) {
		t.Errorf("expected path ending in %s, got %s", authorID, lastPath)
	}
	if tokHeader := lastHeaders.Get("X-Eggs-Artisan-Token"); tokHeader != "" {
		t.Errorf("expected no X-Eggs-Artisan-Token header on TOFU, got %s", tokHeader)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(lastBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal request payload: %v", err)
	}
	if payload["author_id"] != authorID {
		t.Errorf("expected payload author_id %s, got %v", authorID, payload["author_id"])
	}
	if payload["email"] != email {
		t.Errorf("expected payload email %s, got %v", email, payload["email"])
	}
}

func TestSyncCommunityProfile_CaseB_WithToken(t *testing.T) {
	var lastHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	t.Setenv("EGGS_COMMUNITY_ENDPOINT", srv.URL+"/api/community")

	email := "existing.artisan@example.com"
	data := CommunityData{
		AuthorID:     ComputeAuthorID(email),
		Email:        email,
		Enabled:      true,
		ArtisanToken: "egg_pat_existing_secret_999",
	}

	_, err := SyncCommunityProfile(data)
	if err != nil {
		t.Fatalf("SyncCommunityProfile Case B failed: %v", err)
	}

	if sent := lastHeaders.Get("X-Eggs-Artisan-Token"); sent != "egg_pat_existing_secret_999" {
		t.Errorf("expected X-Eggs-Artisan-Token 'egg_pat_existing_secret_999', got %q", sent)
	}
	if auth := lastHeaders.Get("Authorization"); auth != "Bearer egg_pat_existing_secret_999" {
		t.Errorf("expected Authorization 'Bearer egg_pat_existing_secret_999', got %q", auth)
	}
}

func TestSyncCommunityProfile_CaseB_401Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	t.Setenv("EGGS_COMMUNITY_ENDPOINT", srv.URL+"/api/community")

	data := CommunityData{
		Email:        "someone_else@example.com",
		Enabled:      true,
		ArtisanToken: "",
	}

	_, err := SyncCommunityProfile(data)
	if !errors.Is(err, ErrUnauthorizedEmail) {
		t.Errorf("expected ErrUnauthorizedEmail, got %v", err)
	}
}

func TestSyncCommunityProfile_CaseC_Deactivation(t *testing.T) {
	var lastMethod string
	var lastHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("EGGS_COMMUNITY_ENDPOINT", srv.URL+"/api/community")

	email := "artisan@example.com"
	// Missing token on deactivation must fail
	dataNoToken := CommunityData{
		AuthorID:     ComputeAuthorID(email),
		Email:        email,
		Enabled:      false,
		ArtisanToken: "",
	}
	_, errNoTok := SyncCommunityProfile(dataNoToken)
	if !errors.Is(errNoTok, ErrTokenRequired) {
		t.Errorf("expected ErrTokenRequired, got %v", errNoTok)
	}

	// Deactivation with token
	dataWithToken := CommunityData{
		AuthorID:     ComputeAuthorID(email),
		Email:        email,
		Enabled:      false,
		ArtisanToken: "egg_pat_delete_auth_123",
	}
	_, err := SyncCommunityProfile(dataWithToken)
	if err != nil {
		t.Fatalf("deactivation failed: %v", err)
	}

	if lastMethod != http.MethodDelete {
		t.Errorf("expected DELETE method, got %s", lastMethod)
	}
	if sent := lastHeaders.Get("X-Eggs-Artisan-Token"); sent != "egg_pat_delete_auth_123" {
		t.Errorf("expected token on DELETE header, got %q", sent)
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
