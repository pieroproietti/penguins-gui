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
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/pieroproietti/penguins-gui/internal/community"
)

type socialField struct {
	key         string
	label       string
	placeholder string
	field       func(*community.CommunitySocials) *string
	check       *widget.Check
	entry       *widget.Entry
}

// buildCommunityPanel creates the UI container for the Community configuration panel.
func buildCommunityPanel(w fyne.Window, onUpdate func(community.CommunityData)) (fyne.CanvasObject, func() community.CommunityData) {
	var isEnabled bool
	var toggleBtn *widget.Button

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Community Token (egg_pat_...) — saved in ~/.ssh/penguins-community.secret")

	nicknameEntry := widget.NewEntry()
	nicknameEntry.SetPlaceHolder("e.g. tux_artisan or @nickname")

	locationEntry := widget.NewEntry()
	locationEntry.SetPlaceHolder("e.g. Rome, IT")

	sourceURLEntry := widget.NewEntry()
	sourceURLEntry.SetPlaceHolder("e.g. https://github.com/user/distro or https://mydistro.org")

	bioEntry := widget.NewMultiLineEntry()
	bioEntry.SetPlaceHolder("Short description of your project, distro, or interests (max 200 characters)...")
	bioEntry.Wrapping = fyne.TextWrapWord
	bioEntry.SetMinRowsVisible(3)

	bioCounterLabel := widget.NewLabel(fmt.Sprintf("0/%d characters", community.MaxBioLength))
	bioCounterLabel.TextStyle = fyne.TextStyle{Italic: true}

	socials := []*socialField{
		{
			key:         "facebook",
			label:       "Facebook",
			placeholder: "e.g. username or facebook.com/username",
			field:       func(s *community.CommunitySocials) *string { return &s.Facebook },
		},
		{
			key:         "gitee",
			label:       "Gitee",
			placeholder: "e.g. username or gitee.com/username",
			field:       func(s *community.CommunitySocials) *string { return &s.Gitee },
		},
		{
			key:         "github",
			label:       "GitHub",
			placeholder: "e.g. username or github.com/username",
			field:       func(s *community.CommunitySocials) *string { return &s.GitHub },
		},
		{
			key:         "telegram",
			label:       "Telegram",
			placeholder: "e.g. @username or t.me/username",
			field:       func(s *community.CommunitySocials) *string { return &s.Telegram },
		},
		{
			key:         "wechat",
			label:       "WeChat",
			placeholder: "e.g. WeChat ID",
			field:       func(s *community.CommunitySocials) *string { return &s.WeChat },
		},
		{
			key:         "whatsapp",
			label:       "WhatsApp",
			placeholder: "e.g. +39 333 1234567 or wa.me/number",
			field:       func(s *community.CommunitySocials) *string { return &s.WhatsApp },
		},
		{
			key:         "x",
			label:       "X (Twitter)",
			placeholder: "e.g. @username or x.com/username",
			field:       func(s *community.CommunitySocials) *string { return &s.X },
		},
	}

	socialFormObjects := make([]fyne.CanvasObject, 0, len(socials)*2)
	for _, sf := range socials {
		sf.check = widget.NewCheck(sf.label, nil)
		sf.entry = widget.NewEntry()
		sf.entry.SetPlaceHolder(sf.placeholder)
		socialFormObjects = append(socialFormObjects, sf.check, sf.entry)
	}

	socialHelp := widget.NewLabel("ℹ️ Optional: enter your public social channels. Email addresses are strictly forbidden; Fediverse handles (@user@instance.social) are allowed.")
	socialHelp.TextStyle = fyne.TextStyle{Italic: true}
	socialHelp.Wrapping = fyne.TextWrapWord

	validationError := widget.NewLabel("")
	validationError.Wrapping = fyne.TextWrapWord
	validationError.TextStyle = fyne.TextStyle{Bold: true}
	validationError.Hide()

	var currentID string
	var suppressEvents bool
	var syncTimer *time.Timer
	var syncSeq uint64
	var updateFieldStates func()

	setEnabledSilent := func(enabled bool) {
		isEnabled = enabled
		if toggleBtn != nil {
			if enabled {
				toggleBtn.SetText("Join the Penguins Eggs Community: ON")
				toggleBtn.SetIcon(theme.ConfirmIcon())
				toggleBtn.Importance = widget.HighImportance
			} else {
				toggleBtn.SetText("Join the Penguins Eggs Community: OFF")
				toggleBtn.SetIcon(theme.AccountIcon())
				toggleBtn.Importance = widget.MediumImportance
			}
			toggleBtn.Refresh()
		}
	}

	currentData := func() community.CommunityData {
		var soc community.CommunitySocials
		for _, sf := range socials {
			if sf.check.Checked {
				*sf.field(&soc) = community.CleanPlatformHandle(sf.key, sf.entry.Text)
			}
		}

		return community.CommunityData{
			ID:           currentID,
			ArtisanToken: strings.TrimSpace(tokenEntry.Text),
			Enabled:      isEnabled,
			Nickname:     strings.TrimSpace(nicknameEntry.Text),
			Location:     strings.TrimSpace(locationEntry.Text),
			SourceURL:    community.CleanSourceURL(sourceURLEntry.Text),
			Bio:          strings.TrimSpace(bioEntry.Text),
			Contact:      community.ComputePrimaryContact(soc),
			Socials:      soc,
		}
	}

	hasAnyProfileField := func() bool {
		if strings.TrimSpace(nicknameEntry.Text) != "" ||
			strings.TrimSpace(locationEntry.Text) != "" ||
			strings.TrimSpace(sourceURLEntry.Text) != "" ||
			strings.TrimSpace(bioEntry.Text) != "" {
			return true
		}
		for _, sf := range socials {
			if sf.check.Checked && strings.TrimSpace(sf.entry.Text) != "" {
				return true
			}
		}
		return false
	}

	validateInputs := func() bool {
		data := currentData()
		data.Enabled = true
		if err := community.ValidateCommunityData(data); err != nil {
			validationError.SetText("⚠️ " + err.Error())
			validationError.Show()
			return false
		}
		validationError.SetText("")
		validationError.Hide()
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

		for _, sf := range socials {
			val := *sf.field(&d.Socials)
			sf.check.SetChecked(val != "")
			sf.entry.SetText(val)
		}
	}

	// When OFF: fields remain populated (visible) but disabled.
	// When ON: fields are enabled and editable.
	updateFieldStates = func() {
		textEntries := []*widget.Entry{nicknameEntry, locationEntry, sourceURLEntry, bioEntry}
		if !isEnabled {
			tokenEntry.Disable()
			for _, e := range textEntries {
				e.Disable()
			}
			for _, sf := range socials {
				sf.check.Disable()
				sf.entry.Disable()
			}
			validationError.SetText("")
			validationError.Hide()
			return
		}

		tokenEntry.Enable()
		for _, e := range textEntries {
			e.Enable()
		}
		for _, sf := range socials {
			sf.check.Enable()
			if sf.check.Checked {
				sf.entry.Enable()
			} else {
				sf.entry.Disable()
			}
		}

		if hasAnyProfileField() {
			validateInputs()
		} else {
			validationError.SetText("")
			validationError.Hide()
		}
	}

	saveAndSyncOnEdit := func() {
		if suppressEvents || !isEnabled {
			return
		}

		valid := true
		if hasAnyProfileField() {
			valid = validateInputs()
		} else {
			validationError.SetText("")
			validationError.Hide()
		}

		data := currentData()
		_ = community.SaveCommunityData(data)
		if onUpdate != nil {
			onUpdate(data)
		}

		if !valid {
			if syncTimer != nil {
				syncTimer.Stop()
			}
			return
		}

		tok := strings.TrimSpace(tokenEntry.Text)
		onlyToken := !hasAnyProfileField() && tok != ""
		if !onlyToken && !hasAnyProfileField() {
			if syncTimer != nil {
				syncTimer.Stop()
			}
			return
		}

		syncSeq++
		seq := syncSeq
		if syncTimer != nil {
			syncTimer.Stop()
		}
		syncTimer = time.AfterFunc(800*time.Millisecond, func() {
			if onlyToken {
				imported, err := community.FetchProfileWithTokenData(tok)
				fyne.Do(func() {
					if seq != syncSeq || !isEnabled || err != nil {
						return
					}
					imported.Enabled = true
					_ = community.SaveCommunityData(*imported)
					populateFields(*imported)
					updateFieldStates()
					if onUpdate != nil {
						onUpdate(*imported)
					}
				})
				return
			}

			err := community.SyncCommunityProfile(data)
			fyne.Do(func() {
				if seq != syncSeq || !isEnabled || err != nil {
					return
				}
				saved := community.LoadCommunityData()
				if saved.ArtisanToken != "" && saved.ArtisanToken != strings.TrimSpace(tokenEntry.Text) {
					prev := suppressEvents
					suppressEvents = true
					tokenEntry.SetText(saved.ArtisanToken)
					suppressEvents = prev
				}
				if saved.ID != "" {
					currentID = saved.ID
				}
				if onUpdate != nil {
					onUpdate(saved)
				}
			})
		})
	}

	onTextChange := func(string) {
		saveAndSyncOnEdit()
	}

	for _, sf := range socials {
		s := sf
		s.check.OnChanged = func(checked bool) {
			if checked && isEnabled {
				s.entry.Enable()
			} else {
				s.entry.Disable()
			}
			saveAndSyncOnEdit()
		}
		s.entry.OnChanged = onTextChange
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
		saveAndSyncOnEdit()
	}

	// Single ON/OFF button drives community participation:
	// - ON (true):
	//   * Enables fields for editing.
	//   * If only Token is present (profile fields empty): imports profile from server via token and populates fields.
	//   * If profile fields are present: validates, saves locally, and publishes (PUT) profile to Community server.
	// - OFF (false):
	//   * Disables fields (keeping their values visible and saved locally) and removes public profile from Community server (DELETE).
	toggleBtn = widget.NewButtonWithIcon("Join the Penguins Eggs Community: OFF", theme.AccountIcon(), func() {
		if suppressEvents {
			return
		}

		targetState := !isEnabled
		tok := strings.TrimSpace(tokenEntry.Text)

		if targetState {
			// Case 1: Only token is filled and profile fields are empty -> Import from server
			if !hasAnyProfileField() && tok != "" {
				setEnabledSilent(true)
				updateFieldStates()
				toggleBtn.Disable()
				go func() {
					imported, err := community.FetchProfileWithTokenData(tok)
					fyne.Do(func() {
						toggleBtn.Enable()
						if err != nil {
							dialog.ShowError(err, w)
							return
						}
						imported.Enabled = true
						_ = community.SaveCommunityData(*imported)
						populateFields(*imported)
						setEnabledSilent(true)
						updateFieldStates()
						if onUpdate != nil {
							onUpdate(*imported)
						}
					})
				}()
				return
			}

			// Case 2: Neither profile fields nor token are filled yet -> Enable fields so the user can fill them in
			if !hasAnyProfileField() && tok == "" {
				setEnabledSilent(true)
				updateFieldStates()
				data := currentData()
				data.Enabled = true
				_ = community.SaveCommunityData(data)
				if onUpdate != nil {
					onUpdate(data)
				}
				return
			}

			// Case 3: Profile fields are already filled -> Enable fields, validate, save locally, and publish to server
			setEnabledSilent(true)
			updateFieldStates()
			if !validateInputs() {
				return
			}

			data := currentData()
			data.Enabled = true
			if err := community.SaveCommunityData(data); err != nil {
				dialog.ShowError(fmt.Errorf("failed to save local preferences: %w", err), w)
				return
			}
			if onUpdate != nil {
				onUpdate(data)
			}

			toggleBtn.Disable()
			go func() {
				err := community.SyncCommunityProfile(data)
				fyne.Do(func() {
					toggleBtn.Enable()
					if err != nil {
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
				})
			}()
			return
		}

		// OFF (targetState == false): Disable fields (keeping values visible & saved locally) and remove from community server
		if syncTimer != nil {
			syncTimer.Stop()
		}
		syncSeq++
		idToDelete := currentID
		setEnabledSilent(false)
		localDraft := currentData()
		localDraft.Enabled = false
		_ = community.SaveCommunityData(localDraft)
		updateFieldStates()
		toggleBtn.Disable()

		go func() {
			delData := community.CommunityData{
				ID:           idToDelete,
				ArtisanToken: tok,
				Enabled:      false,
			}
			_ = community.SyncCommunityProfile(delData)
			fyne.Do(func() {
				toggleBtn.Enable()
				if onUpdate != nil {
					onUpdate(localDraft)
				}
			})
		}()
	})

	// Load existing configuration
	initial := community.LoadCommunityData()
	populateFields(initial)
	setEnabledSilent(initial.Enabled)
	updateFieldStates()

	socialsForm := container.New(layout.NewFormLayout(), socialFormObjects...)

	infoBtn := widget.NewButtonWithIcon("Info", theme.InfoIcon(), func() {
		dialog.ShowInformation(
			"Penguins Community",
			"Join the global Penguins Eggs Community!\n\n"+
				"• When 'Join the Penguins Eggs Community' is OFF, your saved fields remain visible but disabled, and your public profile is removed from the server.\n"+
				"• Switch 'Join the Penguins Eggs Community' to ON to enable editing and automatically save and publish your profile (or import an existing profile if only the Token field is filled).\n"+
				"• Your private token is stored in ~/.ssh/penguins-community.secret (never in community.json/yaml).\n"+
				"• No email address is requested or stored (anti-spam).",
			w,
		)
	})

	headerRow := container.NewHBox(toggleBtn, layout.NewSpacer(), infoBtn)

	topBar := container.NewVBox(
		headerRow,
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
