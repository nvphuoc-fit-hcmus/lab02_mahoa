package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Create Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("🔐 Secure Note Sharing")
	
	// Initialize API client
	apiClient := &APIClient{}
	
	// Create GUI
	gui := &GUI{
		app:       myApp,
		window:    myWindow,
		apiClient: apiClient,
	}
	
	// Show login screen
	gui.ShowLoginScreen()
	
	// Set window size and show
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}
