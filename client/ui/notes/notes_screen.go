package notes

import (
	"fmt"
	"image/color"
	"io"
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/crypto" // Import thêm để mã hóa
	"path/filepath"             // Import thêm để xử lý tên file

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage" // Import thêm để lọc file
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

	// Success card (GIỮ NGUYÊN)
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

	// ---------------------------------------------------------
	// [PHẦN THÊM MỚI] UPLOAD FORM CARD (Code thêm bắt đầu từ đây)
	// ---------------------------------------------------------

	uploadCardBg := canvas.NewRectangle(color.White) // Nền trắng giống card trên

	// UI: Tiêu đề form
	lblFormTitle := canvas.NewText("Upload New Encrypted Note", color.RGBA{R: 55, G: 65, B: 81, A: 255})
	lblFormTitle.TextStyle = fyne.TextStyle{Bold: true}
	lblFormTitle.TextSize = 16
	lblFormTitle.Alignment = fyne.TextAlignCenter

	// UI: Ô nhập
	titleEntry := widget.NewEntry()
	titleEntry.PlaceHolder = "Enter note title..."

	// UI: Chọn file
	var fileContent []byte // Biến lưu file RAM
	statusLabel := widget.NewLabel("No file selected")
	statusLabel.Alignment = fyne.TextAlignCenter

	btnSelectFile := widget.NewButton("📂 Select File", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			data, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			fileContent = data
			statusLabel.SetText("Selected: " + filepath.Base(reader.URI().Path()))
		}, window)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".pdf", ".png", ".jpg"}))
		fileDialog.Show()
	})

	// UI: Nút Upload
	btnUpload := widget.NewButton("🔒 Encrypt & Upload", func() {
		// 1. Validate
		if titleEntry.Text == "" || len(fileContent) == 0 {
			dialog.ShowError(fmt.Errorf("Please enter title and select a file"), window)
			return
		}

		// 2. Client-side Encryption
		statusLabel.SetText("Encrypting...")
		
		// Sinh khóa file
		fileKey, err := crypto.GenerateKey()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		// Mã hóa nội dung
		encryptedContent, iv, err := crypto.EncryptAES(string(fileContent), fileKey)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		// Mã hóa khóa file bằng password (lấy từ api.CurrentPassword)
		if api.CurrentPassword == "" {
			dialog.ShowError(fmt.Errorf("Session error. Please relogin"), window)
			return
		}
		masterKey := crypto.DeriveKeyFromPassword(api.CurrentPassword, api.CurrentUsername)
		encryptedKey, err := crypto.WrapKey(fileKey, masterKey)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		// 3. Send to Server
		statusLabel.SetText("Uploading...")
		err = apiClient.CreateNote(titleEntry.Text, encryptedContent, encryptedKey, iv)
		if err != nil {
			dialog.ShowError(err, window)
			statusLabel.SetText("Upload Failed")
		} else {
			dialog.ShowInformation("Success", "Note encrypted and uploaded!", window)
			titleEntry.SetText("")
			fileContent = nil
			statusLabel.SetText("Ready for next file")
		}
	})
	btnUpload.Importance = widget.HighImportance

	// Gom nhóm phần Upload vào 1 Card
	uploadFormContent := container.NewVBox(
		lblFormTitle,
		widget.NewSeparator(),
		widget.NewLabel("Title:"),
		titleEntry,
		widget.NewLabel("Content:"),
		btnSelectFile,
		statusLabel,
		layout.NewSpacer(),
		btnUpload,
	)

	uploadCard := container.NewMax(
		uploadCardBg,
		container.NewPadded(
			container.NewPadded(uploadFormContent),
		),
	)
	// ---------------------------------------------------------
	// [KẾT THÚC PHẦN THÊM MỚI]
	// ---------------------------------------------------------

	// Logout button (GIỮ NGUYÊN)
	logoutBtn := widget.NewButton("🚪 Logout", func() {
		api.AuthToken = ""
		onLogout()
	})
	logoutBtn.Importance = widget.DangerImportance

	// Main content layout (CẬP NHẬT: Thêm uploadCard vào danh sách)
	mainContent := container.NewVBox(
		layout.NewSpacer(),
		header,
		layout.NewSpacer(),
		container.NewPadded(
			container.NewPadded(successCard), // Card cũ
		),
		container.NewPadded(
			container.NewPadded(uploadCard),  // Card mới thêm vào
		),
		layout.NewSpacer(),
		container.NewCenter(logoutBtn),
		layout.NewSpacer(),
	)

	// Scroll Container (Thêm cái này để nếu màn hình nhỏ thì cuộn được)
	scrollContainer := container.NewVScroll(mainContent)

	// Final layout
	content := container.NewMax(
		gradientBg,
		scrollContainer, // Dùng scroll thay vì mainContent trực tiếp
	)

	window.SetContent(content)
}