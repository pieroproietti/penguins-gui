/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

// Package community manages artisan identity (via email hash and Gravatar),
// machine UUID derivation, and synchronization with the community API.
package community

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

const (
	DefaultCommunityEndpoint  = "https://api.penguins-eggs.net/api/community"
	FallbackCommunityEndpoint = "https://penguins-eggs.net/api/community"
)

// CommunityData represents local state for community participation.
// Note: ArtisanToken (egg-pat) is managed strictly in memory and is NEVER saved to disk.
type CommunityData struct {
	AuthorID     string `yaml:"author_id,omitempty" json:"author_id,omitempty"`
	Email        string `yaml:"email,omitempty" json:"email,omitempty"`
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	ArtisanToken string `yaml:"-" json:"-"` // in-memory only
}

// ComputeAuthorID normalizes the email (lowercase and trimmed) and returns its hexadecimal SHA-256 hash.
func ComputeAuthorID(email string) string {
	cleaned := strings.ToLower(strings.TrimSpace(email))
	if cleaned == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(cleaned))
	return hex.EncodeToString(hash[:])
}

// GravatarURL returns the Gravatar avatar URL for the given email with fallback identicon.
func GravatarURL(email string, size int) string {
	hash := ComputeAuthorID(email)
	if hash == "" {
		hash = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	if size <= 0 {
		size = 96
	}
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=%d&d=identicon", hash, size)
}

// Endpoint returns the active community API endpoint URL.
func Endpoint() string {
	if ep := strings.TrimSpace(os.Getenv("EGGS_COMMUNITY_ENDPOINT")); ep != "" {
		return ep
	}
	if ep := strings.TrimSpace(os.Getenv("CENSUS_COMMUNITY_URL")); ep != "" {
		return ep
	}
	return DefaultCommunityEndpoint
}
