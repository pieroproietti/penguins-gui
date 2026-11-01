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
	"regexp"
	"strings"
)

var (
	emailValidationPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// IsValidEmail verifies if an email address conforms to standard syntax.
func IsValidEmail(email string) bool {
	return emailValidationPattern.MatchString(strings.TrimSpace(email))
}

// ValidateCommunityInput validates email and token input based on connection state.
func ValidateCommunityInput(email, token string, enabled bool) error {
	if enabled {
		if !IsValidEmail(email) {
			return errors.New("please enter a valid email address")
		}
		return nil
	}
	// When disconnecting/removing from server, the token must be provided
	if strings.TrimSpace(token) == "" {
		return errors.New("your egg-pat token is required to deactivate and remove this machine from the server")
	}
	return nil
}
