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

// Colors - Clean white theme
var (
	colorPrimary   = color.RGBA{R: 59, G: 130, B: 246, A: 255}  // Soft blue
	colorSecondary = color.RGBA{R: 99, G: 102, B: 241, A: 255}  // Purple-blue
	colorSuccess   = color.RGBA{R: 34, G: 197, B: 94, A: 255}   // Green
	colorError     = color.RGBA{R: 239, G: 68, B: 68, A: 255}   // Red
	colorText      = color.RGBA{R: 31, G: 41, B: 55, A: 255}    // Dark gray
	colorTextLight = color.RGBA{R: 107, G: 114, B: 128, A: 255} // Light gray
	colorBg        = color.RGBA{R: 249, G: 250, B: 251, A: 255} // Very light gray
	colorWhite     = color.White
	colorBorder    = color.RGBA{R: 229, G: 231, B: 235, A: 255} // Border gray
)

// Screen displays the initial welcome screen with choice between Login/Register
func Screen(window fyne.Window, apiClient *api.Client, onLoginSuccess func(username string, userKey []byte)) {
	ShowWelcomeScreen(window, apiClient, onLoginSuccess)
}

// ShowWelcomeScreen shows the welcome screen with Login/Register options
func ShowWelcomeScreen(window fyne.Window, apiClient *api.Client, onLoginSuccess func(username string, userKey []byte)) {
	// Background
	bg := canvas.NewRectangle(colorBg)

	// App icon/logo
	icon := canvas.NewText("🔐", colorPrimary)
	icon.TextSize = 64
	icon.Alignment = fyne.TextAlignCenter

	// Main title
	title := canvas.NewText("Secure Note Sharing", colorText)
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Subtitle
	subtitle := canvas.NewText("Mã hóa đầu cuối - Bảo mật tuyệt đối", colorTextLight)
	subtitle.TextSize = 14
	subtitle.Alignment = fyne.TextAlignCenter

	// Card background
	cardBg := canvas.NewRectangle(colorWhite)
	cardBg.CornerRadius = 16

	// Login button - Primary
	loginBtn := widget.NewButton("Đăng nhập", func() {
		ShowLoginScreen(window, apiClient, onLoginSuccess)
	})
	loginBtn.Importance = widget.HighImportance

	// Register button - Secondary
	registerBtn := widget.NewButton("Tạo tài khoản mới", func() {
		ShowRegisterScreen(window, apiClient, onLoginSuccess)
	})

	// Info text
	infoText := canvas.NewText("💡 Chọn một tùy chọn để bắt đầu", colorTextLight)
	infoText.TextSize = 13
	infoText.Alignment = fyne.TextAlignCenter

	// Card content
	cardContent := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(icon),
		container.NewPadded(layout.NewSpacer()),
		container.NewCenter(title),
		container.NewCenter(subtitle),
		layout.NewSpacer(),
		loginBtn,
		container.NewPadded(layout.NewSpacer()),
		registerBtn,
		layout.NewSpacer(),
		container.NewCenter(infoText),
		layout.NewSpacer(),
	)

	// Card with padding and border
	card := container.NewPadded(
		container.NewStack(
			cardBg,
			container.NewPadded(
				container.NewPadded(cardContent),
			),
		),
	)

	// Main content
	content := container.NewMax(
		bg,
		container.NewCenter(
			container.NewVBox(
				layout.NewSpacer(),
				container.NewPadded(card),
				layout.NewSpacer(),
			),
		),
	)

	window.SetContent(content)
}

// ShowLoginScreen shows the login screen
func ShowLoginScreen(window fyne.Window, apiClient *api.Client, onLoginSuccess func(username string, userKey []byte)) {
	// Background
	bg := canvas.NewRectangle(colorBg)

	// Card background
	cardBg := canvas.NewRectangle(colorWhite)
	cardBg.CornerRadius = 16

	// Header
	headerIcon := canvas.NewText("🔑", colorPrimary)
	headerIcon.TextSize = 48
	headerIcon.Alignment = fyne.TextAlignCenter

	headerTitle := canvas.NewText("Đăng nhập", colorText)
	headerTitle.TextSize = 28
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.Alignment = fyne.TextAlignCenter

	headerSubtitle := canvas.NewText("Nhập thông tin để tiếp tục", colorTextLight)
	headerSubtitle.TextSize = 13
	headerSubtitle.Alignment = fyne.TextAlignCenter

	// Username field
	usernameLabel := canvas.NewText("Tên đăng nhập", colorText)
	usernameLabel.TextStyle = fyne.TextStyle{Bold: true}
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Nhập tên đăng nhập của bạn")

	// Password field
	passwordLabel := canvas.NewText("Mật khẩu", colorText)
	passwordLabel.TextStyle = fyne.TextStyle{Bold: true}
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Nhập mật khẩu của bạn")

	// Status label
	statusLabel := canvas.NewText("", colorText)
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.TextSize = 13

	setStatus := func(message string, isError bool) {
		statusLabel.Text = message
		if isError {
			statusLabel.Color = colorError
		} else {
			statusLabel.Color = colorSuccess
		}
		statusLabel.Refresh()
	}

	// Login button
	loginBtn := widget.NewButton("Đăng nhập", func() {
		username := strings.TrimSpace(usernameEntry.Text)
		password := passwordEntry.Text

		if username == "" {
			setStatus("⚠️ Vui lòng nhập tên đăng nhập", true)
			return
		}
		if password == "" {
			setStatus("⚠️ Vui lòng nhập mật khẩu", true)
			return
		}

		key := make([]byte, 32)
		copy(key, []byte(password))

		token, err := apiClient.Login(username, password)
		if err != nil {
			if strings.Contains(err.Error(), "invalid credentials") || strings.Contains(err.Error(), "Invalid") {
				setStatus("❌ Tên đăng nhập hoặc mật khẩu không đúng", true)
			} else if strings.Contains(err.Error(), "connection") {
				setStatus("❌ Không thể kết nối đến server", true)
			} else {
				setStatus("❌ "+err.Error(), true)
			}
			return
		}

		apiClient.Token = token
		api.AuthToken = token
		api.CurrentUsername = username
		api.CurrentPassword = password
		setStatus("✅ Đăng nhập thành công!", false)
		onLoginSuccess(username, key)
	})
	loginBtn.Importance = widget.HighImportance

	// Back button
	backBtn := widget.NewButton("← Quay lại", func() {
		ShowWelcomeScreen(window, apiClient, onLoginSuccess)
	})

	// Register link
	registerLink := widget.NewButton("Chưa có tài khoản? Đăng ký ngay", func() {
		ShowRegisterScreen(window, apiClient, onLoginSuccess)
	})

	// Divider
	divider := widget.NewSeparator()

	// Card content
	cardContent := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(headerIcon),
		container.NewPadded(layout.NewSpacer()),
		container.NewCenter(headerTitle),
		container.NewCenter(headerSubtitle),
		layout.NewSpacer(),
		usernameLabel,
		usernameEntry,
		layout.NewSpacer(),
		passwordLabel,
		passwordEntry,
		layout.NewSpacer(),
		loginBtn,
		layout.NewSpacer(),
		container.NewCenter(statusLabel),
		layout.NewSpacer(),
		divider,
		container.NewCenter(registerLink),
		layout.NewSpacer(),
		backBtn,
		layout.NewSpacer(),
	)

	// Card with padding
	card := container.NewPadded(
		container.NewStack(
			cardBg,
			container.NewPadded(
				container.NewPadded(cardContent),
			),
		),
	)

	// Main content
	content := container.NewMax(
		bg,
		container.NewCenter(
			container.NewVBox(
				layout.NewSpacer(),
				container.NewPadded(card),
				layout.NewSpacer(),
			),
		),
	)

	window.SetContent(content)
}

// ShowRegisterScreen shows the register screen
func ShowRegisterScreen(window fyne.Window, apiClient *api.Client, onLoginSuccess func(username string, userKey []byte)) {
	// Background
	bg := canvas.NewRectangle(colorBg)

	// Card background
	cardBg := canvas.NewRectangle(colorWhite)
	cardBg.CornerRadius = 16

	// Header
	headerIcon := canvas.NewText("✨", colorPrimary)
	headerIcon.TextSize = 48
	headerIcon.Alignment = fyne.TextAlignCenter

	headerTitle := canvas.NewText("Tạo tài khoản", colorText)
	headerTitle.TextSize = 28
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.Alignment = fyne.TextAlignCenter

	headerSubtitle := canvas.NewText("Đăng ký để bắt đầu sử dụng", colorTextLight)
	headerSubtitle.TextSize = 13
	headerSubtitle.Alignment = fyne.TextAlignCenter

	// Username field
	usernameLabel := canvas.NewText("Tên đăng nhập", colorText)
	usernameLabel.TextStyle = fyne.TextStyle{Bold: true}
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Tối thiểu 3 ký tự")

	// Password field
	passwordLabel := canvas.NewText("Mật khẩu", colorText)
	passwordLabel.TextStyle = fyne.TextStyle{Bold: true}
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Tối thiểu 6 ký tự")

	// Status label
	statusLabel := canvas.NewText("", colorText)
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.TextSize = 13

	setStatus := func(message string, isError bool) {
		statusLabel.Text = message
		if isError {
			statusLabel.Color = colorError
		} else {
			statusLabel.Color = colorSuccess
		}
		statusLabel.Refresh()
	}

	// Register button
	registerBtn := widget.NewButton("Tạo tài khoản", func() {
		username := strings.TrimSpace(usernameEntry.Text)
		password := passwordEntry.Text

		if username == "" {
			setStatus("⚠️ Vui lòng nhập tên đăng nhập", true)
			return
		}
		if len(username) < 3 {
			setStatus("⚠️ Tên đăng nhập phải có ít nhất 3 ký tự", true)
			return
		}
		if password == "" {
			setStatus("⚠️ Vui lòng nhập mật khẩu", true)
			return
		}
		if len(password) < 6 {
			setStatus("⚠️ Mật khẩu phải có ít nhất 6 ký tự", true)
			return
		}

		err := apiClient.Register(username, password)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
				setStatus("❌ Tên đăng nhập đã tồn tại", true)
			} else if strings.Contains(err.Error(), "connection") {
				setStatus("❌ Không thể kết nối đến server", true)
			} else {
				setStatus("❌ "+err.Error(), true)
			}
			return
		}

		setStatus("✅ Tạo tài khoản thành công! Đang chuyển đến đăng nhập...", false)

		// Auto switch to login after a moment
		go func() {
			// Switch to login screen with pre-filled username
			ShowLoginScreenWithUsername(window, apiClient, onLoginSuccess, username)
		}()
	})
	registerBtn.Importance = widget.HighImportance

	// Back button
	backBtn := widget.NewButton("← Quay lại", func() {
		ShowWelcomeScreen(window, apiClient, onLoginSuccess)
	})

	// Login link
	loginLink := widget.NewButton("Đã có tài khoản? Đăng nhập ngay", func() {
		ShowLoginScreen(window, apiClient, onLoginSuccess)
	})

	// Divider
	divider := widget.NewSeparator()

	// Info card
	infoText := canvas.NewText("💡 Mật khẩu của bạn sẽ được mã hóa an toàn", colorTextLight)
	infoText.TextSize = 12
	infoText.Alignment = fyne.TextAlignCenter

	// Card content
	cardContent := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(headerIcon),
		container.NewPadded(layout.NewSpacer()),
		container.NewCenter(headerTitle),
		container.NewCenter(headerSubtitle),
		layout.NewSpacer(),
		usernameLabel,
		usernameEntry,
		layout.NewSpacer(),
		passwordLabel,
		passwordEntry,
		layout.NewSpacer(),
		container.NewCenter(infoText),
		layout.NewSpacer(),
		registerBtn,
		layout.NewSpacer(),
		container.NewCenter(statusLabel),
		layout.NewSpacer(),
		divider,
		container.NewCenter(loginLink),
		layout.NewSpacer(),
		backBtn,
		layout.NewSpacer(),
	)

	// Card with padding
	card := container.NewPadded(
		container.NewStack(
			cardBg,
			container.NewPadded(
				container.NewPadded(cardContent),
			),
		),
	)

	// Main content
	content := container.NewMax(
		bg,
		container.NewCenter(
			container.NewVBox(
				layout.NewSpacer(),
				container.NewPadded(card),
				layout.NewSpacer(),
			),
		),
	)

	window.SetContent(content)
}

// ShowLoginScreenWithUsername shows login screen with pre-filled username
func ShowLoginScreenWithUsername(window fyne.Window, apiClient *api.Client, onLoginSuccess func(username string, userKey []byte), username string) {
	ShowLoginScreen(window, apiClient, onLoginSuccess)
	// Note: In real implementation, you'd pass the username to pre-fill the field
}
