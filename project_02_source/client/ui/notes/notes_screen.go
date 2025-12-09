package notes

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"io"
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/crypto"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Screen displays the notes page after login
func Screen(window fyne.Window, apiClient *api.Client, username string, onLogout func()) {
	// Modern gradient background (softer purple-blue)
	gradientBg := canvas.NewRectangle(color.RGBA{R: 139, G: 92, B: 246, A: 255})

	// Header card with rounded appearance
	headerBg := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 250})
	
	headerTitle := canvas.NewText("🔐 Secure Notes", color.RGBA{R: 139, G: 92, B: 246, A: 255})
	headerTitle.TextSize = 32
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.Alignment = fyne.TextAlignCenter

	subTitle := canvas.NewText("Your encrypted notes, protected by AES-256", color.RGBA{R: 107, G: 114, B: 128, A: 255})
	subTitle.TextSize = 12
	subTitle.Alignment = fyne.TextAlignCenter

	userInfo := canvas.NewText("👤 "+username, color.RGBA{R: 59, G: 130, B: 246, A: 255})
	userInfo.TextSize = 14
	userInfo.TextStyle = fyne.TextStyle{Bold: true}
	userInfo.Alignment = fyne.TextAlignCenter

	headerContent := container.NewVBox(
		container.NewCenter(headerTitle),
		container.NewCenter(subTitle),
		widget.NewSeparator(),
		container.NewCenter(userInfo),
	)

	header := container.NewMax(
		headerBg,
		container.NewPadded(headerContent),
	)

	// Status label for feedback
	statusLabel := widget.NewLabel("")

	// Notes list container
	notesContainer := container.NewVBox()
	notesScroll := container.NewScroll(notesContainer)
	notesScroll.SetMinSize(fyne.NewSize(700, 300))

	// Refresh function - will be defined recursively
	var refreshNotes func()
	refreshNotes = func() {
		// Call API to get notes
		notes, err := apiClient.ListNotes()
		
		// Update UI in main thread
		fyne.Do(func() {
			// Clear previous list
			notesContainer.RemoveAll()
			
			if err != nil {
				statusLabel.SetText("❌ Error loading notes: " + err.Error())
				notesContainer.Add(widget.NewLabel("❌ Error loading notes"))
				notesContainer.Refresh()
				return
			}

			if len(notes) == 0 {
				notesContainer.Add(
					container.NewCenter(
						widget.NewLabel("📭 No notes yet. Create your first note!"),
					),
				)
				statusLabel.SetText("✅ No notes found")
			} else {
				statusLabel.SetText(fmt.Sprintf("✅ %d notes loaded", len(notes)))

				for _, note := range notes {
					// Create note card
					noteCard := createNoteCard(note, apiClient, window, refreshNotes)
					notesContainer.Add(noteCard)
				}
			}

			notesContainer.Refresh()
			notesScroll.Refresh()
		})
	}

	// Upload section with card background
	uploadCardBg := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 250})
	
	uploadTitle := canvas.NewText("📤 Upload New Note", color.RGBA{R: 139, G: 92, B: 246, A: 255})
	uploadTitle.TextSize = 18
	uploadTitle.TextStyle = fyne.TextStyle{Bold: true}

	uploadDesc := widget.NewLabel("Select a file to encrypt and upload securely")
	uploadDesc.TextStyle = fyne.TextStyle{Italic: true}

	// File upload button with custom style
	uploadBtn := widget.NewButton("📁 Choose File & Upload", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				statusLabel.SetText("❌ Error: " + err.Error())
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			// Read file
			content, err := io.ReadAll(reader)
			if err != nil {
				statusLabel.SetText("❌ Error reading file: " + err.Error())
				return
			}

			fileName := filepath.Base(reader.URI().String())

			// Generate DEK (Data Encryption Key)
			dek, err := crypto.GenerateKey()
			if err != nil {
				statusLabel.SetText("❌ Key generation error: " + err.Error())
				return
			}

			// Encrypt content with DEK
			encryptedContent, iv, err := crypto.EncryptAES(string(content), dek)
			if err != nil {
				statusLabel.SetText("❌ Encryption error: " + err.Error())
				return
			}

			// Derive KEK (Key Encryption Key) from user password
			kek := crypto.DeriveKeyFromPassword(api.CurrentPassword, nil)

			// Encrypt DEK with KEK
			dekBase64 := base64.StdEncoding.EncodeToString(dek)
			encryptedKey, ivKey, err := crypto.EncryptAES(dekBase64, kek)
			if err != nil {
				statusLabel.SetText("❌ Key encryption error: " + err.Error())
				return
			}

			statusLabel.SetText("⏳ Uploading...")
			
			// Upload to server
			if err := apiClient.CreateNote(fileName, encryptedContent, iv, encryptedKey, ivKey); err != nil {
				statusLabel.SetText("❌ Upload error: " + err.Error())
				return
			}

			statusLabel.SetText("✅ Note uploaded successfully!")
			refreshNotes()
		}, window)
	})
	uploadBtn.Importance = widget.HighImportance

	uploadContent := container.NewVBox(
		uploadTitle,
		uploadDesc,
		uploadBtn,
	)

	uploadForm := container.NewMax(
		uploadCardBg,
		container.NewPadded(uploadContent),
	)

	// Notes list section with card background
	notesCardBg := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 250})
	
	notesTitle := canvas.NewText("📋 Your Notes Collection", color.RGBA{R: 139, G: 92, B: 246, A: 255})
	notesTitle.TextSize = 18
	notesTitle.TextStyle = fyne.TextStyle{Bold: true}

	notesContent := container.NewVBox(
		notesTitle,
		widget.NewSeparator(),
		notesScroll,
	)

	notesSection := container.NewMax(
		notesCardBg,
		container.NewPadded(notesContent),
	)

	// Action buttons with modern style
	logoutBtn := widget.NewButton("🚪 Logout", func() {
		apiClient.Token = ""
		onLogout()
	})
	logoutBtn.Importance = widget.DangerImportance

	actionBar := container.NewHBox(
		layout.NewSpacer(),
		logoutBtn,
	)

	// Status bar with better styling
	statusContainer := container.NewHBox(
		widget.NewIcon(nil),
		statusLabel,
		layout.NewSpacer(),
	)

	// E2EE Shares tab content
	e2eeContainer := container.NewVBox()
	e2eeScroll := container.NewScroll(e2eeContainer)
	e2eeScroll.SetMinSize(fyne.NewSize(700, 300))

	// E2EE status label
	e2eeStatusLabel := widget.NewLabel("")

	// Refresh E2EE shares function
	var refreshE2EEShares func()
	refreshE2EEShares = func() {
		shares, err := apiClient.ListE2EEShares()
		
		fyne.Do(func() {
			e2eeContainer.RemoveAll()
			
			if err != nil {
				e2eeStatusLabel.SetText("❌ Error loading shares: " + err.Error())
				e2eeContainer.Add(widget.NewLabel("❌ Error loading E2EE shares"))
				e2eeContainer.Refresh()
				return
			}

			if len(shares) == 0 {
				e2eeContainer.Add(
					container.NewCenter(
						widget.NewLabel("📭 No E2EE shares received yet"),
					),
				)
				e2eeStatusLabel.SetText("✅ No E2EE shares")
			} else {
				e2eeStatusLabel.SetText(fmt.Sprintf("✅ %d E2EE shares received", len(shares)))

				for _, share := range shares {
					shareCard := createE2EEShareCard(share, apiClient, window, refreshE2EEShares)
					e2eeContainer.Add(shareCard)
				}
			}

			e2eeContainer.Refresh()
			e2eeScroll.Refresh()
		})
	}

	// E2EE shares section
	e2eeCardBg := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 250})
	e2eeTitle := canvas.NewText("🔐 Received E2EE Shares", color.RGBA{R: 139, G: 92, B: 246, A: 255})
	e2eeTitle.TextSize = 18
	e2eeTitle.TextStyle = fyne.TextStyle{Bold: true}

	e2eeDesc := widget.NewLabel("Secure shares from other users using Diffie-Hellman")
	e2eeDesc.TextStyle = fyne.TextStyle{Italic: true}

	refreshE2EEBtn := widget.NewButton("🔄 Refresh", refreshE2EEShares)

	e2eeContent := container.NewVBox(
		e2eeTitle,
		e2eeDesc,
		refreshE2EEBtn,
		widget.NewSeparator(),
		e2eeScroll,
		e2eeStatusLabel,
	)

	e2eeSection := container.NewMax(
		e2eeCardBg,
		container.NewPadded(e2eeContent),
	)

	// Shared Link Viewer tab content
	sharedLinkSection := createSharedLinkViewer(window, apiClient)

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("📋 My Notes", container.NewVBox(
			uploadForm,
			widget.NewLabel(""), // Spacer
			notesSection,
		)),
		container.NewTabItem("🔐 E2EE Shares", e2eeSection),
		container.NewTabItem("🌐 View Shared Link", sharedLinkSection),
	)

	// Main content with tabs
	mainContent := container.NewVBox(
		header,
		widget.NewLabel(""), // Spacer
		tabs,
		widget.NewLabel(""), // Spacer
		statusContainer,
		actionBar,
	)

	// Scrollable main content
	scrollContent := container.NewScroll(mainContent)
	scrollContent.SetMinSize(fyne.NewSize(850, 600))

	// Final layout with padding
	content := container.NewMax(
		gradientBg,
		container.NewPadded(
			container.NewPadded(scrollContent),
		),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(900, 700))

	// Load notes on screen open
	refreshNotes()
	refreshE2EEShares()
}

// createNoteCard creates a card widget for a single note
func createNoteCard(note api.Note, apiClient *api.Client, window fyne.Window, onRefresh func()) fyne.CanvasObject {
	// Card background with shadow effect (lighter background)
	cardBg := canvas.NewRectangle(color.RGBA{R: 249, G: 250, B: 251, A: 255})

	// Note icon based on share status
	var noteIcon string
	if note.IsShared {
		noteIcon = "🌐"
	} else {
		noteIcon = "📄"
	}

	// Note title with icon
	titleText := canvas.NewText(noteIcon+" "+note.Title, color.RGBA{R: 31, G: 41, B: 55, A: 255})
	titleText.TextSize = 16
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	// Share status badge with better styling
	var shareStatusText string
	var shareStatusBg color.Color
	var shareStatusFg color.Color
	if note.IsShared {
		shareStatusText = " SHARED "
		shareStatusBg = color.RGBA{R: 34, G: 197, B: 94, A: 255} // Green
		shareStatusFg = color.White
	} else {
		shareStatusText = " PRIVATE "
		shareStatusBg = color.RGBA{R: 156, G: 163, B: 175, A: 255} // Gray
		shareStatusFg = color.White
	}

	statusBadgeBg := canvas.NewRectangle(shareStatusBg)
	statusBadgeText := canvas.NewText(shareStatusText, shareStatusFg)
	statusBadgeText.TextSize = 10
	statusBadgeText.TextStyle = fyne.TextStyle{Bold: true}
	statusBadge := container.NewMax(
		statusBadgeBg,
		container.NewCenter(statusBadgeText),
	)
	statusBadge.Resize(fyne.NewSize(80, 20))

	titleContainer := container.NewBorder(
		nil, nil, nil, statusBadge,
		titleText,
	)

	// Note info with icons
	sizeText := canvas.NewText("💾 "+fmt.Sprintf("%d bytes", len(note.EncryptedContent)), color.RGBA{R: 107, G: 114, B: 128, A: 255})
	sizeText.TextSize = 11

	createdText := canvas.NewText("🕒 "+note.CreatedAt.Format("Jan 02, 2006 15:04"), color.RGBA{R: 107, G: 114, B: 128, A: 255})
	createdText.TextSize = 11

	infoContainer := container.NewHBox(
		sizeText,
		widget.NewLabel("  •  "),
		createdText,
	)

	// View button (to decrypt and view note)
	viewBtn := widget.NewButton("👁️ View", func() {
		showDecryptDialog(window, apiClient, note)
	})
	viewBtn.Importance = widget.HighImportance

	// Share button (to create share link)
	shareBtn := widget.NewButton("🔗 Share", func() {
		showShareTypeDialog(window, apiClient, note, onRefresh)
	})

	// Revoke button (only enabled if shared)
	revokeBtn := widget.NewButton("🚫 Revoke", func() {
		if err := apiClient.RevokeShare(note.ID); err != nil {
			dialog.ShowError(fmt.Errorf("revoke failed: %w", err), window)
			return
		}
		dialog.ShowInformation("✅ Success", "Sharing revoked successfully", window)
		onRefresh()
	})
	if !note.IsShared {
		revokeBtn.Disable()
	}

	// Delete button
	deleteBtn := widget.NewButton("🗑️ Delete", func() {
		dialog.ShowConfirm("⚠️ Delete Note", 
			fmt.Sprintf("Are you sure you want to permanently delete:\n\n'%s'\n\nThis action cannot be undone!", note.Title),
			func(confirmed bool) {
				if confirmed {
					if err := apiClient.DeleteNote(note.ID); err != nil {
						dialog.ShowError(fmt.Errorf("delete failed: %w", err), window)
						return
					}
					dialog.ShowInformation("✅ Success", "Note deleted successfully", window)
					onRefresh()
				}
			}, window)
	})
	deleteBtn.Importance = widget.DangerImportance

	// Button container with better spacing
	buttonContainer := container.NewHBox(
		viewBtn,
		shareBtn,
		revokeBtn,
		layout.NewSpacer(),
		deleteBtn,
	)

	// Separator line
	separator := widget.NewSeparator()

	// Note content with better layout
	noteInfo := container.NewVBox(
		titleContainer,
		widget.NewLabel(""), // Small spacer
		infoContainer,
		separator,
		buttonContainer,
	)

	// Card container with shadow effect
	shadowBg := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 10})
	
	card := container.NewMax(
		cardBg,
		container.NewPadded(
			container.NewPadded(noteInfo),
		),
	)

	// Add shadow effect
	cardWithShadow := container.NewStack(
		container.NewPadded(shadowBg),
		card,
	)

	return cardWithShadow
}

// showDecryptDialog shows dialog to decrypt and view note content
func showDecryptDialog(window fyne.Window, apiClient *api.Client, note api.Note) {
	// Password entry
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter your password to decrypt")

	// Title
	titleLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("🔓 Decrypt: %s", note.Title),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Info text
	infoLabel := widget.NewLabel("Enter your account password to decrypt and view this note.")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Content
	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		infoLabel,
		widget.NewLabel(""),
		passwordEntry,
	)

	// Decrypt button
	decryptBtn := widget.NewButton("🔓 Decrypt & View", func() {})
	decryptBtn.Importance = widget.HighImportance

	// Download button
	downloadBtn := widget.NewButton("💾 Decrypt & Save", func() {})

	// Dialog
	dlg := dialog.NewCustom("Decrypt Note", "Cancel", content, window)

	// Decrypt button action
	decryptBtn.OnTapped = func() {
		password := passwordEntry.Text
		if password == "" {
			dialog.ShowError(fmt.Errorf("password is required"), window)
			return
		}

		// Get note details from server
		noteDetail, err := apiClient.GetNote(note.ID)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to get note: %w", err), window)
			return
		}

		// Derive KEK from password
		kek := crypto.DeriveKeyFromPassword(password, nil)

		// Decrypt DEK
		dekBase64, err := crypto.DecryptAES(noteDetail.EncryptedKey, noteDetail.EncryptedKeyIV, kek)
		if err != nil {
			dialog.ShowError(fmt.Errorf("❌ Wrong password or corrupted key"), window)
			return
		}

		// Decode DEK from base64
		dek, err := base64.StdEncoding.DecodeString(dekBase64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to decode key: %w", err), window)
			return
		}

		// Decrypt content with DEK
		plaintext, err := crypto.DecryptAES(noteDetail.EncryptedContent, noteDetail.IV, dek)
		if err != nil {
			dialog.ShowError(fmt.Errorf("❌ Failed to decrypt content: %w", err), window)
			return
		}

		// Close decrypt dialog
		dlg.Hide()

		// Show decrypted content
		showDecryptedContent(window, note.Title, plaintext)
	}

	// Download button action
	downloadBtn.OnTapped = func() {
		password := passwordEntry.Text
		if password == "" {
			dialog.ShowError(fmt.Errorf("password is required"), window)
			return
		}

		// Get note details from server
		noteDetail, err := apiClient.GetNote(note.ID)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to get note: %w", err), window)
			return
		}

		// Derive KEK from password
		kek := crypto.DeriveKeyFromPassword(password, nil)

		// Decrypt DEK
		dekBase64, err := crypto.DecryptAES(noteDetail.EncryptedKey, noteDetail.EncryptedKeyIV, kek)
		if err != nil {
			dialog.ShowError(fmt.Errorf("❌ Wrong password or corrupted key"), window)
			return
		}

		// Decode DEK from base64
		dek, err := base64.StdEncoding.DecodeString(dekBase64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to decode key: %w", err), window)
			return
		}

		// Decrypt content with DEK
		plaintext, err := crypto.DecryptAES(noteDetail.EncryptedContent, noteDetail.IV, dek)
		if err != nil {
			dialog.ShowError(fmt.Errorf("❌ Failed to decrypt content: %w", err), window)
			return
		}

		// Close decrypt dialog
		dlg.Hide()

		// Save file dialog
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			// Write decrypted content
			if _, err := writer.Write([]byte(plaintext)); err != nil {
				dialog.ShowError(fmt.Errorf("failed to save file: %w", err), window)
				return
			}

			dialog.ShowInformation("✅ Success", "File decrypted and saved successfully!", window)
		}, window)
	}

	// Add buttons to content
	content.Add(widget.NewLabel(""))
	content.Add(container.NewGridWithColumns(2, decryptBtn, downloadBtn))

	dlg.Show()
}

// showDecryptedContent shows the decrypted content in a dialog
func showDecryptedContent(window fyne.Window, title string, content string) {
	// Title
	titleLabel := widget.NewLabelWithStyle(
		"✅ Decrypted: "+title,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Content display (multiline, scrollable)
	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetText(content)
	contentEntry.Wrapping = fyne.TextWrapWord

	// Make it scrollable
	scrollContent := container.NewScroll(contentEntry)
	scrollContent.SetMinSize(fyne.NewSize(600, 400))

	// Copy button
	copyBtn := widget.NewButton("📋 Copy to Clipboard", func() {
		window.Clipboard().SetContent(content)
		dialog.ShowInformation("✅ Copied", "Content copied to clipboard!", window)
	})

	// Save button
	saveBtn := widget.NewButton("💾 Save to File", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			if _, err := writer.Write([]byte(content)); err != nil {
				dialog.ShowError(fmt.Errorf("failed to save: %w", err), window)
				return
			}

			dialog.ShowInformation("✅ Saved", "File saved successfully!", window)
		}, window)
	})

	// Layout
	layout := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		scrollContent,
		widget.NewLabel(""),
		container.NewGridWithColumns(2, copyBtn, saveBtn),
	)

	// Show dialog
	dialog.NewCustom("Decrypted Content", "Close", layout, window).Show()
}

// showShareDurationDialog shows dialog to select share duration before creating link
func showShareDurationDialog(window fyne.Window, apiClient *api.Client, note api.Note, onRefresh func()) {
	// Duration options (3 minutes for testing purposes)
	durationOptions := []string{"3 minutes (TEST)", "1 hour", "6 hours", "12 hours", "24 hours", "48 hours", "7 days"}
	
	// Default selection
	selectedDuration := "24 hours"
	
	// Radio group for duration selection
	durationRadio := widget.NewRadioGroup(durationOptions, func(selected string) {
		selectedDuration = selected
	})
	durationRadio.SetSelected("24 hours")

	// Password protection
	passwordCheck := widget.NewCheck("🔒 Password protection", nil)
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter password (optional)")
	passwordEntry.Disable()
	
	passwordCheck.OnChanged = func(checked bool) {
		if checked {
			passwordEntry.Enable()
		} else {
			passwordEntry.Disable()
			passwordEntry.SetText("")
		}
	}

	// Max access count
	maxAccessCheck := widget.NewCheck("🔢 Limit access count", nil)
	maxAccessEntry := widget.NewEntry()
	maxAccessEntry.SetPlaceHolder("e.g., 5, 10, 100")
	maxAccessEntry.Disable()
	
	maxAccessCheck.OnChanged = func(checked bool) {
		if checked {
			maxAccessEntry.Enable()
		} else {
			maxAccessEntry.Disable()
			maxAccessEntry.SetText("")
		}
	}

	// Title
	title := widget.NewLabelWithStyle(fmt.Sprintf("📄 Share: %s", note.Title),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true})

	// Info
	infoLabel := widget.NewLabel("⏰ Select how long this link will be valid:")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Content
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		infoLabel,
		durationRadio,
		widget.NewSeparator(),
		passwordCheck,
		passwordEntry,
		widget.NewSeparator(),
		maxAccessCheck,
		maxAccessEntry,
		widget.NewSeparator(),
	)

	// Create and Cancel buttons
	var d dialog.Dialog
	
	createBtn := widget.NewButton("✅ Create Link", func() {
		// Hide options dialog
		d.Hide()
		
		// Show loading
		progressDialog := dialog.NewCustom("", "", 
			container.NewVBox(
				widget.NewProgressBarInfinite(),
				widget.NewLabel("Creating share link..."),
			), window)
		progressDialog.Show()
		
		// Create share in background
		go func() {
			var shareToken string
			var err error
			
			// First, get the note's encryption key (DEK)
			fullNote, err := apiClient.GetNote(note.ID)
			if err != nil {
				fyne.Do(func() {
					progressDialog.Hide()
					dialog.ShowError(fmt.Errorf("failed to fetch note: %w", err), window)
				})
				return
			}
			
			// Decrypt the DEK using user's password
			kek := crypto.DeriveKeyFromPassword(api.CurrentPassword, nil)
			dekBase64, err := crypto.DecryptAES(fullNote.EncryptedKey, fullNote.EncryptedKeyIV, kek)
			if err != nil {
				fyne.Do(func() {
					progressDialog.Hide()
					dialog.ShowError(fmt.Errorf("failed to decrypt encryption key: %w", err), window)
				})
				return
			}
			
			// Get password if enabled
			password := ""
			if passwordCheck.Checked {
				password = passwordEntry.Text
			}
			
			// Get max_access_count if enabled
			maxAccessCount := 0
			if maxAccessCheck.Checked && maxAccessEntry.Text != "" {
				// Parse max_access_count
				fmt.Sscanf(maxAccessEntry.Text, "%d", &maxAccessCount)
			}
			
			// Parse hours from selection
			var hours int
			var useMinutes bool
			if selectedDuration == "3 minutes (TEST)" {
				useMinutes = true
			} else {
				switch selectedDuration {
				case "1 hour":
					hours = 1
				case "6 hours":
					hours = 6
				case "12 hours":
					hours = 12
				case "24 hours":
					hours = 24
				case "48 hours":
					hours = 48
				case "7 days":
					hours = 168
				default:
					hours = 24
				}
			}
			
			// Use appropriate API based on options
			// If password or max_access is set, ALWAYS use CreateShareWithOptions
			if password != "" || maxAccessCount > 0 {
				if useMinutes {
					// For 3 minutes with password/max_access, use CreateShareWithOptions with 1 hour minimum
					// Note: Server doesn't support minutes with options, so use 1 hour instead
					shareToken, err = apiClient.CreateShareWithOptions(note.ID, 1, password, maxAccessCount)
				} else {
					shareToken, err = apiClient.CreateShareWithOptions(note.ID, hours, password, maxAccessCount)
				}
			} else if useMinutes {
				// For 3 minutes test WITHOUT password/max_access
				shareToken, err = apiClient.CreateShareWithMinutes(note.ID, 3)
			} else {
				// Legacy API for normal duration without options
				shareToken, err = apiClient.CreateShare(note.ID, hours)
			}
			
			// Close progress dialog and update UI in main thread
			fyne.Do(func() {
				progressDialog.Hide()
				
				if err != nil {
					dialog.ShowError(fmt.Errorf("share failed: %w", err), window)
					return
				}
				
				// Create share URL with encryption key in fragment
				shareURL := fmt.Sprintf("http://localhost:8080/api/shares/%s#key=%s", shareToken, dekBase64)
				
				// Prepare additional info
				additionalInfo := ""
				if password != "" {
					additionalInfo += "🔒 Password protected\n"
				}
				if maxAccessCount > 0 {
					additionalInfo += fmt.Sprintf("🔢 Max accesses: %d\n", maxAccessCount)
				}
				additionalInfo += "🔑 Encryption key included in URL fragment"
				
				showShareResultDialog(window, shareURL, note.Title, selectedDuration, additionalInfo)
				onRefresh()
			})
		}()
	})
	createBtn.Importance = widget.HighImportance

	// Button row
	buttons := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButton("Cancel", func() { d.Hide() }),
		createBtn,
	)
	
	// Final content with buttons
	finalContent := container.NewVBox(
		content,
		buttons,
	)

	// Create dialog
	d = dialog.NewCustom("Create Share Link", "", finalContent, window)
	d.Resize(fyne.NewSize(450, 400))
	d.Show()
}

// showShareResultDialog displays the final share link with copy functionality
func showShareResultDialog(window fyne.Window, shareURL string, noteTitle string, duration string, additionalInfo string) {
	// Create title
	title := widget.NewLabelWithStyle("🎉 Share Link Created Successfully!", 
		fyne.TextAlignCenter, 
		fyne.TextStyle{Bold: true})

	// Note title
	noteTitleLabel := widget.NewLabelWithStyle(fmt.Sprintf("📄 %s", noteTitle),
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true})

	// URL display (read-only entry for easy selection)
	urlEntry := widget.NewEntry()
	urlEntry.SetText(shareURL)
	urlEntry.Disable() // Make it read-only but still selectable
	urlEntry.TextStyle = fyne.TextStyle{Monospace: true}

	// Status label for copy feedback
	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Copy button with icon
	copyBtn := widget.NewButton("📋 Copy to Clipboard", func() {
		window.Clipboard().SetContent(shareURL)
		statusLabel.SetText("✅ Link copied to clipboard!")
		statusLabel.Refresh()
		
		// Reset status after 3 seconds
		go func() {
			time.Sleep(3 * time.Second)
			fyne.Do(func() {
				statusLabel.SetText("")
				statusLabel.Refresh()
			})
		}()
	})
	copyBtn.Importance = widget.HighImportance

	// Info section with duration and additional info
	infoLines := fmt.Sprintf("⏰ Valid for %s\n🔒 End-to-end encrypted\n", duration)
	if additionalInfo != "" {
		infoLines += additionalInfo
	}
	infoLines += "🌐 Anyone with this link can view the note"
	
	infoText := widget.NewLabel(infoLines)
	infoText.Wrapping = fyne.TextWrapWord
	infoText.Alignment = fyne.TextAlignCenter

	// Info box with background
	infoBg := canvas.NewRectangle(color.RGBA{R: 239, G: 246, B: 255, A: 255}) // Light blue
	infoBox := container.NewMax(
		infoBg,
		container.NewPadded(infoText),
	)

	// Warning box
	warningBg := canvas.NewRectangle(color.RGBA{R: 254, G: 243, B: 199, A: 255}) // Light yellow
	warningText := widget.NewLabel("⚠️ Keep this link secure! Anyone with access can view the encrypted content.")
	warningText.Wrapping = fyne.TextWrapWord
	warningText.Alignment = fyne.TextAlignCenter
	warningBox := container.NewMax(
		warningBg,
		container.NewPadded(warningText),
	)

	// Content layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		noteTitleLabel,
		widget.NewLabel(""), // Spacer
		widget.NewLabelWithStyle("Share this link:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		urlEntry,
		copyBtn,
		statusLabel,
		widget.NewLabel(""), // Spacer
		infoBox,
		widget.NewLabel(""), // Spacer
		warningBox,
	)

	// Create dialog
	customDialog := dialog.NewCustom("", "Close", content, window)
	customDialog.Resize(fyne.NewSize(650, 500))
	customDialog.Show()
}

// showShareDialog displays a custom dialog with share link and copy button (legacy, kept for compatibility)
func showShareDialog(window fyne.Window, shareURL string, noteTitle string) {
	// Create title
	title := widget.NewLabelWithStyle("🎉 Share Link Created Successfully!", 
		fyne.TextAlignCenter, 
		fyne.TextStyle{Bold: true})

	// Note title
	noteTitleLabel := widget.NewLabelWithStyle(fmt.Sprintf("📄 %s", noteTitle),
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true})

	// URL display (read-only entry for easy selection)
	urlEntry := widget.NewEntry()
	urlEntry.SetText(shareURL)
	urlEntry.Disable() // Make it read-only but still selectable

	// Status label for copy feedback
	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Copy button
	copyBtn := widget.NewButton("📋 Copy Link", func() {
		window.Clipboard().SetContent(shareURL)
		statusLabel.SetText("✅ Link copied to clipboard!")
		statusLabel.Refresh()
	})
	copyBtn.Importance = widget.HighImportance

	// Info section
	infoText := widget.NewLabel("⏰ Valid for 24 hours\n🔒 End-to-end encrypted\n🌐 Anyone with this link can view")
	infoText.Wrapping = fyne.TextWrapWord
	infoText.Alignment = fyne.TextAlignCenter

	// Info box with background
	infoBg := canvas.NewRectangle(color.RGBA{R: 239, G: 246, B: 255, A: 255}) // Light blue
	infoBox := container.NewMax(
		infoBg,
		container.NewPadded(infoText),
	)

	// Content layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		noteTitleLabel,
		widget.NewLabel(""), // Spacer
		widget.NewLabelWithStyle("Share this link:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		urlEntry,
		copyBtn,
		statusLabel,
		widget.NewLabel(""), // Spacer
		infoBox,
	)

	// Create dialog
	customDialog := dialog.NewCustom("", "Close", content, window)
	customDialog.Resize(fyne.NewSize(600, 400))
	customDialog.Show()
}

// showShareTypeDialog shows dialog to choose between public URL or E2EE sharing
func showShareTypeDialog(window fyne.Window, apiClient *api.Client, note api.Note, onRefresh func()) {
	title := widget.NewLabelWithStyle("📤 Share Options", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	
	noteInfo := widget.NewLabelWithStyle(fmt.Sprintf("📄 %s", note.Title), fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// Option 1: Public URL
	publicOption := widget.NewButton("🌐 Public Share Link (URL)", func() {
		showShareDurationDialog(window, apiClient, note, onRefresh)
	})
	publicOption.Importance = widget.HighImportance
	
	publicDesc := widget.NewLabel("Anyone with the link can view (key in URL fragment)")
	publicDesc.TextStyle = fyne.TextStyle{Italic: true}
	publicDesc.Wrapping = fyne.TextWrapWord

	// Option 2: E2EE with specific user
	e2eeOption := widget.NewButton("🔐 E2EE Share with User (Diffie-Hellman)", func() {
		showE2EEShareDialog(window, apiClient, note, onRefresh)
	})
	e2eeOption.Importance = widget.SuccessImportance
	
	e2eeDesc := widget.NewLabel("Secure share with specific user using key exchange")
	e2eeDesc.TextStyle = fyne.TextStyle{Italic: true}
	e2eeDesc.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		noteInfo,
		widget.NewLabel(""),
		publicOption,
		publicDesc,
		widget.NewLabel(""),
		e2eeOption,
		e2eeDesc,
	)

	dialog.NewCustom("", "Cancel", content, window).Show()
}

// showE2EEShareDialog shows dialog to create E2EE share with specific user
func showE2EEShareDialog(window fyne.Window, apiClient *api.Client, note api.Note, onRefresh func()) {
	title := widget.NewLabelWithStyle("🔐 Create E2EE Share", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	
	noteInfo := widget.NewLabelWithStyle(fmt.Sprintf("📄 %s", note.Title), fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	infoLabel := widget.NewLabel("Enter the username of the person you want to share with.\nThey will receive this note encrypted with Diffie-Hellman.")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Username entry
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Recipient username")

	// Status label
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	// Share button
	shareBtn := widget.NewButton("🔐 Create E2EE Share", func() {
		recipientUsername := usernameEntry.Text
		if recipientUsername == "" {
			statusLabel.SetText("❌ Please enter recipient username")
			return
		}

		statusLabel.SetText("⏳ Creating E2EE share...")
		
		// Get note content first
		fullNote, err := apiClient.GetNote(note.ID)
		if err != nil {
			statusLabel.SetText("❌ Error fetching note: " + err.Error())
			return
		}

		// Decrypt the note with user's password
		kek := crypto.DeriveKeyFromPassword(api.CurrentPassword, nil)
		dekBase64, err := crypto.DecryptAES(fullNote.EncryptedKey, fullNote.EncryptedKeyIV, kek)
		if err != nil {
			statusLabel.SetText("❌ Wrong password or corrupted key")
			return
		}

		dek, err := base64.StdEncoding.DecodeString(dekBase64)
		if err != nil {
			statusLabel.SetText("❌ Key decode error")
			return
		}

		plaintext, err := crypto.DecryptAES(fullNote.EncryptedContent, fullNote.IV, dek)
		if err != nil {
			statusLabel.SetText("❌ Decryption failed")
			return
		}

		// Check if current user has a DH private key
		if api.CurrentDHPrivateKey == nil {
			statusLabel.SetText("❌ Your DH keypair is not initialized. Please re-login.")
			return
		}

		// Fetch recipient's public key from server
		recipientPubKeyBase64, err := apiClient.GetUserPublicKey(recipientUsername)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no public key") {
				statusLabel.SetText(fmt.Sprintf("❌ User '%s' has not set up their E2EE key. They need to login first.", recipientUsername))
			} else {
				statusLabel.SetText("❌ Failed to fetch recipient's public key: " + err.Error())
			}
			return
		}

		// Convert recipient's public key from base64
		recipientPubKey, err := crypto.PublicKeyFromBase64(recipientPubKeyBase64)
		if err != nil {
			statusLabel.SetText("❌ Invalid recipient public key: " + err.Error())
			return
		}

		// Get sender's public key for sending to server
		senderPubKeyBase64 := crypto.PublicKeyToBase64(api.CurrentDHPrivateKey.PublicKey())
		
		// Debug: Log public keys
		fmt.Printf("DEBUG Create - Sender public key: %s\n", senderPubKeyBase64[:20]+"...")
		fmt.Printf("DEBUG Create - Recipient public key: %s\n", recipientPubKeyBase64[:20]+"...")

		// Compute shared secret using sender's private key and recipient's public key
		sharedSecret, err := crypto.ComputeSharedSecret(api.CurrentDHPrivateKey, recipientPubKey)
		if err != nil {
			statusLabel.SetText("❌ Shared secret computation failed: " + err.Error())
			return
		}

		// IMPORTANT: Destroy shared secret after use to ensure forward secrecy
		defer func() {
			// Zero out the shared secret from memory
			for i := range sharedSecret {
				sharedSecret[i] = 0
			}
		}()

		fmt.Printf("DEBUG Create - Shared secret: %x\n", sharedSecret[:8])

		// Encrypt content with shared secret
		encryptedContent, contentIV, err := crypto.EncryptWithSharedSecret(plaintext, sharedSecret)
		if err != nil {
			statusLabel.SetText("❌ Encryption failed: " + err.Error())
			return
		}

		// Send to server - include sender's public key so recipient can compute shared secret
		shareID, err := apiClient.CreateE2EEShare(note.ID, recipientUsername, senderPubKeyBase64, encryptedContent, contentIV, 24)
		if err != nil {
			statusLabel.SetText("❌ Failed to create share: " + err.Error())
			return
		}

		statusLabel.SetText(fmt.Sprintf("✅ E2EE share created! ID: %d", shareID))
		
		// Show success dialog
		dialog.ShowInformation("✅ Success", 
			fmt.Sprintf("E2EE share created successfully with %s!\n\nThey can view it in their 'E2EE Shares' tab.", recipientUsername), 
			window)
		
		if onRefresh != nil {
			onRefresh()
		}
	})
	shareBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		noteInfo,
		widget.NewLabel(""),
		infoLabel,
		widget.NewLabel(""),
		widget.NewLabel("Recipient Username:"),
		usernameEntry,
		widget.NewLabel(""),
		shareBtn,
		statusLabel,
	)

	dialog.NewCustom("", "Cancel", content, window).Show()
}

// createE2EEShareCard creates a card for displaying received E2EE shares
func createE2EEShareCard(share api.E2EEShare, apiClient *api.Client, window fyne.Window, onRefresh func()) fyne.CanvasObject {
	cardBg := canvas.NewRectangle(color.RGBA{R: 249, G: 250, B: 251, A: 255})

	// Title with icon
	titleText := canvas.NewText("🔐 "+share.NoteTitle, color.RGBA{R: 31, G: 41, B: 55, A: 255})
	titleText.TextSize = 16
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	// Sender info
	senderText := canvas.NewText("👤 From: "+share.SenderUsername, color.RGBA{R: 107, G: 114, B: 128, A: 255})
	senderText.TextSize = 12

	// Timestamp
	timeText := canvas.NewText("🕒 "+share.CreatedAt.Format("Jan 02, 2006 15:04"), color.RGBA{R: 107, G: 114, B: 128, A: 255})
	timeText.TextSize = 11

	// Expiry info
	expiryText := canvas.NewText("⏰ Expires: "+share.ExpiresAt.Format("Jan 02 15:04"), color.RGBA{R: 239, G: 68, B: 68, A: 255})
	expiryText.TextSize = 11
	expiryText.TextStyle = fyne.TextStyle{Bold: true}

	infoContainer := container.NewVBox(
		titleText,
		senderText,
		container.NewHBox(timeText, widget.NewLabel("  •  "), expiryText),
	)

	// Decrypt button
	decryptBtn := widget.NewButton("🔓 Decrypt & View", func() {
		showE2EEDecryptDialog(window, apiClient, share, onRefresh)
	})
	decryptBtn.Importance = widget.HighImportance

	// Button container
	buttonContainer := container.NewHBox(
		decryptBtn,
		layout.NewSpacer(),
	)

	separator := widget.NewSeparator()

	cardContent := container.NewVBox(
		infoContainer,
		separator,
		buttonContainer,
	)

	card := container.NewMax(
		cardBg,
		container.NewPadded(container.NewPadded(cardContent)),
	)

	return card
}

// showE2EEDecryptDialog shows dialog to decrypt E2EE share
func showE2EEDecryptDialog(window fyne.Window, apiClient *api.Client, share api.E2EEShare, onRefresh func()) {
	title := widget.NewLabelWithStyle("🔓 Decrypt E2EE Share", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	
	noteInfo := widget.NewLabelWithStyle(fmt.Sprintf("📄 %s (from %s)", share.NoteTitle, share.SenderUsername), 
		fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	infoLabel := widget.NewLabel("This note was shared using Diffie-Hellman key exchange.\nGenerating shared secret to decrypt...")
	infoLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	// Create a text area for decrypted content
	contentArea := widget.NewMultiLineEntry()
	contentArea.SetMinRowsVisible(10)
	contentArea.Wrapping = fyne.TextWrapWord
	
	// Make it read-only but keep text visible (don't use Disable())
	var isUpdating bool
	var lastValidContent string
	contentArea.OnChanged = func(newText string) {
		if !isUpdating && newText != lastValidContent {
			// Prevent user edits by reverting
			isUpdating = true
			contentArea.SetText(lastValidContent)
			isUpdating = false
		}
	}

	// Decrypt button
	decryptBtn := widget.NewButton("🔐 Decrypt with DH", func() {
		statusLabel.SetText("⏳ Computing shared secret...")

		// Check if current user has a DH private key
		if api.CurrentDHPrivateKey == nil {
			statusLabel.SetText("❌ Your DH keypair is not initialized. Please re-login.")
			return
		}

		// Debug: Log recipient's public key
		recipientPubKey := api.CurrentDHPrivateKey.PublicKey()
		recipientPubKeyBase64 := crypto.PublicKeyToBase64(recipientPubKey)
		fmt.Printf("DEBUG Decrypt - Recipient public key: %s\n", recipientPubKeyBase64[:20]+"...")
		fmt.Printf("DEBUG Decrypt - Sender public key: %s\n", share.SenderPublicKey[:20]+"...")

		// Parse sender's public key
		senderPubKey, err := crypto.PublicKeyFromBase64(share.SenderPublicKey)
		if err != nil {
			statusLabel.SetText("❌ Failed to parse sender public key: " + err.Error())
			return
		}

		// Compute shared secret using recipient's private key (stored in memory) and sender's public key
		sharedSecret, err := crypto.ComputeSharedSecret(api.CurrentDHPrivateKey, senderPubKey)
		if err != nil {
			statusLabel.SetText("❌ Shared secret computation failed: " + err.Error())
			return
		}

		// IMPORTANT: Destroy shared secret after use to ensure forward secrecy
		defer func() {
			// Zero out the shared secret from memory
			for i := range sharedSecret {
				sharedSecret[i] = 0
			}
		}()

		fmt.Printf("DEBUG Decrypt - Shared secret: %x\n", sharedSecret[:8])

		statusLabel.SetText("⏳ Decrypting content...")

		// Decrypt content with shared secret
		plaintext, err := crypto.DecryptWithSharedSecret(share.EncryptedContent, share.ContentIV, sharedSecret)
		if err != nil {
			statusLabel.SetText("❌ Decryption failed: " + err.Error())
			return
		}

		statusLabel.SetText("✅ Decrypted successfully!")
		isUpdating = true
		lastValidContent = plaintext
		contentArea.SetText(plaintext)
		isUpdating = false
	})
	decryptBtn.Importance = widget.HighImportance

	// Copy button
	copyBtn := widget.NewButton("📋 Copy Content", func() {
		if contentArea.Text != "" {
			window.Clipboard().SetContent(contentArea.Text)
			statusLabel.SetText("✅ Content copied to clipboard!")
		}
	})

	buttons := container.NewHBox(
		decryptBtn,
		copyBtn,
	)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		noteInfo,
		widget.NewLabel(""),
		infoLabel,
		widget.NewLabel(""),
		buttons,
		statusLabel,
		widget.NewLabel(""),
		widget.NewLabel("Decrypted Content:"),
		contentArea,
	)

	customDialog := dialog.NewCustom("", "Close", content, window)
	customDialog.Resize(fyne.NewSize(700, 600))
	customDialog.Show()
}

// createSharedLinkViewer creates the shared link viewer section
func createSharedLinkViewer(window fyne.Window, apiClient *api.Client) fyne.CanvasObject {
	// Background
	bg := canvas.NewRectangle(color.RGBA{R: 255, G: 255, B: 255, A: 250})
	
	// Title
	title := canvas.NewText("🌐 View Shared Link", color.RGBA{R: 139, G: 92, B: 246, A: 255})
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}
	
	// Description
	desc := widget.NewLabel("Enter a share token or URL to view a shared note")
	desc.TextStyle = fyne.TextStyle{Italic: true}
	desc.Wrapping = fyne.TextWrapWord
	
	// Share token entry
	tokenEntry := widget.NewEntry()
	tokenEntry.SetPlaceHolder("Enter share token or full URL with #key=...")
	tokenEntry.MultiLine = false
	
	// Encryption key entry (will be shown if needed or auto-filled from URL)
	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("Encryption key (auto-filled from URL or enter manually)")
	keyEntry.Hide()
	
	// Password entry (will be shown if needed)
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password (if required)")
	passwordEntry.Hide()
	
	// Status label
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	
	// Content display area
	contentCard := container.NewVBox()
	contentCard.Hide()
	
	// View button
	viewBtn := widget.NewButton("🔍 View Shared Note", func() {
		tokenInput := strings.TrimSpace(tokenEntry.Text)
		if tokenInput == "" {
			statusLabel.SetText("❌ Please enter a share token or URL")
			return
		}
		
		// Extract encryption key from URL fragment (#key=...)
		encryptionKey := strings.TrimSpace(keyEntry.Text)
		if strings.Contains(tokenInput, "#key=") {
			parts := strings.Split(tokenInput, "#key=")
			if len(parts) == 2 {
				encryptionKey = parts[1]
				keyEntry.SetText(encryptionKey)
				keyEntry.Show()
			}
			tokenInput = parts[0] // Remove fragment from URL
		}
		
		// Extract token from URL if full URL provided
		shareToken := tokenInput
		if strings.Contains(tokenInput, "/shares/") {
			parts := strings.Split(tokenInput, "/shares/")
			if len(parts) == 2 {
				shareToken = parts[1]
			}
		}
		
		// Get password if provided
		password := ""
		if passwordEntry.Visible() {
			password = strings.TrimSpace(passwordEntry.Text)
		}
		
		// Show loading
		statusLabel.SetText("🔄 Loading shared note...")
		contentCard.Hide()
		
		// Fetch in background
		go func() {
			sharedNote, err := apiClient.GetSharedNote(shareToken, password)
			
			fyne.Do(func() {
				if err != nil {
					errMsg := err.Error()
					if strings.Contains(errMsg, "unauthorized") {
						// Show password field
						passwordEntry.Show()
						statusLabel.SetText("🔒 This share requires a password. Please enter it above.")
					} else if strings.Contains(errMsg, "expired") {
						statusLabel.SetText("❌ " + errMsg)
					} else {
						statusLabel.SetText("❌ Error: " + errMsg)
					}
					contentCard.Hide()
					return
				}
				
				// Success - display the note with optional decryption
				displaySharedNote(window, sharedNote, encryptionKey, contentCard, statusLabel, keyEntry)
			})
		}()
	})
	viewBtn.Importance = widget.HighImportance
	
	// Instructions
	instructions := widget.NewLabel("📝 Instructions:\n" +
		"1. Copy the share link from the sender\n" +
		"2. Paste it in the field above (full URL or just the token)\n" +
		"3. If password-protected, enter the password\n" +
		"4. Click 'View Shared Note' to decrypt and view")
	instructions.Wrapping = fyne.TextWrapWord
	
	instrCard := canvas.NewRectangle(color.RGBA{R: 239, G: 246, B: 255, A: 255})
	instrBox := container.NewMax(
		instrCard,
		container.NewPadded(instructions),
	)
	
	// Main content
	content := container.NewVBox(
		title,
		desc,
		widget.NewSeparator(),
		widget.NewLabel("Share Token or URL:"),
		tokenEntry,
		keyEntry,
		passwordEntry,
		viewBtn,
		widget.NewSeparator(),
		statusLabel,
		contentCard,
		widget.NewLabel(""),
		instrBox,
	)
	
	return container.NewMax(
		bg,
		container.NewPadded(content),
	)
}

// displaySharedNote displays the fetched shared note with optional decryption
func displaySharedNote(window fyne.Window, sharedNote api.SharedNote, encryptionKey string, contentCard *fyne.Container, statusLabel *widget.Label, keyEntry *widget.Entry) {
	statusLabel.SetText(fmt.Sprintf("✅ Loaded: %s", sharedNote.Title))
	
	// Clear previous content
	contentCard.RemoveAll()
	
	// Note info card
	infoBg := canvas.NewRectangle(color.RGBA{R: 243, G: 244, B: 246, A: 255})
	
	noteTitle := canvas.NewText(sharedNote.Title, color.RGBA{R: 59, G: 130, B: 246, A: 255})
	noteTitle.TextSize = 16
	noteTitle.TextStyle = fyne.TextStyle{Bold: true}
	
	ownerLabel := widget.NewLabel(fmt.Sprintf("👤 Shared by: %s", sharedNote.OwnerUsername))
	createdLabel := widget.NewLabel(fmt.Sprintf("📅 Created: %s", sharedNote.CreatedAt.Format("2006-01-02 15:04")))
	expiresLabel := widget.NewLabel(fmt.Sprintf("⏰ Expires: %s", sharedNote.ExpiresAt.Format("2006-01-02 15:04")))
	
	// Check if expired
	if time.Now().After(sharedNote.ExpiresAt) {
		expiresLabel.Text = "⏰ Expires: ⚠️ EXPIRED"
	}
	
	infoContent := container.NewVBox(
		noteTitle,
		widget.NewSeparator(),
		ownerLabel,
		createdLabel,
		expiresLabel,
	)
	
	infoCard := container.NewMax(
		infoBg,
		container.NewPadded(infoContent),
	)
	
	contentCard.Add(infoCard)
	contentCard.Add(widget.NewLabel(""))
	
	// Try to decrypt if encryption key is provided
	var decryptedContent string
	var decryptionError error
	
	if encryptionKey != "" {
		// Decode the key from base64
		keyBytes, err := base64.StdEncoding.DecodeString(encryptionKey)
		if err != nil {
			decryptionError = fmt.Errorf("Invalid encryption key format: %v", err)
		} else {
			// Decrypt the content
			decryptedContent, decryptionError = crypto.DecryptAES(sharedNote.EncryptedContent, sharedNote.IV, keyBytes)
		}
	}
	
	// Display decrypted content or encrypted content
	if decryptionError == nil && decryptedContent != "" {
		// Successfully decrypted - show plain text
		contentLabel := widget.NewLabel("📄 Decrypted Content:")
		contentLabel.TextStyle = fyne.TextStyle{Bold: true}
		
		// Use RichText with better visibility instead of disabled Entry
		contentText := widget.NewRichTextFromMarkdown("```\n" + decryptedContent + "\n```")
		contentText.Wrapping = fyne.TextWrapWord
		
		// Or use a scrollable container with Label for better readability
		contentDisplay := widget.NewLabel(decryptedContent)
		contentDisplay.Wrapping = fyne.TextWrapWord
		contentScroll := container.NewScroll(contentDisplay)
		contentScroll.SetMinSize(fyne.NewSize(0, 200))
		
		// Success indicator
		successBg := canvas.NewRectangle(color.RGBA{R: 209, G: 250, B: 229, A: 255})
		successText := widget.NewLabel("✅ Content successfully decrypted!")
		successText.TextStyle = fyne.TextStyle{Bold: true}
		
		successCard := container.NewMax(
			successBg,
			container.NewPadded(successText),
		)
		
		// Copy button for decrypted content
		copyBtn := widget.NewButton("📋 Copy Decrypted Content", func() {
			window.Clipboard().SetContent(decryptedContent)
			statusLabel.SetText("✅ Decrypted content copied to clipboard!")
		})
		
		contentCard.Add(successCard)
		contentCard.Add(widget.NewLabel(""))
		contentCard.Add(contentLabel)
		contentCard.Add(contentScroll)
		contentCard.Add(copyBtn)
	} else {
		// Show encrypted content (no key or decryption failed)
		contentLabel := widget.NewLabel("🔒 Encrypted Content:")
		contentLabel.TextStyle = fyne.TextStyle{Bold: true}
		
		contentText := widget.NewMultiLineEntry()
		contentText.SetText(sharedNote.EncryptedContent)
		contentText.Disable()
		contentText.SetMinRowsVisible(5)
		
		// Warning about encryption key
		warningBg := canvas.NewRectangle(color.RGBA{R: 254, G: 243, B: 199, A: 255})
		var warningMessage string
		
		if decryptionError != nil {
			warningMessage = fmt.Sprintf("❌ Decryption failed: %v\n\nPlease check that the encryption key is correct.", decryptionError)
		} else if encryptionKey == "" {
			warningMessage = "⚠️ No encryption key provided.\n\nTo decrypt this content:\n" +
				"1. Include #key=... at the end of the URL when pasting\n" +
				"2. Or enter the encryption key manually in the field above and click 'View' again"
			
			// Show the key entry field
			keyEntry.Show()
		}
		
		warningText := widget.NewLabel(warningMessage)
		warningText.Wrapping = fyne.TextWrapWord
		
		warningCard := container.NewMax(
			warningBg,
			container.NewPadded(warningText),
		)
		
		// Copy button
		copyBtn := widget.NewButton("📋 Copy Encrypted Content", func() {
			window.Clipboard().SetContent(sharedNote.EncryptedContent)
			statusLabel.SetText("✅ Encrypted content copied to clipboard!")
		})
		
		contentCard.Add(contentLabel)
		contentCard.Add(contentText)
		contentCard.Add(copyBtn)
		contentCard.Add(widget.NewLabel(""))
		contentCard.Add(warningCard)
	}
	
	contentCard.Show()
	contentCard.Refresh()
}
