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
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	emailValidationPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	emailForbiddenPattern  = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
)

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
