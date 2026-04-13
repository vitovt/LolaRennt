package main

import "fyne.io/fyne/v2/app"

func main() {
	a := app.New()
	ui := newAppUI(a)
	ui.window.ShowAndRun()
}
