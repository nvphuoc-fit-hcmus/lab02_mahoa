package main

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	app       fyne.App
	window    fyne.Window
	apiClient *APIClient
	userKey   []byte // User's encryption key (derived from password)
}

// ShowLoginScreen displays the login/register screen
func (g *GUI) ShowLoginScreen() {
	// Main background with gradient-like effect
	bgTop := canvas.NewRectangle(color.RGBA{R: 99, G: 102, B: 241, A: 255})

	// Create card background (white)
	cardBg := canvas.NewRectangle(color.White)

	// Shadow effect (light gray border simulation)
	shadowBg := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 100})

	// Title with emoji and styling
	title := canvas.NewText("🔐 Secure Note Sharing", color.RGBA{R: 99, G: 102, B: 241, A: 255})
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Your privacy, our priority", color.RGBA{R: 107, G: 114, B: 128, A: 255})
	subtitle.TextSize = 14
	subtitle.Alignment = fyne.TextAlignCenter

	// Username field
	usernameLabel := canvas.NewText("Username", color.RGBA{R: 55, G: 65, B: 81, A: 255})
	usernameLabel.TextStyle = fyne.TextStyle{Bold: true}
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Enter your username")

	// Password field
	passwordLabel := canvas.NewText("Password", color.RGBA{R: 55, G: 65, B: 81, A: 255})
	passwordLabel.TextStyle = fyne.TextStyle{Bold: true}
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter your password")

	// Status label
	statusLabel := canvas.NewText("", color.Black)
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.TextSize = 13

	setStatus := func(message string, isError bool) {
		statusLabel.Text = message
		if isError {
			statusLabel.Color = color.RGBA{R: 239, G: 68, B: 68, A: 255}
		} else {
			statusLabel.Color = color.RGBA{R: 34, G: 197, B: 94, A: 255}
		}
		statusLabel.Refresh()
	}

	// Login button (primary - gradient blue)
	loginBtn := widget.NewButton("Login", func() {
		username := strings.TrimSpace(usernameEntry.Text)
		password := passwordEntry.Text

		if username == "" {
			setStatus("⚠️  Please enter your username", true)
			return
		}
		if password == "" {
			setStatus("⚠️  Please enter your password", true)
			return
		}

		key := make([]byte, 32)
		copy(key, []byte(password))
		g.userKey = key

		token, err := g.apiClient.Login(username, password)
		if err != nil {
			if strings.Contains(err.Error(), "invalid credentials") || strings.Contains(err.Error(), "Invalid") {
				setStatus("❌ Invalid username or password", true)
			} else if strings.Contains(err.Error(), "connection") {
				setStatus("❌ Server connection failed", true)
			} else {
				setStatus("❌ "+err.Error(), true)
			}
			return
		}

		g.apiClient.Token = token
		setStatus("✅ Login successful!", false)
		g.ShowNotesScreen(username)
	})
	loginBtn.Importance = widget.HighImportance

	// Register button (secondary)
	registerBtn := widget.NewButton("Create Account", func() {
		username := strings.TrimSpace(usernameEntry.Text)
		password := passwordEntry.Text

		if username == "" {
			setStatus("⚠️  Username is required", true)
			return
		}
		if len(username) < 3 {
			setStatus("⚠️  Username must be at least 3 characters", true)
			return
		}
		if password == "" {
			setStatus("⚠️  Password is required", true)
			return
		}
		if len(password) < 6 {
			setStatus("⚠️  Password must be at least 6 characters", true)
			return
		}

		err := g.apiClient.Register(username, password)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
				setStatus("❌ Username already taken", true)
			} else if strings.Contains(err.Error(), "connection") {
				setStatus("❌ Server connection failed", true)
			} else {
				setStatus("❌ "+err.Error(), true)
			}
			return
		}

		setStatus("✅ Account created! Please login", false)
		passwordEntry.SetText("")
	})

	// Card content
	cardContent := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(title),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
		layout.NewSpacer(),
		usernameLabel,
		usernameEntry,
		layout.NewSpacer(),
		passwordLabel,
		passwordEntry,
		layout.NewSpacer(),
		loginBtn,
		registerBtn,
		layout.NewSpacer(),
		container.NewCenter(statusLabel),
		layout.NewSpacer(),
	)

	// Card with shadow effect
	card := container.NewPadded(
		container.NewPadded(
			container.NewMax(
				shadowBg,
				container.NewPadded(
					container.NewMax(
						cardBg,
						container.NewPadded(cardContent),
					),
				),
			),
		),
	)

	// Main layout with gradient background
	content := container.NewMax(
		bgTop,
		container.NewCenter(
			container.NewVBox(
				layout.NewSpacer(),
				card,
				layout.NewSpacer(),
			),
		),
	)

	g.window.SetContent(content)
}

// ShowNotesScreen displays a simple notes page after login
func (g *GUI) ShowNotesScreen(username string) {
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

	// Success card
	successCardBg := canvas.NewRectangle(color.White)

	successIcon := canvas.NewText("✅", color.RGBA{R: 34, G: 197, B: 94, A: 255})
	successIcon.TextSize = 64
	successIcon.Alignment = fyne.TextAlignCenter

	successTitle := canvas.NewText("Welcome Back!", color.RGBA{R: 31, G: 41, B: 55, A: 255})
	successTitle.TextSize = 24
	successTitle.TextStyle = fyne.TextStyle{Bold: true}
	successTitle.Alignment = fyne.TextAlignCenter

	successMsg := canvas.NewText("You have successfully logged in", color.RGBA{R: 107, G: 114, B: 128, A: 255})
	successMsg.TextSize = 14
	successMsg.Alignment = fyne.TextAlignCenter

	secureInfo := canvas.NewText("🔒 Your data is end-to-end encrypted", color.RGBA{R: 99, G: 102, B: 241, A: 255})
	secureInfo.TextSize = 13
	secureInfo.Alignment = fyne.TextAlignCenter

	successCard := container.NewMax(
		successCardBg,
		container.NewPadded(
			container.NewVBox(
				layout.NewSpacer(),
				container.NewCenter(successIcon),
				layout.NewSpacer(),
				container.NewCenter(successTitle),
				layout.NewSpacer(),
				container.NewCenter(successMsg),
				container.NewCenter(secureInfo),
				layout.NewSpacer(),
			),
		),
	)

	// Logout button
	logoutBtn := widget.NewButton("🚪 Logout", func() {
		g.apiClient.Token = ""
		g.userKey = nil
		g.ShowLoginScreen()
	})
	logoutBtn.Importance = widget.DangerImportance

	// Main content layout
	mainContent := container.NewVBox(
		layout.NewSpacer(),
		header,
		layout.NewSpacer(),
		container.NewPadded(
			container.NewPadded(successCard),
		),
		layout.NewSpacer(),
		container.NewCenter(logoutBtn),
		layout.NewSpacer(),
	)

	// Final layout
	content := container.NewMax(
		gradientBg,
		mainContent,
	)

	g.window.SetContent(content)
}
