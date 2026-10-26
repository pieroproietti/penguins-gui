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

	"github.com/pieroproietti/penguins-gui/internal/census"
)

// buildCommunityPanel creates the UI container for the Community configuration panel.
func buildCommunityPanel(w fyne.Window, onUpdate func(census.CommunityData)) (fyne.CanvasObject, func() census.CommunityData) {
	checkEnabled := widget.NewCheck("Join the Penguins Eggs Community", nil)

	// Artisan Token (for multi-machine synchronization)
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Personal Artisan Token (egg_pat_...) - Optional / Auto-generated")

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

	bioCounterLabel := widget.NewLabel(fmt.Sprintf("0/%d characters", census.MaxBioLength))
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

	var saveBtn *widget.Button
	var currentID string
	var updateFieldStates func()

	importTokenBtn := widget.NewButtonWithIcon("Import with Token", theme.DownloadIcon(), func() {
		tok := strings.TrimSpace(tokenEntry.Text)
		if tok == "" {
			dialog.ShowError(errors.New("insert an Artisan Token to import an existing profile"), w)
			return
		}
		statusLabel.SetText("Importing profile from community server...")
		go func() {
			imported, err := census.FetchProfileWithTokenData(tok)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText("⚠️ Import failed: " + err.Error())
					dialog.ShowError(err, w)
					return
				}
				currentID = imported.ID
				checkEnabled.SetChecked(imported.Enabled)
				nicknameEntry.SetText(imported.Nickname)
				locationEntry.SetText(imported.Location)
				sourceURLEntry.SetText(imported.SourceURL)
				bioEntry.SetText(imported.Bio)
				bioCounterLabel.SetText(fmt.Sprintf("%d/%d characters", len([]rune(imported.Bio)), census.MaxBioLength))

				checkFacebook.SetChecked(imported.Socials.Facebook != "")
				entryFacebook.SetText(imported.Socials.Facebook)
				checkGitee.SetChecked(imported.Socials.Gitee != "")
				entryGitee.SetText(imported.Socials.Gitee)
				checkGitHub.SetChecked(imported.Socials.GitHub != "")
				entryGitHub.SetText(imported.Socials.GitHub)
				checkTelegram.SetChecked(imported.Socials.Telegram != "")
				entryTelegram.SetText(imported.Socials.Telegram)
				checkWeChat.SetChecked(imported.Socials.WeChat != "")
				entryWeChat.SetText(imported.Socials.WeChat)
				checkWhatsApp.SetChecked(imported.Socials.WhatsApp != "")
				entryWhatsApp.SetText(imported.Socials.WhatsApp)
				checkX.SetChecked(imported.Socials.X != "")
				entryX.SetText(imported.Socials.X)

				updateFieldStates()
				if onUpdate != nil {
					onUpdate(*imported)
				}
				statusLabel.SetText(fmt.Sprintf("✅ Profile @%s imported successfully!", imported.Nickname))
			})
		}()
	})

	currentData := func() census.CommunityData {
		socials := census.CommunitySocials{}
		if checkFacebook.Checked {
			socials.Facebook = census.CleanPlatformHandle("facebook", entryFacebook.Text)
		}
		if checkGitee.Checked {
			socials.Gitee = census.CleanPlatformHandle("gitee", entryGitee.Text)
		}
		if checkGitHub.Checked {
			socials.GitHub = census.CleanPlatformHandle("github", entryGitHub.Text)
		}
		if checkTelegram.Checked {
			socials.Telegram = census.CleanPlatformHandle("telegram", entryTelegram.Text)
		}
		if checkWeChat.Checked {
			socials.WeChat = census.CleanPlatformHandle("wechat", entryWeChat.Text)
		}
		if checkWhatsApp.Checked {
			socials.WhatsApp = census.CleanPlatformHandle("whatsapp", entryWhatsApp.Text)
		}
		if checkX.Checked {
			socials.X = census.CleanPlatformHandle("x", entryX.Text)
		}

		primary := census.ComputePrimaryContact(socials)

		return census.CommunityData{
			ID:           currentID,
			ArtisanToken: strings.TrimSpace(tokenEntry.Text),
			Enabled:      checkEnabled.Checked,
			Nickname:     strings.TrimSpace(nicknameEntry.Text),
			Location:     strings.TrimSpace(locationEntry.Text),
			SourceURL:    census.CleanSourceURL(sourceURLEntry.Text),
			Bio:          strings.TrimSpace(bioEntry.Text),
			Contact:      primary,
			Socials:      socials,
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
		if err := census.ValidateCommunityData(data); err != nil {
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

	updateFieldStates = func() {
		enabled := checkEnabled.Checked
		if enabled {
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

			statusLabel.SetText("Partecipazione attiva: compila i campi e premi Salva per sincronizzare il profilo con la community.")
		} else {
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
			statusLabel.SetText("Partecipazione alla community disattivata.")
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
	nicknameEntry.OnChanged = onTextChange
	locationEntry.OnChanged = onTextChange
	sourceURLEntry.OnChanged = onTextChange

	// Bio length constraint and visual counter
	bioEntry.OnChanged = func(s string) {
		runes := []rune(s)
		if len(runes) > census.MaxBioLength {
			bioEntry.SetText(string(runes[:census.MaxBioLength]))
			runes = []rune(bioEntry.Text)
		}
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d caratteri", len(runes), census.MaxBioLength))
		validateInputs()
	}

	entryFacebook.OnChanged = onTextChange
	entryGitee.OnChanged = onTextChange
	entryGitHub.OnChanged = onTextChange
	entryTelegram.OnChanged = onTextChange
	entryWeChat.OnChanged = onTextChange
	entryWhatsApp.OnChanged = onTextChange
	entryX.OnChanged = onTextChange

	saveBtn = widget.NewButtonWithIcon("Salva / Aggiorna profilo Community", theme.DocumentSaveIcon(), func() {
		if checkEnabled.Checked && !validateInputs() {
			dialog.ShowError(errors.New("correggi gli errori di validazione prima di salvare"), w)
			return
		}
		data := currentData()
		if err := census.SaveCommunityData(data); err != nil {
			dialog.ShowError(fmt.Errorf("impossibile salvare le preferenze locali: %w", err), w)
			return
		}

		statusLabel.SetText("Sincronizzazione in corso con il server community...")
		if onUpdate != nil {
			onUpdate(data)
		}

		go func() {
			err := census.SyncCommunityProfile(data)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText("⚠️ Salvato localmente, errore server: " + err.Error())
				} else {
					saved := census.LoadCommunityData()
					if saved.ArtisanToken != "" {
						tokenEntry.SetText(saved.ArtisanToken)
					}
					currentID = saved.ID
					if data.Enabled {
						statusLabel.SetText("✅ Profilo salvato e sincronizzato con successo su " + census.Endpoint())
					} else {
						statusLabel.SetText("✅ Profilo disattivato e rimosso dal server.")
					}
				}
			})
		}()
	})

	clearBtn := widget.NewButtonWithIcon("Disabilita / Cancella Campi", theme.ContentClearIcon(), func() {
		checkEnabled.SetChecked(false)
		tokenEntry.SetText("")
		currentID = ""
		nicknameEntry.SetText("")
		locationEntry.SetText("")
		sourceURLEntry.SetText("")
		bioEntry.SetText("")
		bioCounterLabel.SetText(fmt.Sprintf("0/%d caratteri", census.MaxBioLength))
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
		data := census.CommunityData{Enabled: false}
		_ = census.SaveCommunityData(data)
		if onUpdate != nil {
			onUpdate(data)
		}

		statusLabel.SetText("Rimozione profilo dal server in corso...")
		go func() {
			err := census.SyncCommunityProfile(data)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText("⚠️ Preferenze azzerate localmente, errore server: " + err.Error())
				} else {
					statusLabel.SetText("✅ Profilo disabilitato e rimosso dal server.")
				}
			})
		}()
	})

	checkEnabled.OnChanged = func(checked bool) {
		updateFieldStates()
		if !checked {
			data := currentData()
			_ = census.SaveCommunityData(data)
			if onUpdate != nil {
				onUpdate(data)
			}
			go func() {
				_ = census.SyncCommunityProfile(data)
			}()
		}
	}

	// Load existing configuration
	initial := census.LoadCommunityData()
	currentID = initial.ID
	if initial.ArtisanToken != "" {
		tokenEntry.SetText(initial.ArtisanToken)
	}
	if initial.Enabled {
		checkEnabled.SetChecked(true)
		nicknameEntry.SetText(initial.Nickname)
		locationEntry.SetText(initial.Location)
		sourceURLEntry.SetText(initial.SourceURL)
		bioEntry.SetText(initial.Bio)
		bioCounterLabel.SetText(fmt.Sprintf("%d/%d caratteri", len([]rune(initial.Bio)), census.MaxBioLength))

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
			"Penguins Community",
			"Partecipa alla community globale di Penguins Eggs!\n\n"+
				"Connettiti con altri sviluppatori e artigiani di distribuzioni Linux.\n\n"+
				"• Nessun indirizzo email viene richiesto o salvato (anti-spam).\n"+
				"• Il profilo è pubblicato direttamente su "+census.Endpoint()+"\n"+
				"• La bio ha un limite massimo di 200 caratteri.\n"+
				"• I link social sono opzionali e supportano handle Fediverse.",
			w,
		)
	})

	btnRow := container.NewHBox(saveBtn, clearBtn, infoBtn, layout.NewSpacer())

	topBar := container.NewVBox(
		btnRow,
		statusLabel,
		validationError,
		widget.NewSeparator(),
	)

	tokenRow := container.NewBorder(nil, nil, nil, importTokenBtn, tokenEntry)

	formContent := container.NewVBox(
		checkEnabled,
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Artisan Token (optional / multi-machine sync):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tokenRow,
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
