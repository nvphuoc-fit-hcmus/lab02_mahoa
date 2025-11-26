package login

import (
	"image/color"
	"lab02_mahoa/client/api"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Screen displays the login/register screen
func Screen(window fyne.Window, apiClient *api.Client, onLoginSuccess func(username string, userKey []byte)) {
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

		token, err := apiClient.Login(username, password)
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

		apiClient.Token = token
		setStatus("✅ Login successful!", false)
		onLoginSuccess(username, key)
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

		err := apiClient.Register(username, password)
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

	window.SetContent(content)
}
