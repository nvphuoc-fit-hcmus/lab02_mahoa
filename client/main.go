package main

import (
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Create Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("Secure Note Sharing")

	// Initialize API client
	apiClient := &api.Client{}

	// Create GUI
	gui := &ui.GUI{
		App:       myApp,
		Window:    myWindow,
		ApiClient: apiClient,
	}

	// Show login screen
	gui.ShowLoginScreen()

	// Set window size and show
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}
