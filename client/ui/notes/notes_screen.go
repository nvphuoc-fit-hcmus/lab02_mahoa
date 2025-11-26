package notes

import (
	"image/color"
	"lab02_mahoa/client/api"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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
		apiClient.Token = ""
		onLogout()
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

	window.SetContent(content)
}
