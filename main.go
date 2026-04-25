package main

import "fyne.io/fyne/v2/app"

func main() {
	a := app.NewWithID("com.vitovt.lolarennt")
	ui := newAppUI(a)
	ui.window.ShowAndRun()
}
