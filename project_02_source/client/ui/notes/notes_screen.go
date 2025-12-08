package notes

import (
	"fmt"
	"image/color"
	"io"
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/crypto"
	"path/filepath"
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
		showShareOptionsDialog(window, apiClient, note, onRefresh)
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

// showShareOptionsDialog shows dialog to select share duration before creating link
func showShareOptionsDialog(window fyne.Window, apiClient *api.Client, note api.Note, onRefresh func()) {
	// Duration options (3 minutes for testing purposes)
	durationOptions := []string{"3 minutes (TEST)", "1 hour", "6 hours", "12 hours", "24 hours", "48 hours", "7 days"}
	
	// Default selection
	selectedDuration := "24 hours"
	
	// Radio group for duration selection
	durationRadio := widget.NewRadioGroup(durationOptions, func(selected string) {
		selectedDuration = selected
	})
	durationRadio.SetSelected("24 hours")

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
			
			// Use appropriate API based on selected duration
			if selectedDuration == "3 minutes" {
				shareToken, err = apiClient.CreateShareWithMinutes(note.ID, 3)
			} else {
				// Parse hours from selection
				var hours int
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
				shareToken, err = apiClient.CreateShare(note.ID, hours)
			}
			
			// Close progress dialog and update UI in main thread
			fyne.Do(func() {
				progressDialog.Hide()
				
				if err != nil {
					dialog.ShowError(fmt.Errorf("share failed: %w", err), window)
					return
				}
				
				shareURL := fmt.Sprintf("http://localhost:8080/api/shares/%s", shareToken)
				showShareResultDialog(window, shareURL, note.Title, selectedDuration)
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
func showShareResultDialog(window fyne.Window, shareURL string, noteTitle string, duration string) {
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

	// Info section with duration
	infoLines := fmt.Sprintf("⏰ Valid for %s\n🔒 End-to-end encrypted\n🌐 Anyone with this link can view the note", duration)
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
