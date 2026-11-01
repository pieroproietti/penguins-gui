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
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/pieroproietti/penguins-gui/internal/community"
)

// buildCommunityPanel creates the UI container for the simplified Community configuration panel.
// Identity is strictly based on email hash (Gravatar compatible) and authorization via egg-pat in memory.
func buildCommunityPanel(w fyne.Window, onUpdate func(community.CommunityData)) (fyne.CanvasObject, func() community.CommunityData) {
	initial := community.LoadCommunityData()
	isEnabled := initial.Enabled

	// Avatar preview
	avatarImg := canvas.NewImageFromResource(theme.AccountIcon())
	avatarImg.FillMode = canvas.ImageFillContain
	avatarImg.SetMinSize(fyne.NewSize(96, 96))
	avatarBox := container.NewCenter(container.NewGridWrap(fyne.NewSize(96, 96), avatarImg))

	// Email Entry & Author ID
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Your email (e.g. artisan@example.com)")

	authorIDLabel := widget.NewLabel("")
	authorIDLabel.TextStyle = fyne.TextStyle{Italic: true, Monospace: true}

	// Token Entry & Copy Button (in-memory only)
	tokenEntry := widget.NewEntry()
	tokenEntry.SetPlaceHolder("Paste your egg_pat_... token here (leave blank on first run)")

	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		tok := strings.TrimSpace(tokenEntry.Text)
		if tok == "" {
			if w != nil {
				dialog.ShowInformation("Token", "No token available to copy.", w)
			}
			return
		}
		if w != nil && w.Clipboard() != nil {
			w.Clipboard().SetContent(tok)
			dialog.ShowInformation("Copy", "Token copied to clipboard!", w)
		}
	})

	tokenHelp := widget.NewLabel("ℹ️ Leave blank if this is your first time connecting to generate a new token.\n⚠️ The token is held in memory only and never saved to disk: keep it safe.")
	tokenHelp.Wrapping = fyne.TextWrapWord
	tokenHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Status and Action Controls
	statusLabel := widget.NewLabel("")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	feedbackLabel := widget.NewLabel("")
	feedbackLabel.Wrapping = fyne.TextWrapWord
	feedbackLabel.Hide()

	var actionBtn *widget.Button

	updateUIState := func() {
		if isEnabled {
			statusLabel.SetText("Status: Machine connected to Community ✅")
			if actionBtn != nil {
				actionBtn.SetText("Remove Machine from Server")
				actionBtn.SetIcon(theme.CancelIcon())
				actionBtn.Importance = widget.DangerImportance
			}
			emailEntry.Disable()
		} else {
			statusLabel.SetText("Status: Machine not connected ⚪")
			if actionBtn != nil {
				actionBtn.SetText("Connect Machine to Community")
				actionBtn.SetIcon(theme.ConfirmIcon())
				actionBtn.Importance = widget.HighImportance
			}
			emailEntry.Enable()
		}
		if actionBtn != nil {
			actionBtn.Refresh()
		}
	}

	var gravatarTimer *time.Timer
	updateGravatar := func(email string) {
		clean := strings.TrimSpace(strings.ToLower(email))
		if clean == "" {
			authorIDLabel.SetText("Author ID: unavailable")
			avatarImg.Resource = theme.AccountIcon()
			avatarImg.Refresh()
			return
		}
		authID := community.ComputeAuthorID(clean)
		if len(authID) >= 16 {
			authorIDLabel.SetText(fmt.Sprintf("Author ID (SHA-256): %s...", authID[:16]))
		} else {
			authorIDLabel.SetText(fmt.Sprintf("Author ID: %s", authID))
		}

		if gravatarTimer != nil {
			gravatarTimer.Stop()
		}
		gravatarTimer = time.AfterFunc(350*time.Millisecond, func() {
			data, err := community.FetchGravatarImage(clean, 96)
			fyne.Do(func() {
				if err == nil && len(data) > 0 {
					avatarImg.Resource = fyne.NewStaticResource("gravatar.png", data)
					avatarImg.Refresh()
				}
			})
		})
	}

	emailEntry.OnChanged = func(s string) {
		feedbackLabel.Hide()
		updateGravatar(s)
	}

	actionBtn = widget.NewButtonWithIcon("Connect Machine to Community", theme.ConfirmIcon(), func() {
		feedbackLabel.Hide()
		email := strings.TrimSpace(strings.ToLower(emailEntry.Text))
		tok := strings.TrimSpace(tokenEntry.Text)

		if !isEnabled {
			// Case A (TOFU) or Case B (Existing author with token)
			if !community.IsValidEmail(email) {
				feedbackLabel.SetText("⚠️ Please enter a valid email address before proceeding.")
				feedbackLabel.Show()
				return
			}

			actionBtn.Disable()
			statusLabel.SetText("Connecting to server...")

			go func() {
				commData := community.CommunityData{
					AuthorID:     community.ComputeAuthorID(email),
					Email:        email,
					ArtisanToken: tok,
					Enabled:      true,
				}
				newTok, err := community.SyncCommunityProfile(commData)
				fyne.Do(func() {
					actionBtn.Enable()
					if err != nil {
						updateUIState()
						if errors.Is(err, community.ErrUnauthorizedEmail) {
							msg := "⚠️ This email is already registered. Please paste your valid egg-pat token to associate this machine."
							feedbackLabel.SetText(msg)
							feedbackLabel.Show()
							if w != nil {
								dialog.ShowError(errors.New("This email is already registered.\nPlease paste your valid egg-pat token to associate this machine."), w)
							}
						} else {
							feedbackLabel.SetText(fmt.Sprintf("⚠️ Error: %v", err))
							feedbackLabel.Show()
							if w != nil {
								dialog.ShowError(err, w)
							}
						}
						return
					}

					isEnabled = true
					commData.Enabled = true
					_ = community.SaveCommunityData(commData) // Persists author_id, email, enabled (NO TOKEN)
					if onUpdate != nil {
						onUpdate(commData)
					}
					updateUIState()

					if newTok != "" {
						tokenEntry.SetText(newTok)
						showGeneratedTokenDialog(w, newTok)
					} else {
						if w != nil {
							dialog.ShowInformation("Community", "✅ Machine successfully connected to your workshop!", w)
						}
					}
				})
			}()
		} else {
			// Case C: Deactivation / Removal
			if tok == "" {
				msg := "⚠️ Your egg-pat token is required in the Community Token field to remove this machine from the server."
				feedbackLabel.SetText(msg)
				feedbackLabel.Show()
				if w != nil {
					dialog.ShowError(errors.New("Your egg-pat token is required in the Community Token field to remove this machine from the server."), w)
				}
				return
			}

			if w == nil {
				return
			}

			confirmDialog := dialog.NewConfirm("Remove Machine",
				"Are you sure you want to disconnect and remove this machine from your workshop in the Community?",
				func(confirmed bool) {
					if !confirmed {
						return
					}
					actionBtn.Disable()
					statusLabel.SetText("Removing from server...")

					go func() {
						commData := community.CommunityData{
							AuthorID:     community.ComputeAuthorID(email),
							Email:        email,
							ArtisanToken: tok,
							Enabled:      false,
						}
						_, err := community.SyncCommunityProfile(commData)
						fyne.Do(func() {
							actionBtn.Enable()
							if err != nil {
								updateUIState()
								feedbackLabel.SetText(fmt.Sprintf("⚠️ Removal error: %v", err))
								feedbackLabel.Show()
								if w != nil {
									dialog.ShowError(err, w)
								}
								return
							}

							isEnabled = false
							commData.Enabled = false
							_ = community.SaveCommunityData(commData)
							if onUpdate != nil {
								onUpdate(commData)
							}
							updateUIState()
							if w != nil {
								dialog.ShowInformation("Community", "Machine successfully removed from Community.", w)
							}
						})
					}()
				}, w)
			confirmDialog.Show()
		}
	})

	// Prepopulate if saved preferences exist
	if initial.Email != "" {
		emailEntry.SetText(initial.Email)
		updateGravatar(initial.Email)
	} else if initial.AuthorID != "" {
		authorIDLabel.SetText(fmt.Sprintf("Author ID: %s", initial.AuthorID))
	}
	updateUIState()

	// Layout
	headerTitle := widget.NewLabelWithStyle("Penguins Eggs Community", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	headerDesc := widget.NewLabel("Connect this machine to your artisan workshop in the global Penguins' Eggs Community.\nIdentity is derived from your email SHA-256 hash (Gravatar compatible).\nThe egg-pat authorization token is kept in memory only and never saved to disk.")
	headerDesc.Wrapping = fyne.TextWrapWord
	headerDesc.TextStyle = fyne.TextStyle{Italic: true}

	emailSection := container.NewVBox(
		widget.NewLabelWithStyle("Author Email:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		emailEntry,
		authorIDLabel,
	)

	identityRow := container.NewBorder(nil, nil, avatarBox, nil, container.NewPadded(emailSection))

	tokenRow := container.NewBorder(nil, nil, nil, copyBtn, tokenEntry)
	tokenSection := container.NewVBox(
		widget.NewLabelWithStyle("Community Token (egg-pat):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tokenRow,
		tokenHelp,
	)

	actionSection := container.NewVBox(
		statusLabel,
		actionBtn,
		feedbackLabel,
	)

	content := container.NewVBox(
		headerTitle,
		headerDesc,
		widget.NewSeparator(),
		identityRow,
		widget.NewSeparator(),
		tokenSection,
		widget.NewSeparator(),
		actionSection,
	)

	currentDataFunc := func() community.CommunityData {
		em := strings.TrimSpace(strings.ToLower(emailEntry.Text))
		authID := ""
		if em != "" {
			authID = community.ComputeAuthorID(em)
		} else {
			authID = initial.AuthorID
		}
		return community.CommunityData{
			AuthorID:     authID,
			Email:        em,
			Enabled:      isEnabled,
			ArtisanToken: strings.TrimSpace(tokenEntry.Text),
		}
	}

	return container.NewVScroll(container.NewPadded(content)), currentDataFunc
}

func showGeneratedTokenDialog(parent fyne.Window, token string) {
	if parent == nil {
		return
	}

	copyToClipboard := func(text string) {
		if fyne.CurrentApp() != nil && fyne.CurrentApp().Clipboard() != nil {
			fyne.CurrentApp().Clipboard().SetContent(text)
		} else if parent.Clipboard() != nil {
			parent.Clipboard().SetContent(text)
		}
	}

	copyToClipboard(token)

	headerLabel := widget.NewLabelWithStyle("🎉 Machine successfully connected!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	infoLabel := widget.NewLabel("Your new Community Token (egg-pat) has been generated:")
	infoLabel.Wrapping = fyne.TextWrapWord

	tokenEntry := widget.NewEntry()
	tokenEntry.TextStyle = fyne.TextStyle{Monospace: true}
	tokenEntry.SetText(token)
	tokenEntry.OnChanged = func(s string) {
		if s != token {
			tokenEntry.SetText(token)
		}
	}

	statusLabel := widget.NewLabel("✅ Token automatically copied to clipboard.")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	copyTokenBtn := widget.NewButtonWithIcon("Copy Token", theme.ContentCopyIcon(), func() {
		copyToClipboard(token)
		statusLabel.SetText("✅ Token copied to clipboard!")
	})

	fullMessageText := fmt.Sprintf(
		"Penguins Eggs Community Token (egg-pat):\n%s\n\n⚠️ IMPORTANT: This token is held in memory only and never saved to disk.\nKeep it safe: you will need it to connect other machines or deactivate the profile.",
		token,
	)

	copyAllBtn := widget.NewButtonWithIcon("Copy All", theme.FileTextIcon(), func() {
		copyToClipboard(fullMessageText)
		statusLabel.SetText("✅ Full message copied to clipboard!")
	})

	tokenRow := container.NewBorder(nil, nil, nil, container.NewHBox(copyTokenBtn, copyAllBtn), tokenEntry)

	warningLabel := widget.NewLabel("⚠️ IMPORTANT: This token is held in memory only and never saved to disk.\nKeep it safe: you will need it to connect other machines or deactivate the profile.")
	warningLabel.Wrapping = fyne.TextWrapWord
	warningLabel.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewVBox(
		headerLabel,
		widget.NewSeparator(),
		infoLabel,
		tokenRow,
		statusLabel,
		widget.NewSeparator(),
		warningLabel,
	)

	d := dialog.NewCustom("Generated Token (egg-pat)", "I saved the token", content, parent)
	d.SetIcon(theme.InfoIcon())
	d.Resize(fyne.NewSize(620, 300))
	d.Show()
}
