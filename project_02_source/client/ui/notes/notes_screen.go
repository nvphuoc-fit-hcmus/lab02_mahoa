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

	// Main content with better spacing
	mainContent := container.NewVBox(
		header,
		widget.NewLabel(""), // Spacer
		uploadForm,
		widget.NewLabel(""), // Spacer
		notesSection,
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

	// Share button (to create share link)
	shareBtn := widget.NewButton("🔗 Share", func() {
		if shareToken, err := apiClient.CreateShare(note.ID, 24); err != nil {
			dialog.ShowError(fmt.Errorf("share failed: %w", err), window)
		} else {
			shareURL := fmt.Sprintf("http://localhost:8080/share/%s", shareToken)
			dialog.ShowInformation("✅ Share Link Created", 
				fmt.Sprintf("Share URL:\n\n%s\n\n⏰ Valid for 24 hours\n🔒 Encrypted end-to-end", shareURL), 
				window)
			onRefresh()
		}
	})
	shareBtn.Importance = widget.HighImportance

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
