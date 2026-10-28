/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package ui

import (
	"errors"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/pieroproietti/penguins-gui/internal/community"
)

// buildCommunityPanel creates the UI container for the Community configuration panel.
func buildCommunityPanel(w fyne.Window, onUpdate func(community.CommunityData)) (fyne.CanvasObject, func() community.CommunityData) {
	checkEnabled := widget.NewCheck("Join the Penguins Eggs Community", nil)

	// Community Token (stored in ~/.ssh/penguins-community.secret for multi-machine synchronization)
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Community Token (egg_pat_...) — saved in ~/.ssh/penguins-community.secret")

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

	bioCounterLabel := widget.NewLabel(fmt.Sprintf("0/%d characters", community.MaxBioLength))
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

	var currentID string
	var suppressEvents bool
	var updateFieldStates func()

	setCheckedSilent := func(checked bool) {
		prev := suppressEvents
		suppressEvents = true
		checkEnabled.SetChecked(checked)
		suppressEvents = prev
	}

	currentData := func() community.CommunityData {
		socials := community.CommunitySocials{}
		if checkFacebook.Checked {
			socials.Facebook = community.CleanPlatformHandle("facebook", entryFacebook.Text)
		}
		if checkGitee.Checked {
			socials.Gitee = community.CleanPlatformHandle("gitee", entryGitee.Text)
		}
		if checkGitHub.Checked {
			socials.GitHub = community.CleanPlatformHandle("github", entryGitHub.Text)
		}
		if checkTelegram.Checked {
			socials.Telegram = community.CleanPlatformHandle("telegram", entryTelegram.Text)
		}
		if checkWeChat.Checked {
			socials.WeChat = community.CleanPlatformHandle("wechat", entryWeChat.Text)
		}
		if checkWhatsApp.Checked {
			socials.WhatsApp = community.CleanPlatformHandle("whatsapp", entryWhatsApp.Text)
		}
		if checkX.Checked {
			socials.X = community.CleanPlatformHandle("x", entryX.Text)
		}

		primary := community.ComputePrimaryContact(socials)

		return community.CommunityData{
			ID:           currentID,
			ArtisanToken: strings.TrimSpace(tokenEntry.Text),
			Enabled:      checkEnabled.Checked,
			Nickname:     strings.TrimSpace(nicknameEntry.Text),
			Location:     strings.TrimSpace(locationEntry.Text),
			SourceURL:    community.CleanSourceURL(sourceURLEntry.Text),
			Bio:          strings.TrimSpace(bioEntry.Text),
			Contact:      primary,
			Socials:      socials,
		}
	}

	hasAnyProfileField := func() bool {
		return strings.TrimSpace(nicknameEntry.Text) != "" ||
			strings.TrimSpace(locationEntry.Text) != "" ||
			strings.TrimSpace(sourceURLEntry.Text) != "" ||
			strings.TrimSpace(bioEntry.Text) != "" ||
			(checkFacebook.Checked && strings.TrimSpace(entryFacebook.Text) != "") ||
			(checkGitee.Checked && strings.TrimSpace(entryGitee.Text) != "") ||
			(checkGitHub.Checked && strings.TrimSpace(entryGitHub.Text) != "") ||
			(checkTelegram.Checked && strings.TrimSpace(entryTelegram.Text) != "") ||
			(checkWeChat.Checked && strings.TrimSpace(entryWeChat.Text) != "") ||
			(checkWhatsApp.Checked && strings.TrimSpace(entryWhatsApp.Text) != "") ||
			(checkX.Checked && strings.TrimSpace(entryX.Text) != "")
	}

	validateInputs := func() bool {
		data := currentData()
		data.Enabled = true
		if err := community.ValidateCommunityData(data); err != nil {
			validationError.SetText("⚠️ " + err.Error())
			return false
		}
		validationError.SetText("")
		return true
	}

	populateFields := func(d community.CommunityData) {
		prev := suppressEvents
		suppressEvents = true
		defer func() { suppressEvents = prev }()

		currentID = d.ID
		if d.ArtisanToken != "" {
			tokenEntry.SetText(d.ArtisanToken)
		}
		nicknameEntry.SetText(d.Nickname)
		locationEntry.SetText(d.Location)
		sourceURLEntry.SetText(d.SourceURL)
		bioEntry.SetText(d.Bio)
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d characters", len([]rune(d.Bio)), community.MaxBioLength))

		checkFacebook.SetChecked(d.Socials.Facebook != "")
		entryFacebook.SetText(d.Socials.Facebook)
		checkGitee.SetChecked(d.Socials.Gitee != "")
		entryGitee.SetText(d.Socials.Gitee)
		checkGitHub.SetChecked(d.Socials.GitHub != "")
		entryGitHub.SetText(d.Socials.GitHub)
		checkTelegram.SetChecked(d.Socials.Telegram != "")
		entryTelegram.SetText(d.Socials.Telegram)
		checkWeChat.SetChecked(d.Socials.WeChat != "")
		entryWeChat.SetText(d.Socials.WeChat)
		checkWhatsApp.SetChecked(d.Socials.WhatsApp != "")
		entryWhatsApp.SetText(d.Socials.WhatsApp)
		checkX.SetChecked(d.Socials.X != "")
		entryX.SetText(d.Socials.X)
	}

	// When checked (published to Community): fields are locked (read-only).
	// When unchecked (removed from server / local mode): fields are editable.
	updateFieldStates = func() {
		published := checkEnabled.Checked
		if published {
			tokenEntry.Disable()
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
		} else {
			tokenEntry.Enable()
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

			validateInputs()
		}
	}

	saveLocalDraftQuietly := func() {
		if suppressEvents || checkEnabled.Checked {
			return
		}
		if validateInputs() {
			draft := currentData()
			draft.Enabled = false
			_ = community.SaveCommunityData(draft)
		}
	}

	// Wire social checkbox toggles
	checkFacebook.OnChanged = func(checked bool) {
		if checked && !checkEnabled.Checked {
			entryFacebook.Enable()
		} else {
			entryFacebook.Disable()
		}
		saveLocalDraftQuietly()
	}
	checkGitee.OnChanged = func(checked bool) {
		if checked && !checkEnabled.Checked {
			entryGitee.Enable()
		} else {
			entryGitee.Disable()
		}
		saveLocalDraftQuietly()
	}
	checkGitHub.OnChanged = func(checked bool) {
		if checked && !checkEnabled.Checked {
			entryGitHub.Enable()
		} else {
			entryGitHub.Disable()
		}
		saveLocalDraftQuietly()
	}
	checkTelegram.OnChanged = func(checked bool) {
		if checked && !checkEnabled.Checked {
			entryTelegram.Enable()
		} else {
			entryTelegram.Disable()
		}
		saveLocalDraftQuietly()
	}
	checkWeChat.OnChanged = func(checked bool) {
		if checked && !checkEnabled.Checked {
			entryWeChat.Enable()
		} else {
			entryWeChat.Disable()
		}
		saveLocalDraftQuietly()
	}
	checkWhatsApp.OnChanged = func(checked bool) {
		if checked && !checkEnabled.Checked {
			entryWhatsApp.Enable()
		} else {
			entryWhatsApp.Disable()
		}
		saveLocalDraftQuietly()
	}
	checkX.OnChanged = func(checked bool) {
		if checked && !checkEnabled.Checked {
			entryX.Enable()
		} else {
			entryX.Disable()
		}
		saveLocalDraftQuietly()
	}

	onTextChange := func(string) {
		saveLocalDraftQuietly()
	}
	tokenEntry.OnChanged = onTextChange
	nicknameEntry.OnChanged = onTextChange
	locationEntry.OnChanged = onTextChange
	sourceURLEntry.OnChanged = onTextChange

	bioEntry.OnChanged = func(s string) {
		runes := []rune(s)
		if len(runes) > community.MaxBioLength {
			bioEntry.SetText(string(runes[:community.MaxBioLength]))
			runes = []rune(bioEntry.Text)
		}
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d characters", len(runes), community.MaxBioLength))
		saveLocalDraftQuietly()
	}

	entryFacebook.OnChanged = onTextChange
	entryGitee.OnChanged = onTextChange
	entryGitHub.OnChanged = onTextChange
	entryTelegram.OnChanged = onTextChange
	entryWeChat.OnChanged = onTextChange
	entryWhatsApp.OnChanged = onTextChange
	entryX.OnChanged = onTextChange

	// Single checkbox drives everything:
	// - Checked (true):
	//   * If only Token is present (profile fields empty): imports profile from server via token, populates & enables.
	//   * Otherwise: validates fields, saves locally, and publishes (PUT) profile to Community server.
	// - Unchecked (false):
	//   * Removes profile data from Community server (DELETE) while keeping all local fields & token editable.
	checkEnabled.OnChanged = func(checked bool) {
		if suppressEvents {
			return
		}

		tok := strings.TrimSpace(tokenEntry.Text)

		if checked {
			// Case 1: Only token is filled and profile fields are empty -> Import from server
			if !hasAnyProfileField() && tok != "" {
				updateFieldStates()
				statusLabel.SetText("Importing profile from community server with token...")
				go func() {
					imported, err := community.FetchProfileWithTokenData(tok)
					fyne.Do(func() {
						if err != nil {
							setCheckedSilent(false)
							updateFieldStates()
							statusLabel.SetText("⚠️ Token import failed: " + err.Error())
							dialog.ShowError(err, w)
							return
						}
						imported.Enabled = true
						_ = community.SaveCommunityData(*imported)
						populateFields(*imported)
						setCheckedSilent(true)
						updateFieldStates()
						if onUpdate != nil {
							onUpdate(*imported)
						}
						statusLabel.SetText(fmt.Sprintf("✅ Profile @%s imported and active in the Community!", imported.Nickname))
					})
				}()
				return
			}

			// Case 2: Neither profile fields nor token are filled
			if !hasAnyProfileField() && tok == "" {
				setCheckedSilent(false)
				updateFieldStates()
				err := errors.New("please fill in your profile fields (or enter an existing Community Token) before joining the community")
				statusLabel.SetText("⚠️ " + err.Error())
				dialog.ShowError(err, w)
				return
			}

			// Case 3: Profile fields are filled -> Validate, save locally, and publish to server
			if !validateInputs() {
				setCheckedSilent(false)
				updateFieldStates()
				dialog.ShowError(errors.New("please fix validation errors before joining the community"), w)
				return
			}

			data := currentData()
			data.Enabled = true
			if err := community.SaveCommunityData(data); err != nil {
				setCheckedSilent(false)
				updateFieldStates()
				dialog.ShowError(fmt.Errorf("failed to save local preferences: %w", err), w)
				return
			}

			updateFieldStates()
			statusLabel.SetText("Publishing profile to community server...")
			go func() {
				err := community.SyncCommunityProfile(data)
				fyne.Do(func() {
					if err != nil {
						setCheckedSilent(false)
						data.Enabled = false
						_ = community.SaveCommunityData(data)
						updateFieldStates()
						statusLabel.SetText("⚠️ Failed to publish profile: " + err.Error())
						dialog.ShowError(err, w)
						return
					}
					saved := community.LoadCommunityData()
					if saved.ArtisanToken != "" {
						prev := suppressEvents
						suppressEvents = true
						tokenEntry.SetText(saved.ArtisanToken)
						suppressEvents = prev
					}
					currentID = saved.ID
					if onUpdate != nil {
						onUpdate(saved)
					}
					statusLabel.SetText("✅ Profile published to " + community.Endpoint() + " (uncheck to edit or remove from server).")
				})
			}()
			return
		}

		// Unchecked (checked == false): Remove from community server, keep local fields & token editable
		idToDelete := currentID
		localDraft := currentData()
		localDraft.Enabled = false
		_ = community.SaveCommunityData(localDraft)
		updateFieldStates()

		statusLabel.SetText("Removing profile data from community server...")
		go func() {
			delData := community.CommunityData{
				ID:           idToDelete,
				ArtisanToken: tok,
				Enabled:      false,
			}
			err := community.SyncCommunityProfile(delData)
			fyne.Do(func() {
				if onUpdate != nil {
					onUpdate(localDraft)
				}
				if err != nil {
					statusLabel.SetText("⚠️ Disabled locally (fields editable), server error: " + err.Error())
				} else {
					statusLabel.SetText("✅ Data removed from community server. Local fields are preserved and editable; check the box to publish again.")
				}
			})
		}()
	}

	// Load existing configuration
	initial := community.LoadCommunityData()
	populateFields(initial)
	setCheckedSilent(initial.Enabled)
	updateFieldStates()
	if initial.Enabled {
		statusLabel.SetText("✅ Community participation active (uncheck 'Join the Penguins Eggs Community' to edit or remove data from server).")
	} else {
		statusLabel.SetText("Edit your fields below and check 'Join the Penguins Eggs Community' to publish (or paste only your Token and check to import).")
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
			"Join the global Penguins Eggs Community!\n\n"+
				"• Check 'Join the Penguins Eggs Community' to publish your profile (or import an existing profile if only the Token field is filled).\n"+
				"• Uncheck 'Join the Penguins Eggs Community' to remove your public data from the server while keeping your fields saved and editable locally.\n"+
				"• Your private token is stored in ~/.ssh/penguins-community.secret (never in community.json/yaml).\n"+
				"• No email address is requested or stored (anti-spam).",
			w,
		)
	})

	headerRow := container.NewHBox(checkEnabled, layout.NewSpacer(), infoBtn)

	topBar := container.NewVBox(
		headerRow,
		statusLabel,
		validationError,
		widget.NewSeparator(),
	)

	formContent := container.NewVBox(
		widget.NewLabelWithStyle("Community Token (~/.ssh/penguins-community.secret):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tokenEntry,
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Nickname / Public handle:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nicknameEntry,

		widget.NewLabelWithStyle("Location / City (e.g. Roma, IT):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		locationEntry,

		widget.NewLabelWithStyle("Repository / Website URL (source_url):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sourceURLEntry,

		widget.NewLabelWithStyle("Bio / Description (max 200 characters):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		bioEntry,
		bioCounterLabel,
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Social Networks / Contact (optional):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		socialHelp,
		socialsForm,
	)

	return container.NewBorder(topBar, nil, nil, nil, container.NewScroll(formContent)), currentData
}
