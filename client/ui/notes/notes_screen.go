package notes

import (
	"fmt"
	"image/color"
	"io"
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/crypto"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Screen displays the notes page after login
func Screen(window fyne.Window, apiClient *api.Client, username string, onLogout func()) {
	// Gradient background
	gradientBg := canvas.NewRectangle(color.RGBA{R: 99, G: 102, B: 241, A: 255})

	// Header section
	headerTitle := canvas.NewText("📝 Secure Notes", color.White)
	headerTitle.TextSize = 28
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.Alignment = fyne.TextAlignCenter

	userInfo := canvas.NewText("👤 "+username, color.RGBA{R: 255, G: 255, B: 255, A: 200})
	userInfo.TextSize = 14
	userInfo.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(headerTitle),
		container.NewCenter(userInfo),
		layout.NewSpacer(),
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
		// Clear previous list
		notesContainer.RemoveAll()

		// Call API to get notes
		notes, err := apiClient.ListNotes()
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
	}

	// Upload section
	uploadTitle := canvas.NewText("📤 Upload New Note", color.RGBA{R: 31, G: 41, B: 55, A: 255})
	uploadTitle.TextSize = 16
	uploadTitle.TextStyle = fyne.TextStyle{Bold: true}

	// File upload button
	uploadBtn := widget.NewButton("Choose File & Upload", func() {
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

			// Generate key
			key, err := crypto.GenerateKey()
			if err != nil {
				statusLabel.SetText("❌ Key generation error: " + err.Error())
				return
			}

			// Encrypt content
			encryptedContent, iv, err := crypto.EncryptAES(string(content), key)
			if err != nil {
				statusLabel.SetText("❌ Encryption error: " + err.Error())
				return
			}

			// Encrypt key
			keyStr := fmt.Sprintf("key_%s", fileName)
			encryptedKey, _, err := crypto.EncryptAES(keyStr, key)
			if err != nil {
				statusLabel.SetText("❌ Key encryption error: " + err.Error())
				return
			}

			statusLabel.SetText("⏳ Uploading...")
			
			// Upload to server
			if err := apiClient.CreateNote(fileName, encryptedContent, encryptedKey, iv); err != nil {
				statusLabel.SetText("❌ Upload error: " + err.Error())
				return
			}

			statusLabel.SetText("✅ Note uploaded successfully!")
			refreshNotes()
		}, window)
	})

	uploadForm := container.NewVBox(
		uploadTitle,
		uploadBtn,
	)

	// Action buttons
	logoutBtn := widget.NewButton("🚪 Logout", func() {
		apiClient.Token = ""
		onLogout()
	})
	logoutBtn.Importance = widget.DangerImportance

	actionBar := container.NewHBox(
		layout.NewSpacer(),
		logoutBtn,
	)

	// Notes list section
	notesTitle := canvas.NewText("📋 Your Notes", color.RGBA{R: 31, G: 41, B: 55, A: 255})
	notesTitle.TextSize = 16
	notesTitle.TextStyle = fyne.TextStyle{Bold: true}

	notesSection := container.NewVBox(
		notesTitle,
		notesScroll,
	)

	// Main content
	mainContent := container.NewVBox(
		header,
		layout.NewSpacer(),
		uploadForm,
		layout.NewSpacer(),
		notesSection,
		layout.NewSpacer(),
		statusLabel,
		actionBar,
	)

	// Scrollable main content
	scrollContent := container.NewScroll(mainContent)

	// Final layout
	content := container.NewMax(
		gradientBg,
		container.NewPadded(scrollContent),
	)

	window.SetContent(content)

	// Load notes on screen open
	refreshNotes()
}

// createNoteCard creates a card widget for a single note
func createNoteCard(note api.Note, apiClient *api.Client, window fyne.Window, onRefresh func()) fyne.CanvasObject {
	// Card background
	cardBg := canvas.NewRectangle(color.White)

	// Note title with share status icon
	titleText := canvas.NewText(note.Title, color.RGBA{R: 31, G: 41, B: 55, A: 255})
	titleText.TextSize = 14
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	// Share status icon
	var shareStatusIcon string
	var shareStatusColor color.Color
	if note.IsShared {
		shareStatusIcon = "🌐 Shared"
		shareStatusColor = color.RGBA{R: 34, G: 197, B: 94, A: 255} // Green
	} else {
		shareStatusIcon = "🔒 Private"
		shareStatusColor = color.RGBA{R: 107, G: 114, B: 128, A: 255} // Gray
	}

	shareStatus := canvas.NewText(shareStatusIcon, shareStatusColor)
	shareStatus.TextSize = 12
	shareStatus.TextStyle = fyne.TextStyle{Bold: true}

	titleContainer := container.NewHBox(
		titleText,
		layout.NewSpacer(),
		shareStatus,
	)

	// Note info
	sizeText := canvas.NewText(fmt.Sprintf("Size: %d bytes", len(note.EncryptedContent)), color.RGBA{R: 107, G: 114, B: 128, A: 255})
	sizeText.TextSize = 12

	createdText := canvas.NewText(fmt.Sprintf("Created: %s", note.CreatedAt.Format("2006-01-02 15:04")), color.RGBA{R: 107, G: 114, B: 128, A: 255})
	createdText.TextSize = 12

	// Delete button
	deleteBtn := widget.NewButton("🗑️ Delete", func() {
		dialog.ShowConfirm("Delete Note", "Are you sure you want to delete this note?", func(confirmed bool) {
			if confirmed {
				if err := apiClient.DeleteNote(note.ID); err != nil {
					dialog.ShowError(fmt.Errorf("delete failed: %w", err), window)
					return
				}
				dialog.ShowInformation("Success", "Note deleted successfully", window)
				onRefresh()
			}
		}, window)
	})
	deleteBtn.Importance = widget.DangerImportance

	// Revoke button (only show if shared)
	var revokeBtn *widget.Button
	if note.IsShared {
		revokeBtn = widget.NewButton("🔐 Revoke Share", func() {
			if err := apiClient.RevokeShare(note.ID); err != nil {
				dialog.ShowError(fmt.Errorf("revoke failed: %w", err), window)
				return
			}
			dialog.ShowInformation("Success", "Sharing revoked successfully", window)
			onRefresh()
		})
	} else {
		revokeBtn = widget.NewButton("🔐 Revoke Share", func() {})
		revokeBtn.Disable() // Disable if not shared
	}

	// Share button (to create share link)
	shareBtn := widget.NewButton("🌐 Share", func() {
		if shareToken, err := apiClient.CreateShare(note.ID, 24); err != nil {
			dialog.ShowError(fmt.Errorf("share failed: %w", err), window)
		} else {
			shareURL := fmt.Sprintf("http://localhost:8080/share/%s", shareToken)
			dialog.ShowInformation("Share Link Created", "Share URL:\n"+shareURL+"\n\nValid for 24 hours", window)
			onRefresh()
		}
	})

	// Button container
	buttonContainer := container.NewHBox(
		deleteBtn,
		shareBtn,
		revokeBtn,
		layout.NewSpacer(),
	)

	// Note content
	noteInfo := container.NewVBox(
		titleContainer,
		sizeText,
		createdText,
		buttonContainer,
	)

	// Card container
	card := container.NewMax(
		cardBg,
		container.NewPadded(noteInfo),
	)

	// Add border effect by wrapping in a box
	border := container.NewBorder(
		nil, nil, nil, nil,
		card,
	)

	return border
}
