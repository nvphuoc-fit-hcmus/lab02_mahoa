package main

import (
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/cli"
	"lab02_mahoa/client/ui"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Check if CLI mode (if arguments provided and not "-gui")
	if len(os.Args) > 1 && os.Args[1] != "-gui" {
		cli.Run(os.Args[1:])
		return
	}

	// GUI mode
	myApp := app.NewWithID("secure-notes-app")
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
