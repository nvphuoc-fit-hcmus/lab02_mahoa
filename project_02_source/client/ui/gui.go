package ui

import (
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/ui/login"
	"lab02_mahoa/client/ui/notes"

	"fyne.io/fyne/v2"
)

type GUI struct {
	App       fyne.App
	Window    fyne.Window
	ApiClient *api.Client
	UserKey   []byte // User's encryption key (derived from password)
}

// ShowLoginScreen displays the login/register screen
func (g *GUI) ShowLoginScreen() {
	login.Screen(g.Window, g.ApiClient, func(username string, userKey []byte) {
		g.UserKey = userKey
		g.ShowNotesScreen(username)
	})
}

// ShowNotesScreen displays a simple notes page after login
func (g *GUI) ShowNotesScreen(username string) {
	notes.Screen(g.Window, g.ApiClient, username, func() {
		g.UserKey = nil
		g.ShowLoginScreen()
	})
}
